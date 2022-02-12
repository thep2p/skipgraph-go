package mocknet

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"github/yhassanzadeh13/skipgraph-go/model/messages"
	"github/yhassanzadeh13/skipgraph-go/model/skipgraph"
	"sync"
	"testing"
)

type NetworkStub struct {
	l sync.Mutex
	underlays map[skipgraph.Identifier]*MockUnderlay
}

func NewNetworkStub() *NetworkStub{
	return &NetworkStub{underlays: make(map[skipgraph.Identifier]*MockUnderlay)}
}

// NewMockUnderlay creates and returns a mock underlay connected to this network stub.
func (n *NetworkStub) NewMockUnderlay(t *testing.T, id skipgraph.Identifier) *MockUnderlay{
	n.l.Lock()
	defer n.l.Unlock()

	_, exists := n.underlays[id]
	require.False(t, exists, "attempting to create mock underlay for already existing identifier")

	u := newMockUnderlay(n)
	n.underlays[id] = u

	return u
}

// routeMessageTo imitates routing the message in the underlying network to the target identifier's mock underlay.
func (n *NetworkStub) routeMessageTo(msg messages.Message, target skipgraph.Identifier) error {
	n.l.Lock()
	defer n.l.Unlock()

	u, exists := n.underlays[target]
	if !exists {
		return fmt.Errorf("no mock underlay exists for %x", target)
	}

	h, exists := u.messageHandlers[msg.Type]
	if !exists {
		return fmt.Errorf("no handler exists for message type %v", msg.Type)
	}

	err := h(msg)
	if err != nil {
		return fmt.Errorf("mock underlay handler could not handler message %w", err)
	}

	return nil
}



