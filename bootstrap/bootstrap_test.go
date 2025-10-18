package bootstrap

import (
	"fmt"
	"github.com/thep2p/skipgraph-go/bootstrap/internal"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/core"
	"github.com/thep2p/skipgraph-go/core/model"
	"github.com/thep2p/skipgraph-go/unittest"
)

// hasNeighbor checks if a bootstrap entry has a valid neighbor in the given direction and level
func hasNeighbor(entry *BootstrapEntry, dir core.Direction, level core.Level) bool {
	neighbor, err := entry.LookupTable.GetEntry(dir, level)
	return err == nil && neighbor != nil
}

// TestBootstrapSingleNode tests bootstrap with a single node
func TestBootstrapSingleNode(t *testing.T) {
	logger := unittest.Logger(zerolog.TraceLevel)
	bootstrapper := NewBootstrapper(logger, 1)

	entries, err := bootstrapper.Bootstrap()
	require.NoError(t, err)
	require.NotNil(t, entries)
	assert.Len(t, entries, 1)

	// Single node should have no neighbors
	entry := entries[0]
	leftNeighbor, err := entry.LookupTable.GetEntry(core.LeftDirection, 0)
	require.NoError(t, err)
	assert.Nil(t, leftNeighbor, "Single node should have no left neighbor")

	rightNeighbor, err := entry.LookupTable.GetEntry(core.RightDirection, 0)
	require.NoError(t, err)
	assert.Nil(t, rightNeighbor, "Single node should have no right neighbor")
}

// TestBootstrapSmallGraph tests bootstrap with a small number of nodes
func TestBootstrapSmallGraph(t *testing.T) {
	nodeCount := 5
	logger := unittest.Logger(zerolog.TraceLevel)
	bootstrapper := NewBootstrapper(logger, nodeCount)

	entries, err := bootstrapper.Bootstrap()
	require.NoError(t, err)
	require.NotNil(t, entries)
	assert.Len(t, entries, nodeCount)

	// Verify level 0 is properly sorted and linked
	t.Run(
		"Level0Ordering", func(t *testing.T) {
			verifyLevel0Ordering(t, entries)
		},
	)

	// Verify neighbor consistency
	t.Run(
		"NeighborConsistency", func(t *testing.T) {
			verifyNeighborConsistency(t, entries)
		},
	)

	// Verify membership vector prefixes
	t.Run(
		"MembershipVectorPrefixes", func(t *testing.T) {
			verifyMembershipVectorPrefixes(t, entries)
		},
	)
}

// TestBootstrapMediumGraph tests bootstrap with a medium number of nodes
func TestBootstrapMediumGraph(t *testing.T) {
	nodeCount := 100
	logger := unittest.Logger(zerolog.InfoLevel)
	bootstrapper := NewBootstrapper(logger, 100)

	entries, err := bootstrapper.Bootstrap()
	require.NoError(t, err)
	require.NotNil(t, entries)
	assert.Len(t, entries, nodeCount)

	// Verify level 0 is properly sorted and linked
	t.Run(
		"Level0Ordering", func(t *testing.T) {
			verifyLevel0Ordering(t, entries)
		},
	)

	// Verify neighbor consistency
	t.Run(
		"NeighborConsistency", func(t *testing.T) {
			verifyNeighborConsistency(t, entries)
		},
	)

	// Verify connected components at each level
	t.Run(
		"ConnectedComponents", func(t *testing.T) {
			verifyConnectedComponents(t, entries)
		},
	)
}

// TestBootstrapLargeGraph tests bootstrap with a large number of nodes
func TestBootstrapLargeGraph(t *testing.T) {
	nodeCount := 100

	logger := unittest.Logger(zerolog.WarnLevel)
	bootstrapper := NewBootstrapper(logger, nodeCount)

	entries, err := bootstrapper.Bootstrap()
	require.NoError(t, err)
	require.NotNil(t, entries)
	assert.Len(t, entries, nodeCount)

	// Verify basic properties
	t.Run(
		"Level0Ordering", func(t *testing.T) {
			verifyLevel0Ordering(t, entries)
		},
	)

	// Verify neighbor consistency
	t.Run(
		"NeighborConsistency", func(t *testing.T) {
			verifyNeighborConsistency(t, entries)
		},
	)

	// Verify connected components
	t.Run(
		"ConnectedComponents", func(t *testing.T) {
			verifyConnectedComponents(t, entries)
		},
	)
}

