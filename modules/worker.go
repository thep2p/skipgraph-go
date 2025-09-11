package modules

import "github.com/thep2p/skipgraph-go/modules/throwable"

// Job represents a unit of work that can be executed by a worker.
// Jobs should be self-contained and include all necessary data for execution.
type Job interface {
	// Execute performs the job's work.
	// As jobs execute concurrently, there is no returned error. Benign errors should be logged internally.
	// Fatal errors that should crash the process by throwing at ctx.
	// The context provided may be used for cancellation and timeouts.
	// This method should be safe to call concurrently from multiple goroutines.
	// It represents an atomic unit of work; it should not depend on external state that may change during execution.
	// If the job cannot be completed successfully, it should return a non-nil error.
	// The caller is responsible for handling retries or failure logging as needed.
	// A job should not be cancelled mid-execution; it should either complete successfully or fail with an error.
	Execute(ctx throwable.Context)
}

// WorkerPool manages a pool of workers for concurrent job execution.
// The pool distributes submitted jobs among available workers and provides
// lifecycle management through the Component interface.
type WorkerPool interface {
	Component

	// Submit adds a job to the work queue for processing.
	// Returns an error if the pool is not running or cannot accept more jobs.
	// This method should be non-blocking; it may return an error if the
	// submission queue is full rather than blocking indefinitely.
	Submit(job Job) error

	// WorkerCount returns the current number of active workers in the pool.
	WorkerCount() int

	// QueueSize returns the current number of jobs waiting to be processed.
	QueueSize() int
}
