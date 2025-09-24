package unittest

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/modules"
	"testing"
)

// MockThrowableContext is a mock implementation of modules.ThrowableContext for testing purposes.
// It fails the test if ThrowIrrecoverable is called.
// Other than that it behaves like a no-op context.
type MockThrowableContext struct {
	context.Context
	cancel context.CancelFunc
	t      *testing.T
	throw  func(err error) // Optional logic to run when ThrowIrrecoverable is called
}

func WithThrowLogic(throwLogic func(err error)) func(*MockThrowableContext) {
	return func(m *MockThrowableContext) {
		m.throw = throwLogic
	}
}

func NewMockThrowableContext(t *testing.T, opts ...func(*MockThrowableContext)) *MockThrowableContext {
	ctx, cancel := context.WithCancel(context.Background())
	throwCtx := &MockThrowableContext{
		Context: ctx,
		cancel:  cancel,
		t:       t,
		throw: func(err error) {
			require.Fail(t, "irrecoverable error: "+err.Error())
		},
	}
	for _, opt := range opts {
		opt(throwCtx)
	}

	return throwCtx
}

func (m *MockThrowableContext) Cancel() {
	m.cancel()
}

func (m *MockThrowableContext) ThrowIrrecoverable(err error) {
	m.throw(err)
}

var _ modules.ThrowableContext = (*MockThrowableContext)(nil)
