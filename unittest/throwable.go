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
}

func NewMockThrowableContext(t *testing.T) *MockThrowableContext {
	ctx, cancel := context.WithCancel(context.Background())
	return &MockThrowableContext{
		Context: ctx,
		cancel:  cancel,
		t:       t,
	}
}

func (m *MockThrowableContext) Cancel() {
	m.cancel()
}

func (m *MockThrowableContext) ThrowIrrecoverable(err error) {
	require.Fail(m.t, "irrecoverable error: "+err.Error())
}

var _ modules.ThrowableContext = (*MockThrowableContext)(nil)
