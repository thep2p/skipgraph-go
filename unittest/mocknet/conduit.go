package mocknet

import (
	"github/thep2p/skipgraph-go/core/model"
	"github/thep2p/skipgraph-go/net"
)

type MockConduit struct {
	stub    *NetworkStub
	channel net.Channel
	id      model.Identifier
}

func (m MockConduit) Send(targetId model.Identifier, message net.Message) error {
	return m.stub.routeMessageTo(m.channel, m.id, message, targetId)
}

var _ net.Conduit = (*MockConduit)(nil)