// TestBootstrapInvalidInput tests bootstrap with invalid input
func TestBootstrapInvalidInput(t *testing.T) {
	logger := unittest.Logger(zerolog.ErrorLevel)

	testCases := []struct {
		name     string
		numNodes int
	}{
		{"Zero nodes", 0},
		{"Negative nodes", -1},
		{"Large negative", -100},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				bootstrapper := NewBootstrapper(logger, tc.numNodes)
				result, err := bootstrapper.Bootstrap()
				assert.Error(t, err)
				assert.Nil(t, result)
			},
		)
	}
}

// verifyLevel0Ordering verifies that level 0 forms a sorted doubly-linked list
func verifyLevel0Ordering(t *testing.T, entries []*BootstrapEntry) {
	t.Helper()

	// Verify entries are sorted by identifier
	for i := 1; i < len(entries); i++ {
		idPrev := entries[i-1].Identity.GetIdentifier()
		idCurr := entries[i].Identity.GetIdentifier()
		comp := idPrev.Compare(&idCurr)
		assert.Equal(
			t, model.CompareLess, comp.GetComparisonResult(),
			"Entries should be sorted in ascending order at index %d", i,
		)
	}

	// Traverse from left to right and verify we visit all entries
	visited := make(map[model.Identifier]bool)
	current := entries[0]
	visited[current.Identity.GetIdentifier()] = true

	for {
		if !hasNeighbor(current, core.RightDirection, 0) {
			break // Reached the end
		}

		rightNeighbor, _ := current.LookupTable.GetEntry(core.RightDirection, 0)
		if rightNeighbor != nil {
			rightId := rightNeighbor.GetIdentifier()
			assert.False(t, visited[rightId], "Should not visit same entry twice")
			visited[rightId] = true

			// Find the entry with this identifier
			found := false
			for _, e := range entries {
				if e.Identity.GetIdentifier() == rightId {
					current = e
					found = true
					break
				}
			}
			assert.True(t, found, "Neighbor should exist in entries array")
		}
	}

	assert.Len(t, visited, len(entries), "Should visit all entries when traversing level 0")
}

// verifyNeighborConsistency verifies that neighbor relationships are bidirectional
func verifyNeighborConsistency(t *testing.T, entries []*BootstrapEntry) {
	t.Helper()

	for level := core.Level(0); level <= core.MaxLookupTableLevel; level++ {
		for _, e := range entries {

			// Check left neighbor consistency
			// If entry e has a left neighbor, verify that the left neighbor points back to e as its right neighbor
			if hasNeighbor(e, core.LeftDirection, level) {
				leftNeighbor, _ := e.LookupTable.GetEntry(core.LeftDirection, level)
				if leftNeighbor != nil {
					leftId := leftNeighbor.GetIdentifier()
					// Find the left neighbor entry
					for _, other := range entries {
						if other.Identity.GetIdentifier() == leftId {
							// Verify that the left neighbor points back to this entry as its right neighbor
							assert.True(
								t, hasNeighbor(other, core.RightDirection, level), "Left neighbor should have a right neighbor at level %d", level,
							)
							rightOfLeft, _ := other.LookupTable.GetEntry(core.RightDirection, level)
							require.NotNil(t, rightOfLeft, "Right neighbor of left should not be nil")
							assert.Equal(
								t, e.Identity.GetIdentifier(), rightOfLeft.GetIdentifier(),
								"Bidirectional neighbor relationship broken at level %d", level,
							)
							break
						}
					}
				}
			}

			// Check right neighbor consistency
			if hasNeighbor(e, core.RightDirection, level) {
				rightNeighbor, _ := e.LookupTable.GetEntry(core.RightDirection, level)
				if rightNeighbor != nil {
					rightId := rightNeighbor.GetIdentifier()
					// Find the right neighbor entry
					for _, other := range entries {
						if other.Identity.GetIdentifier() == rightId {
							// Verify that the right neighbor points back to this entry as its left neighbor
							assert.True(
								t, hasNeighbor(other, core.LeftDirection, level),
								"Right neighbor should have a left neighbor at level %d", level,
							)
							leftOfRight, _ := other.LookupTable.GetEntry(core.LeftDirection, level)
							require.NotNil(t, leftOfRight, "Left neighbor of right should not be nil")
							assert.Equal(
								t, e.Identity.GetIdentifier(), leftOfRight.GetIdentifier(),
								"Bidirectional neighbor relationship broken at level %d", level,
							)
							break
						}
					}
				}
			}
		}
	}
}

