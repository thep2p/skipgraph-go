package bootstrap

import (
	"crypto/rand"
	"fmt"
	"sort"

	"github.com/rs/zerolog"
	"github.com/thep2p/skipgraph-go/core"
	"github.com/thep2p/skipgraph-go/core/lookup"
	"github.com/thep2p/skipgraph-go/core/model"
	"github.com/thep2p/skipgraph-go/node"
)

// isEmptyIdentity checks if an identifier is empty (all zeros)
func isEmptyIdentity(id model.Identifier) bool {
	empty := model.Identifier{}
	return id == empty
}

// NodeReference represents a reference to a node in the array using both identifier and array index.
// This is used for testing purposes to enable graph traversal validation.
type NodeReference struct {
	Identifier model.Identifier
	ArrayIndex int
}

// BootstrapResult contains the result of bootstrap operation
type BootstrapResult struct {
	Nodes    []*node.SkipGraphNode
	MaxLevel int
	Stats    BootstrapStats
}

// BootstrapStats contains statistics about the bootstrapped skip graph
type BootstrapStats struct {
	TotalLevels          int
	AverageNeighbors     float64
	ConnectedComponents  map[int]int // level -> component count
}

// Bootstrap creates a skip graph with the specified number of nodes using centralized insert (Algorithm 2).
// Returns an array of nodes where each node's lookup table contains references to other nodes in the array.
func Bootstrap(logger zerolog.Logger, numNodes int) (*BootstrapResult, error) {
	logger = logger.With().Str("component", "bootstrap").Logger()

	if numNodes <= 0 {
		return nil, fmt.Errorf("number of nodes must be positive, got %d", numNodes)
	}

	logger.Info().Int("numNodes", numNodes).Msg("Starting bootstrap")

	// Create nodes with unique identifiers and random membership vectors
	nodes, err := createNodes(logger, numNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to create nodes: %w", err)
	}

	// Sort nodes by identifier for level 0
	sortNodesByIdentifier(nodes)

	// Insert each node into the skip graph structure using Algorithm 2
	maxLevel := 0
	for i, n := range nodes {
		level := insertNode(logger, nodes, i, n)
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Calculate statistics
	stats := calculateStats(nodes, maxLevel)

	logger.Info().
		Int("nodes", len(nodes)).
		Int("maxLevel", maxLevel).
		Float64("avgNeighbors", stats.AverageNeighbors).
		Msg("Bootstrap completed")

	return &BootstrapResult{
		Nodes:    nodes,
		MaxLevel: maxLevel,
		Stats:    stats,
	}, nil
}

// createNodes creates numNodes nodes with unique identifiers and random membership vectors
func createNodes(logger zerolog.Logger, numNodes int) ([]*node.SkipGraphNode, error) {
	nodes := make([]*node.SkipGraphNode, numNodes)
	identifierSet := make(map[model.Identifier]bool)

	for i := 0; i < numNodes; i++ {
		// Generate unique identifier
		var id model.Identifier
		for {
			if _, err := rand.Read(id[:]); err != nil {
				return nil, fmt.Errorf("failed to generate identifier: %w", err)
			}
			if !identifierSet[id] {
				identifierSet[id] = true
				break
			}
		}

		// Generate random membership vector
		var mv model.MembershipVector
		if _, err := rand.Read(mv[:]); err != nil {
			return nil, fmt.Errorf("failed to generate membership vector: %w", err)
		}

		// Create identity with dummy address (not used in bootstrap)
		addr := model.NewAddress("localhost", fmt.Sprintf("800%d", i))
		identity := model.NewIdentity(id, mv, addr)

		// Create lookup table
		lt := &lookup.Table{}

		// Create node
		nodes[i] = node.NewSkipGraphNode(identity, lt)

		logger.Debug().
			Int("index", i).
			Str("identifier", id.String()).
			Str("membershipVector", mv.String()).
			Msg("Created node")
	}

	return nodes, nil
}

// sortNodesByIdentifier sorts nodes in ascending order by identifier
func sortNodesByIdentifier(nodes []*node.SkipGraphNode) {
	sort.Slice(nodes, func(i, j int) bool {
		idI := nodes[i].Identifier()
		idJ := nodes[j].Identifier()
		comparison := idI.Compare(&idJ)
		return comparison.GetComparisonResult() == model.CompareLess
	})
}

// insertNode implements Algorithm 2 insert operation for a single node
// Returns the maximum level at which this node has neighbors
func insertNode(logger zerolog.Logger, nodes []*node.SkipGraphNode, nodeIndex int, n *node.SkipGraphNode) int {
	nodeId := n.Identifier()
	logger = logger.With().
		Int("nodeIndex", nodeIndex).
		Str("identifier", nodeId.String()).
		Logger()

	// Start at level 0
	level := core.Level(0)
	maxLevel := 0

	// Link at level 0 (all nodes are connected in sorted order)
	linkLevel0(logger, nodes, nodeIndex, n)

	// Process higher levels
	for level < core.MaxLookupTableLevel {
		level++

		// Find nodes at this level with matching membership vector prefix
		leftNeighbor, rightNeighbor := findNeighborsAtLevel(nodes, nodeIndex, n, int(level))

		if leftNeighbor == -1 && rightNeighbor == -1 {
			// Node is in a singleton list at this level
			break
		}

		// Link with neighbors at this level
		if leftNeighbor != -1 {
			leftNode := nodes[leftNeighbor]
			leftIdentity := model.NewIdentity(
				leftNode.Identifier(),
				leftNode.MembershipVector(),
				model.NewAddress("localhost", fmt.Sprintf("800%d", leftNeighbor)),
			)
			// Add left neighbor to this node's lookup table
			if err := n.SetNeighbor(core.LeftDirection, level, leftIdentity); err != nil {
				logger.Error().Err(err).Msg("Failed to add left neighbor")
			}

			// Update left neighbor's right pointer to this node
			nodeIdentity := model.NewIdentity(
				n.Identifier(),
				n.MembershipVector(),
				model.NewAddress("localhost", fmt.Sprintf("800%d", nodeIndex)),
			)
			if err := leftNode.SetNeighbor(core.RightDirection, level, nodeIdentity); err != nil {
				logger.Error().Err(err).Msg("Failed to update left neighbor's right pointer")
			}
		}

		if rightNeighbor != -1 {
			rightNode := nodes[rightNeighbor]
			rightIdentity := model.NewIdentity(
				rightNode.Identifier(),
				rightNode.MembershipVector(),
				model.NewAddress("localhost", fmt.Sprintf("800%d", rightNeighbor)),
			)
			// Add right neighbor to this node's lookup table
			if err := n.SetNeighbor(core.RightDirection, level, rightIdentity); err != nil {
				logger.Error().Err(err).Msg("Failed to add right neighbor")
			}

			// Update right neighbor's left pointer to this node
			nodeIdentity := model.NewIdentity(
				n.Identifier(),
				n.MembershipVector(),
				model.NewAddress("localhost", fmt.Sprintf("800%d", nodeIndex)),
			)
			if err := rightNode.SetNeighbor(core.LeftDirection, level, nodeIdentity); err != nil {
				logger.Error().Err(err).Msg("Failed to update right neighbor's left pointer")
			}
		}

		maxLevel = int(level)
	}

	logger.Debug().Int("maxLevel", maxLevel).Msg("Node inserted")
	return maxLevel
}

// linkLevel0 links a node at level 0 with its immediate neighbors in sorted order
func linkLevel0(logger zerolog.Logger, nodes []*node.SkipGraphNode, nodeIndex int, n *node.SkipGraphNode) {
	level := core.Level(0)

	// Link with left neighbor
	if nodeIndex > 0 {
		leftNode := nodes[nodeIndex-1]
		leftIdentity := model.NewIdentity(
			leftNode.Identifier(),
			leftNode.MembershipVector(),
			model.NewAddress("localhost", fmt.Sprintf("800%d", nodeIndex-1)),
		)
		if err := n.SetNeighbor(core.LeftDirection, level, leftIdentity); err != nil {
			logger.Error().Err(err).Msg("Failed to set left neighbor at level 0")
		}
	}

	// Link with right neighbor
	if nodeIndex < len(nodes)-1 {
		rightNode := nodes[nodeIndex+1]
		rightIdentity := model.NewIdentity(
			rightNode.Identifier(),
			rightNode.MembershipVector(),
			model.NewAddress("localhost", fmt.Sprintf("800%d", nodeIndex+1)),
		)
		if err := n.SetNeighbor(core.RightDirection, level, rightIdentity); err != nil {
			logger.Error().Err(err).Msg("Failed to set right neighbor at level 0")
		}
	}
}

// findNeighborsAtLevel finds the left and right neighbors for a node at a specific level
// based on membership vector prefix matching
func findNeighborsAtLevel(nodes []*node.SkipGraphNode, nodeIndex int, n *node.SkipGraphNode, level int) (int, int) {
	leftNeighbor := -1
	rightNeighbor := -1

	nodeMV := n.MembershipVector()

	// Search left for the closest node with matching prefix
	for i := nodeIndex - 1; i >= 0; i-- {
		if hasMatchingPrefix(nodeMV, nodes[i].MembershipVector(), level) {
			leftNeighbor = i
			break
		}
	}

	// Search right for the closest node with matching prefix
	for i := nodeIndex + 1; i < len(nodes); i++ {
		if hasMatchingPrefix(nodeMV, nodes[i].MembershipVector(), level) {
			rightNeighbor = i
			break
		}
	}

	return leftNeighbor, rightNeighbor
}

// hasMatchingPrefix checks if two membership vectors have matching prefix up to the specified level (in bits)
func hasMatchingPrefix(mv1, mv2 model.MembershipVector, level int) bool {
	commonPrefixLength := mv1.CommonPrefix(mv2)
	return commonPrefixLength >= level
}

// calculateStats calculates statistics about the bootstrapped skip graph
func calculateStats(nodes []*node.SkipGraphNode, maxLevel int) BootstrapStats {
	totalNeighbors := 0
	connectedComponents := make(map[int]int)

	for level := 0; level <= maxLevel; level++ {
		components := countConnectedComponents(nodes, core.Level(level))
		connectedComponents[level] = components
	}

	// Count total neighbors across all nodes and levels
	for _, n := range nodes {
		for level := core.Level(0); level <= core.Level(maxLevel); level++ {
			leftNeighbor, err := n.GetNeighbor(core.LeftDirection, level)
			if err == nil && !isEmptyIdentity(leftNeighbor.GetIdentifier()) {
				totalNeighbors++
			}
			rightNeighbor, err := n.GetNeighbor(core.RightDirection, level)
			if err == nil && !isEmptyIdentity(rightNeighbor.GetIdentifier()) {
				totalNeighbors++
			}
		}
	}

	avgNeighbors := float64(totalNeighbors) / float64(len(nodes))

	return BootstrapStats{
		TotalLevels:         maxLevel + 1,
		AverageNeighbors:    avgNeighbors,
		ConnectedComponents: connectedComponents,
	}
}

// countConnectedComponents counts the number of connected components at a given level
func countConnectedComponents(nodes []*node.SkipGraphNode, level core.Level) int {
	visited := make(map[int]bool)
	components := 0

	for i := range nodes {
		if !visited[i] {
			// Start a new component
			components++
			// DFS to mark all nodes in this component
			dfs(nodes, i, level, visited)
		}
	}

	return components
}

// dfs performs depth-first search to mark all nodes in a connected component
func dfs(nodes []*node.SkipGraphNode, nodeIndex int, level core.Level, visited map[int]bool) {
	visited[nodeIndex] = true
	n := nodes[nodeIndex]

	// Check left neighbor
	if leftNeighbor, err := n.GetNeighbor(core.LeftDirection, level); err == nil {
		leftId := leftNeighbor.GetIdentifier()
		if !isEmptyIdentity(leftId) {
			// Find the index of this neighbor
			for i, other := range nodes {
				if other.Identifier() == leftId && !visited[i] {
					dfs(nodes, i, level, visited)
					break
				}
			}
		}
	}

	// Check right neighbor
	if rightNeighbor, err := n.GetNeighbor(core.RightDirection, level); err == nil {
		rightId := rightNeighbor.GetIdentifier()
		if !isEmptyIdentity(rightId) {
			// Find the index of this neighbor
			for i, other := range nodes {
				if other.Identifier() == rightId && !visited[i] {
					dfs(nodes, i, level, visited)
					break
				}
			}
		}
	}
}