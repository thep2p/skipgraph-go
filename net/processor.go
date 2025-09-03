package net

import (
	"github/thep2p/skipgraph-go/model/skipgraph"
)

// MessageProcessor is the interface of an Engine that processes incoming messages from the network layer.
type MessageProcessor interface {
	// ProcessIncomingMessage is called by the network layer when a new message is received.
	// It processes the message and performs some actions.
	// Any error must be handled internally and not be returned.
	// Panics must be avoided at all costs as it exposes the whole node to a DoS attack.
	// Args:
	//  - channel: the channel on which the message was received.
	//  - originID: the identifier of the sender of the message.
	//  - msg: the message received.
	ProcessIncomingMessage(channel Channel, originID skipgraph.Identifier, msg interface{})
}
