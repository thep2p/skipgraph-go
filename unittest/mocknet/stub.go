package mocknet

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github/thep2p/skipgraph-go/core/model"
	"github/thep2p/skipgraph-go/net"
	"sync"
	"testing"
)

// NetworkStub acts as a router to connect a set of MockNetwork
// it needs to be locked using its l field before being accessed
type NetworkStub struct {
	l        sync.Mutex
	networks map[model.Identifier]*MockNetwork
}

// NewNetworkStub creates an empty NetworkStub
func NewNetworkStub() *NetworkStub {
	return &NetworkStub{networks: make(map[model.Identifier]*MockNetwork)}
}

// NewMockNetwork creates and returns a mock network connected to this network stub for a non-existing Identifier.
func (n *NetworkStub) NewMockNetwork(t *testing.T, id model.Identifier) *MockNetwork {
	n.l.Lock()
	defer n.l.Unlock()

	_, exists := n.networks[id]
	require.False(t, exists, "attempting to create mock network for already existing identifier")

	u := newMockNetwork(id, n)
	n.networks[id] = u

	return u
}

// routeMessageTo imitates routing the message in the underlying network to the target identifier's mock network.
func (n *NetworkStub) routeMessageTo(channel net.Channel, originId model.Identifier, msg net.Message, target model.Identifier) error {
	n.l.Lock()
	defer n.l.Unlock()

	u, exists := n.networks[target]
	if !exists {
		return fmt.Errorf("no mock network exists for %x", target)
	}

	h, exists := u.messageProcessors[channel]
	if !exists {
		return fmt.Errorf("no handler exists for channel %v", channel)
	}

	h.ProcessIncomingMessage(channel, originId, msg)

	return nil
}
