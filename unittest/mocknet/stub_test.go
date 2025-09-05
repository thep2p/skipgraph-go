package mocknet_test

import (
	"github/thep2p/skipgraph-go/model/messages"
	"github/thep2p/skipgraph-go/model/skipgraph"
	"github/thep2p/skipgraph-go/net"
	"github/thep2p/skipgraph-go/unittest"
	"github/thep2p/skipgraph-go/unittest/mocknet"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestTwoNetworks checks two mock networks can send message to each other
func TestTwoNetworks(t *testing.T) {
	// construct an empty mocked network
	stub := mocknet.NewNetworkStub()

	// create a random identifier
	id1 := unittest.IdentifierFixture(t)
	u1 := stub.NewMockNetwork(t, id1)

	// create a random identifier
	id2 := unittest.IdentifierFixture(t)
	u2 := stub.NewMockNetwork(t, id2)

	// make sure they are not equal
	require.NotEqual(t, id1, id2)

	tCtx := unittest.NewMockThrowableContext(t)
	u1.Start(tCtx)
	u2.Start(tCtx)

	// starts network
	unittest.ChannelsMustCloseWithinTimeout(
		t,
		100*time.Millisecond, "could not start networks on time", u1.Ready(), u2.Ready(),
	)

	// sets message handler at u1
	received := false
	var receivedPayload interface{}
	f := func(channel net.Channel, originId skipgraph.Identifier, msg messages.Message) {
		received = true
		receivedPayload = msg.Payload
		require.Equal(t, id2, originId)
	}
	_, err := u1.Register(net.TestChannel, mocknet.NewMockMessageProcessor(f))
	require.NoError(t, err)

	// sends message from u2 -> u1
	con2, err := u2.Register(
		net.TestChannel, mocknet.NewMockMessageProcessor(
			func(channel net.Channel, originID skipgraph.Identifier, msg messages.Message) {
				// No-op, just to satisfy the interface, u2 does not expect to receive messages in this test
			},
		),
	)
	require.NoError(t, err)
	msg := unittest.TestMessageFixture(t)
	// TODO: refactor message as an interface
	// TODO: add test for u1 -> u2
	require.NoError(t, con2.Send(id1, *msg))

	// the handler is called
	require.True(t, received)
	require.Equal(t, msg.Payload, receivedPayload)

	// stops network
	unittest.ChannelsMustCloseWithinTimeout(
		t,
		100*time.Millisecond, "could not stop network on time", u1.Done(), u2.Done(),
	)
}
