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
