package mocknet

import (
	"github/thep2p/skipgraph-go/model/messages"
	"github/thep2p/skipgraph-go/model/skipgraph"
	"github/thep2p/skipgraph-go/net"
)

// MockMessageProcessor is a mock implementation of the MessageProcessor interface.
// It allows custom processing logic to be injected via a function.
type MockMessageProcessor struct {
	processLogic func(channel net.Channel, originID skipgraph.Identifier, msg messages.Message)
}

func NewMockMessageProcessor(processLogic func(channel net.Channel, originID skipgraph.Identifier, msg messages.Message)) *MockMessageProcessor {
	return &MockMessageProcessor{
		processLogic: processLogic,
	}
}

func (m *MockMessageProcessor) ProcessIncomingMessage(channel net.Channel, originID skipgraph.Identifier, msg messages.Message) {
	m.processLogic(channel, originID, msg)
}

var _ net.MessageProcessor = (*MockMessageProcessor)(nil)
