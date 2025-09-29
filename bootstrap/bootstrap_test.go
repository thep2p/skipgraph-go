package bootstrap

import (
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/core"
	"github.com/thep2p/skipgraph-go/core/model"
	"github.com/thep2p/skipgraph-go/node"
	"github.com/thep2p/skipgraph-go/unittest"
)

// isEmptyIdentity checks if an identifier is empty (all zeros)
func isEmptyIdentity(id model.Identifier) bool {
	empty := model.Identifier{}
	return id == empty
}

// hasNeighbor checks if a node has a valid neighbor in the given direction and level
func hasNeighbor(n *node.SkipGraphNode, dir core.Direction, level core.Level) bool {
	neighbor, err := n.GetNeighbor(dir, level)
	if err != nil {
		return false
	}
	return !isEmptyIdentity(neighbor.GetIdentifier())
}

// TestBootstrapSingleNode tests bootstrap with a single node
func TestBootstrapSingleNode(t *testing.T) {
	logger := unittest.Logger(zerolog.TraceLevel)
	bootstrapper := NewBootstrapper(logger)

	result, err := bootstrapper.Bootstrap(1)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Nodes, 1)

	// Single node should have no neighbors
	n := result.Nodes[0]
	leftNeighbor, err := n.GetNeighbor(core.LeftDirection, 0)
	require.NoError(t, err) // GetNeighbor doesn't return error for empty neighbor
	assert.Equal(t, model.Identifier{}, leftNeighbor.GetIdentifier(), "Single node should have no left neighbor")

	rightNeighbor, err := n.GetNeighbor(core.RightDirection, 0)
	require.NoError(t, err) // GetNeighbor doesn't return error for empty neighbor
	assert.Equal(t, model.Identifier{}, rightNeighbor.GetIdentifier(), "Single node should have no right neighbor")
}

// TestBootstrapSmallGraph tests bootstrap with a small number of nodes
func TestBootstrapSmallGraph(t *testing.T) {
	logger := unittest.Logger(zerolog.TraceLevel)
	bootstrapper := NewBootstrapper(logger)

	result, err := bootstrapper.Bootstrap(5)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Nodes, 5)

	// Verify level 0 is properly sorted and linked
	t.Run("Level0Ordering", func(t *testing.T) {
		verifyLevel0Ordering(t, result.Nodes)
	})

	// Verify neighbor consistency
	t.Run("NeighborConsistency", func(t *testing.T) {
		verifyNeighborConsistency(t, result.Nodes, result.MaxLevel)
	})

	// Verify membership vector prefixes
	t.Run("MembershipVectorPrefixes", func(t *testing.T) {
		verifyMembershipVectorPrefixes(t, result.Nodes, result.MaxLevel)
	})
}

// TestBootstrapMediumGraph tests bootstrap with a medium number of nodes
func TestBootstrapMediumGraph(t *testing.T) {
	logger := unittest.Logger(zerolog.InfoLevel)
	bootstrapper := NewBootstrapper(logger)

	result, err := bootstrapper.Bootstrap(100)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Nodes, 100)

	// Verify level 0 is properly sorted and linked
	t.Run("Level0Ordering", func(t *testing.T) {
		verifyLevel0Ordering(t, result.Nodes)
	})

	// Verify neighbor consistency
	t.Run("NeighborConsistency", func(t *testing.T) {
		verifyNeighborConsistency(t, result.Nodes, result.MaxLevel)
	})

	// Verify connected components at each level
	t.Run("ConnectedComponents", func(t *testing.T) {
		verifyConnectedComponents(t, result.Nodes, result.MaxLevel)
	})

	// Verify statistics
	t.Run("Statistics", func(t *testing.T) {
		assert.Greater(t, result.Stats.AverageNeighbors, 0.0)
		assert.Greater(t, result.Stats.TotalLevels, 0)
		assert.NotEmpty(t, result.Stats.ConnectedComponents)

		// Level 0 should have exactly 1 connected component
		assert.Equal(t, 1, result.Stats.ConnectedComponents[0], "Level 0 should have exactly one connected component")
	})
}

// TestBootstrapLargeGraph tests bootstrap with a large number of nodes
func TestBootstrapLargeGraph(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large graph test in short mode")
	}

	logger := unittest.Logger(zerolog.WarnLevel)
	bootstrapper := NewBootstrapper(logger)

	result, err := bootstrapper.Bootstrap(1000)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Nodes, 1000)

	// Verify basic properties
	t.Run("Level0Ordering", func(t *testing.T) {
		verifyLevel0Ordering(t, result.Nodes)
	})

	// Verify statistics
	t.Run("Statistics", func(t *testing.T) {
		assert.Greater(t, result.Stats.AverageNeighbors, 0.0)
		assert.Greater(t, result.MaxLevel, 5, "Large graph should have multiple levels")

		// Check that higher levels have more components
		level0Components := result.Stats.ConnectedComponents[0]
		highLevelComponents := result.Stats.ConnectedComponents[result.MaxLevel]
		assert.Equal(t, 1, level0Components)
		assert.Greater(t, highLevelComponents, level0Components, "Higher levels should have more components")
	})
}

