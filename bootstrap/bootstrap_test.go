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
	"github.com/thep2p/skipgraph-go/node"
	"github.com/thep2p/skipgraph-go/unittest"
)

// hasNeighbor checks if a node has a valid neighbor in the given direction and level
func hasNeighbor(n *node.SkipGraphNode, dir core.Direction, level core.Level) bool {
	neighbor, err := n.GetNeighbor(dir, level)
	return err == nil && neighbor != nil
}

// TestBootstrapSingleNode tests bootstrap with a single node
func TestBootstrapSingleNode(t *testing.T) {
	logger := unittest.Logger(zerolog.TraceLevel)
	bootstrapper := NewBootstrapper(logger, 1)

	nodes, err := bootstrapper.Bootstrap()
	require.NoError(t, err)
	require.NotNil(t, nodes)
	assert.Len(t, nodes, 1)

	// Single node should have no neighbors
	n := nodes[0]
	leftNeighbor, err := n.GetNeighbor(core.LeftDirection, 0)
	require.NoError(t, err)
	assert.Nil(t, leftNeighbor, "Single node should have no left neighbor")

	rightNeighbor, err := n.GetNeighbor(core.RightDirection, 0)
	require.NoError(t, err)
	assert.Nil(t, rightNeighbor, "Single node should have no right neighbor")
}

// TestBootstrapSmallGraph tests bootstrap with a small number of nodes
func TestBootstrapSmallGraph(t *testing.T) {
	nodeCount := 5
	logger := unittest.Logger(zerolog.TraceLevel)
	bootstrapper := NewBootstrapper(logger, nodeCount)

	nodes, err := bootstrapper.Bootstrap()
	require.NoError(t, err)
	require.NotNil(t, nodes)
	assert.Len(t, nodes, nodeCount)

	// Verify level 0 is properly sorted and linked
	t.Run(
		"Level0Ordering", func(t *testing.T) {
			verifyLevel0Ordering(t, nodes)
		},
	)

	// Verify neighbor consistency
	t.Run(
		"NeighborConsistency", func(t *testing.T) {
			verifyNeighborConsistency(t, nodes)
		},
	)

	// Verify membership vector prefixes
	t.Run(
		"MembershipVectorPrefixes", func(t *testing.T) {
			verifyMembershipVectorPrefixes(t, nodes)
		},
	)
}

// TestBootstrapMediumGraph tests bootstrap with a medium number of nodes
func TestBootstrapMediumGraph(t *testing.T) {
	nodeCount := 100
	logger := unittest.Logger(zerolog.InfoLevel)
	bootstrapper := NewBootstrapper(logger, 100)

	nodes, err := bootstrapper.Bootstrap()
	require.NoError(t, err)
	require.NotNil(t, nodes)
	assert.Len(t, nodes, nodeCount)

	// Verify level 0 is properly sorted and linked
	t.Run(
		"Level0Ordering", func(t *testing.T) {
			verifyLevel0Ordering(t, nodes)
		},
	)

	// Verify neighbor consistency
	t.Run(
		"NeighborConsistency", func(t *testing.T) {
			verifyNeighborConsistency(t, nodes)
		},
	)

	// Verify connected components at each level
	t.Run(
		"ConnectedComponents", func(t *testing.T) {
			verifyConnectedComponents(t, nodes)
		},
	)
}

