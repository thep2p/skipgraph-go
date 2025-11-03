package unittest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/core"
	"github.com/thep2p/skipgraph-go/core/model"
	"github.com/thep2p/skipgraph-go/core/types"
)

// TestRandomLevelFixture tests the RandomLevelFixture function.
func TestRandomLevelFixture(t *testing.T) {
	// Generate multiple random levels to ensure they're all valid
	for i := 0; i < 100; i++ {
		level := RandomLevelFixture(t)
		require.GreaterOrEqual(t, level, types.Level(0))
		require.Less(t, level, core.MaxLookupTableLevel)
	}
}

// TestRandomLevelWithMaxFixture tests the RandomLevelWithMaxFixture function.
func TestRandomLevelWithMaxFixture(t *testing.T) {
	t.Run(
		"with custom max", func(t *testing.T) {
			max := types.Level(10)
			// Generate multiple random levels to ensure they're all valid
			for i := 0; i < 100; i++ {
				level := RandomLevelWithMaxFixture(t, max)
				require.GreaterOrEqual(t, level, types.Level(0))
				require.Less(t, level, max)
			}
		},
	)

	t.Run(
		"with max of 1", func(t *testing.T) {
			max := types.Level(1)
			// Should always return 0
			for i := 0; i < 10; i++ {
				level := RandomLevelWithMaxFixture(t, max)
				require.Equal(t, types.Level(0), level)
			}
		},
	)

	t.Run(
		"with max of 2", func(t *testing.T) {
			max := types.Level(2)
			foundZero := false
			foundOne := false

			// Generate enough samples to likely get both 0 and 1
			for i := 0; i < 100 && !(foundZero && foundOne); i++ {
				level := RandomLevelWithMaxFixture(t, max)
				require.GreaterOrEqual(t, level, types.Level(0))
				require.Less(t, level, max)

				if level == 0 {
					foundZero = true
				}
				if level == 1 {
					foundOne = true
				}
			}

			// With 100 samples, we should have seen both values
			require.True(t, foundZero, "should have generated 0 at least once")
			require.True(t, foundOne, "should have generated 1 at least once")
		},
	)

	t.Run(
		"with MaxLookupTableLevel", func(t *testing.T) {
			// Should work with the actual max level
			for i := 0; i < 100; i++ {
				level := RandomLevelWithMaxFixture(t, core.MaxLookupTableLevel)
				require.GreaterOrEqual(t, level, types.Level(0))
				require.Less(t, level, core.MaxLookupTableLevel)
			}
		},
	)
}

// TestRandomDirectionFixture tests the RandomDirectionFixture function.
func TestRandomDirectionFixture(t *testing.T) {
	t.Run(
		"generates valid directions", func(t *testing.T) {
			// Generate multiple random directions to ensure they're all valid
			for i := 0; i < 100; i++ {
				direction := RandomDirectionFixture(t)
				// Direction must be either DirectionLeft or DirectionRight
				require.True(
					t,
					direction == types.DirectionLeft || direction == types.DirectionRight,
					"direction must be either DirectionLeft or DirectionRight",
				)
			}
		},
	)

	t.Run(
		"generates both directions", func(t *testing.T) {
			foundLeft := false
			foundRight := false

			// Generate enough samples to likely get both directions
			for i := 0; i < 100 && !(foundLeft && foundRight); i++ {
				direction := RandomDirectionFixture(t)

				if direction == types.DirectionLeft {
					foundLeft = true
				}
				if direction == types.DirectionRight {
					foundRight = true
				}
			}

			// With 100 samples, we should have seen both directions
			require.True(t, foundLeft, "should have generated DirectionLeft at least once")
			require.True(t, foundRight, "should have generated DirectionRight at least once")
		},
	)
}

