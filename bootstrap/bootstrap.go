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
	TotalLevels         int
	AverageNeighbors    float64
	ConnectedComponents map[int]int // level -> component count
}

// bootstrapEntry is an internal structure used during bootstrap process
// It contains the identity information and lookup table for a node being bootstrapped
type bootstrapEntry struct {
	identity model.Identity
	lt       core.MutableLookupTable
}

// Bootstrapper encapsulates all bootstrap logic for creating a skip graph with centralized insert.
// This ensures bootstrap logic is only used for bootstrapping and not borrowed for other purposes.
type Bootstrapper struct {
	logger zerolog.Logger
}

// NewBootstrapper creates a new Bootstrapper instance.
func NewBootstrapper(logger zerolog.Logger) *Bootstrapper {
	return &Bootstrapper{
		logger: logger.With().Str("component", "bootstrap").Logger(),
	}
}

// Bootstrap creates a skip graph with the specified number of nodes using centralized insert (Algorithm 2).
// Returns an array of nodes where each node's lookup table contains references to other nodes in the array.
func (b *Bootstrapper) Bootstrap(numNodes int) (*BootstrapResult, error) {
	if numNodes <= 0 {
		return nil, fmt.Errorf("number of nodes must be positive, got %d", numNodes)
	}

	b.logger.Info().Int("numNodes", numNodes).Msg("Starting bootstrap")

	// Create bootstrap entries with unique identifiers and random membership vectors
	entries, err := b.createBootstrapEntries(numNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to create bootstrap entries: %w", err)
	}

	// Sort entries by identifier for level 0
	b.sortEntriesByIdentifier(entries)

	// Insert each entry into the skip graph structure using Algorithm 2 of the Skip Graphs paper.
	maxLevel := 0
	for i, entry := range entries {
		level := b.insertEntry(entries, i, entry)
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Convert bootstrap entries to nodes
	nodes := b.createNodesFromEntries(entries)

	// Calculate statistics
	stats := b.calculateStats(nodes, maxLevel)

	b.logger.Info().
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

// createBootstrapEntries creates numNodes bootstrap entries with unique identifiers and random membership vectors
func (b *Bootstrapper) createBootstrapEntries(numNodes int) ([]*bootstrapEntry, error) {
	entries := make([]*bootstrapEntry, numNodes)
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

		// Create bootstrap entry
		entries[i] = &bootstrapEntry{
			identity: identity,
			lt:       lt,
		}

		b.logger.Debug().
			Int("index", i).
			Str("identifier", id.String()).
			Str("membershipVector", mv.String()).
			Msg("Created bootstrap entry")
	}

	return entries, nil
}

// sortEntriesByIdentifier sorts bootstrap entries in ascending order by identifier
func (b *Bootstrapper) sortEntriesByIdentifier(entries []*bootstrapEntry) {
	sort.Slice(
		entries, func(i, j int) bool {
			idI := entries[i].identity.GetIdentifier()
			idJ := entries[j].identity.GetIdentifier()
			comparison := idI.Compare(&idJ)
			return comparison.GetComparisonResult() == model.CompareLess
		},
	)
}

// insertEntry implements Algorithm 2 insert operation for a single bootstrap entry
// Returns the maximum level at which this entry has neighbors
func (b *Bootstrapper) insertEntry(entries []*bootstrapEntry, entryIndex int, entry *bootstrapEntry) int {
	entryId := entry.identity.GetIdentifier()
	logger := b.logger.With().
		Int("entryIndex", entryIndex).
		Str("identifier", entryId.String()).
		Logger()

	// Start at level 0
	level := core.Level(0)
	maxLevel := 0

	// Link at level 0 (all entries are connected in sorted order)
	b.linkLevel0(logger, entries, entryIndex, entry)

	// Process higher levels
	for level < core.MaxLookupTableLevel {
		level++

		// Find entries at this level with matching membership vector prefix
		leftNeighbor, rightNeighbor := b.findNeighborsAtLevel(entries, entryIndex, entry, int(level))

		if leftNeighbor == -1 && rightNeighbor == -1 {
			// Entry is in a singleton list at this level
			break
		}

		// Link with neighbors at this level
		if leftNeighbor != -1 {
			leftEntry := entries[leftNeighbor]
			// Add left neighbor to this entry's lookup table
			if err := entry.lt.AddEntry(core.LeftDirection, level, leftEntry.identity); err != nil {
				logger.Error().Err(err).Msg("Failed to add left neighbor")
			}

			// Update left neighbor's right pointer to this entry
			if err := leftEntry.lt.AddEntry(core.RightDirection, level, entry.identity); err != nil {
				logger.Error().Err(err).Msg("Failed to update left neighbor's right pointer")
			}
		}

		if rightNeighbor != -1 {
			rightEntry := entries[rightNeighbor]
			// Add right neighbor to this entry's lookup table
			if err := entry.lt.AddEntry(core.RightDirection, level, rightEntry.identity); err != nil {
				logger.Error().Err(err).Msg("Failed to add right neighbor")
			}

			// Update right neighbor's left pointer to this entry
			if err := rightEntry.lt.AddEntry(core.LeftDirection, level, entry.identity); err != nil {
				logger.Error().Err(err).Msg("Failed to update right neighbor's left pointer")
			}
		}

		maxLevel = int(level)
	}

	logger.Debug().Int("maxLevel", maxLevel).Msg("Entry inserted")
	return maxLevel
}

// linkLevel0 links an entry at level 0 with its immediate neighbors in sorted order
func (b *Bootstrapper) linkLevel0(logger zerolog.Logger, entries []*bootstrapEntry, entryIndex int, entry *bootstrapEntry) {
	level := core.Level(0)

	// Link with left neighbor
	if entryIndex > 0 {
		leftEntry := entries[entryIndex-1]
		if err := entry.lt.AddEntry(core.LeftDirection, level, leftEntry.identity); err != nil {
			logger.Error().Err(err).Msg("Failed to set left neighbor at level 0")
		}
	}

	// Link with right neighbor
	if entryIndex < len(entries)-1 {
		rightEntry := entries[entryIndex+1]
		if err := entry.lt.AddEntry(core.RightDirection, level, rightEntry.identity); err != nil {
			logger.Error().Err(err).Msg("Failed to set right neighbor at level 0")
		}
	}
}

// findNeighborsAtLevel finds the left and right neighbors for an entry at a specific level
// based on membership vector prefix matching
func (b *Bootstrapper) findNeighborsAtLevel(entries []*bootstrapEntry, entryIndex int, entry *bootstrapEntry, level int) (int, int) {
	leftNeighbor := -1
	rightNeighbor := -1

	entryMV := entry.identity.GetMembershipVector()

	// Search left for the closest entry with matching prefix
	for i := entryIndex - 1; i >= 0; i-- {
		if b.hasMatchingPrefix(entryMV, entries[i].identity.GetMembershipVector(), level) {
			leftNeighbor = i
			break
		}
	}

	// Search right for the closest entry with matching prefix
	for i := entryIndex + 1; i < len(entries); i++ {
		if b.hasMatchingPrefix(entryMV, entries[i].identity.GetMembershipVector(), level) {
			rightNeighbor = i
			break
		}
	}

	return leftNeighbor, rightNeighbor
}

// hasMatchingPrefix checks if two membership vectors have matching prefix up to the specified level (in bits)
func (b *Bootstrapper) hasMatchingPrefix(mv1, mv2 model.MembershipVector, level int) bool {
	commonPrefixLength := mv1.CommonPrefix(mv2)
	return commonPrefixLength >= level
}

// createNodesFromEntries converts bootstrap entries to SkipGraphNodes
func (b *Bootstrapper) createNodesFromEntries(entries []*bootstrapEntry) []*node.SkipGraphNode {
	nodes := make([]*node.SkipGraphNode, len(entries))
	for i, entry := range entries {
		nodes[i] = node.NewSkipGraphNode(entry.identity, entry.lt)
	}
	return nodes
}

// calculateStats calculates statistics about the bootstrapped skip graph
func (b *Bootstrapper) calculateStats(nodes []*node.SkipGraphNode, maxLevel int) BootstrapStats {
	totalNeighbors := 0
	connectedComponents := make(map[int]int)

	for level := 0; level <= maxLevel; level++ {
		components := b.countConnectedComponents(nodes, core.Level(level))
		connectedComponents[level] = components
	}

	// Count total neighbors across all nodes and levels
	for _, n := range nodes {
		for level := core.Level(0); level <= core.Level(maxLevel); level++ {
			leftNeighbor, err := n.GetNeighbor(core.LeftDirection, level)
			if err == nil && !b.isEmptyIdentity(leftNeighbor.GetIdentifier()) {
				totalNeighbors++
			}
			rightNeighbor, err := n.GetNeighbor(core.RightDirection, level)
			if err == nil && !b.isEmptyIdentity(rightNeighbor.GetIdentifier()) {
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

// isEmptyIdentity checks if an identifier is empty (all zeros)
func (b *Bootstrapper) isEmptyIdentity(id model.Identifier) bool {
	empty := model.Identifier{}
	return id == empty
}

// countConnectedComponents counts the number of connected components at a given level
func (b *Bootstrapper) countConnectedComponents(nodes []*node.SkipGraphNode, level core.Level) int {
	visited := make(map[int]bool)
	components := 0

	for i := range nodes {
		if !visited[i] {
			// Start a new component
			components++
			// DFS to mark all nodes in this component
			b.dfs(nodes, i, level, visited)
		}
	}

	return components
}

// dfs performs depth-first search to mark all nodes in a connected component
func (b *Bootstrapper) dfs(nodes []*node.SkipGraphNode, nodeIndex int, level core.Level, visited map[int]bool) {
	visited[nodeIndex] = true
	n := nodes[nodeIndex]

	// Check left neighbor
	if leftNeighbor, err := n.GetNeighbor(core.LeftDirection, level); err == nil {
		leftId := leftNeighbor.GetIdentifier()
		if !b.isEmptyIdentity(leftId) {
			// Find the index of this neighbor
			for i, other := range nodes {
				if other.Identifier() == leftId && !visited[i] {
					b.dfs(nodes, i, level, visited)
					break
				}
			}
		}
	}

	// Check right neighbor
	if rightNeighbor, err := n.GetNeighbor(core.RightDirection, level); err == nil {
		rightId := rightNeighbor.GetIdentifier()
		if !b.isEmptyIdentity(rightId) {
			// Find the index of this neighbor
			for i, other := range nodes {
				if other.Identifier() == rightId && !visited[i] {
					b.dfs(nodes, i, level, visited)
					break
				}
			}
		}
	}
}
