package net

import (
	"github/thep2p/skipgraph-go/model/messages"
	"github/thep2p/skipgraph-go/model/skipgraph"
	"github/thep2p/skipgraph-go/modules"
)

// Network represents the underlying networking layer of a skip graph node.
type Network interface {
	modules.Startable
	modules.ReadyDoneAware

	// Register registers a MessageProcessor for a specific channel.
	// There must be exactly one MessageProcessor per channel on a node.
	// If a MessageProcessor is already registered for the given channel, an error is returned.
	// Any returned error must be treated as fatal.
	Register(Channel, MessageProcessor) (Conduit, error)
}

type Channel string

// Conduit is a high-level abstraction for sending messages to other nodes in the skip graph.
// It abstracts away the details of connection management and message serialization.
// Each conduit is associated with a specific channel.
type Conduit interface {
	// Send sends a message to the specified destination node defined by its identifier.
	// It establishes a connection to the destination node if one does not already exist.
	// Any returned error must be treated as benign, it should not cause the node to crash.
	Send(skipgraph.Identifier, messages.Message) error
}