// TestRandomLookupTable tests the RandomLookupTable function.
func TestRandomLookupTable(t *testing.T) {
	t.Run(
		"generates full lookup table without constraints", func(t *testing.T) {
			// Generate multiple lookup tables to ensure they're all valid
			for i := 0; i < 10; i++ {
				table := RandomLookupTable(t)
				require.NotNil(t, table, "lookup table should not be nil")

				// Verify all levels have neighbors in both directions
				for level := types.Level(0); level < core.MaxLookupTableLevel; level++ {
					// Left neighbor should exist
					leftEntry, err := table.GetEntry(types.DirectionLeft, level)
					require.NoError(t, err, "GetEntry should not return error for valid level")
					require.NotNil(t, leftEntry, "left entry should exist at level %d", level)

					// Right neighbor should exist
					rightEntry, err := table.GetEntry(types.DirectionRight, level)
					require.NoError(t, err, "GetEntry should not return error for valid level")
					require.NotNil(t, rightEntry, "right entry should exist at level %d", level)
				}
			}
		},
	)

	t.Run(
		"WithIdsGreaterThan generates all IDs greater than constraint", func(t *testing.T) {
			// Create a constraint ID
			constraintID := IdentifierFixture(t)

			// Generate multiple lookup tables with the constraint
			for i := 0; i < 10; i++ {
				table := RandomLookupTable(t, WithIdsGreaterThan(constraintID))
				require.NotNil(t, table)

				// Check all neighbors in the table
				for level := types.Level(0); level < core.MaxLookupTableLevel; level++ {
					// Check left neighbor
					leftEntry, err := table.GetEntry(types.DirectionLeft, level)
					require.NoError(t, err)
					require.NotNil(t, leftEntry, "left entry should exist at level %d", level)
					id := leftEntry.GetIdentifier()
					comparison := id.Compare(&constraintID)
					require.Equal(
						t, model.CompareGreater, comparison.GetComparisonResult(),
						"left neighbor ID at level %d should be greater than constraint: %s",
						level, comparison.DebugInfo(),
					)

					// Check right neighbor
					rightEntry, err := table.GetEntry(types.DirectionRight, level)
					require.NoError(t, err)
					require.NotNil(t, rightEntry, "right entry should exist at level %d", level)
					id = rightEntry.GetIdentifier()
					comparison = id.Compare(&constraintID)
					require.Equal(
						t, model.CompareGreater, comparison.GetComparisonResult(),
						"right neighbor ID at level %d should be greater than constraint: %s",
						level, comparison.DebugInfo(),
					)
				}
			}
		},
	)

	t.Run(
		"WithIdsLessThan generates all IDs less than constraint", func(t *testing.T) {
			// Create a constraint ID
			constraintID := IdentifierFixture(t)

			// Generate multiple lookup tables with the constraint
			for i := 0; i < 10; i++ {
				table := RandomLookupTable(t, WithIdsLessThan(constraintID))
				require.NotNil(t, table)

				// Check all neighbors in the table
				for level := types.Level(0); level < core.MaxLookupTableLevel; level++ {
					// Check left neighbor
					leftEntry, err := table.GetEntry(types.DirectionLeft, level)
					require.NoError(t, err)
					require.NotNil(t, leftEntry, "left entry should exist at level %d", level)
					id := leftEntry.GetIdentifier()
					comparison := id.Compare(&constraintID)
					require.Equal(
						t, model.CompareLess, comparison.GetComparisonResult(),
						"left neighbor ID at level %d should be less than constraint: %s",
						level, comparison.DebugInfo(),
					)

					// Check right neighbor
					rightEntry, err := table.GetEntry(types.DirectionRight, level)
					require.NoError(t, err)
					require.NotNil(t, rightEntry, "right entry should exist at level %d", level)
					id = rightEntry.GetIdentifier()
					comparison = id.Compare(&constraintID)
					require.Equal(
						t, model.CompareLess, comparison.GetComparisonResult(),
						"right neighbor ID at level %d should be less than constraint: %s",
						level, comparison.DebugInfo(),
					)
				}
			}
		},
	)

	t.Run(
		"WithIdsGreaterThan and WithIdsLessThan combined", func(t *testing.T) {
			// Create two constraint IDs where minID < maxID
			minID := IdentifierFixture(t)
			maxID := IdentifierFixture(t)

			// Ensure minID < maxID
			comparison := minID.Compare(&maxID)
			if comparison.GetComparisonResult() == model.CompareGreater {
				// Swap them
				minID, maxID = maxID, minID
			} else if comparison.GetComparisonResult() == model.CompareEqual {
				// Generate a completely new ID instead of incrementing
				maxID = IdentifierFixture(t)
				// Ensure it's different
				newComparison := maxID.Compare(&minID)
				for newComparison.GetComparisonResult() == model.CompareEqual {
					maxID = IdentifierFixture(t)
					newComparison = maxID.Compare(&minID)
				}
				// Ensure minID < maxID by swapping if needed
				if newComparison.GetComparisonResult() == model.CompareLess {
					minID, maxID = maxID, minID
				}
			}

			// Generate multiple lookup tables with both constraints
			for i := 0; i < 10; i++ {
				table := RandomLookupTable(t, WithIdsGreaterThan(minID), WithIdsLessThan(maxID))
				require.NotNil(t, table)

				// Check all neighbors in the table
				for level := types.Level(0); level < core.MaxLookupTableLevel; level++ {
					// Check left neighbor
					leftEntry, err := table.GetEntry(types.DirectionLeft, level)
					require.NoError(t, err)
					require.NotNil(t, leftEntry, "left entry should exist at level %d", level)
					id := leftEntry.GetIdentifier()

					// Check greater than minID
					minComparison := id.Compare(&minID)
					require.Equal(
						t, model.CompareGreater, minComparison.GetComparisonResult(),
						"left neighbor ID at level %d should be greater than minID: %s",
						level, minComparison.DebugInfo(),
					)

					// Check less than maxID
					maxComparison := id.Compare(&maxID)
					require.Equal(
						t, model.CompareLess, maxComparison.GetComparisonResult(),
						"left neighbor ID at level %d should be less than maxID: %s",
						level, maxComparison.DebugInfo(),
					)

					// Check right neighbor
					rightEntry, err := table.GetEntry(types.DirectionRight, level)
					require.NoError(t, err)
					require.NotNil(t, rightEntry, "right entry should exist at level %d", level)
					id = rightEntry.GetIdentifier()

					// Check greater than minID
					minComparison = id.Compare(&minID)
					require.Equal(
						t, model.CompareGreater, minComparison.GetComparisonResult(),
						"right neighbor ID at level %d should be greater than minID: %s",
						level, minComparison.DebugInfo(),
					)

					// Check less than maxID
					maxComparison = id.Compare(&maxID)
					require.Equal(
						t, model.CompareLess, maxComparison.GetComparisonResult(),
						"right neighbor ID at level %d should be less than maxID: %s",
						level, maxComparison.DebugInfo(),
					)
				}
			}
		},
	)

	t.Run(
		"neighbors have complete identities", func(t *testing.T) {
			table := RandomLookupTable(t)
			require.NotNil(t, table)

			// Check that all neighbors have complete identities
			for level := types.Level(0); level < core.MaxLookupTableLevel; level++ {
				leftEntry, err := table.GetEntry(types.DirectionLeft, level)
				require.NoError(t, err)
				require.NotNil(t, leftEntry, "left entry should exist at level %d", level)

				// Verify identifier is not all zeros
				id := leftEntry.GetIdentifier()
				idAllZeros := true
				for _, b := range id {
					if b != 0 {
						idAllZeros = false
						break
					}
				}
				require.False(
					t,
					idAllZeros,
					"left neighbor identifier at level %d should not be all zeros",
					level,
				)

				// Verify membership vector is not all zeros
				memVec := leftEntry.GetMembershipVector()
				mvAllZeros := true
				for _, b := range memVec {
					if b != 0 {
						mvAllZeros = false
						break
					}
				}
				require.False(
					t,
					mvAllZeros,
					"left neighbor membership vector at level %d should not be all zeros",
					level,
				)

				// Verify address has valid hostname and port
				addr := leftEntry.GetAddress()
				require.NotEmpty(t, addr.HostName(), "left neighbor address should have hostname at level %d", level)
				require.NotEmpty(t, addr.Port(), "left neighbor address should have port at level %d", level)

				rightEntry, err := table.GetEntry(types.DirectionRight, level)
				require.NoError(t, err)
				require.NotNil(t, rightEntry, "right entry should exist at level %d", level)

				// Verify identifier is not all zeros
				id = rightEntry.GetIdentifier()
				idAllZeros = true
				for _, b := range id {
					if b != 0 {
						idAllZeros = false
						break
					}
				}
				require.False(
					t,
					idAllZeros,
					"right neighbor identifier at level %d should not be all zeros",
					level,
				)

				// Verify membership vector is not all zeros
				memVec = rightEntry.GetMembershipVector()
				mvAllZeros = true
				for _, b := range memVec {
					if b != 0 {
						mvAllZeros = false
						break
					}
				}
				require.False(
					t,
					mvAllZeros,
					"right neighbor membership vector at level %d should not be all zeros",
					level,
				)

				// Verify address has valid hostname and port
				addr = rightEntry.GetAddress()
				require.NotEmpty(t, addr.HostName(), "right neighbor address should have hostname at level %d", level)
				require.NotEmpty(t, addr.Port(), "right neighbor address should have port at level %d", level)
			}
		},
	)
}