// verifyMembershipVectorPrefixes verifies that neighbors at each level have matching membership vector prefixes
func verifyMembershipVectorPrefixes(t *testing.T, entries []*BootstrapEntry) {
	t.Helper()

	for level := core.Level(1); level <= core.MaxLookupTableLevel; level++ {
		for _, e := range entries {
			entryMV := e.Identity.GetMembershipVector()

			// Check left neighbor
			// If entry e has a left neighbor, verify that the left neighbor shares at least 'level' bits of prefix
			if hasNeighbor(e, core.LeftDirection, level) {
				leftNeighbor, _ := e.LookupTable.GetEntry(core.LeftDirection, level)
				if leftNeighbor != nil {
					leftMV := leftNeighbor.GetMembershipVector()
					commonPrefix := entryMV.CommonPrefix(leftMV)
					require.GreaterOrEqual(t, commonPrefix, int(level), "Left neighbor at level %d should have at least %d bits common prefix, got %d bits", level, level, commonPrefix)
				}
			}

			// Check right neighbor
			// If entry e has a right neighbor, verify that the right neighbor shares at least 'level' bits of prefix
			if hasNeighbor(e, core.RightDirection, level) {
				rightNeighbor, _ := e.LookupTable.GetEntry(core.RightDirection, level)
				if rightNeighbor != nil {
					rightMV := rightNeighbor.GetMembershipVector()
					commonPrefix := entryMV.CommonPrefix(rightMV)
					require.GreaterOrEqual(t, commonPrefix, int(level), "Right neighbor at level %d should have at least %d bits common prefix, got %d bits", level, level, commonPrefix)
				}
			}
		}
	}
}

// verifyConnectedComponents verifies that entries with matching prefixes form connected components
func verifyConnectedComponents(t *testing.T, entries []*BootstrapEntry) {
	t.Helper()

	// Create identifier to index map once for O(1) lookups
	idToIndex := make(map[model.Identifier]int)
	for i, entry := range entries {
		idToIndex[entry.Identity.GetIdentifier()] = i
	}

	for level := core.Level(1); level <= core.MaxLookupTableLevel; level++ {
		// Group entries by their membership vector prefix at this level
		prefixGroups := make(map[string][]*BootstrapEntry)
		for _, e := range entries {
			mv := e.Identity.GetMembershipVector()
			prefix, err := mv.GetPrefixBits(int(level))
			require.NoError(t, err)
			prefixGroups[prefix] = append(prefixGroups[prefix], e)
		}

		// For each group, verify they form a connected component
		for prefix, group := range prefixGroups {
			if len(group) <= 1 {
				continue // Single entry is trivially connected
			}

			// Pick the first entry and verify all others are reachable
			startId := group[0].Identity.GetIdentifier()
			reachable := make(map[model.Identifier]bool)
			dfsReachable(entries, startId, level, reachable, idToIndex)

			for _, e := range group {
				assert.True(
					t, reachable[e.Identity.GetIdentifier()],
					"Entry with prefix %s should be reachable at level %d", prefix, level,
				)
			}
		}
	}
}

// dfsReachable performs DFS to find all reachable entries from a starting identifier at a given level.
// The idToIndex map is passed in to avoid redundant map creation on each call.
func dfsReachable(entries []*BootstrapEntry, startId model.Identifier, level core.Level, visited map[model.Identifier]bool, idToIndex map[model.Identifier]int) {
	// Find the starting entry's index
	startIndex, exists := idToIndex[startId]
	if !exists {
		return // Entry not found
	}

	// Convert visited map from Identifier->bool to int->bool for TraverseConnectedEntries
	visitedIndices := make(map[int]bool)

	// Use the consolidated traversal function
	logger := unittest.Logger(zerolog.TraceLevel)
	bootstrapper := NewBootstrapper(logger, len(entries))
	bootstrapper.TraverseConnectedEntries(entries, startIndex, level, visitedIndices, idToIndex)

	// Convert visitedIndices back to visited identifiers
	for index := range visitedIndices {
		visited[entries[index].Identity.GetIdentifier()] = true
	}
}

