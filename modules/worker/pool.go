package worker

import (
	"fmt"
	"sync"

	"github.com/thep2p/skipgraph-go/modules"
)

type Pool struct {
	workerCount int
	queue       chan modules.Job
	ready       chan interface{}
	done        chan interface{}
	wg          sync.WaitGroup
	ctx         modules.ThrowableContext
}

func NewWorkerPool(queueSize int, workerCount int) *Pool {
	return &Pool{
		workerCount: workerCount,
		queue:       make(chan modules.Job, queueSize),
		ready:       make(chan interface{}),
		done:        make(chan interface{}),
	}
}

func (p *Pool) Start(ctx modules.ThrowableContext) {
	p.ctx = ctx

	// Start all workers
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.runWorker(i)
	}

	// Signal ready immediately after workers start
	close(p.ready)

	// Monitor for shutdown
	go p.monitorShutdown()
}

func (p *Pool) runWorker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case job, ok := <-p.queue:
			if !ok {
				// Queue closed, worker should exit
				return
			}
			// Execute job - it handles its own errors via the throwable context
			job.Execute(p.ctx)
		}
	}
}

func (p *Pool) monitorShutdown() {
	// Wait for context cancellation
	<-p.ctx.Done()

	// Close queue to signal workers to stop
	close(p.queue)

	// Wait for all workers to finish processing
	p.wg.Wait()

	// Signal done after all workers have finished
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
		return nil
	default:
		return fmt.Errorf("queue full")
	}
}

func (p *Pool) WorkerCount() int {
	return p.workerCount
}

func (p *Pool) QueueSize() int {
	return len(p.queue)
}
