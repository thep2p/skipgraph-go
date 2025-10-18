package bootstrap

import (
	"crypto/rand"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/thep2p/skipgraph-go/bootstrap/internal"
	"github.com/thep2p/skipgraph-go/core"
	"github.com/thep2p/skipgraph-go/core/lookup"
	"github.com/thep2p/skipgraph-go/core/model"
	"github.com/thep2p/skipgraph-go/node"
)

const (
	// DefaultSkipGraphPort is the default port for Skip Graph nodes.
	// In bootstrap context, this is used as a placeholder since actual network
	// communication doesn't occur during the bootstrap phase.
	DefaultSkipGraphPort = "5555"

	// maxIdentifierGenerationRetries is the maximum number of attempts to generate
	// a unique identifier or membership vector before returning an error.
	// This prevents infinite loops in the unlikely event of hash collisions.
	//
	// The value 1000 is set as a defensive upper bound. With 256-bit identifiers
	// and membership vectors, collision probability is negligible (~10^-71 for 1000 nodes
	// based on birthday paradox calculations). This limit ensures guaranteed termination
	// without impacting normal operation, as collisions should never occur in practice.
	maxIdentifierGenerationRetries = 1000
)

// Bootstrapper encapsulates all bootstrap logic for creating a skip graph with centralized insert.
// This ensures bootstrap logic is only used for bootstrapping and not borrowed for other purposes.
type Bootstrapper struct {
	logger   zerolog.Logger
	numNodes int // number of nodes to bootstrap
}

// NewBootstrapper creates a new Bootstrapper instance.
func NewBootstrapper(logger zerolog.Logger, numNodes int) *Bootstrapper {
	return &Bootstrapper{
		logger:   logger.With().Str("component", "bootstrap").Logger(),
		numNodes: numNodes,
	}
}

// Bootstrap creates a skip graph with the specified number of nodes using centralized insert (Algorithm 2).
// Returns an array of nodes where each node's lookup table contains references to other nodes in the array.
func (b *Bootstrapper) Bootstrap() ([]*node.SkipGraphNode, error) {
	if b.numNodes <= 0 {
		return nil, fmt.Errorf("number of nodes must be positive, got %d", b.numNodes)
	}

	lg := b.logger.With().Int("numNodes", b.numNodes).Logger()
	lg.Info().Msg("bootstrapping skip graph started")

	// Create bootstrap entries with unique identifiers and random membership vectors
	entries, err := b.createBootstrapEntries()
	if err != nil {
		return nil, fmt.Errorf("failed to create bootstrap entries: %w", err)
	}

	// Insert each entry into the skip graph structure using Algorithm 2 of the Skip Graphs paper.
	nodes, err := entries.InsertAll()
	if err != nil {
		return nil, fmt.Errorf("failed to insert entries into skip graph: %w", err)
	}

	b.logger.Info().
		Int("nodes", len(nodes)).
		Msg("bootstrap completed")

	return nodes, nil
}

// createBootstrapEntries creates numNodes bootstrap entries with unique identifiers and random membership vectors
func (b *Bootstrapper) createBootstrapEntries() (*internal.SortedEntryList, error) {
	entries := internal.NewSortedEntryList()
	identifierSet := make(map[model.Identifier]bool)
	membershipVectorSet := make(map[model.MembershipVector]bool)

	for i := 0; i < b.numNodes; i++ {
		// Generate unique identifier
		var id model.Identifier
		var generated bool
		for attempt := 0; attempt < maxIdentifierGenerationRetries; attempt++ {
			if _, err := rand.Read(id[:]); err != nil {
				return nil, fmt.Errorf("failed to generate identifier: %w", err)
			}
			if !identifierSet[id] {
				identifierSet[id] = true
				generated = true
				break
			}
		}
		if !generated {
			return nil, fmt.Errorf("failed to generate unique identifier after %d attempts for node %d", maxIdentifierGenerationRetries, i)
		}

		// Generate unique membership vector
		// Design Decision: While Skip Graph theory doesn't strictly require unique membership vectors,
		// this implementation enforces uniqueness to guarantee better performance characteristics.
		// Non-unique membership vectors can lead to unbalanced skip graph structures and degraded
		// search performance. With 256-bit vectors, enforcing uniqueness is practical and provides
		// stronger structural guarantees without meaningful overhead.
		var mv model.MembershipVector
		generated = false
		for attempt := 0; attempt < maxIdentifierGenerationRetries; attempt++ {
			if _, err := rand.Read(mv[:]); err != nil {
				return nil, fmt.Errorf("failed to generate membership vector: %w", err)
			}
			if !membershipVectorSet[mv] {
				membershipVectorSet[mv] = true
				generated = true
				break
			}
		}
		if !generated {
			return nil, fmt.Errorf("failed to generate unique membership vector after %d attempts for node %d", maxIdentifierGenerationRetries, i)
		}

		// Create Identity with placeholder address (not used in bootstrap)
		// Using the default port since actual network communication doesn't occur during bootstrap
		addr := model.NewAddress("localhost", DefaultSkipGraphPort)
		identity := model.NewIdentity(id, mv, addr)

		// Create lookup table
		lt := &lookup.Table{}

		// Create bootstrap entry
		entries.Add(
			&internal.Entry{
				Identity:    identity,
				LookupTable: lt,
			},
		)

		b.logger.Debug().
			Int("index", i).
			Str("identifier", id.String()).
			Str("membershipVector", mv.String()).
			Msg("created bootstrap entry")
	}

	return entries, nil
}

// TraverseConnectedNodes performs a depth-first traversal of connected nodes at a given level.
// It starts from the specified node and marks all reachable nodes as visited.
// The idToIndex map provides O(1) lookup from identifier to node index.
// This is a reusable DFS function used by both CountConnectedComponents and test utilities.
func (b *Bootstrapper) TraverseConnectedNodes(
	nodes []*node.SkipGraphNode,
	startIndex int,
	level core.Level,
	visited map[int]bool,
	idToIndex map[model.Identifier]int,
) {
	visited[startIndex] = true
	currentNode := nodes[startIndex]

	// Helper function to visit a neighbor
	visitNeighbor := func(neighbor *model.Identity) {
		if neighbor != nil {
			neighborId := neighbor.GetIdentifier()
			if neighborIndex, exists := idToIndex[neighborId]; exists && !visited[neighborIndex] {
				b.TraverseConnectedNodes(nodes, neighborIndex, level, visited, idToIndex)
			}
		}
	}

	// Check left neighbor
	if leftNeighbor, err := currentNode.GetNeighbor(core.LeftDirection, level); err == nil {
		visitNeighbor(leftNeighbor)
	}

	// Check right neighbor
	if rightNeighbor, err := currentNode.GetNeighbor(core.RightDirection, level); err == nil {
		visitNeighbor(rightNeighbor)
	}
}

// CountConnectedComponents counts the number of connected components at a given level.
// This is useful for verifying skip graph properties during testing.
func (b *Bootstrapper) CountConnectedComponents(nodes []*node.SkipGraphNode, level core.Level) int {
	// Create identifier to index map for O(1) lookups
	idToIndex := make(map[model.Identifier]int)
	for i, n := range nodes {
		idToIndex[n.Identifier()] = i
	}

	visited := make(map[int]bool)
	components := 0

	for i := range nodes {
		if !visited[i] {
			// Start a new component
			components++
			// DFS to mark all nodes in this component
			b.TraverseConnectedNodes(nodes, i, level, visited, idToIndex)
		}
	}

	return components
}
