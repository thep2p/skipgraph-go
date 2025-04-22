package mocknet

import (
	"fmt"
	"github/thep2p/skipgraph-go/model/messages"
	"github/thep2p/skipgraph-go/model/skipgraph"
	"github/thep2p/skipgraph-go/network"
	"sync"
)

// MockUnderlay keeps data necessary for processing of incoming network messages in a mock network
type MockUnderlay struct {
	l sync.Mutex
	// there is only one handler per message type (but not per caller)
	messageHandlers map[messages.Type]network.MessageHandler
	stub            *NetworkStub
}

// NewMockUnderlay initializes an empty MockUnderlay and returns a pointer to it
func newMockUnderlay(stub *NetworkStub) *MockUnderlay {
	return &MockUnderlay{
		stub:            stub,
		messageHandlers: make(map[messages.Type]network.MessageHandler),
	}
}

// SetMessageHandler determines the handler of a message based on its message type.
func (m *MockUnderlay) SetMessageHandler(msgType messages.Type, handler network.MessageHandler) error {
	m.l.Lock()
	defer m.l.Unlock()

	// check whether a handler exists for the supplied message type
	_, ok := m.messageHandlers[msgType]
	if ok {
		return fmt.Errorf("a handler exists for the attempted message type: %s", msgType)
	}
	m.messageHandlers[msgType] = handler
	return nil
}

// Send sends a message to a list of target recipients in the underlying network.
func (m *MockUnderlay) Send(msg messages.Message, target skipgraph.Identifier) error {
	m.l.Lock()
	defer m.l.Unlock()

	return m.stub.routeMessageTo(msg, target)
}

// Start starts a MockUnderlay
func (m *MockUnderlay) Start() <-chan interface{} {
	ch := make(chan interface{})
	close(ch)
	return ch
}

// Stop stops a MockUnderlay
func (m *MockUnderlay) Stop() <-chan interface{} {
	ch := make(chan interface{})
	close(ch)
	return ch
}
