package node

import (
	"fmt"
	"github.com/thep2p/skipgraph-go/core"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/core/lookup"
	"github.com/thep2p/skipgraph-go/core/model"
	"github.com/thep2p/skipgraph-go/core/types"
	"github.com/thep2p/skipgraph-go/unittest"
)

// TestSearchByIDSingletonFallback tests fallback behavior when no neighbors exist (empty lookup table).
// All searches should return terminationLevel = 0, result = node's own ID.
func TestSearchByIDSingletonFallback(t *testing.T) {
	// Create node with ID and empty lookup table
	nodeID := unittest.IdentifierFixture(t)

	memVec := unittest.MembershipVectorFixture(t)
	identity := model.NewIdentity(nodeID, memVec, model.NewAddress("localhost", "8000"))
	node := NewSkipGraphNode(unittest.Logger(zerolog.TraceLevel), identity, &lookup.Table{})

	for i := 0; i < 100; i++ {
		target := unittest.IdentifierFixture(t)

		level := unittest.RandomLevelFixture(t)

		req, err := model.NewIdSearchReq(target, level, unittest.RandomDirectionFixture(t))
		require.NoError(t, err)
		res, err := node.SearchByID(req)

		require.NoError(t, err)
		require.Equal(
			t,
			types.Level(0),
			res.TerminationLevel(),
			"expected fallback to level 0",
		)
		require.Equal(t, nodeID, res.Result(), "expected fallback to own ID")
	}

}

// TestSearchByIDFoundLeftDirection verifies correct candidate selection in left direction
// (smallest ID >= target).
func TestSearchByIDFoundLeftDirection(t *testing.T) {
	// Test for all levels
	for testLevel := types.Level(0); testLevel < core.MaxLookupTableLevel; testLevel++ {
		// Create node
		nodeID := unittest.IdentifierFixture(t)
		memVec := unittest.MembershipVectorFixture(t)
		identity := model.NewIdentity(nodeID, memVec, unittest.AddressFixture(t))

		// Generate a random target
		target := unittest.IdentifierFixture(t)

		// Populate lookup table with random neighbors with IDs greater than target
		// to guarantee at least one valid candidate exists >= target on left direction
		lt := unittest.RandomLookupTable(t, unittest.WithIdsGreaterThan(target))
		node := NewSkipGraphNode(unittest.Logger(zerolog.TraceLevel), identity, lt)

		// Perform search
		req, err := model.NewIdSearchReq(target, testLevel, types.DirectionLeft)
		require.NoError(t, err)
		res, err := node.SearchByID(req)
		require.NoError(t, err)

		// Manually compute expected result: smallest ID >= target
		var expectedLevel types.Level
		var expectedID model.Identifier
		foundCandidate := false

		for level := types.Level(0); level <= testLevel; level++ {
			neighbor, err := lt.GetEntry(types.DirectionLeft, level)
			require.NoError(t, err)
			neighborId := neighbor.GetIdentifier()
			cmp := neighborId.Compare(&target)
			if cmp.GetComparisonResult() == model.CompareGreater || cmp.GetComparisonResult() == model.CompareEqual {
				if !foundCandidate {
					expectedID = neighborId
					expectedLevel = level
					foundCandidate = true
				} else {
					// Check if this neighbor is smaller than current best
					bestCmp := neighborId.Compare(&expectedID)
					if bestCmp.GetComparisonResult() == model.CompareLess {
						expectedID = neighborId
						expectedLevel = level
					}
				}
			}
		}

		require.True(t, foundCandidate, "expected to find a candidate >= target")
		require.Equal(
			t,
			expectedLevel,
			res.TerminationLevel(),
			"termination level mismatch, expected %d, got %d", expectedLevel,
			res.TerminationLevel(),
		)
		require.Equal(t, expectedID, res.Result(), "result identifier mismatch")
	}
}

