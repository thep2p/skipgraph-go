package modules

// Job represents a unit of work that can be executed by a worker.
// Jobs should be self-contained and include all necessary data for execution.
type Job interface {
	// Execute performs the job's work.
	// As jobs execute concurrently, there is no returned error.
	// Benign errors should be logged internally within the job implementation.
	// Fatal errors that require process termination should be thrown via ctx.ThrowIrrecoverable().
	// The context provided may be used for cancellation and timeouts.
	// This method should be safe to call concurrently from multiple goroutines.
	// It represents an atomic unit of work and should not depend on external state that may change during execution.
	// Jobs are responsible for their own error handling - either logging benign errors or throwing irrecoverable ones.
	// The worker pool will continue operating unless an irrecoverable error is thrown.
	Execute(ctx ThrowableContext)
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
