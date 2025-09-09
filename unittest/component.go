package unittest

import (
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/modules"
	"sync"
	"testing"
)

// MockComponent is a mock implementation of modules.Component for testing
type MockComponent struct {
	readyChan   chan interface{}
	doneChan    chan interface{}
	startCalled bool
	mu          sync.Mutex
	readyOnce   sync.Once
	doneOnce    sync.Once
	t           *testing.T
	readyLogic  func() // Optional logic to run when ready
	doneLogic   func() // Optional logic to run when done
}

func NewMockComponent(t *testing.T) *MockComponent {
	return &MockComponent{
		readyChan:  make(chan interface{}),
		doneChan:   make(chan interface{}),
		t:          t,
		readyLogic: func() {},
		doneLogic:  func() {},
	}
}

func NewMockComponentWithLogic(t *testing.T, readyLogic, doneLogic func()) *MockComponent {
	return &MockComponent{
		readyChan:  make(chan interface{}),
		doneChan:   make(chan interface{}),
		t:          t,
		readyLogic: readyLogic,
		doneLogic:  doneLogic,
	}
}

func (m *MockComponent) Start(ctx modules.ThrowableContext) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.startCalled {
		require.Fail(m.t, "component.Start() called multiple times")
	}
	m.startCalled = true

	// Execute ready logic in a separate goroutine and then close ready channel
	go func() {
		m.readyLogic() // Execute the ready blocking logic
		m.readyOnce.Do(
			func() {
				close(m.readyChan)
			},
		)
	}()

	// Wait for context to be done in a separate goroutine
	go func() {
		<-ctx.Done()
		m.doneLogic() // Execute the done blocking logic
		m.doneOnce.Do(
			func() {
				close(m.doneChan)
			},
		)
	}()
}

func (m *MockComponent) Ready() <-chan interface{} {
	return m.readyChan
}

func (m *MockComponent) Done() <-chan interface{} {
	return m.doneChan
}

var _ modules.Component = (*MockComponent)(nil)