// TestBootstrapLargeGraph tests bootstrap with a large number of nodes
func TestBootstrapLargeGraph(t *testing.T) {
	nodeCount := 100

	logger := unittest.Logger(zerolog.WarnLevel)
	bootstrapper := NewBootstrapper(logger, nodeCount)

	nodes, err := bootstrapper.Bootstrap()
	require.NoError(t, err)
	require.NotNil(t, nodes)
	assert.Len(t, nodes, nodeCount)

	// Verify basic properties
	t.Run(
		"Level0Ordering", func(t *testing.T) {
			verifyLevel0Ordering(t, nodes)
		},
	)

	// TODO: verify neighbor consistency
	// TODO: verify connected components
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
func verifyLevel0Ordering(t *testing.T, nodes []*node.SkipGraphNode) {
	t.Helper()

	// Verify nodes are sorted by identifier
	for i := 1; i < len(nodes); i++ {
		idPrev := nodes[i-1].Identifier()
		idCurr := nodes[i].Identifier()
		comp := idPrev.Compare(&idCurr)
		assert.Equal(
			t, model.CompareLess, comp.GetComparisonResult(),
			"Nodes should be sorted in ascending order at index %d", i,
		)
	}

	// Traverse from left to right and verify we visit all nodes
	visited := make(map[model.Identifier]bool)
	current := nodes[0]
	visited[current.Identifier()] = true

	for {
		if !hasNeighbor(current, core.RightDirection, 0) {
			break // Reached the end
		}

		rightNeighbor, _ := current.GetNeighbor(core.RightDirection, 0)
		if rightNeighbor != nil {
			rightId := rightNeighbor.GetIdentifier()
			assert.False(t, visited[rightId], "Should not visit same node twice")
			visited[rightId] = true

			// Find the node with this identifier
			found := false
			for _, n := range nodes {
				if n.Identifier() == rightId {
					current = n
					found = true
					break
				}
			}
			assert.True(t, found, "Neighbor should exist in nodes array")
		}
	}

	assert.Len(t, visited, len(nodes), "Should visit all nodes when traversing level 0")
}

// verifyNeighborConsistency verifies that neighbor relationships are bidirectional
func verifyNeighborConsistency(t *testing.T, nodes []*node.SkipGraphNode) {
	t.Helper()

	for level := core.Level(0); level <= core.MaxLookupTableLevel; level++ {
		for _, n := range nodes {

			// Check left neighbor consistency
			// If node n has a left neighbor, verify that the left neighbor points back to n as its right neighbor
			if hasNeighbor(n, core.LeftDirection, level) {
				leftNeighbor, _ := n.GetNeighbor(core.LeftDirection, level)
				if leftNeighbor != nil {
					leftId := leftNeighbor.GetIdentifier()
					// Find the left neighbor node
					for _, other := range nodes {
						if other.Identifier() == leftId {
							// Verify that the left neighbor points back to this node as its right neighbor
							assert.True(
								t, hasNeighbor(other, core.RightDirection, level), "Left neighbor should have a right neighbor at level %d", level,
							)
							rightOfLeft, _ := other.GetNeighbor(core.RightDirection, level)
							require.NotNil(t, rightOfLeft, "Right neighbor of left should not be nil")
							assert.Equal(
								t, n.Identifier(), rightOfLeft.GetIdentifier(),
								"Bidirectional neighbor relationship broken at level %d", level,
							)
							break
						}
					}
				}
			}

			// Check right neighbor consistency
			if hasNeighbor(n, core.RightDirection, level) {
				rightNeighbor, _ := n.GetNeighbor(core.RightDirection, level)
				if rightNeighbor != nil {
					rightId := rightNeighbor.GetIdentifier()
					// Find the right neighbor node
					for _, other := range nodes {
						if other.Identifier() == rightId {
							// Verify that the right neighbor points back to this node as its left neighbor
							assert.True(
								t, hasNeighbor(other, core.LeftDirection, level),
								"Right neighbor should have a left neighbor at level %d", level,
							)
							leftOfRight, _ := other.GetNeighbor(core.LeftDirection, level)
							require.NotNil(t, leftOfRight, "Left neighbor of right should not be nil")
							assert.Equal(
								t, n.Identifier(), leftOfRight.GetIdentifier(),
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
func verifyMembershipVectorPrefixes(t *testing.T, nodes []*node.SkipGraphNode) {
	t.Helper()

	for level := core.Level(1); level <= core.MaxLookupTableLevel; level++ {
		for _, n := range nodes {
			nodeMV := n.MembershipVector()

			// Check left neighbor
			// If node n has a left neighbor, verify that the left neighbor shares at least 'level' bits of prefix
			if hasNeighbor(n, core.LeftDirection, level) {
				leftNeighbor, _ := n.GetNeighbor(core.LeftDirection, level)
				if leftNeighbor != nil {
					leftMV := leftNeighbor.GetMembershipVector()
					commonPrefix := nodeMV.CommonPrefix(leftMV)
					require.GreaterOrEqual(t, commonPrefix, int(level), "Left neighbor at level %d should have at least %d bits common prefix, got %d bits", level, level, commonPrefix)
				}
			}

			// Check right neighbor
			// If node n has a right neighbor, verify that the right neighbor shares at least 'level' bits of prefix
			if hasNeighbor(n, core.RightDirection, level) {
				rightNeighbor, _ := n.GetNeighbor(core.RightDirection, level)
				if rightNeighbor != nil {
					rightMV := rightNeighbor.GetMembershipVector()
					commonPrefix := nodeMV.CommonPrefix(rightMV)
					require.GreaterOrEqual(t, commonPrefix, int(level), "Right neighbor at level %d should have at least %d bits common prefix, got %d bits", level, level, commonPrefix)
				}
			}
		}
	}
}

// verifyConnectedComponents verifies that nodes with matching prefixes form connected components
func verifyConnectedComponents(t *testing.T, nodes []*node.SkipGraphNode) {
	t.Helper()

	for level := core.Level(1); level <= core.MaxLookupTableLevel; level++ {
		// Group nodes by their membership vector prefix at this level
		prefixGroups := make(map[string][]*node.SkipGraphNode)
		for _, n := range nodes {
			mv := n.MembershipVector()
			prefix := mv.GetPrefixBits(int(level))
			prefixGroups[prefix] = append(prefixGroups[prefix], n)
		}

		// For each group, verify they form a connected component
		for prefix, group := range prefixGroups {
			if len(group) <= 1 {
				continue // Single node is trivially connected
			}

			// Pick the first node and verify all others are reachable
			start := group[0]
			reachable := make(map[model.Identifier]bool)
			dfsReachable(nodes, start, level, reachable)

			for _, n := range group {
				assert.True(
					t, reachable[n.Identifier()],
					"Node with prefix %s should be reachable at level %d", prefix, level,
				)
			}
		}
	}
}

// dfsReachable performs DFS to find all reachable nodes from a starting node at a given level
func dfsReachable(nodes []*node.SkipGraphNode, start *node.SkipGraphNode, level core.Level, visited map[model.Identifier]bool) {
	// Create identifier to index map for O(1) lookups
	idToIndex := make(map[model.Identifier]int)
	for i, n := range nodes {
		idToIndex[n.Identifier()] = i
	}

	// Find the starting node's index
	startIndex := -1
	for i, n := range nodes {
		if n.Identifier() == start.Identifier() {
			startIndex = i
			break
		}
	}
	if startIndex == -1 {
		return // Node not found in array
	}

	// Convert visited map from Identifier->bool to int->bool for TraverseConnectedNodes
	visitedIndices := make(map[int]bool)

	// Use the consolidated traversal function
	logger := unittest.Logger(zerolog.TraceLevel)
	bootstrapper := NewBootstrapper(logger, len(nodes))
	bootstrapper.TraverseConnectedNodes(nodes, startIndex, level, visitedIndices, idToIndex)

	// Convert visitedIndices back to visited identifiers
	for index := range visitedIndices {
		visited[nodes[index].Identifier()] = true
	}
}

// TestTraversalWithNodeReference tests traversal using (identifier, array_index) pairs
func TestTraversalWithNodeReference(t *testing.T) {
	nodeCount := 10
	logger := unittest.Logger(zerolog.InfoLevel)
	bootstrapper := NewBootstrapper(logger, nodeCount)

	nodes, err := bootstrapper.Bootstrap()
	require.NoError(t, err)

	// Create node references for testing
	nodeRefs := make([]internal.NodeReference, len(nodes))
	for i, n := range nodes {
		nodeRefs[i] = internal.NodeReference{
			Identifier: n.Identifier(),
			ArrayIndex: i,
		}
	}

	// Test traversal at level 0
	t.Run(
		"TraverseLevel0", func(t *testing.T) {
			traversed := traverseLevel(nodes, nodeRefs[0], core.Level(0))
			assert.Len(t, traversed, len(nodes), "Should traverse all nodes at level 0")

			// Verify order; identifiers at level zero should be in acsending order
			for i := 1; i < len(traversed); i++ {
				idPrev := traversed[i-1].Identifier
				idCurr := traversed[i].Identifier
				comp := idPrev.Compare(&idCurr)
				assert.Equal(t, model.CompareLess, comp.GetComparisonResult(), "Nodes should be in sorted order")
			}
		},
	)

	// Test traversal at higher levels
	t.Run("TraverseHigherLevels", func(t *testing.T) {
		for level := core.Level(1); level <= core.MaxLookupTableLevel; level++ {
			// Find a node that has neighbors at this level
			var startRef internal.NodeReference
			hasNeighborAtLevel := false
			for i, n := range nodes {
				if hasNeighbor(n, core.RightDirection, level) {
					startRef = nodeRefs[i]
					hasNeighborAtLevel = true
					break
				}
			}

			if hasNeighborAtLevel {
				traversed := traverseLevel(nodes, startRef, level)
				require.NotEmpty(t, traversed, "Should traverse at least one node at level %d", level)

				// Verify all traversed nodes have matching prefix
				startMV := nodes[startRef.ArrayIndex].MembershipVector()
				prefix := startMV.GetPrefixBits(int(level))
				for _, ref := range traversed {
					nodeMV := nodes[ref.ArrayIndex].MembershipVector()
					nodePrefix := nodeMV.GetPrefixBits(int(level))
					assert.Equal(
						t, prefix, nodePrefix,
						"All traversed nodes should have same prefix at level %d", level,
					)
				}
			}
		}
	},
	)
}

// traverseLevel traverses all connected nodes at a given level starting from a node reference
func traverseLevel(nodes []*node.SkipGraphNode, start internal.NodeReference, level core.Level) []internal.NodeReference {
	visited := make(map[model.Identifier]bool)
	result := []internal.NodeReference{}

	// Traverse left
	current := start
	for {
		if visited[current.Identifier] {
			break
		}
		visited[current.Identifier] = true

		n := nodes[current.ArrayIndex]
		if hasNeighbor(n, core.LeftDirection, level) {
			leftNeighbor, _ := n.GetNeighbor(core.LeftDirection, level)
			if leftNeighbor != nil {
				leftId := leftNeighbor.GetIdentifier()
				// Find the array index of this neighbor
				found := false
				for i, other := range nodes {
					if other.Identifier() == leftId {
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

	// Now traverse right from the leftmost node
	leftmost := current
	current = leftmost
	for {
		if !visited[current.Identifier] {
			visited[current.Identifier] = true
		}
		result = append(result, current)

		n := nodes[current.ArrayIndex]
		if hasNeighbor(n, core.RightDirection, level) {
			rightNeighbor, _ := n.GetNeighbor(core.RightDirection, level)
			if rightNeighbor != nil {
				rightId := rightNeighbor.GetIdentifier()
				if visited[rightId] {
					break // Avoid cycles
				}
				// Find the array index of this neighbor
				found := false
				for i, other := range nodes {
					if other.Identifier() == rightId {
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

			nodes, err := bootstrapper.Bootstrap()
			require.NoError(t, err)
			require.NotNil(t, nodes)
			assert.Len(t, nodes, tc.nodeCount)

			// For each level, verify that the number of connected components is at most 2^i
			for level := core.Level(0); level <= tc.maxLevel && level < core.MaxLookupTableLevel; level++ {
				componentCount := bootstrapper.CountConnectedComponents(nodes, level)
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

	nodes, err := bootstrapper.Bootstrap()
	require.NoError(t, err)

	// Collect statistics about connected components at each level
	stats := make(map[core.Level]int)
	for level := core.Level(0); level <= 10 && level < core.MaxLookupTableLevel; level++ {
		componentCount := bootstrapper.CountConnectedComponents(nodes, level)
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
