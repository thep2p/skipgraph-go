package mocknet

import (
	"fmt"
	"github/thep2p/skipgraph-go/modules"
	"github/thep2p/skipgraph-go/net"
	"sync"
)

// MockNetwork keeps data necessary for processing of incoming network messages in a mock network
type MockNetwork struct {
	l sync.Mutex
	// there is only one handler per message type (but not per caller)
	messageProcessors map[net.Channel]net.MessageProcessor
	stub              *NetworkStub
}

// Start is a no-op for the mock network.
func (m *MockNetwork) Start(ctx modules.ThrowableContext) {
	// No-op
}

// Ready returns a closed channel as there is nothing to wait for in the mock network
func (m *MockNetwork) Ready() <-chan interface{} {
	ch := make(chan interface{})
	close(ch)
	return ch
}

// Done returns a closed channel as there is nothing to wait for in the mock network
func (m *MockNetwork) Done() <-chan interface{} {
	ch := make(chan interface{})
	close(ch)
	return ch
}

func (m *MockNetwork) Register(channel net.Channel, processor net.MessageProcessor) (net.Conduit, error) {
	if _, exists := m.messageProcessors[channel]; exists {
		return nil, fmt.Errorf("message processor for channel %v already exists", channel)
	}
	m.l.Lock()
	m.messageProcessors[channel] = processor
	m.l.Unlock()
	return &MockConduit{
		channel: channel,
		stub:    m.stub,
	}, nil
}

// NewMockUnderlay initializes an empty MockNetwork and returns a pointer to it
func newMockNetwork(stub *NetworkStub) *MockNetwork {
	return &MockNetwork{
		stub:              stub,
		messageProcessors: make(map[net.Channel]net.MessageProcessor),
	}
}

var _ net.Network = (*MockNetwork)(nil)