// TestTraversalWithNodeReference tests traversal using (identifier, array_index) pairs
func TestTraversalWithNodeReference(t *testing.T) {
	nodeCount := 10
	logger := unittest.Logger(zerolog.InfoLevel)
	bootstrapper := NewBootstrapper(logger, nodeCount)

	entries, err := bootstrapper.Bootstrap()
	require.NoError(t, err)

	// Create entry references for testing
	entryRefs := make([]internal.NodeReference, len(entries))
	for i, e := range entries {
		entryRefs[i] = internal.NodeReference{
			Identifier: e.Identity.GetIdentifier(),
			ArrayIndex: i,
		}
	}

	// Test traversal at level 0
	t.Run(
		"TraverseLevel0", func(t *testing.T) {
			traversed := traverseLevel(entries, entryRefs[0], core.Level(0))
			assert.Len(t, traversed, len(entries), "Should traverse all entries at level 0")

			// Verify order; identifiers at level zero should be in ascending order
			for i := 1; i < len(traversed); i++ {
				idPrev := traversed[i-1].Identifier
				idCurr := traversed[i].Identifier
				comp := idPrev.Compare(&idCurr)
				assert.Equal(t, model.CompareLess, comp.GetComparisonResult(), "Entries should be in sorted order")
			}
		},
	)

	// Test traversal at higher levels
	t.Run("TraverseHigherLevels", func(t *testing.T) {
		for level := core.Level(1); level <= core.MaxLookupTableLevel; level++ {
			// Find an entry that has neighbors at this level
			var startRef internal.NodeReference
			hasNeighborAtLevel := false
			for i, e := range entries {
				if hasNeighbor(e, core.RightDirection, level) {
					startRef = entryRefs[i]
					hasNeighborAtLevel = true
					break
				}
			}

			if hasNeighborAtLevel {
				traversed := traverseLevel(entries, startRef, level)
				require.NotEmpty(t, traversed, "Should traverse at least one entry at level %d", level)

				// Verify all traversed entries have matching prefix
				startMV := entries[startRef.ArrayIndex].Identity.GetMembershipVector()
				prefix, err := startMV.GetPrefixBits(int(level))
				require.NoError(t, err)
				for _, ref := range traversed {
					entryMV := entries[ref.ArrayIndex].Identity.GetMembershipVector()
					entryPrefix, err := entryMV.GetPrefixBits(int(level))
					require.NoError(t, err)
					assert.Equal(
						t, prefix, entryPrefix,
						"All traversed entries should have same prefix at level %d", level,
					)
				}
			}
		}
	},
	)
}

// traverseLevel traverses all connected entries at a given level starting from a node reference
func traverseLevel(entries []*BootstrapEntry, start internal.NodeReference, level core.Level) []internal.NodeReference {
	visited := make(map[model.Identifier]bool)
	result := []internal.NodeReference{}

	// Traverse left
	current := start
	for {
		if visited[current.Identifier] {
			break
		}
		visited[current.Identifier] = true

		entry := entries[current.ArrayIndex]
		if hasNeighbor(entry, core.LeftDirection, level) {
			leftNeighbor, _ := entry.LookupTable.GetEntry(core.LeftDirection, level)
			if leftNeighbor != nil {
				leftId := leftNeighbor.GetIdentifier()
				// Find the array index of this neighbor
				found := false
				for i, other := range entries {
					if other.Identity.GetIdentifier() == leftId {
						current = internal.NodeReference{
							Identifier: leftId,
							ArrayIndex: i,
						}
						found = true
						break
					}
				}
				if !found {
					break
				}
			} else {
				break
			}
		} else {
			break
		}
	}

	// Now traverse right from the leftmost entry
	leftmost := current
	current = leftmost
	for {
		if !visited[current.Identifier] {
			visited[current.Identifier] = true
		}
		result = append(result, current)

		entry := entries[current.ArrayIndex]
		if hasNeighbor(entry, core.RightDirection, level) {
			rightNeighbor, _ := entry.LookupTable.GetEntry(core.RightDirection, level)
			if rightNeighbor != nil {
				rightId := rightNeighbor.GetIdentifier()
				if visited[rightId] {
					break // Avoid cycles
				}
				// Find the array index of this neighbor
				found := false
				for i, other := range entries {
					if other.Identity.GetIdentifier() == rightId {
						current = internal.NodeReference{
							Identifier: rightId,
							ArrayIndex: i,
						}
						found = true
						break
					}
				}
				if !found {
					break
				}
			} else {
				break
			}
		} else {
			break
		}
	}

	return result
}

