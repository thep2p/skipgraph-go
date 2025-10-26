package unittest

import (
	"crypto/rand"
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/core"
	"github.com/thep2p/skipgraph-go/core/lookup"
	"github.com/thep2p/skipgraph-go/core/model"
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
func IdentifierFixture(t *testing.T) model.Identifier {
	var id model.Identifier
	bytes := RandomBytesFixture(t, model.IdentifierSizeBytes)

	for i := 0; i < model.IdentifierSizeBytes; i++ {
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
func MembershipVectorFixture(t *testing.T) model.MembershipVector {
	bytes := RandomBytesFixture(t, model.MembershipVectorSize)

	var mv model.MembershipVector
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
func IdentityFixture(t *testing.T) model.Identity {
	id := IdentifierFixture(t)
	memVec := MembershipVectorFixture(t)
	addr := AddressFixture(t)
	identity := model.NewIdentity(id, memVec, addr)
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

// RandomDirectionFixture generates a random direction (either DirectionLeft or DirectionRight).
// This is useful for testing Skip Graph operations that require direction values.
// The function uses cryptographic randomness to ensure fair distribution between the two directions.
//
// Args:
//   - t: the testing context
//
// Returns:
//   - Either types.DirectionLeft or types.DirectionRight with equal probability
func RandomDirectionFixture(t *testing.T) types.Direction {
	// Generate random bit (0 or 1)
	maxBig := big.NewInt(2)
	randomBig, err := rand.Int(rand.Reader, maxBig)
	require.NoError(t, err, "failed to generate random direction")

	if randomBig.Int64() == 0 {
		return types.DirectionLeft
	}
	return types.DirectionRight
}

// LookupTableOption is a functional option for configuring RandomLookupTable generation.
type LookupTableOption func(*lookupTableConfig)

// lookupTableConfig holds configuration for generating random lookup tables.
type lookupTableConfig struct {
	minID *model.Identifier // if set, all generated IDs must be greater than this
	maxID *model.Identifier // if set, all generated IDs must be less than this
}

// WithIdsGreaterThan configures the RandomLookupTable to generate all identifiers
// greater than the specified ID. This is useful for testing scenarios where nodes
// must have identifiers within a specific range.
//
// Args:
//   - id: the lower bound (exclusive) for generated identifiers
//
// Returns:
//   - A LookupTableOption that can be passed to RandomLookupTable
func WithIdsGreaterThan(id model.Identifier) LookupTableOption {
	return func(config *lookupTableConfig) {
		config.minID = &id
	}
}

// WithIdsLessThan configures the RandomLookupTable to generate all identifiers
// less than the specified ID. This is useful for testing scenarios where nodes
// must have identifiers within a specific range.
//
// Args:
//   - id: the upper bound (exclusive) for generated identifiers
//
// Returns:
//   - A LookupTableOption that can be passed to RandomLookupTable
func WithIdsLessThan(id model.Identifier) LookupTableOption {
	return func(config *lookupTableConfig) {
		config.maxID = &id
	}
}

// RandomLookupTable generates a random lookup table with neighbors at random levels.
// All neighbors have random identities (ID, membership vector, and address).
// The lookup table will have a random number of neighbors (0 to MaxLookupTableLevel)
// at random levels in random directions.
//
// Options:
//   - WithIdsGreaterThan: constrains all generated IDs to be greater than the specified ID
//   - WithIdsLessThan: constrains all generated IDs to be less than the specified ID
//
// Args:
//   - t: the testing context
//   - opts: optional configuration options
//
// Returns:
//   - A pointer to a randomly generated lookup.Table
//
// Example:
//
//	// Generate a lookup table with any random IDs
//	table := unittest.RandomLookupTable(t)
//
//	// Generate a lookup table with all IDs greater than someID
//	table := unittest.RandomLookupTable(t, unittest.WithIdsGreaterThan(someID))
//
//	// Generate a lookup table with IDs in a specific range
//	table := unittest.RandomLookupTable(t,
//	    unittest.WithIdsGreaterThan(minID),
//	    unittest.WithIdsLessThan(maxID))
func RandomLookupTable(t *testing.T, opts ...LookupTableOption) *lookup.Table {
	// Apply options
	config := &lookupTableConfig{}
	for _, opt := range opts {
		opt(config)
	}

	// Validate that minID < maxID if both are set
	if config.minID != nil && config.maxID != nil {
		comparison := config.minID.Compare(config.maxID)
		require.NotEqual(
			t, model.CompareGreater, comparison.GetComparisonResult(),
			"minID must be less than maxID",
		)
		require.NotEqual(
			t, model.CompareEqual, comparison.GetComparisonResult(),
			"minID must be less than maxID (cannot be equal)",
		)
	}

	table := &lookup.Table{}

	// Generate a random number of neighbors (0 to MaxLookupTableLevel)
	maxNeighbors := big.NewInt(int64(core.MaxLookupTableLevel + 1))
	neighborCountBig, err := rand.Int(rand.Reader, maxNeighbors)
	require.NoError(t, err, "failed to generate random neighbor count")
	neighborCount := int(neighborCountBig.Int64())

	// Generate random neighbors at random levels in random directions
	for i := 0; i < neighborCount; i++ {
		level := RandomLevelFixture(t)
		direction := RandomDirectionFixture(t)
		identity := randomIdentityWithConstraints(t, config)

		err := table.AddEntry(direction, level, identity)
		require.NoError(t, err, "failed to add entry to lookup table")
	}

	return table
}

// randomIdentityWithConstraints generates a random identity that satisfies the given constraints.
func randomIdentityWithConstraints(t *testing.T, config *lookupTableConfig) model.Identity {
	var id model.Identifier

	// If we have constraints, generate an ID that satisfies them
	if config.minID != nil || config.maxID != nil {
		id = randomIdentifierWithConstraints(t, config.minID, config.maxID)
	} else {
		// No constraints, generate a completely random ID
		id = IdentifierFixture(t)
	}

	memVec := MembershipVectorFixture(t)
	addr := AddressFixture(t)
	return model.NewIdentity(id, memVec, addr)
}

// randomIdentifierWithConstraints generates a random identifier that satisfies min/max constraints.
// If minID is set, the generated ID will be greater than minID.
// If maxID is set, the generated ID will be less than maxID.
// If both are set, the generated ID will be in the range (minID, maxID).
func randomIdentifierWithConstraints(
	t *testing.T,
	minID, maxID *model.Identifier,
) model.Identifier {
	maxAttempts := 1000
	for attempt := 0; attempt < maxAttempts; attempt++ {
		id := IdentifierFixture(t)

		// Check if ID satisfies minID constraint
		if minID != nil {
			comparison := id.Compare(minID)
			if comparison.GetComparisonResult() != model.CompareGreater {
				continue // ID is not greater than minID, try again
			}
		}

		// Check if ID satisfies maxID constraint
		if maxID != nil {
			comparison := id.Compare(maxID)
			if comparison.GetComparisonResult() != model.CompareLess {
				continue // ID is not less than maxID, try again
			}
		}

		// ID satisfies all constraints
		return id
	}

	// If we failed to generate a valid ID after maxAttempts, fail the test
	require.FailNow(
		t,
		"failed to generate identifier within constraints after %d attempts",
		maxAttempts,
	)
	return model.Identifier{} // unreachable
}
