package mocknet

import (
	"github/thep2p/skipgraph-go/core/model"
	"github/thep2p/skipgraph-go/net"
)

// MockMessageProcessor is a mock implementation of the MessageProcessor interface.
// It allows custom processing logic to be injected via a function.
type MockMessageProcessor struct {
	processLogic func(channel net.Channel, originID model.Identifier, msg net.Message)
}

func NewMockMessageProcessor(processLogic func(channel net.Channel, originID model.Identifier, msg net.Message)) *MockMessageProcessor {
	return &MockMessageProcessor{
		processLogic: processLogic,
	}
}

func (m *MockMessageProcessor) ProcessIncomingMessage(channel net.Channel, originID model.Identifier, msg net.Message) {
	m.processLogic(channel, originID, msg)
}

var _ net.MessageProcessor = (*MockMessageProcessor)(nil)
