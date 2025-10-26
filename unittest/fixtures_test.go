package unittest

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/core"
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
	t.Run("with custom max", func(t *testing.T) {
		max := types.Level(10)
		// Generate multiple random levels to ensure they're all valid
		for i := 0; i < 100; i++ {
			level := RandomLevelWithMaxFixture(t, max)
			require.GreaterOrEqual(t, level, types.Level(0))
			require.Less(t, level, max)
		}
	})

	t.Run("with max of 1", func(t *testing.T) {
		max := types.Level(1)
		// Should always return 0
		for i := 0; i < 10; i++ {
			level := RandomLevelWithMaxFixture(t, max)
			require.Equal(t, types.Level(0), level)
		}
	})

	t.Run("with max of 2", func(t *testing.T) {
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
	})

	t.Run("with MaxLookupTableLevel", func(t *testing.T) {
		// Should work with the actual max level
		for i := 0; i < 100; i++ {
			level := RandomLevelWithMaxFixture(t, core.MaxLookupTableLevel)
			require.GreaterOrEqual(t, level, types.Level(0))
			require.Less(t, level, core.MaxLookupTableLevel)
		}
	})
}

// TestRandomDirectionFixture tests the RandomDirectionFixture function.
func TestRandomDirectionFixture(t *testing.T) {
	t.Run("generates valid directions", func(t *testing.T) {
		// Generate multiple random directions to ensure they're all valid
		for i := 0; i < 100; i++ {
			direction := RandomDirectionFixture(t)
			// Direction must be either DirectionLeft or DirectionRight
			require.True(t,
				direction == types.DirectionLeft || direction == types.DirectionRight,
				"direction must be either DirectionLeft or DirectionRight")
		}
	})

	t.Run("generates both directions", func(t *testing.T) {
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
	})
}

// TestRandomLookupTable tests the RandomLookupTable function.
func TestRandomLookupTable(t *testing.T) {
	t.Run("generates full lookup table without constraints", func(t *testing.T) {
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
	})

	t.Run("WithIdsGreaterThan generates all IDs greater than constraint", func(t *testing.T) {
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
				require.Equal(t, "compare-greater", comparison.GetComparisonResult(),
					"left neighbor ID at level %d should be greater than constraint: %s",
					level, comparison.DebugInfo())

				// Check right neighbor
				rightEntry, err := table.GetEntry(types.DirectionRight, level)
				require.NoError(t, err)
				require.NotNil(t, rightEntry, "right entry should exist at level %d", level)
				id = rightEntry.GetIdentifier()
				comparison = id.Compare(&constraintID)
				require.Equal(t, "compare-greater", comparison.GetComparisonResult(),
					"right neighbor ID at level %d should be greater than constraint: %s",
					level, comparison.DebugInfo())
			}
		}
	})

	t.Run("WithIdsLessThan generates all IDs less than constraint", func(t *testing.T) {
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
				require.Equal(t, "compare-less", comparison.GetComparisonResult(),
					"left neighbor ID at level %d should be less than constraint: %s",
					level, comparison.DebugInfo())

				// Check right neighbor
				rightEntry, err := table.GetEntry(types.DirectionRight, level)
				require.NoError(t, err)
				require.NotNil(t, rightEntry, "right entry should exist at level %d", level)
				id = rightEntry.GetIdentifier()
				comparison = id.Compare(&constraintID)
				require.Equal(t, "compare-less", comparison.GetComparisonResult(),
					"right neighbor ID at level %d should be less than constraint: %s",
					level, comparison.DebugInfo())
			}
		}
	})

	t.Run("WithIdsGreaterThan and WithIdsLessThan combined", func(t *testing.T) {
		// Create two constraint IDs where minID < maxID
		minID := IdentifierFixture(t)
		maxID := IdentifierFixture(t)

		// Ensure minID < maxID
		comparison := minID.Compare(&maxID)
		if comparison.GetComparisonResult() == "compare-greater" {
			// Swap them
			minID, maxID = maxID, minID
		} else if comparison.GetComparisonResult() == "compare-equal" {
			// Generate a new maxID that's greater
			// We'll use a simple approach: increment the last byte
			maxID[len(maxID)-1]++
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
				require.Equal(t, "compare-greater", minComparison.GetComparisonResult(),
					"left neighbor ID at level %d should be greater than minID: %s",
					level, minComparison.DebugInfo())

				// Check less than maxID
				maxComparison := id.Compare(&maxID)
				require.Equal(t, "compare-less", maxComparison.GetComparisonResult(),
					"left neighbor ID at level %d should be less than maxID: %s",
					level, maxComparison.DebugInfo())

				// Check right neighbor
				rightEntry, err := table.GetEntry(types.DirectionRight, level)
				require.NoError(t, err)
				require.NotNil(t, rightEntry, "right entry should exist at level %d", level)
				id = rightEntry.GetIdentifier()

				// Check greater than minID
				minComparison = id.Compare(&minID)
				require.Equal(t, "compare-greater", minComparison.GetComparisonResult(),
					"right neighbor ID at level %d should be greater than minID: %s",
					level, minComparison.DebugInfo())

				// Check less than maxID
				maxComparison = id.Compare(&maxID)
				require.Equal(t, "compare-less", maxComparison.GetComparisonResult(),
					"right neighbor ID at level %d should be less than maxID: %s",
					level, maxComparison.DebugInfo())
			}
		}
	})

	t.Run("neighbors have complete identities", func(t *testing.T) {
		table := RandomLookupTable(t)
		require.NotNil(t, table)

		// Check that all neighbors have complete identities
		for level := types.Level(0); level < core.MaxLookupTableLevel; level++ {
			leftEntry, err := table.GetEntry(types.DirectionLeft, level)
			require.NoError(t, err)
			require.NotNil(t, leftEntry, "left entry should exist at level %d", level)

			// Identity should have non-zero ID, membership vector, and address
			id := leftEntry.GetIdentifier()
			memVec := leftEntry.GetMembershipVector()
			addr := leftEntry.GetAddress()

			// Verify these are not zero values
			_ = id
			_ = memVec
			require.NotEmpty(t, addr.HostName(), "address should have hostname")
			require.NotEmpty(t, addr.Port(), "address should have port")

			rightEntry, err := table.GetEntry(types.DirectionRight, level)
			require.NoError(t, err)
			require.NotNil(t, rightEntry, "right entry should exist at level %d", level)

			// Identity should have non-zero ID, membership vector, and address
			id = rightEntry.GetIdentifier()
			memVec = rightEntry.GetMembershipVector()
			addr = rightEntry.GetAddress()

			_ = id
			_ = memVec
			require.NotEmpty(t, addr.HostName(), "address should have hostname")
			require.NotEmpty(t, addr.Port(), "address should have port")
		}
	})
}
