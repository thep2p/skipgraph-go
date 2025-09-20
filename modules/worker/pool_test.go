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
	picked   chan struct{} // closed when picked up by worker
	executed chan struct{} // closed when executed
	block    chan struct{} // if non-nil, job blocks until channel is closed
	panic    bool          // if true, job panics when executed
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

func TestPool_HappyPath(t *testing.T) {
	throwCtx := unittest.NewMockThrowableContext(t)
	defer throwCtx.Cancel()
	pool := NewWorkerPool(10, 3)

	// Start pool
	pool.Start(throwCtx)

	// Wait for ready
	unittest.AllReady(t, pool)

	// Submit and execute jobs
	jobs := make([]*mockJob, 5)
	for i := range jobs {
		jobs[i] = &mockJob{}
		require.NoError(t, pool.Submit(jobs[i]))
	}

	// Wait for execution
	require.Eventually(
		t, func() bool {
			for _, job := range jobs {
				if !job.executed.Load() {
					return false
				}
			}
			return true
		}, time.Second, 10*time.Millisecond,
	)

	// Verify pool state; 3 workers, and empty queue at end.
	assert.Equal(t, 3, pool.WorkerCount())
	assert.Equal(t, 0, pool.QueueSize())
}

func TestPool_QueueFull(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	throwCtx := &cancellableThrowableContext{Context: ctx}
	pool := NewWorkerPool(1, 1)

	pool.Start(throwCtx)
	<-pool.Ready()

	// Block the worker
	blocker := &mockJob{block: make(chan struct{})}
	require.NoError(t, pool.Submit(blocker))

	// Wait for worker to pick up blocker
	time.Sleep(10 * time.Millisecond)

	// Fill queue
	require.NoError(t, pool.Submit(&mockJob{}))

	// Queue full - should error
	err := pool.Submit(&mockJob{})
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
	job := &mockJob{block: make(chan struct{})}
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
	panicJob := &mockJob{panic: true}
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
	normalJob := &mockJob{}
	require.NoError(t, pool.Submit(normalJob))

	require.Eventually(
		t, func() bool {
			return normalJob.executed.Load()
		}, time.Second, 10*time.Millisecond,
	)

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
	blocker := &mockJob{block: make(chan struct{})}
	require.NoError(t, pool.Submit(blocker))

	// Wait for worker to pick up blocker
	time.Sleep(10 * time.Millisecond)

	// Add to queue
	require.NoError(t, pool.Submit(&mockJob{}))
	require.Eventually(
		t, func() bool {
			return pool.QueueSize() == 1
		}, 100*time.Millisecond, 10*time.Millisecond,
	)

	require.NoError(t, pool.Submit(&mockJob{}))
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
		jobs[i] = &mockJob{}
		wg.Add(1)
		go func(job *mockJob) {
			defer wg.Done()
			_ = pool.Submit(job)
		}(jobs[i])
	}

	wg.Wait()

	// All should execute
	require.Eventually(
		t, func() bool {
			for _, job := range jobs {
				if !job.executed.Load() {
					return false
				}
			}
			return true
		}, 2*time.Second, 10*time.Millisecond,
	)
}