// TestIdentifierFixtureConstraints tests the IdentifierFixture function with various constraints
// to ensure the deterministic approach works correctly in all cases.
func TestIdentifierFixtureConstraints(t *testing.T) {
	t.Run(
		"generates IDs greater than minID", func(t *testing.T) {
			minID := IdentifierFixture(t)

			// Generate 100 IDs and verify all are greater than minID
			for i := 0; i < 100; i++ {
				id := IdentifierFixture(t, WithIdsGreaterThan(minID))
				comparison := id.Compare(&minID)
				require.Equal(
					t, model.CompareGreater, comparison.GetComparisonResult(),
					"generated ID should be greater than minID: %s", comparison.DebugInfo(),
				)
			}
		},
	)

	t.Run(
		"generates IDs less than maxID", func(t *testing.T) {
			maxID := IdentifierFixture(t)

			// Generate 100 IDs and verify all are less than maxID
			for i := 0; i < 100; i++ {
				id := IdentifierFixture(t, WithIdsLessThan(maxID))
				comparison := id.Compare(&maxID)
				require.Equal(
					t, model.CompareLess, comparison.GetComparisonResult(),
					"generated ID should be less than maxID: %s", comparison.DebugInfo(),
				)
			}
		},
	)

	t.Run(
		"generates IDs in range (minID, maxID)", func(t *testing.T) {
			// Create two random IDs and ensure minID < maxID
			id1 := IdentifierFixture(t)
			id2 := IdentifierFixture(t)

			comparison := id1.Compare(&id2)
			var minID, maxID model.Identifier
			if comparison.GetComparisonResult() == model.CompareLess {
				minID = id1
				maxID = id2
			} else if comparison.GetComparisonResult() == model.CompareGreater {
				minID = id2
				maxID = id1
			} else {
				// IDs are equal, generate a completely new ID instead of incrementing
				minID = id1
				maxID = IdentifierFixture(t)
				// Ensure it's different
				newComparison := maxID.Compare(&minID)
				for newComparison.GetComparisonResult() == model.CompareEqual {
					maxID = IdentifierFixture(t)
					newComparison = maxID.Compare(&minID)
				}
				// Ensure minID < maxID by swapping if needed
				if newComparison.GetComparisonResult() == model.CompareLess {
					minID, maxID = maxID, minID
				}
			}

			// Generate 100 IDs and verify all are in range
			for i := 0; i < 100; i++ {
				id := IdentifierFixture(t, WithIdsGreaterThan(minID), WithIdsLessThan(maxID))

				// Verify ID > minID
				minComparison := id.Compare(&minID)
				require.Equal(
					t, model.CompareGreater, minComparison.GetComparisonResult(),
					"generated ID should be greater than minID: %s", minComparison.DebugInfo(),
				)

				// Verify ID < maxID
				maxComparison := id.Compare(&maxID)
				require.Equal(
					t, model.CompareLess, maxComparison.GetComparisonResult(),
					"generated ID should be less than maxID: %s", maxComparison.DebugInfo(),
				)
			}
		},
	)
}