// TestSearchByIDFoundRightDirection verifies correct candidate selection in right direction
// (greatest ID <= target).
func TestSearchByIDFoundRightDirection(t *testing.T) {
	// Test for all levels
	for testLevel := types.Level(0); testLevel < core.MaxLookupTableLevel; testLevel++ {
		// Create node
		nodeID := unittest.IdentifierFixture(t)
		memVec := unittest.MembershipVectorFixture(t)
		identity := model.NewIdentity(nodeID, memVec, unittest.AddressFixture(t))

		// Populate lookup table with random neighbors
		lt := unittest.RandomLookupTable(t)

		// Get one right neighbor and use it to set a baseline for target
		// This ensures at least one valid candidate exists
		baseNeighbor, err := lt.GetEntry(types.DirectionRight, types.Level(0))
		require.NoError(t, err)
		baseNeighborID := baseNeighbor.GetIdentifier()

		// Generate target that is >= baseNeighborID OR use baseNeighborID itself
		// This guarantees at least one right neighbor will be <= target
		target := baseNeighborID
		node := NewSkipGraphNode(unittest.Logger(zerolog.TraceLevel), identity, lt)

		// Perform search
		req, err := model.NewIdSearchReq(target, testLevel, types.DirectionRight)
		require.NoError(t, err)
		res, err := node.SearchByID(req)
		require.NoError(t, err)

		// Manually compute expected result: greatest ID <= target
		var expectedLevel types.Level
		var expectedID model.Identifier
		foundCandidate := false

		for level := types.Level(0); level <= testLevel; level++ {
			neighbor, err := lt.GetEntry(types.DirectionRight, level)
			require.NoError(t, err)
			neighborId := neighbor.GetIdentifier()
			cmp := neighborId.Compare(&target)
			if cmp.GetComparisonResult() == model.CompareLess || cmp.GetComparisonResult() == model.CompareEqual {
				if !foundCandidate {
					expectedID = neighborId
					expectedLevel = level
					foundCandidate = true
				} else {
					// Check if this neighbor is greater than current best
					bestCmp := neighborId.Compare(&expectedID)
					if bestCmp.GetComparisonResult() == model.CompareGreater {
						expectedID = neighborId
						expectedLevel = level
					}
				}
			}
		}

		require.True(t, foundCandidate, "expected to find a candidate <= target")
		require.Equal(
			t,
			expectedLevel,
			res.TerminationLevel(),
			"termination level mismatch, expected %d, got %d", expectedLevel,
			res.TerminationLevel(),
		)
		require.Equal(t, expectedID, res.Result(), "result identifier mismatch")
	}
}

// TestSearchByIDNotFoundLeftDirection verifies fallback when no valid candidates exist in left direction.
// All left neighbors have IDs less than target, so should fallback to own ID.
func TestSearchByIDNotFoundLeftDirection(t *testing.T) {
	// Test for various levels
	for testLevel := types.Level(0); testLevel < core.MaxLookupTableLevel; testLevel++ {
		// Create node
		nodeID := unittest.IdentifierFixture(t)
		memVec := unittest.MembershipVectorFixture(t)
		identity := model.NewIdentity(nodeID, memVec, unittest.AddressFixture(t))

		// Generate a random target
		target := unittest.IdentifierFixture(t)

		// Populate ALL left neighbors with IDs less than target
		lt := unittest.RandomLookupTable(t, unittest.WithIdsLessThan(target))
		node := NewSkipGraphNode(unittest.Logger(zerolog.TraceLevel), identity, lt)

		// Perform search - should fallback to own ID
		req, err := model.NewIdSearchReq(target, testLevel, types.DirectionLeft)
		require.NoError(t, err)
		res, err := node.SearchByID(req)

		require.NoError(t, err)
		require.Equal(
			t,
			types.Level(0),
			res.TerminationLevel(),
			"expected fallback to level 0",
		)
		require.Equal(t, nodeID, res.Result(), "expected fallback to own ID")
	}
}

// TestSearchByIDNotFoundRightDirection verifies fallback when no valid candidates exist in right direction.
// All right neighbors have IDs greater than target, so should fallback to own ID.
func TestSearchByIDNotFoundRightDirection(t *testing.T) {
	// Test for various levels
	for testLevel := types.Level(0); testLevel < core.MaxLookupTableLevel; testLevel++ {
		// Create node
		nodeID := unittest.IdentifierFixture(t)
		memVec := unittest.MembershipVectorFixture(t)
		identity := model.NewIdentity(nodeID, memVec, unittest.AddressFixture(t))

		// Generate a random target
		target := unittest.IdentifierFixture(t)

		// Populate ALL right neighbors with IDs greater than target
		lt := unittest.RandomLookupTable(t, unittest.WithIdsGreaterThan(target))
		node := NewSkipGraphNode(unittest.Logger(zerolog.TraceLevel), identity, lt)

		// Perform search - should fallback to own ID
		req, err := model.NewIdSearchReq(target, testLevel, types.DirectionRight)
		require.NoError(t, err)
		res, err := node.SearchByID(req)

		require.NoError(t, err)
		require.Equal(
			t,
			types.Level(0),
			res.TerminationLevel(),
			"expected fallback to level 0",
		)
		require.Equal(t, nodeID, res.Result(), "expected fallback to own ID")
	}
}

