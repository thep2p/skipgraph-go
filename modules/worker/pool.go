package worker

import (
	"fmt"
	"github.com/thep2p/skipgraph-go/modules"
)

type Pool struct {
	workers []func(modules.Job)
	queue   chan modules.Job
	ready   chan interface{}
	done    chan interface{}
}

func NewWorkerPool(queueSize int, workerCount int, workerLogic func(modules.Job)) *Pool {
	pool := &Pool{
		workers: make([]func(modules.Job), workerCount),
		queue:   make(chan modules.Job, queueSize),
	}

	for i := range pool.workers {
		pool.workers[i] = workerLogic
	}

	return pool
}

func (p *Pool) Start(ctx modules.ThrowableContext) {
	go func() {
		close(p.ready)
		for _, w := range p.workers {
			select {
			case <-ctx.Done():
				// TODO: Log termination
				close(p.done)
				return
			case job := <-p.queue:
				w(job)
			}
		}
	}()
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
	return len(p.workers)
}

func (p *Pool) QueueSize() int {
	return len(p.queue)
}
