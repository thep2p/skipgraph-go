package unittest

import (
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/core"
	"github.com/thep2p/skipgraph-go/core/model"
	"github.com/thep2p/skipgraph-go/core/types"
	"testing"
)

// SmallestIdLessThanOrEqualTo finds the smallest identifier >= target across levels 0 to level in the given direction.
// Returns (found, level, identifier) where found indicates if a candidate was found.
func SmallestIdLessThanOrEqualTo(
	t *testing.T,
	target model.Identifier,
	level types.Level,
	dir types.Direction,
	table core.ImmutableLookupTable,
) (bool, types.Level, model.Identifier) {
	var expectedLevel types.Level
	var expectedID model.Identifier
	foundCandidate := false

	for l := types.Level(0); l <= level; l++ {
		neighbor, err := table.GetEntry(dir, l)
		require.NoError(t, err)
		neighborID := neighbor.GetIdentifier()
		cmp := neighborID.Compare(&target)
		if cmp.GetComparisonResult() == model.CompareGreater || cmp.GetComparisonResult() == model.CompareEqual {
			if !foundCandidate {
				expectedID = neighborID
				expectedLevel = l
				foundCandidate = true
			} else {
				bestCmp := neighborID.Compare(&expectedID)
				if bestCmp.GetComparisonResult() == model.CompareLess {
					expectedID = neighborID
					expectedLevel = l
				}
			}
		}
	}

	return foundCandidate, expectedLevel, expectedID
}