// TestConnectedComponentsConstraint verifies that at level i there are at most 2^i connected components
func TestConnectedComponentsConstraint(t *testing.T) {
	testCases := []struct {
		name      string
		nodeCount int
		maxLevel  core.Level
	}{
		{"Small graph (10 nodes)", 10, 4},
		{"Medium graph (50 nodes)", 50, 6},
		{"Large graph (100 nodes)", 100, 7},
		{"Very large graph (500 nodes)", 500, 9},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := unittest.Logger(zerolog.WarnLevel)
			bootstrapper := NewBootstrapper(logger, tc.nodeCount)

			entries, err := bootstrapper.Bootstrap()
			require.NoError(t, err)
			require.NotNil(t, entries)
			assert.Len(t, entries, tc.nodeCount)

			// For each level, verify that the number of connected components is at most 2^i
			for level := core.Level(0); level <= tc.maxLevel && level < core.MaxLookupTableLevel; level++ {
				componentCount := bootstrapper.CountConnectedComponents(entries, level)
				maxComponents := 1 << level // 2^level

				assert.LessOrEqual(
					t, componentCount, maxComponents,
					"At level %d, found %d connected components but expected at most %d (2^%d)",
					level, componentCount, maxComponents, level,
				)

				t.Logf("Level %d: %d connected components (max allowed: %d)",
					level, componentCount, maxComponents)
			}
		})
	}
}

// TestConnectedComponentsDistribution verifies the distribution of connected components across first 10 levels
func TestConnectedComponentsDistribution(t *testing.T) {
	nodeCount := 200
	logger := unittest.Logger(zerolog.InfoLevel)
	bootstrapper := NewBootstrapper(logger, nodeCount)

	entries, err := bootstrapper.Bootstrap()
	require.NoError(t, err)

	// Collect statistics about connected components at each level
	stats := make(map[core.Level]int)
	for level := core.Level(0); level <= 10 && level < core.MaxLookupTableLevel; level++ {
		componentCount := bootstrapper.CountConnectedComponents(entries, level)
		stats[level] = componentCount
	}

	// Verify the constraint and print distribution
	t.Log("Connected components distribution:")
	for level := core.Level(0); level <= 10 && level < core.MaxLookupTableLevel; level++ {
		componentCount := stats[level]
		maxComponents := 1 << level // 2^level

		require.LessOrEqual(
			t, componentCount, maxComponents,
			"Level %d violates constraint: %d components > %d max",
			level, componentCount, maxComponents,
		)

		// Calculate utilization percentage
		utilization := float64(componentCount) / float64(maxComponents) * 100
		t.Logf("Level %2d: %3d components / %4d max (%.1f%% utilization)",
			level, componentCount, maxComponents, utilization)

		// At higher levels, we expect to approach the maximum as nodes become more distributed
		if level > 0 && componentCount == nodeCount {
			// All nodes are in separate components - we've reached maximum fragmentation
			t.Logf("  -> Maximum fragmentation reached at level %d", level)
			break
		}
	}
}

// BenchmarkBootstrap benchmarks bootstrap performance
func BenchmarkBootstrap(b *testing.B) {
	logger := unittest.Logger(zerolog.ErrorLevel)

	nodeCounts := []int{10, 100, 1000}
	for _, nodeCount := range nodeCounts {
		b.Run(
			fmt.Sprintf("Size-%d", nodeCount), func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					bootstrapper := NewBootstrapper(logger, nodeCount)
					_, err := bootstrapper.Bootstrap()
					if err != nil {
						b.Fatal(err)
					}
				}
			},
		)
	}
}
