package throwable

import (
	"context"
	"fmt"
	"time"
)

// Context is a context that can propagate irrecoverable errors up the context chain.
// If an irrecoverable error is thrown, it will propagate to the parent context if it exists.
// If there is no parent context, it will log the error and terminate the program.
// This is useful for components that need to signal fatal errors that should stop the entire application.
// Application: any error during startup that should stop the application from running.
// This streamlines error handling during startup by avoiding repetitive error checks and propagations.
type Context struct {
	ctx context.Context
}

func NewContext(ctx context.Context) *Context {
	return &Context{ctx: ctx}
}

var _ context.Context = (*Context)(nil)

// ThrowIrrecoverable propagates an irrecoverable error up the context chain.
// When it reaches the top-level context, it panics with the error.
func (t *Context) ThrowIrrecoverable(err error) {
	// Propagate the error to the parent context if it exists
	if parent, ok := t.ctx.(*Context); ok {
		parent.ThrowIrrecoverable(err)
	}
	// If there is no parent context, panic with the error.
	panic(fmt.Errorf("irrecoverable error: %w", err))
}

// Deadline returns the underlying context's deadline.
func (t *Context) Deadline() (deadline time.Time, ok bool) {
	return t.ctx.Deadline()
}

// Done returns the underlying context's done channel.
func (t *Context) Done() <-chan struct{} {
	return t.ctx.Done()
}

// Err returns the underlying context's error.
func (t *Context) Err() error {
	return t.ctx.Err()
}

// Value returns the value associated with the key in the underlying context.
func (t *Context) Value(key any) any {
	return t.ctx.Value(key)
}