// TestBootstrapInvalidInput tests bootstrap with invalid input
func TestBootstrapInvalidInput(t *testing.T) {
	logger := unittest.Logger(zerolog.ErrorLevel)
	bootstrapper := NewBootstrapper(logger)

	testCases := []struct {
		name     string
		numNodes int
	}{
		{"Zero nodes", 0},
		{"Negative nodes", -1},
		{"Large negative", -100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := bootstrapper.Bootstrap(tc.numNodes)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
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
		assert.Equal(t, model.CompareLess, comp.GetComparisonResult(),
			"Nodes should be sorted in ascending order at index %d", i)
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

	assert.Len(t, visited, len(nodes), "Should visit all nodes when traversing level 0")
}

// verifyNeighborConsistency verifies that neighbor relationships are bidirectional
func verifyNeighborConsistency(t *testing.T, nodes []*node.SkipGraphNode, maxLevel int) {
	t.Helper()

	for level := core.Level(0); level <= core.Level(maxLevel); level++ {
		for _, n := range nodes {
			// Check left neighbor consistency
			if hasNeighbor(n, core.LeftDirection, level) {
				leftNeighbor, _ := n.GetNeighbor(core.LeftDirection, level)
				leftId := leftNeighbor.GetIdentifier()
				// Find the left neighbor node
				for _, other := range nodes {
					if other.Identifier() == leftId {
						// Verify that the left neighbor points back to this node as its right neighbor
						assert.True(t, hasNeighbor(other, core.RightDirection, level),
							"Left neighbor should have a right neighbor at level %d", level)
						rightOfLeft, _ := other.GetNeighbor(core.RightDirection, level)
						assert.Equal(t, n.Identifier(), rightOfLeft.GetIdentifier(),
							"Bidirectional neighbor relationship broken at level %d", level)
						break
					}
				}
			}

			// Check right neighbor consistency
			if hasNeighbor(n, core.RightDirection, level) {
				rightNeighbor, _ := n.GetNeighbor(core.RightDirection, level)
				rightId := rightNeighbor.GetIdentifier()
				// Find the right neighbor node
				for _, other := range nodes {
					if other.Identifier() == rightId {
						// Verify that the right neighbor points back to this node as its left neighbor
						assert.True(t, hasNeighbor(other, core.LeftDirection, level),
							"Right neighbor should have a left neighbor at level %d", level)
						leftOfRight, _ := other.GetNeighbor(core.LeftDirection, level)
						assert.Equal(t, n.Identifier(), leftOfRight.GetIdentifier(),
							"Bidirectional neighbor relationship broken at level %d", level)
						break
					}
				}
			}
		}
	}
}

// verifyMembershipVectorPrefixes verifies that neighbors at each level have matching membership vector prefixes
func verifyMembershipVectorPrefixes(t *testing.T, nodes []*node.SkipGraphNode, maxLevel int) {
	t.Helper()

	for level := 1; level <= maxLevel; level++ {
		for _, n := range nodes {
			nodeMV := n.MembershipVector()

			// Check left neighbor
			if hasNeighbor(n, core.LeftDirection, core.Level(level)) {
				leftNeighbor, _ := n.GetNeighbor(core.LeftDirection, core.Level(level))
				leftMV := leftNeighbor.GetMembershipVector()
				commonPrefix := nodeMV.CommonPrefix(leftMV)
				assert.GreaterOrEqual(t, commonPrefix, level,
					"Left neighbor at level %d should have at least %d bits common prefix", level, level)
			}

			// Check right neighbor
			if hasNeighbor(n, core.RightDirection, core.Level(level)) {
				rightNeighbor, _ := n.GetNeighbor(core.RightDirection, core.Level(level))
				rightMV := rightNeighbor.GetMembershipVector()
				commonPrefix := nodeMV.CommonPrefix(rightMV)
				assert.GreaterOrEqual(t, commonPrefix, level,
					"Right neighbor at level %d should have at least %d bits common prefix", level, level)
			}
		}
	}
}

// verifyConnectedComponents verifies that nodes with matching prefixes form connected components
func verifyConnectedComponents(t *testing.T, nodes []*node.SkipGraphNode, maxLevel int) {
	t.Helper()

	for level := 1; level <= maxLevel; level++ {
		// Group nodes by their membership vector prefix at this level
		prefixGroups := make(map[string][]*node.SkipGraphNode)
		for _, n := range nodes {
			mv := n.MembershipVector()
			prefix := getPrefixBits(mv, level)
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
			dfsReachable(nodes, start, core.Level(level), reachable)

			for _, n := range group {
				assert.True(t, reachable[n.Identifier()],
					"Node with prefix %s should be reachable at level %d", prefix, level)
			}
		}
	}
}

// getPrefixBits returns the first numBits bits of a membership vector as a string
func getPrefixBits(mv model.MembershipVector, numBits int) string {
	binaryStr := mv.ToBinaryString()
	if numBits > len(binaryStr) {
		return binaryStr
	}
	return binaryStr[:numBits]
}

// dfsReachable performs DFS to find all reachable nodes from a starting node at a given level
func dfsReachable(nodes []*node.SkipGraphNode, start *node.SkipGraphNode, level core.Level, visited map[model.Identifier]bool) {
	visited[start.Identifier()] = true

	// Check left neighbor
	if hasNeighbor(start, core.LeftDirection, level) {
		leftNeighbor, _ := start.GetNeighbor(core.LeftDirection, level)
		leftId := leftNeighbor.GetIdentifier()
		if !visited[leftId] {
			// Find the node with this identifier
			for _, n := range nodes {
				if n.Identifier() == leftId {
					dfsReachable(nodes, n, level, visited)
					break
				}
			}
		}
	}

	// Check right neighbor
	if hasNeighbor(start, core.RightDirection, level) {
		rightNeighbor, _ := start.GetNeighbor(core.RightDirection, level)
		rightId := rightNeighbor.GetIdentifier()
		if !visited[rightId] {
			// Find the node with this identifier
			for _, n := range nodes {
				if n.Identifier() == rightId {
					dfsReachable(nodes, n, level, visited)
					break
				}
			}
		}
	}
}

// TestTraversalWithNodeReference tests traversal using (identifier, array_index) pairs
func TestTraversalWithNodeReference(t *testing.T) {
	logger := unittest.Logger(zerolog.InfoLevel)
	bootstrapper := NewBootstrapper(logger)

	result, err := bootstrapper.Bootstrap(10)
	require.NoError(t, err)

	// Create node references for testing
	nodeRefs := make([]NodeReference, len(result.Nodes))
	for i, n := range result.Nodes {
		nodeRefs[i] = NodeReference{
			Identifier: n.Identifier(),
			ArrayIndex: i,
		}
	}

	// Test traversal at level 0
	t.Run("TraverseLevel0", func(t *testing.T) {
		traversed := traverseLevel(result.Nodes, nodeRefs[0], core.Level(0))
		assert.Len(t, traversed, len(result.Nodes), "Should traverse all nodes at level 0")

		// Verify order
		for i := 1; i < len(traversed); i++ {
			idPrev := traversed[i-1].Identifier
			idCurr := traversed[i].Identifier
			comp := idPrev.Compare(&idCurr)
			assert.Equal(t, model.CompareLess, comp.GetComparisonResult(), "Nodes should be in sorted order")
		}
	})

	// Test traversal at higher levels
	t.Run("TraverseHigherLevels", func(t *testing.T) {
		for level := 1; level <= result.MaxLevel; level++ {
			// Find a node that has neighbors at this level
			var startRef NodeReference
			hasNeighborAtLevel := false
			for i, n := range result.Nodes {
				if hasNeighbor(n, core.RightDirection, core.Level(level)) {
					startRef = nodeRefs[i]
					hasNeighborAtLevel = true
					break
				}
			}

			if hasNeighborAtLevel {
				traversed := traverseLevel(result.Nodes, startRef, core.Level(level))
				assert.NotEmpty(t, traversed, "Should traverse at least one node at level %d", level)

				// Verify all traversed nodes have matching prefix
				startMV := result.Nodes[startRef.ArrayIndex].MembershipVector()
				prefix := getPrefixBits(startMV, level)
				for _, ref := range traversed {
					nodeMV := result.Nodes[ref.ArrayIndex].MembershipVector()
					nodePrefix := getPrefixBits(nodeMV, level)
					assert.Equal(t, prefix, nodePrefix,
						"All traversed nodes should have same prefix at level %d", level)
				}
			}
		}
	})
}

// traverseLevel traverses all connected nodes at a given level starting from a node reference
func traverseLevel(nodes []*node.SkipGraphNode, start NodeReference, level core.Level) []NodeReference {
	visited := make(map[model.Identifier]bool)
	result := []NodeReference{}

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
			leftId := leftNeighbor.GetIdentifier()
			// Find the array index of this neighbor
			found := false
			for i, other := range nodes {
				if other.Identifier() == leftId {
					current = NodeReference{
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
			rightId := rightNeighbor.GetIdentifier()
			if visited[rightId] {
				break // Avoid cycles
			}
			// Find the array index of this neighbor
			found := false
			for i, other := range nodes {
				if other.Identifier() == rightId {
					current = NodeReference{
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
	}

	return result
}

// BenchmarkBootstrap benchmarks bootstrap performance
func BenchmarkBootstrap(b *testing.B) {
	logger := unittest.Logger(zerolog.ErrorLevel)
	bootstrapper := NewBootstrapper(logger)

	sizes := []int{10, 100, 1000}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := bootstrapper.Bootstrap(size)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
