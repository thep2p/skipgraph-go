package network

import (
	"github/thep2p/skipgraph-go/model/messages"
	"github/thep2p/skipgraph-go/model/skipgraph"
)

// Underlay represents the underlying network and services therein
// A skip graph node interacts with other nodes using its Underlay.
type Underlay interface {
	// Start starts the networking layer.
	Start() <-chan interface{}

	// Stop stops the networking layer.
	Stop() <-chan interface{}

	// SetMessageHandler determines the handler of a message based on its message type.
	SetMessageHandler(messages.Type, MessageHandler) error

	// Send sends a message to a target recipient in the underlying network.
	Send(messages.Message, skipgraph.Identifier) error
}

// MessageHandler processes a Message and performs some actions
type MessageHandler func(messages.Message) error
