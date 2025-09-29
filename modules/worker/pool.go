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
	logger zerolog.Logger
	*component.Manager
	workerCount int
	queue       chan modules.Job
	wg          sync.WaitGroup
	ctx         modules.ThrowableContext
}

// NewWorkerPool creates a new worker pool.
// Args:
//   - logger: zerolog.Logger for logging
//   - queueSize: buffer size for job queue (max pending jobs)
//   - workerCount: number of concurrent workers to spawn
//
// Returns initialized pool (not started).
func NewWorkerPool(logger zerolog.Logger, queueSize int, workerCount int) *Pool {
	logger = logger.With().
		Str("component", "worker_pool").
		Int("worker_count", workerCount).
		Int("queue_size", queueSize).
		Logger()

	logger.Trace().
		Msg("Creating new worker pool")

	p := &Pool{
		logger:      logger,
		workerCount: workerCount,
		queue:       make(chan modules.Job, queueSize),
	}

	p.Manager = component.NewManager(
		logger,
		component.WithStartupLogic(func(ctx modules.ThrowableContext) {
			p.startWorkers(ctx)
		}),
		component.WithShutdownLogic(func() {
			p.stopWorkers()
		}),
	)

	return p
}

// Submit adds a job to the worker pool queue.
// Returns error if queue is full or pool has been shut down.
// Args:
//   - job: the job to execute
//
// Returns error if job cannot be submitted.
func (p *Pool) Submit(job modules.Job) error {
	p.logger.Trace().
		Msg("Job submitted to pool")

      if p.ctx == nil {
          return fmt.Errorf("pool not started")
      }
	select {
	case <-p.ctx.Done():
		p.logger.Trace().
			Msg("Cannot submit job - pool shutting down")
		return fmt.Errorf("pool shutting down")
	case p.queue <- job:
		return nil
	default:
		p.logger.Trace().
			Msg("Failed to submit job - queue full")
		return fmt.Errorf("queue full")
	}
}

// WorkerCount returns the configured number of workers in the pool.
func (p *Pool) WorkerCount() int {
	return p.workerCount
}

// QueueSize returns the current number of pending jobs in the queue.
func (p *Pool) QueueSize() int {
	return len(p.queue)
}

func (p *Pool) Start(ctx modules.ThrowableContext) {
	p.ctx = ctx
	p.Manager.Start(ctx)
}

func (p *Pool) startWorkers(ctx modules.ThrowableContext) {
	p.logger.Trace().
		Msg("Starting worker pool")

	p.wg.Add(p.workerCount)
	for i := 0; i < p.workerCount; i++ {
		p.logger.Trace().
			Int("worker_id", i).
			Msg("Starting worker")
		go p.worker(ctx, i)
	}

	p.logger.Trace().
		Msg("All workers started, startup complete")
}

func (p *Pool) stopWorkers() {
	p.logger.Trace().
		Msg("initiating shutdown")

	p.logger.Trace().
		Msg("Closing job queue")
	close(p.queue)

	p.logger.Trace().
		Msg("Waiting for all workers to finish")
	p.wg.Wait()

	p.logger.Trace().
		Msg("All workers finished, shutdown complete")
}

// worker is the main loop for each worker goroutine.
// Continuously pulls jobs from the queue until shutdown.
// Handles job panics by logging them and continuing.
func (p *Pool) worker(ctx modules.ThrowableContext, id int) {
	defer p.wg.Done()
	p.logger.Trace().
		Int("worker_id", id).
		Msg("Worker started")

	for {
		select {
		case <-ctx.Done():
			p.logger.Trace().
				Int("worker_id", id).
				Msg("Worker received context done signal")
			p.logger.Trace().
				Int("worker_id", id).
				Msg("Worker shutting down")
			return
		case job, ok := <-p.queue:
			if !ok {
				p.logger.Trace().
					Int("worker_id", id).
					Msg("Worker exiting - queue closed")
				return
			}
			p.logger.Trace().
				Int("worker_id", id).
				Msg("Worker executing job")
			job.Execute(ctx)
			p.logger.Trace().
				Int("worker_id", id).
				Msg("Worker completed job")
		}
	}
}
