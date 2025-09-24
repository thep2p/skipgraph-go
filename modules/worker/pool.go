package worker

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	workerCount int
	queue       chan modules.Job
	ready       chan interface{}
	done        chan interface{}
	started     chan interface{}
	wg          sync.WaitGroup
	ctx         modules.ThrowableContext
	logger      zerolog.Logger
}

// NewWorkerPool creates a new worker pool.
// Args:
//   - queueSize: buffer size for job queue (max pending jobs)
//   - workerCount: number of concurrent workers to spawn
//
// Returns initialized pool (not started).
func NewWorkerPool(queueSize int, workerCount int) *Pool {
	logger := log.With().
		Str("component", "worker_pool").
		Int("worker_count", workerCount).
		Int("queue_size", queueSize).
		Logger()

	logger.Trace().
		Msg("Creating new worker pool")

	return &Pool{
		workerCount: workerCount,
		queue:       make(chan modules.Job, queueSize),
		ready:       make(chan interface{}),
		done:        make(chan interface{}),
		started:     make(chan interface{}),
		logger:      logger,
	}
}

// Start initializes and begins worker execution.
// Args:
//   - ctx: context for cancellation and error propagation
//
// Spawns workers, signals ready, and monitors for shutdown.
func (p *Pool) Start(ctx modules.ThrowableContext) {
	// Prevent multiple starts
	select {
	case <-p.started:
		ctx.ThrowIrrecoverable(fmt.Errorf("worker pool already started"))
		return
	default:
		close(p.started)
	}

	p.ctx = ctx

	p.logger.Trace().
		Msg("Starting worker pool")

	// Start all workers
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		workerID := i
		p.logger.Trace().
			Int("worker_id", workerID).
			Msg("Starting worker")
		go p.runWorker(workerID)
	}

	// Signal ready immediately after workers start
	p.logger.Trace().
		Msg("All workers started, signaling ready")
	close(p.ready)

	// Monitor for shutdown
	go p.monitorShutdown()
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

// monitorShutdown coordinates graceful shutdown.
// Waits for context cancellation, closes queue, waits for workers, signals done.
func (p *Pool) monitorShutdown() {
	p.logger.Trace().
		Msg("Shutdown monitor started")

	// Wait for context cancellation
	<-p.ctx.Done()
	p.logger.Trace().
		Msg("Context cancelled, initiating shutdown")

	// Close queue to signal workers to stop
	p.logger.Trace().
		Msg("Closing job queue")
	close(p.queue)

	// Wait for all workers to finish processing
	p.logger.Trace().
		Msg("Waiting for all workers to finish")
	p.wg.Wait()

	// Signal done after all workers have finished
	p.logger.Trace().
		Msg("All workers finished, signaling done")
	close(p.done)
}

// Ready returns channel signaled when all workers have started.
// Returns read-only channel closed after successful startup.
func (p *Pool) Ready() <-chan interface{} {
	return p.ready
}

// Done returns channel signaled when all workers have stopped.
// Returns read-only channel closed after complete shutdown.
func (p *Pool) Done() <-chan interface{} {
	return p.done
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