// TestSearchByIDExactResult verifies exact match when target exists in lookup table.
// When we search for a neighbor's ID, we should get that exact neighbor back.
func TestSearchByIDExactResult(t *testing.T) {
	// Create node
	nodeID := unittest.IdentifierFixture(t)
	memVec := unittest.MembershipVectorFixture(t)
	identity := model.NewIdentity(nodeID, memVec, unittest.AddressFixture(t))
	lt := unittest.RandomLookupTable(t)
	node := NewSkipGraphNode(unittest.Logger(zerolog.TraceLevel), identity, lt)

	// Test left direction - search for each left neighbor's ID
	for level := types.Level(0); level < core.MaxLookupTableLevel; level++ {
		target, err := lt.GetEntry(types.DirectionLeft, level)
		require.NoError(t, err)
		req, err := model.NewIdSearchReq(target.GetIdentifier(), level, types.DirectionLeft)
		require.NoError(t, err)
		res, err := node.SearchByID(req)

		require.NoError(t, err)
		require.Equal(t, level, res.TerminationLevel(), "expected to find at same level")
		require.Equal(t, target.GetIdentifier(), res.Result(), "expected exact match")
	}

	// Test right direction - search for each right neighbor's ID
	for level := types.Level(0); level < core.MaxLookupTableLevel; level++ {
		target, err := lt.GetEntry(types.DirectionRight, level)
		req, err := model.NewIdSearchReq(
			target.GetIdentifier(),
			level,
			types.DirectionRight,
		)
		require.NoError(t, err)
		res, err := node.SearchByID(req)

		require.NoError(t, err)
		require.Equal(t, level, res.TerminationLevel(), "expected to find at same level")
		require.Equal(t, target.GetIdentifier(), res.Result(), "expected exact match")
	}
}

// TestSearchByIDConcurrentFoundLeftDirection tests thread safety with concurrent searches in left direction.
func TestSearchByIDConcurrentFoundLeftDirection(t *testing.T) {
	// Create node with populated lookup table
	nodeID := unittest.IdentifierFixture(t)
	memVec := unittest.MembershipVectorFixture(t)
	identity := model.NewIdentity(nodeID, memVec, unittest.AddressFixture(t))
	lt := &lookup.Table{}

	// Populate lookup table
	const maxTestLevel = types.Level(10)
	neighbors := make(map[types.Level]model.Identifier)

	// Use a fixed target
	targetBytes := []byte{0x50} // 80 in decimal
	target, err := model.ByteToId(targetBytes)
	require.NoError(t, err)

	// Populate with neighbors, ensuring at least one is >= target
	for level := types.Level(0); level < maxTestLevel; level++ {
		var neighborID model.Identifier
		if level == 0 {
			// Guarantee level 0 has a neighbor >= target
			neighborID = unittest.IdentifierFixture(t, unittest.WithIdsGreaterThan(target))
		} else {
			neighborID = unittest.IdentifierFixture(t)
		}
		neighbors[level] = neighborID
		neighborIdentity := model.NewIdentity(
			neighborID,
			unittest.MembershipVectorFixture(t),
			unittest.AddressFixture(t),
		)
		err := lt.AddEntry(types.DirectionLeft, types.Level(level), neighborIdentity)
		require.NoError(t, err)
	}

	node := NewSkipGraphNode(unittest.Logger(zerolog.TraceLevel), identity, lt)

	// Spawn 20 goroutines
	const numGoroutines = 20
	var wg sync.WaitGroup
	barrier := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			<-barrier // Wait for all goroutines to be ready

			// Pick random level
			level := types.Level(rand.Intn(int(maxTestLevel)))
			req, err := model.NewIdSearchReq(target, level, types.DirectionLeft)
			require.NoError(t, err)
			res, err := node.SearchByID(req)
			require.NoError(t, err)

			// Compute expected result: smallest ID >= target from levels 0 to level
			var expectedLevel types.Level
			var expectedID model.Identifier
			foundCandidate := false

			for l := types.Level(0); l <= level; l++ {
				neighborID := neighbors[l]
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

			if foundCandidate {
				require.Equal(
					t,
					expectedLevel,
					res.TerminationLevel(),
					"goroutine %d: termination level mismatch",
					goroutineID,
				)
				require.Equal(
					t,
					expectedID,
					res.Result(),
					"goroutine %d: result identifier mismatch",
					goroutineID,
				)
			} else {
				// Fallback to own ID
				require.Equal(
					t,
					types.Level(0),
					res.TerminationLevel(),
					"goroutine %d: expected fallback to level 0",
					goroutineID,
				)
				require.Equal(
					t,
					nodeID,
					res.Result(),
					"goroutine %d: expected fallback to own ID",
					goroutineID,
				)
			}
		}(i)
	}

	close(barrier) // Release all goroutines at once

	// Wait for all goroutines to complete within timeout
	unittest.CallMustReturnWithinTimeout(
		t,
		wg.Wait,
		1*time.Second,
		"concurrent searches should complete within 1s",
	)
}

