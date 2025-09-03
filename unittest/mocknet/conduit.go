package mocknet

import (
	"github/thep2p/skipgraph-go/model/messages"
	"github/thep2p/skipgraph-go/model/skipgraph"
	"github/thep2p/skipgraph-go/net"
)

type MockConduit struct {
	stub    *NetworkStub
	channel net.Channel
	id      skipgraph.Identifier
}

func (m MockConduit) Send(targetId skipgraph.Identifier, message messages.Message) error {
	return m.stub.routeMessageTo(m.channel, m.id, message, targetId)
}

var _ net.Conduit = (*MockConduit)(nil)