// BenchmarkIdentifierFixture benchmarks the IdentifierFixture function
// to demonstrate the performance improvement of the deterministic approach.
func BenchmarkIdentifierFixture(b *testing.B) {
	b.Run(
		"unconstrained", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = IdentifierFixture(&testing.T{})
			}
		},
	)

	b.Run(
		"with_greater_than", func(b *testing.B) {
			minID := IdentifierFixture(&testing.T{})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = IdentifierFixture(&testing.T{}, WithIdsGreaterThan(minID))
			}
		},
	)

	b.Run(
		"with_less_than", func(b *testing.B) {
			maxID := IdentifierFixture(&testing.T{})
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = IdentifierFixture(&testing.T{}, WithIdsLessThan(maxID))
			}
		},
	)

	b.Run(
		"with_range", func(b *testing.B) {
			minID := IdentifierFixture(&testing.T{})
			maxID := IdentifierFixture(&testing.T{})

			// Ensure minID < maxID
			comparison := minID.Compare(&maxID)
			if comparison.GetComparisonResult() != model.CompareLess {
				minID, maxID = maxID, minID
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = IdentifierFixture(
					&testing.T{},
					WithIdsGreaterThan(minID),
					WithIdsLessThan(maxID),
				)
			}
		},
	)
}

// TestTestMessageFixture tests the TestMessageFixture function.
func TestTestMessageFixture(t *testing.T) {
	t.Run(
		"generates message with 100-byte payload", func(t *testing.T) {
			msg := TestMessageFixture(t)
			require.NotNil(t, msg, "message should not be nil")
			require.NotNil(t, msg.Payload, "message payload should not be nil")

			// Verify payload is a byte slice
			payloadBytes, ok := msg.Payload.([]byte)
			require.True(t, ok, "payload should be a byte slice")
			require.Len(t, payloadBytes, 100, "payload should be 100 bytes")
		},
	)

	t.Run(
		"generates random payload", func(t *testing.T) {
			msg := TestMessageFixture(t)
			payloadBytes := msg.Payload.([]byte)

			// Check that the payload is not all zeros
			allZeros := true
			for _, b := range payloadBytes {
				if b != 0 {
					allZeros = false
					break
				}
			}
			require.False(t, allZeros, "payload should not be all zeros")
		},
	)

	t.Run(
		"generates different messages on multiple calls", func(t *testing.T) {
			msg1 := TestMessageFixture(t)
			msg2 := TestMessageFixture(t)

			payload1 := msg1.Payload.([]byte)
			payload2 := msg2.Payload.([]byte)

			// Check that the two payloads are different
			require.NotEqual(
				t,
				payload1,
				payload2,
				"multiple calls should generate different payloads",
			)
		},
	)
}