// TestSearchByIDConcurrentRightDirection tests thread safety with concurrent searches in right direction.
func TestSearchByIDConcurrentRightDirection(t *testing.T) {
	// Create node with populated lookup table
	nodeID := unittest.IdentifierFixture(t)
	memVec := unittest.MembershipVectorFixture(t)
	identity := model.NewIdentity(nodeID, memVec, unittest.AddressFixture(t))
	lt := &lookup.Table{}

	// Populate lookup table
	const maxTestLevel = types.Level(10)
	neighbors := make(map[types.Level]model.Identifier)

	// Use a fixed target
	targetBytes := []byte{0x80} // 128 in decimal
	target, err := model.ByteToId(targetBytes)
	require.NoError(t, err)

	// Populate with neighbors, ensuring at least one is <= target
	for level := types.Level(0); level < maxTestLevel; level++ {
		var neighborID model.Identifier
		if level == 0 {
			// Guarantee level 0 has a neighbor <= target
			neighborID = unittest.IdentifierFixture(t, unittest.WithIdsLessThan(target))
		} else {
			neighborID = unittest.IdentifierFixture(t)
		}
		neighbors[level] = neighborID
		neighborIdentity := model.NewIdentity(
			neighborID,
			unittest.MembershipVectorFixture(t),
			unittest.AddressFixture(t),
		)
		err := lt.AddEntry(types.DirectionRight, types.Level(level), neighborIdentity)
		require.NoError(t, err)
	}

	node := NewSkipGraphNode(unittest.Logger(zerolog.TraceLevel), identity, lt)

	// Spawn 20 goroutines
	const numGoroutines = 20
	var wg sync.WaitGroup
	barrier := make(chan struct{})

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			<-barrier // Wait for all goroutines to be ready

			// Pick random level
			level := types.Level(rand.Intn(int(maxTestLevel)))
			req, err := model.NewIdSearchReq(target, level, types.DirectionRight)
			require.NoError(t, err)
			res, err := node.SearchByID(req)
			require.NoError(t, err)

			// Compute expected result: greatest ID <= target from levels 0 to level
			var expectedLevel types.Level
			var expectedID model.Identifier
			foundCandidate := false

			for l := types.Level(0); l <= level; l++ {
				neighborID := neighbors[l]
				cmp := neighborID.Compare(&target)
				if cmp.GetComparisonResult() == model.CompareLess || cmp.GetComparisonResult() == model.CompareEqual {
					if !foundCandidate {
						expectedID = neighborID
						expectedLevel = l
						foundCandidate = true
					} else {
						bestCmp := neighborID.Compare(&expectedID)
						if bestCmp.GetComparisonResult() == model.CompareGreater {
							expectedID = neighborID
							expectedLevel = l
						}
					}
				}
			}

			if foundCandidate {
				require.Equal(
					t,
					expectedLevel,
					res.TerminationLevel(),
					"goroutine %d: termination level mismatch",
					goroutineID,
				)
				require.Equal(
					t,
					expectedID,
					res.Result(),
					"goroutine %d: result identifier mismatch",
					goroutineID,
				)
			} else {
				// Fallback to own ID
				require.Equal(
					t,
					types.Level(0),
					res.TerminationLevel(),
					"goroutine %d: expected fallback to level 0",
					goroutineID,
				)
				require.Equal(
					t,
					nodeID,
					res.Result(),
					"goroutine %d: expected fallback to own ID",
					goroutineID,
				)
			}
		}(i)
	}

	close(barrier) // Release all goroutines at once

	// Wait for all goroutines to complete within timeout
	unittest.CallMustReturnWithinTimeout(
		t,
		wg.Wait,
		1*time.Second,
		"concurrent searches should complete within 1s",
	)
}

