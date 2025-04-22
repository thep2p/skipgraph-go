package mocknet_test

import (
	"github/thep2p/skipgraph-go/model/messages"
	"github/thep2p/skipgraph-go/unittest"
	"github/thep2p/skipgraph-go/unittest/mocknet"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestTwoUnderlays checks two mock underlays can send message to each other
func TestTwoUnderlays(t *testing.T) {
	// construct an empty mocked underlay
	stub := mocknet.NewNetworkStub()

	// create a random identifier
	id1 := unittest.IdentifierFixture(t)
	u1 := stub.NewMockUnderlay(t, id1)

	// create a random identifier
	id2 := unittest.IdentifierFixture(t)
	u2 := stub.NewMockUnderlay(t, id2)

	// make sure they are not equal
	require.NotEqual(t, id1, id2)

	// starts underlay
	unittest.ChannelsMustCloseWithinTimeout(t,
		100*time.Millisecond, "could not start underlays on time", u1.Start(), u2.Start())

	// sets message handler at u1
	received := false
	var receivedPayload interface{}
	f := func(msg messages.Message) error {
		received = true
		receivedPayload = msg.Payload
		return nil
	}
	require.NoError(t, u1.SetMessageHandler(unittest.TestMessageType, f))

	// sends message from u2 -> u1
	msg := unittest.TestMessageFixture(t)
	// TODO: refactor message as an interface
	// TODO: add test for u1 -> u2
	require.NoError(t, u2.Send(*msg, id1))

	// the handler is called
	require.True(t, received)
	require.Equal(t, msg.Payload, receivedPayload)

	// stops underlay
	unittest.ChannelsMustCloseWithinTimeout(t,
		100*time.Millisecond, "could not stop underlay on time", u1.Stop(), u2.Stop())
}
