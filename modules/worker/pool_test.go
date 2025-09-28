package worker

import (
	"context"
	"github.com/thep2p/skipgraph-go/modules/throwable"
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

// TestPool_HappyPath tests normal operation of the worker pool.
// It starts the pool, submits jobs, and verifies they execute.
// Also verifies pool state (worker count, queue size) at the end.
func TestPool_HappyPath(t *testing.T) {
	throwCtx := unittest.NewMockThrowableContext(t)
	logger := unittest.Logger(zerolog.TraceLevel)
	pool := NewWorkerPool(logger, 10, 3)
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

// TestPool_QueueFull tests that when the job queue is full,
// submitting a new job returns an error and does not block.
func TestPool_QueueFull(t *testing.T) {
	throwCtx := unittest.NewMockThrowableContext(t)
	logger := unittest.Logger(zerolog.TraceLevel)
	pool := NewWorkerPool(logger, 1, 1)
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

	// Wait for worker to pick up the blocker job, hence blocking the only worker of the pool.
	unittest.ChannelMustCloseWithinTimeout(t, blocker.picked, 100*time.Millisecond, "blocker job not picked up on time")

	// Fill queue
	secondJob := &mockJob{
		picked:   make(chan interface{}),
		executed: make(chan interface{}),
	}
	require.NoError(
		t, pool.Submit(
			secondJob,
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

	// Wait for blocker job to finish
	unittest.ChannelMustCloseWithinTimeout(t, blocker.executed, 100*time.Millisecond, "blocker job not executed on time")

	// Wait for second job to finish
	unittest.ChannelMustCloseWithinTimeout(t, secondJob.picked, 100*time.Millisecond, "second job not executed on time")
	unittest.ChannelMustCloseWithinTimeout(t, secondJob.executed, 100*time.Millisecond, "second job not executed on time")

	// Wait for queue to drain
	require.Eventually(
		t, func() bool {
			return pool.QueueSize() == 0
		}, time.Second, 10*time.Millisecond,
	)

	assert.Equal(t, 0, pool.QueueSize())
}

// TestPool_ContextCancellation tests that when the context is cancelled,
// the pool stops accepting new jobs, finishes executing current jobs,
// and then shuts down gracefully.
func TestPool_ContextCancellation(t *testing.T) {
	throwCtx := unittest.NewMockThrowableContext(t)
	logger := unittest.Logger(zerolog.TraceLevel)

	pool := NewWorkerPool(logger, 10, 2)
	pool.Start(throwCtx)

	unittest.RequireAllReady(t, pool)

	// Submit blocking job
	job := &mockJob{
		picked:   make(chan interface{}),
		executed: make(chan interface{}),
		block:    make(chan interface{}),
	}
	require.NoError(t, pool.Submit(job))

	// Wait for job to be picked up
	unittest.ChannelMustCloseWithinTimeout(t, job.picked, 100*time.Millisecond, "job not picked up on time")

	// Cancel context and unblock job
	throwCtx.Cancel()
	close(job.block)

	// Job should execute before pool shuts down
	unittest.ChannelMustCloseWithinTimeout(t, job.executed, 100*time.Millisecond, "job not executed on time")
	unittest.RequireAllDone(t, pool)
}

// TestPool_JobPanic tests that if a job panics (throws irrecoverable error),
// the pool captures the throw and continues operating.
func TestPool_JobPanic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	throwCtx := throwable.NewContext(ctx)

	logger := unittest.Logger(zerolog.TraceLevel)
	pool := NewWorkerPool(logger, 10, 2)
	defer func() {
		cancel()
		unittest.RequireAllDone(t, pool)
	}()

	pool.Start(throwCtx)

	// Submit job that throws
	panicJob := &mockJob{
		picked:   make(chan interface{}),
		executed: make(chan interface{}),
		panic:    true,
	}
	require.NoError(t, pool.Submit(panicJob))

	// Wait for job to be picked up
	unittest.ChannelMustCloseWithinTimeout(t, panicJob.picked, 100*time.Millisecond, "job not picked up on time")

	// The job should panic, which will be recovered by the worker
	// We verify the panic by checking that the job was executed (which means the panic was recovered)
	unittest.ChannelMustCloseWithinTimeout(t, panicJob.executed, 100*time.Millisecond, "job panic not handled properly")
}

// TestPool_QueueSize tests that the QueueSize method accurately reflects
// the number of pending jobs in the queue as jobs are submitted and processed.
func TestPool_QueueSize(t *testing.T) {
	throwCtx := unittest.NewMockThrowableContext(t)
	logger := unittest.Logger(zerolog.TraceLevel)
	pool := NewWorkerPool(logger, 10, 1)

	defer func() {
		throwCtx.Cancel()
		unittest.RequireAllDone(t, pool)
	}()

	pool.Start(throwCtx)
	unittest.RequireAllReady(t, pool)

	assert.Equal(t, 0, pool.QueueSize())

	// Block worker
	blocker := &mockJob{
		picked:   make(chan interface{}),
		executed: make(chan interface{}),
		block:    make(chan interface{}),
	}
	require.NoError(t, pool.Submit(blocker))

	// Wait for worker to pick up the blocker job, hence blocking the only worker of the pool.
	unittest.ChannelMustCloseWithinTimeout(t, blocker.picked, 100*time.Millisecond, "blocker job not picked up on time")

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

	secondJob := &mockJob{
		picked:   make(chan interface{}),
		executed: make(chan interface{}),
	}

	require.NoError(
		t, pool.Submit(
			secondJob,
		),
	)
	require.Eventually(
		t, func() bool {
			return pool.QueueSize() == 2
		}, 100*time.Millisecond, 10*time.Millisecond,
	)

	// Unblock
	close(blocker.block)

	// require blocked job to finish
	unittest.ChannelMustCloseWithinTimeout(t, blocker.executed, 100*time.Millisecond, "blocker job not executed on time")

	// require second job to be picked up and executed
	unittest.ChannelMustCloseWithinTimeout(t, secondJob.picked, 100*time.Millisecond, "second job not picked up on time")
	unittest.ChannelMustCloseWithinTimeout(t, secondJob.executed, 100*time.Millisecond, "second job not executed on time")

	// Queue drains
	require.Eventually(
		t, func() bool {
			return pool.QueueSize() == 0
		}, time.Second, 10*time.Millisecond,
	)
}

// TestPool_ConcurrentSubmit tests that multiple goroutines can concurrently
// submit jobs to the pool without errors or deadlocks, and all jobs execute.
func TestPool_ConcurrentSubmit(t *testing.T) {
	throwCtx := unittest.NewMockThrowableContext(t)
	logger := unittest.Logger(zerolog.TraceLevel)
	pool := NewWorkerPool(logger, 100, 5)

	defer func() {
		throwCtx.Cancel()
		unittest.RequireAllDone(t, pool)
	}()

	pool.Start(throwCtx)
	unittest.RequireAllReady(t, pool)

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

	unittest.CallMustReturnWithinTimeout(t, wg.Wait, 2*time.Second, "concurrent submissions did not complete on time")

	// All should execute
	executedChannels := make([]<-chan interface{}, len(jobs))
	for i, job := range jobs {
		executedChannels[i] = job.executed
	}
	unittest.ChannelsMustCloseWithinTimeout(t, 2*time.Second, "not all jobs executed on time", executedChannels...)
}

// TestPool_StartAlreadyStarted tests that starting an already started pool
// throws an irrecoverable error.
func TestPool_StartAlreadyStarted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	throwCtx := throwable.NewContext(ctx)
	logger := unittest.Logger(zerolog.TraceLevel)

	pool := NewWorkerPool(logger, 10, 3)
	defer func() {
		cancel()
		unittest.RequireAllDone(t, pool)
	}()

	// Start pool first time - should succeed
	pool.Start(throwCtx)
	unittest.RequireAllReady(t, pool)

	// Create a second context for the second start attempt
	ctx2, cancel2 := context.WithCancel(context.Background())
	throwCtx2 := throwable.NewContext(ctx2)
	defer cancel2()

	// Start pool second time - should panic
	assert.Panics(t, func() {
		pool.Start(throwCtx2)
	}, "expected panic for starting already started pool")
}