// TestIdentifierFixture tests the IdentifierFixture function without constraints.
func TestIdentifierFixture(t *testing.T) {
	t.Run(
		"generates valid 32-byte identifiers", func(t *testing.T) {
			// Generate multiple identifiers and verify they're all valid
			for i := 0; i < 10; i++ {
				id := IdentifierFixture(t)
				require.Len(t, id[:], model.IdentifierSizeBytes, "identifier should be 32 bytes")
			}
		},
	)

	t.Run(
		"generates different identifiers on multiple calls", func(t *testing.T) {
			// Generate two identifiers and verify they're different
			id1 := IdentifierFixture(t)
			id2 := IdentifierFixture(t)

			require.NotEqual(t, id1, id2, "multiple calls should generate different identifiers")
		},
	)

	t.Run(
		"WithIdsGreaterThan generates IDs greater than constraint", func(t *testing.T) {
			minID := IdentifierFixture(t)

			// Generate multiple IDs and verify all are greater than minID
			for i := 0; i < 50; i++ {
				id := IdentifierFixture(t, WithIdsGreaterThan(minID))
				comparison := id.Compare(&minID)
				require.Equal(
					t, model.CompareGreater, comparison.GetComparisonResult(),
					"generated ID should be greater than minID: %s", comparison.DebugInfo(),
				)
			}
		},
	)

	t.Run(
		"WithIdsLessThan generates IDs less than constraint", func(t *testing.T) {
			maxID := IdentifierFixture(t)

			// Generate multiple IDs and verify all are less than maxID
			for i := 0; i < 50; i++ {
				id := IdentifierFixture(t, WithIdsLessThan(maxID))
				comparison := id.Compare(&maxID)
				require.Equal(
					t, model.CompareLess, comparison.GetComparisonResult(),
					"generated ID should be less than maxID: %s", comparison.DebugInfo(),
				)
			}
		},
	)

	t.Run(
		"combined constraints generate IDs in range", func(t *testing.T) {
			// Create two random IDs and ensure minID < maxID
			id1 := IdentifierFixture(t)
			id2 := IdentifierFixture(t)

			comparison := id1.Compare(&id2)
			var minID, maxID model.Identifier
			if comparison.GetComparisonResult() == model.CompareLess {
				minID = id1
				maxID = id2
			} else if comparison.GetComparisonResult() == model.CompareGreater {
				minID = id2
				maxID = id1
			} else {
				// IDs are equal, generate a completely new ID instead of incrementing
				minID = id1
				maxID = IdentifierFixture(t)
				// Ensure it's different
				newComparison := maxID.Compare(&minID)
				for newComparison.GetComparisonResult() == model.CompareEqual {
					maxID = IdentifierFixture(t)
					newComparison = maxID.Compare(&minID)
				}
				// Ensure minID < maxID by swapping if needed
				if newComparison.GetComparisonResult() == model.CompareLess {
					minID, maxID = maxID, minID
				}
			}

			// Generate multiple IDs and verify all are in range
			for i := 0; i < 50; i++ {
				id := IdentifierFixture(t, WithIdsGreaterThan(minID), WithIdsLessThan(maxID))

				// Verify ID > minID
				minComparison := id.Compare(&minID)
				require.Equal(
					t, model.CompareGreater, minComparison.GetComparisonResult(),
					"generated ID should be greater than minID: %s", minComparison.DebugInfo(),
				)

				// Verify ID < maxID
				maxComparison := id.Compare(&maxID)
				require.Equal(
					t, model.CompareLess, maxComparison.GetComparisonResult(),
					"generated ID should be less than maxID: %s", maxComparison.DebugInfo(),
				)
			}
		},
	)
}

