package worker

import (
	"fmt"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/thep2p/skipgraph-go/modules"
)

type Pool struct {
	workerCount int
	queue       chan modules.Job
	ready       chan interface{}
	done        chan interface{}
	wg          sync.WaitGroup
	ctx         modules.ThrowableContext
	logger      zerolog.Logger
}

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
		logger:      logger,
	}
}

func (p *Pool) Start(ctx modules.ThrowableContext) {
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

func (p *Pool) Ready() <-chan interface{} {
	return p.ready
}

func (p *Pool) Done() <-chan interface{} {
	return p.done
}

func (p *Pool) Submit(job modules.Job) error {
	select {
	case p.queue <- job:
		p.logger.Trace().
			Int("queue_size", len(p.queue)).
			Msg("Job submitted to pool")
		return nil
	default:
		p.logger.Trace().
			Int("queue_size", len(p.queue)).
			Msg("Failed to submit job - queue full")
		return fmt.Errorf("queue full")
	}
}

func (p *Pool) WorkerCount() int {
	return p.workerCount
}

func (p *Pool) QueueSize() int {
	return len(p.queue)
}
