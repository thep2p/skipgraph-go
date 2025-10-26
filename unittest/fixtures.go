package unittest

import (
	"crypto/rand"
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/core"
	model2 "github.com/thep2p/skipgraph-go/core/model"
	"github.com/thep2p/skipgraph-go/core/types"
	"github.com/thep2p/skipgraph-go/net"
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

// RandomLevelFixture generates a random level between 0 and MaxLookupTableLevel-1 (inclusive).
// This is useful for testing Skip Graph operations that require valid level values.
// The returned level is guaranteed to be within the valid range for Skip Graph lookup tables.
func RandomLevelFixture(t *testing.T) types.Level {
	return RandomLevelWithMaxFixture(t, core.MaxLookupTableLevel)
}

// RandomLevelWithMaxFixture generates a random level between 0 and max-1 (inclusive).
// This allows testing with custom maximum level bounds.
// The max parameter must be greater than 0, otherwise the function will fail the test.
//
// Args:
//   - t: the testing context
//   - max: the exclusive upper bound for level generation (must be > 0)
//
// Returns:
//   - A random level in the range [0, max-1]
func RandomLevelWithMaxFixture(t *testing.T, max types.Level) types.Level {
	require.Greater(t, max, types.Level(0), "max must be greater than 0")

	// Generate random number in range [0, max-1]
	maxBig := big.NewInt(int64(max))
	randomBig, err := rand.Int(rand.Reader, maxBig)
	require.NoError(t, err, "failed to generate random level")

	level := types.Level(randomBig.Int64())

	// Verify the generated level is within bounds
	require.GreaterOrEqual(t, level, types.Level(0), "generated level must be >= 0")
	require.Less(t, level, max, "generated level must be < max")

	return level
}
