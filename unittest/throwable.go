package unittest

import (
	"github.com/stretchr/testify/require"
	"github/thep2p/skipgraph-go/modules"
	"testing"
	"time"
)

// MockThrowableContext is a mock implementation of modules.ThrowableContext for testing purposes.
// It fails the test if ThrowIrrecoverable is called.
// Other than that it behaves like a no-op context.
type MockThrowableContext struct {
	t *testing.T
}

func NewMockThrowableContext(t *testing.T) *MockThrowableContext {
	return &MockThrowableContext{t: t}
}

func (m *MockThrowableContext) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (m *MockThrowableContext) Done() <-chan struct{} {
	done := make(chan struct{})
	close(done)
	return done
}

func (m *MockThrowableContext) Err() error {
	return nil
}

func (m *MockThrowableContext) Value(_ any) any {
	return nil
}

func (m *MockThrowableContext) ThrowIrrecoverable(err error) {
	require.Fail(m.t, "irrecoverable error: "+err.Error())
}

var _ modules.ThrowableContext = (*MockThrowableContext)(nil)
