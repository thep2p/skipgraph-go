package unittest

import (
	"crypto/rand"
	"github.com/stretchr/testify/require"
	model2 "github/thep2p/skipgraph-go/core/model"
	"github/thep2p/skipgraph-go/net"
	"math/big"
	"testing"
)

/**
A utility module to generate random values of some certain type
*/

// TestMessageFixture generates a random Message.
func TestMessageFixture(t *testing.T) *net.Message {

	return &net.Message{
		Payload: RandomBytesFixture(t, 100),
	}
}

// IdentifierFixture generates a random Identifier
func IdentifierFixture(t *testing.T) model2.Identifier {
	var id model2.Identifier
	bytes := RandomBytesFixture(t, model2.IdentifierSizeBytes)

	for i := 0; i < model2.IdentifierSizeBytes; i++ {
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
func MembershipVectorFixture(t *testing.T) model2.MembershipVector {
	bytes := RandomBytesFixture(t, model2.MembershipVectorSize)

	var mv model2.MembershipVector
	copy(mv[:], bytes)

	return mv
}

// AddressFixture returns an Address on localhost with a random port number.
func AddressFixture(t *testing.T) model2.Address {
	// pick a random port
	max := big.NewInt(65535)
	randomInt, _ := rand.Int(rand.Reader, max)
	port := randomInt.String()
	addr := model2.NewAddress("localhost", port)
	return addr

}

// IdentityFixture generates a random Identity with an address on localhost.
func IdentityFixture(t *testing.T) model2.Identity {
	id := IdentifierFixture(t)
	memVec := MembershipVectorFixture(t)
	addr := AddressFixture(t)
	identity := model2.NewIdentity(id, memVec, addr)
	return identity
}
