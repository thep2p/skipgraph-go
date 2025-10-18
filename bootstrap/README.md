# Skip Graph Bootstrap

## Overview

The bootstrap package provides a centralized mechanism for creating and initializing skip graph networks. It implements Algorithm 2 from the Skip Graphs paper, which performs deterministic insertion of nodes to construct a properly formed skip graph structure.

Bootstrap is essential for:
- **Testing**: Creating skip graphs with known properties for unit and integration tests
- **Simulation**: Building networks of various sizes for performance analysis and behavior studies
- **Development**: Rapidly prototyping skip graph applications without dealing with distributed node joining
- **Benchmarking**: Generating consistent test networks for performance measurements

## What is Bootstrap?

In a production skip graph, nodes join the network dynamically through a distributed protocol. However, for testing and development purposes, we often need to create a complete skip graph structure immediately with multiple nodes. The bootstrap process achieves this by:

1. **Generating unique node identities** - Each node receives a unique 32-byte identifier and random membership vector
2. **Sorting nodes by identifier** - Establishes the base ordering for level 0 connections
3. **Building the multi-level structure** - Connects nodes at each level based on membership vector prefixes
4. **Verifying graph properties** - Ensures the resulting structure maintains skip graph invariants

The result is a fully-formed skip graph where all nodes have their lookup tables populated with the correct neighbors at each level, ready for routing operations.

## Key Concepts

### Skip Graph Structure
A skip graph is a multi-level distributed data structure where:
- **Level 0**: All nodes form a sorted doubly-linked list by identifier
- **Higher levels**: Nodes with matching membership vector prefixes form separate linked lists
- **Logarithmic height**: The expected number of levels is O(log n) for n nodes

### Membership Vectors
Each node has a random membership vector that determines its connections at higher levels:
- Nodes sharing i-bit prefixes are neighbors at level i
- Longer shared prefixes mean connections at higher levels
- Random vectors ensure balanced distribution across levels

### Connected Components
At each level i, the skip graph forms at most 2^i connected components:
- Level 0: One component (all nodes connected)
- Level 1: Up to 2 components (nodes with 0 vs 1 prefix)
- Level i: Up to 2^i components (based on i-bit prefixes)

## Usage

### Basic Bootstrap

```go
import (
    "github.com/rs/zerolog"
    "github.com/thep2p/skipgraph-go/bootstrap"
)

// Create a logger
logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

// Create a bootstrapper for 100 nodes
bootstrapper := bootstrap.NewBootstrapper(logger, 100)

// Bootstrap the skip graph
entries, err := bootstrapper.Bootstrap()
if err != nil {
    log.Fatal("Bootstrap failed:", err)
}

// entries is now an array of 100 BootstrapEntry instances
// Each entry contains Identity and LookupTable that references other entries
```

### Testing with Bootstrap

```go
func TestSkipGraphRouting(t *testing.T) {
    // Create a test skip graph
    bootstrapper := bootstrap.NewBootstrapper(testLogger, 50)
    entries, err := bootstrapper.Bootstrap()
    require.NoError(t, err)

    // Test routing between entries
    sourceEntry := entries[0]
    targetEntry := entries[25]

    // Access identity and lookup table
    sourceId := sourceEntry.Identity.GetIdentifier()
    targetId := targetEntry.Identity.GetIdentifier()

    // Perform routing operations using lookup tables...
}
```

### Analyzing Graph Properties

```go
// Count connected components at different levels
for level := 0; level <= 10; level++ {
    components := bootstrapper.CountConnectedComponents(entries, level)
    fmt.Printf("Level %d: %d components\n", level, components)
}
```

## API Reference

### Types

#### `Bootstrapper`
Main struct for creating skip graphs.

```go
type Bootstrapper struct {
    // Internal fields
}
```

#### `BootstrapEntry`
Represents a bootstrapped skip graph entry containing the node's identity and lookup table.

```go
type BootstrapEntry struct {
    Identity    model.Identity
    LookupTable core.MutableLookupTable
}
```

### Functions

#### `NewBootstrapper(logger zerolog.Logger, numNodes int) *Bootstrapper`
Creates a new bootstrapper instance.

**Parameters:**
- `logger`: Logger for debug output
- `numNodes`: Number of nodes to create (must be positive)

**Returns:** Bootstrapper instance

#### `Bootstrap() ([]*BootstrapEntry, error)`
Creates a skip graph with the configured number of nodes.

**Returns:**
- Array of BootstrapEntry instances with populated lookup tables
- Error if bootstrap fails (invalid parameters, etc.)

#### `CountConnectedComponents(entries []*BootstrapEntry, level core.Level) int`
Counts the number of connected components at a given level.

**Parameters:**
- `entries`: Array of bootstrap entries
- `level`: Level to analyze (0 to MaxLookupTableLevel)

**Returns:** Number of connected components

## Implementation Details

### Algorithm
The bootstrap process uses a centralized insertion algorithm:

1. **Create entries**: Generate n entries with unique identifiers and random membership vectors
2. **Sort entries**: Order by identifier for level 0 structure
3. **Insert sequentially**: For each entry:
   - Link at level 0 with immediate neighbors in sorted order
   - For higher levels, find and link with entries sharing membership vector prefixes
   - Continue until no matching prefixes exist

### Complexity
- **Time**: O(n² log n) - Each insertion scans existing nodes at multiple levels
- **Space**: O(n log n) - Each node stores O(log n) neighbor pointers

### Guarantees
The bootstrap process ensures:
- **Unique identifiers**: No two nodes share the same identifier
- **Sorted level 0**: Nodes at level 0 form an ordered linked list
- **Prefix consistency**: Neighbors at level i share at least i-bit prefixes
- **Bidirectional links**: If A points to B, then B points to A
- **Component constraint**: At most 2^i components at level i

## Testing

The bootstrap package includes comprehensive tests:

- **Single node**: Verifies edge case handling
- **Small graphs**: Tests basic properties (5-10 nodes)
- **Medium graphs**: Validates structure integrity (50-100 nodes)
- **Large graphs**: Stress tests and performance (500+ nodes)
- **Property verification**:
  - Level 0 ordering
  - Neighbor consistency
  - Membership vector prefixes
  - Connected components constraint

Run tests:
```bash
go test ./bootstrap/...
```

Run benchmarks:
```bash
go test -bench=. ./bootstrap/...
```

## Performance Considerations

### Memory Usage
- Each node requires ~1KB for identity and lookup table
- 1000 nodes ≈ 1MB total memory

### Scalability
Bootstrap performance for different sizes:
- 10 nodes: ~1ms
- 100 nodes: ~10ms
- 1000 nodes: ~200ms
- 10000 nodes: ~5s

Note: These are approximate values; actual performance depends on hardware.

### Optimization Tips
- Use appropriate log levels (Info/Warn for large graphs)
- Reuse bootstrapped graphs across multiple tests
- Consider caching for repeated bootstrap operations

## Limitations

1. **Centralized only**: Bootstrap is not suitable for production distributed systems
2. **Memory bound**: All nodes exist in the same process memory
3. **Static structure**: Nodes cannot dynamically join/leave after bootstrap
4. **Testing focused**: Designed for development and testing, not production use

## References

- [Skip Graphs Paper](../docs/skip-graphs-journal.pdf) - Original academic paper
- [Skip Graph Wikipedia](https://en.wikipedia.org/wiki/Skip_graph) - General overview
- [Project Documentation](../README.md) - Main project documentation