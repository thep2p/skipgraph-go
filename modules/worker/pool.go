package worker

import (
	"fmt"
	"github.com/thep2p/skipgraph-go/modules/component"
	"sync"

	"github.com/rs/zerolog"
	"github.com/thep2p/skipgraph-go/modules"
)

// Pool manages a fixed number of goroutine workers for concurrent job execution.
// Fields:
//   - workerCount: number of concurrent workers
//   - queue: buffered channel holding pending jobs
//   - ready: signaled when all workers have started
//   - done: signaled when all workers have stopped
//   - wg: tracks active worker goroutines
//   - ctx: context for cancellation and error propagation
//   - logger: structured logger for trace-level events
type Pool struct {
	*component.Manager
	workerCount int
	queue       chan modules.Job
	wg          sync.WaitGroup
	ctx         modules.ThrowableContext
	logger      zerolog.Logger
}

// NewWorkerPool creates a new worker pool.
// Args:
//   - queueSize: buffer size for job queue (max pending jobs)
//   - workerCount: number of concurrent workers to spawn
//   - logger: zerolog.Logger for logging
//
// Returns initialized pool (not started).
func NewWorkerPool(queueSize int, workerCount int, logger zerolog.Logger) *Pool {
	logger = logger.With().
		Str("component", "worker_pool").
		Int("worker_count", workerCount).
		Int("queue_size", queueSize).
		Logger()

	logger.Trace().
		Msg("Creating new worker pool")

	p := &Pool{
		workerCount: workerCount,
		queue:       make(chan modules.Job, queueSize),
		logger:      logger,
	}

	p.Manager = component.NewManager(
		component.WithStartupLogic(func(ctx modules.ThrowableContext) {
			// Startup logic - store context
			p.ctx = ctx
			p.logger.Trace().Msg("Starting worker pool")
			// Start all workers
			for i := 0; i < p.workerCount; i++ {
				p.wg.Add(1)
				workerID := i
				p.logger.Trace().Int("worker_id", workerID).Msg("Starting worker")
				go p.runWorker(workerID)
			}
			// Signal ready immediately after workers start
			p.logger.Trace().Msg("all workers started, startup complete")
		}),
		component.WithShutdownLogic(func() {
			p.logger.Trace().Msg("initiating shutdown")

			// Close queue to signal workers to stop
			p.logger.Trace().Msg("Closing job queue")
			close(p.queue)

			// Wait for all workers to finish processing
			p.logger.Trace().Msg("Waiting for all workers to finish")
			p.wg.Wait()

			// Signal done after all workers have finished
			p.logger.Trace().Msg("All workers finished, shutdown complete")
		}),
	)

	return p
}

// runWorker executes jobs from the queue until shutdown.
// Args:
//   - workerID: unique identifier for logging
//
// Exits on context cancellation or queue closure.
func (p *Pool) runWorker(workerID int) {
	defer func() {
		p.logger.Trace().
			Int("worker_id", workerID).
			Msg("Worker shutting down")
		p.wg.Done()
	}()

	p.logger.Trace().
		Int("worker_id", workerID).
		Msg("Worker started")

	for {
		select {
		case <-p.ctx.Done():
			p.logger.Trace().
				Int("worker_id", workerID).
				Msg("Worker received context done signal")
			return
		case job, ok := <-p.queue:
			if !ok {
				// Queue closed, worker should exit
				p.logger.Trace().
					Int("worker_id", workerID).
					Msg("Worker queue closed")
				return
			}
			// Execute job - it handles its own errors via the throwable context
			p.logger.Trace().
				Int("worker_id", workerID).
				Msg("Worker executing job")
			job.Execute(p.ctx)
			p.logger.Trace().
				Int("worker_id", workerID).
				Msg("Worker completed job")
		}
	}
}

// Submit adds a job to the queue for processing.
// Args:
//   - job: work to be executed by a worker
//
// Returns error if queue is full (non-blocking). Caller should handle error.
func (p *Pool) Submit(job modules.Job) error {
	select {
	case p.queue <- job:
		p.logger.Trace().
			Msg("Job submitted to pool")
		return nil
	default:
		p.logger.Trace().
			Msg("Failed to submit job - queue full")
		return fmt.Errorf("queue full")
	}
}

// WorkerCount returns the number of workers in the pool.
// Returns configured worker count.
func (p *Pool) WorkerCount() int {
	return p.workerCount
}

// QueueSize returns current number of pending jobs.
// Returns count of jobs waiting in queue.
func (p *Pool) QueueSize() int {
	return len(p.queue)
}
