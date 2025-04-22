package unittest

import (
	"crypto/rand"
	"github.com/stretchr/testify/require"
	"github/thep2p/skipgraph-go/model"
	"github/thep2p/skipgraph-go/model/messages"
	"github/thep2p/skipgraph-go/model/skipgraph"
	"math/big"
	"testing"
)

/**
A utility module to generate random values of some certain type
*/

// TestMessageType  is a random message type.
const TestMessageType = messages.Type("test-message")

// TestMessageFixture generates a random Message.
func TestMessageFixture(t *testing.T) *messages.Message {

	return &messages.Message{
		Type:    TestMessageType,
		Payload: RandomBytesFixture(t, 100),
	}
}

// IdentifierFixture generates a random Identifier
func IdentifierFixture(t *testing.T) skipgraph.Identifier {
	var id skipgraph.Identifier
	bytes := RandomBytesFixture(t, skipgraph.IdentifierSizeBytes)

	for i := 0; i < skipgraph.IdentifierSizeBytes; i++ {
		id[i] = bytes[i]
	}

	return id
}

// RandomBytesFixture generates a random byte array of the supplied size.
func RandomBytesFixture(t *testing.T, size int) []byte {
	bytes := make([]byte, size)
	n, err := rand.Read(bytes[:])

	require.Equal(t, size, n)
	require.NoError(t, err)
	require.Len(t, bytes, size)

	return bytes
}

// MembershipVectorFixture creates and returns a random MemberShipVector.
func MembershipVectorFixture(t *testing.T) skipgraph.MembershipVector {
	bytes := RandomBytesFixture(t, skipgraph.MembershipVectorSize)

	var mv skipgraph.MembershipVector
	copy(mv[:], bytes)

	return mv
}

// AddressFixture returns an Address on localhost with a random port number.
func AddressFixture(t *testing.T) model.Address {
	// pick a random port
	max := big.NewInt(65535)
	randomInt, _ := rand.Int(rand.Reader, max)
	port := randomInt.String()
	addr := model.NewAddress("localhost", port)
	return addr

}

// IdentityFixture generates a random Identity with an address on localhost.
func IdentityFixture(t *testing.T) skipgraph.Identity {
	id := IdentifierFixture(t)
	memVec := MembershipVectorFixture(t)
	addr := AddressFixture(t)
	identity := skipgraph.NewIdentity(id, memVec, addr)
	return identity
}