// TestSearchByIDErrorPropagation verifies errors from lookup table are propagated correctly.
func TestSearchByIDErrorPropagation(t *testing.T) {
	// Create mock lookup table that returns error from GetEntry
	mockLT := &mockErrorLookupTable{
		errorAtLevel: 2,
	}

	// Create node
	nodeID := unittest.IdentifierFixture(t)
	memVec := unittest.MembershipVectorFixture(t)
	identity := model.NewIdentity(nodeID, memVec, unittest.AddressFixture(t))
	node := NewSkipGraphNode(unittest.Logger(zerolog.TraceLevel), identity, mockLT)

	// Try to search - should return error at level 2
	target := unittest.IdentifierFixture(t)
	req, err := model.NewIdSearchReq(target, 5, types.DirectionLeft)
	require.NoError(t, err)

	res, err := node.SearchByID(req)

	// Verify error is returned
	require.Error(t, err)
	require.Contains(t, err.Error(), "error while searching by id in level 2")
	require.Contains(t, err.Error(), "simulated lookup table error")
	require.Equal(t, model.IdSearchRes{}, res, "expected zero value result on error")
}

// TestSearchByIDNetworkingIntegration is a placeholder for future network integration testing.
// This test is skipped because the network layer and message processing infrastructure
// may not be fully implemented yet.
func TestSearchByIDNetworkingIntegration(t *testing.T) {
	t.Skip("Network integration test - depends on event processing infrastructure not yet implemented")

	// TODO: Implement when network layer is ready
	// Test strategy:
	// 1. Create node with mock network
	// 2. Register node as event processor
	// 3. Send IdSearchRequest event to node
	// 4. Verify node responds with IdSearchResponse event
	// 5. Assert response contains correct result
}

// TestSearchByIDInvalidDirection verifies that NewIdSearchReq rejects invalid direction values.
func TestSearchByIDInvalidDirection(t *testing.T) {
	target := unittest.IdentifierFixture(t)
	invalidDirection := types.Direction("invalid")

	req, err := model.NewIdSearchReq(target, 5, invalidDirection)

	require.Error(t, err)
	require.Contains(t, err.Error(), "direction must be either DirectionLeft or DirectionRight")
	require.Equal(t, model.IdSearchReq{}, req, "expected zero value on error")
}

// TestSearchByIDNegativeLevel verifies that NewIdSearchReq rejects negative level values.
func TestSearchByIDNegativeLevel(t *testing.T) {
	target := unittest.IdentifierFixture(t)
	negativeLevel := types.Level(-1)

	req, err := model.NewIdSearchReq(target, negativeLevel, types.DirectionLeft)

	require.Error(t, err)
	require.Contains(t, err.Error(), "level must be non-negative")
	require.Equal(t, model.IdSearchReq{}, req, "expected zero value on error")
}

// TestSearchByIDLevelExceedsMax verifies that NewIdSearchReq rejects level >= MaxLookupTableLevel.
func TestSearchByIDLevelExceedsMax(t *testing.T) {
	target := unittest.IdentifierFixture(t)

	testCases := []struct {
		name  string
		level types.Level
	}{
		{"exactly_max", model.IdentifierSizeBytes * 8},
		{"exceeds_max", model.IdentifierSizeBytes*8 + 1},
		{"far_exceeds_max", model.IdentifierSizeBytes*8 + 100},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				req, err := model.NewIdSearchReq(target, tc.level, types.DirectionRight)

				require.Error(t, err)
				require.Contains(t, err.Error(), "level must be less than")
				require.Equal(t, model.IdSearchReq{}, req, "expected zero value on error")
			},
		)
	}
}

// mockErrorLookupTable is a mock implementation that returns errors at a specific level.
type mockErrorLookupTable struct {
	errorAtLevel types.Level
}

func (m *mockErrorLookupTable) GetEntry(dir types.Direction, level types.Level) (
	*model.Identity,
	error,
) {
	if level == m.errorAtLevel {
		return nil, fmt.Errorf("simulated lookup table error")
	}
	return nil, nil
}

func (m *mockErrorLookupTable) AddEntry(
	dir types.Direction,
	level types.Level,
	identity model.Identity,
) error {
	return nil
}
