package unittest

import (
	"github/thep2p/skipgraph-go/modules"
	"sync"
)

// MockComponent is a mock implementation of modules.Component for testing
type MockComponent struct {
	readyChan   chan interface{}
	doneChan    chan interface{}
	startCalled bool
	mu          sync.Mutex
	readyOnce   sync.Once
	doneOnce    sync.Once
}

func NewMockComponent() *MockComponent {
	return &MockComponent{
		readyChan: make(chan interface{}),
		doneChan:  make(chan interface{}),
	}
}

func (m *MockComponent) Start(ctx modules.ThrowableContext) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.startCalled {
		panic("Start called multiple times")
	}
	m.startCalled = true
	m.readyOnce.Do(
		func() {
			close(m.readyChan)
		},
	)

	// Wait for context to be done in a separate goroutine
	go func() {
		select {
		case <-ctx.Done():
			m.doneOnce.Do(
				func() {
					close(m.doneChan)
				},
			)
		}
	}()
}

func (m *MockComponent) Ready() <-chan interface{} {
	return m.readyChan
}

func (m *MockComponent) Done() <-chan interface{} {
	return m.doneChan
}

var _ modules.Component = (*MockComponent)(nil)