// TestRandomBytesFixture tests the RandomBytesFixture function.
func TestRandomBytesFixture(t *testing.T) {
	t.Run(
		"generates correct size", func(t *testing.T) {
			sizes := []int{1, 10, 32, 100, 256, 1024}
			for _, size := range sizes {
				bytes := RandomBytesFixture(t, size)
				require.Len(t, bytes, size, "should generate byte array of size %d", size)
			}
		},
	)

	t.Run(
		"generates random non-zero bytes", func(t *testing.T) {
			// Generate a 100-byte array and check it's not all zeros
			bytes := RandomBytesFixture(t, 100)

			// Check that at least some bytes are non-zero
			allZeros := true
			for _, b := range bytes {
				if b != 0 {
					allZeros = false
					break
				}
			}
			require.False(t, allZeros, "should generate non-zero bytes (highly unlikely all zeros)")
		},
	)

	t.Run(
		"generates different byte arrays on multiple calls", func(t *testing.T) {
			// Generate two byte arrays of the same size
			bytes1 := RandomBytesFixture(t, 100)
			bytes2 := RandomBytesFixture(t, 100)

			// They should be different (extremely high probability)
			require.NotEqual(
				t,
				bytes1,
				bytes2,
				"multiple calls should generate different byte arrays",
			)
		},
	)

	t.Run(
		"handles various sizes", func(t *testing.T) {
			// Test edge cases
			bytes0 := RandomBytesFixture(t, 0)
			require.Len(t, bytes0, 0, "should handle size 0")

			bytes1 := RandomBytesFixture(t, 1)
			require.Len(t, bytes1, 1, "should handle size 1")

			bytesLarge := RandomBytesFixture(t, 10000)
			require.Len(t, bytesLarge, 10000, "should handle large sizes")
		},
	)
}

// TestMembershipVectorFixture tests the MembershipVectorFixture function.
func TestMembershipVectorFixture(t *testing.T) {
	t.Run(
		"generates correct size", func(t *testing.T) {
			mv := MembershipVectorFixture(t)
			require.Len(
				t,
				mv[:],
				model.MembershipVectorSize,
				"membership vector should have correct size",
			)
		},
	)

	t.Run(
		"generates random values", func(t *testing.T) {
			mv := MembershipVectorFixture(t)

			// Check that the membership vector is not all zeros
			allZeros := true
			for _, b := range mv {
				if b != 0 {
					allZeros = false
					break
				}
			}
			require.False(
				t,
				allZeros,
				"membership vector should not be all zeros (highly unlikely)",
			)
		},
	)

	t.Run(
		"generates different vectors on multiple calls", func(t *testing.T) {
			mv1 := MembershipVectorFixture(t)
			mv2 := MembershipVectorFixture(t)

			require.NotEqual(
				t,
				mv1,
				mv2,
				"multiple calls should generate different membership vectors",
			)
		},
	)
}

