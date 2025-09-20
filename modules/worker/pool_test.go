package worker

import (
	"context"
	"github.com/thep2p/skipgraph-go/unittest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/modules"
)

func init() {
	// Set trace level for tests to see logs
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

type mockJob struct {
	picked   chan interface{} // closed when picked up by worker
	executed chan interface{} // closed when executed
	block    chan interface{} // if non-nil, job blocks until channel is closed
	panic    bool             // if true, job panics when executed
}

func (m *mockJob) Execute(ctx modules.ThrowableContext) {
	close(m.picked) // signal picked up
	if m.panic {
		ctx.ThrowIrrecoverable(assert.AnError)
	}
	if m.block != nil {
		<-m.block
	}
	close(m.executed) // signal executed
}

type cancellableThrowableContext struct {
	context.Context
	thrown error
	mu     sync.Mutex
}

func (m *cancellableThrowableContext) ThrowIrrecoverable(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.thrown = err
}

// TestPool_HappyPath tests normal operation of the worker pool.
// It starts the pool, submits jobs, and verifies they execute.
// Also verifies pool state (worker count, queue size) at the end.
func TestPool_HappyPath(t *testing.T) {
	throwCtx := unittest.NewMockThrowableContext(t)
	pool := NewWorkerPool(10, 3)
	defer func() {
		throwCtx.Cancel()
		unittest.RequireAllDone(t, pool)
	}()

	// Start pool
	pool.Start(throwCtx)

	// Wait for ready
	unittest.RequireAllReady(t, pool)

	// Submit and execute jobs
	jobsCount := 5
	jobs := make([]*mockJob, jobsCount)
	for i := range jobs {
		jobs[i] = &mockJob{
			picked:   make(chan interface{}),
			executed: make(chan interface{}),
		}
		require.NoError(t, pool.Submit(jobs[i]))
	}

	// Wait for all jobs to execute
	for _, job := range jobs {
		select {
		case <-job.executed:
		case <-time.After(2 * time.Second):
			t.Fatal("job did not complete")
		}
	}

	// Verify pool state; 3 workers, and empty queue at end.
	assert.Equal(t, 3, pool.WorkerCount())
	assert.Equal(t, 0, pool.QueueSize())
}

func TestPool_QueueFull(t *testing.T) {
	throwCtx := unittest.NewMockThrowableContext(t)
	pool := NewWorkerPool(1, 1)
	defer func() {
		throwCtx.Cancel()
		unittest.RequireAllDone(t, pool)
	}()

	pool.Start(throwCtx)
	unittest.RequireAllReady(t, pool)

	// Block the worker
	blocker := &mockJob{
		picked:   make(chan interface{}),
		executed: make(chan interface{}),
		block:    make(chan interface{}),
	}
	require.NoError(t, pool.Submit(blocker))

	// Wait for worker to pick up blocker
	unittest.ChannelMustCloseWithinTimeout(t, blocker.picked, 100*time.Millisecond, "blocker job not picked up on time")

	// Fill queue
	require.NoError(
		t, pool.Submit(
			&mockJob{
				picked:   make(chan interface{}),
				executed: make(chan interface{}),
			},
		),
	)

	// Queue full - should error
	err := pool.Submit(
		&mockJob{
			picked:   make(chan interface{}),
			executed: make(chan interface{}),
		},
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue full")

	// Unblock and cleanup
	close(blocker.block)
}

func TestPool_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	throwCtx := &cancellableThrowableContext{Context: ctx}

	pool := NewWorkerPool(10, 2)
	pool.Start(throwCtx)
	<-pool.Ready()

	// Submit blocking job
	job := &mockJob{
		picked:   make(chan interface{}),
		executed: make(chan interface{}),
		block:    make(chan interface{}),
	}
	require.NoError(t, pool.Submit(job))

	// Wait for worker to pick it up
	time.Sleep(10 * time.Millisecond)

	// Cancel context and unblock job
	cancel()
	close(job.block)

	// Pool should shutdown
	select {
	case <-pool.Done():
	case <-time.After(time.Second):
		t.Fatal("pool failed to shutdown on context cancel")
	}
}

func TestPool_JobPanic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	throwCtx := &cancellableThrowableContext{Context: ctx}
	pool := NewWorkerPool(10, 2)

	pool.Start(throwCtx)
	<-pool.Ready()

	// Submit job that throws
	panicJob := &mockJob{
		picked:   make(chan interface{}),
		executed: make(chan interface{}),
		panic:    true,
	}
	require.NoError(t, pool.Submit(panicJob))

	// Wait for throw
	require.Eventually(
		t, func() bool {
			throwCtx.mu.Lock()
			defer throwCtx.mu.Unlock()
			return throwCtx.thrown != nil
		}, time.Second, 10*time.Millisecond,
	)

	// Verify error was thrown
	throwCtx.mu.Lock()
	assert.Equal(t, assert.AnError, throwCtx.thrown)
	throwCtx.mu.Unlock()

	// Pool continues working
	normalJob := &mockJob{
		picked:   make(chan interface{}),
		executed: make(chan interface{}),
	}
	require.NoError(t, pool.Submit(normalJob))

	select {
	case <-normalJob.executed:
	case <-time.After(time.Second):
		t.Fatal("normal job did not execute")
	}

	cancel()
	<-pool.Done()
}

func TestPool_QueueSize(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	throwCtx := &cancellableThrowableContext{Context: ctx}
	pool := NewWorkerPool(10, 1)

	pool.Start(throwCtx)
	<-pool.Ready()

	assert.Equal(t, 0, pool.QueueSize())

	// Block worker
	blocker := &mockJob{
		picked:   make(chan interface{}),
		executed: make(chan interface{}),
		block:    make(chan interface{}),
	}
	require.NoError(t, pool.Submit(blocker))

	// Wait for worker to pick up blocker
	time.Sleep(10 * time.Millisecond)

	// Add to queue
	require.NoError(
		t, pool.Submit(
			&mockJob{
				picked:   make(chan interface{}),
				executed: make(chan interface{}),
			},
		),
	)
	require.Eventually(
		t, func() bool {
			return pool.QueueSize() == 1
		}, 100*time.Millisecond, 10*time.Millisecond,
	)

	require.NoError(
		t, pool.Submit(
			&mockJob{
				picked:   make(chan interface{}),
				executed: make(chan interface{}),
			},
		),
	)
	require.Eventually(
		t, func() bool {
			return pool.QueueSize() == 2
		}, 100*time.Millisecond, 10*time.Millisecond,
	)

	// Unblock
	close(blocker.block)

	// Queue drains
	require.Eventually(
		t, func() bool {
			return pool.QueueSize() == 0
		}, time.Second, 10*time.Millisecond,
	)
}

func TestPool_ConcurrentSubmit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	throwCtx := &cancellableThrowableContext{Context: ctx}
	pool := NewWorkerPool(100, 5)

	pool.Start(throwCtx)
	<-pool.Ready()

	// Concurrent submissions
	var wg sync.WaitGroup
	jobs := make([]*mockJob, 50)

	for i := range jobs {
		jobs[i] = &mockJob{
			picked:   make(chan interface{}),
			executed: make(chan interface{}),
		}
		wg.Add(1)
		go func(job *mockJob) {
			defer wg.Done()
			_ = pool.Submit(job)
		}(jobs[i])
	}

	wg.Wait()

	// All should execute
	for _, job := range jobs {
		select {
		case <-job.executed:
		case <-time.After(2 * time.Second):
			t.Fatal("job did not complete")
		}
	}
}
