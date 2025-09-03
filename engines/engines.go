package engines

import (
	"github/thep2p/skipgraph-go/modules"
	"github/thep2p/skipgraph-go/net"
)

// Engine represents a separate domain of functionality in a skip graph node.
// It is responsible for a specific aspect of the skip graph protocol.
// Engines on a node should must act independently of each other as much as possible,
// and share as little state as possible.
// Shared state should be injected at the time of engine creation.
type Engine interface {
	modules.ReadyDoneAware
	modules.Startable
	net.MessageProcessor
}