// TestAddressFixture tests the AddressFixture function.
func TestAddressFixture(t *testing.T) {
	t.Run(
		"generates localhost addresses", func(t *testing.T) {
			// Generate multiple addresses and verify they're all on localhost
			for i := 0; i < 10; i++ {
				addr := AddressFixture(t)
				require.Equal(t, "localhost", addr.HostName(), "address should be on localhost")
			}
		},
	)

	t.Run(
		"generates valid ports", func(t *testing.T) {
			// Generate multiple addresses and verify ports are in valid range [0, 65535)
			for i := 0; i < 100; i++ {
				addr := AddressFixture(t)
				port := addr.Port()
				require.NotEmpty(t, port, "port should not be empty")

				// Parse port as integer to verify it's in valid range
				var portNum int
				_, err := fmt.Sscanf(port, "%d", &portNum)
				require.NoError(t, err, "port should be a valid integer")
				require.GreaterOrEqual(t, portNum, 0, "port should be >= 0")
				require.LessOrEqual(t, portNum, 65534, "port should be <= 65534")
			}
		},
	)

	t.Run(
		"generates different ports on multiple calls", func(t *testing.T) {
			// Generate many addresses and verify we get at least some different ports
			ports := make(map[string]bool)
			for i := 0; i < 100; i++ {
				addr := AddressFixture(t)
				require.False(t, ports[addr.Port()], "port should be unique")
				ports[addr.Port()] = true
			}
		},
	)

	t.Run(
		"generates complete addresses", func(t *testing.T) {
			addr := AddressFixture(t)
			require.NotEmpty(t, addr.HostName(), "hostname should not be empty")
			require.NotEmpty(t, addr.Port(), "port should not be empty")
		},
	)
}

// TestIdentityFixture tests the IdentityFixture function.
func TestIdentityFixture(t *testing.T) {
	t.Run(
		"generates complete identities with all fields", func(t *testing.T) {
			identity := IdentityFixture(t)

			// Verify identifier is set
			id := identity.GetIdentifier()
			require.Len(t, id[:], model.IdentifierSizeBytes, "identifier should be 32 bytes")

			// Verify membership vector is set
			memVec := identity.GetMembershipVector()
			require.Len(
				t,
				memVec[:],
				model.MembershipVectorSize,
				"membership vector should have correct size",
			)

			// Verify address is set
			addr := identity.GetAddress()
			require.Equal(t, "localhost", addr.HostName(), "address should be on localhost")
			require.NotEmpty(t, addr.Port(), "port should not be empty")
		},
	)

	t.Run(
		"generates all non-zero fields", func(t *testing.T) {
			identity := IdentityFixture(t)

			// Check identifier is not all zeros
			id := identity.GetIdentifier()
			idAllZeros := true
			for _, b := range id {
				if b != 0 {
					idAllZeros = false
					break
				}
			}
			require.False(t, idAllZeros, "identifier should not be all zeros (highly unlikely)")

			// Check membership vector is not all zeros
			memVec := identity.GetMembershipVector()
			mvAllZeros := true
			for _, b := range memVec {
				if b != 0 {
					mvAllZeros = false
					break
				}
			}
			require.False(
				t,
				mvAllZeros,
				"membership vector should not be all zeros (highly unlikely)",
			)

			// Check address has valid port
			addr := identity.GetAddress()
			require.NotEmpty(t, addr.Port(), "port should not be empty")
		},
	)

	t.Run(
		"generates different identities on multiple calls", func(t *testing.T) {
			identity1 := IdentityFixture(t)
			identity2 := IdentityFixture(t)

			// At least the identifier should be different
			id1 := identity1.GetIdentifier()
			id2 := identity2.GetIdentifier()
			require.NotEqual(t, id1, id2, "identifiers should be different")

			// Membership vectors should also be different (very high probability)
			mv1 := identity1.GetMembershipVector()
			mv2 := identity2.GetMembershipVector()
			require.NotEqual(t, mv1, mv2, "membership vectors should be different")
		},
	)

	t.Run(
		"all components are properly initialized", func(t *testing.T) {
			// Generate multiple identities to ensure consistency
			for i := 0; i < 10; i++ {
				identity := IdentityFixture(t)

				// Verify all components can be accessed without panic
				id := identity.GetIdentifier()
				require.NotNil(t, id)

				memVec := identity.GetMembershipVector()
				require.NotNil(t, memVec)

				addr := identity.GetAddress()
				require.NotEmpty(t, addr.HostName())
				require.NotEmpty(t, addr.Port())
			}
		},
	)
}
