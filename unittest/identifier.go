package unittest

import (
	"bytes"
	"github.com/thep2p/skipgraph-go/core"
	"github.com/thep2p/skipgraph-go/core/model"
	"testing"
)

// IdentifierGreaterThan returns an identifier greater than the given target.
// It increments the target identifier by 1 by finding the rightmost byte < 0xFF
// and incrementing it. If all bytes are 0xFF, it wraps around to all zeros.
func IdentifierGreaterThan(target model.Identifier) model.Identifier {
	byteSlice := make([]byte, model.IdentifierSizeBytes)
	copy(byteSlice, target.Bytes())

	// Increment from the right until we find a byte < 0xFF
	for i := len(byteSlice) - 1; i >= 0; i-- {
		if byteSlice[i] < 0xFF {
			byteSlice[i]++
			break
		}
		// If byte is 0xFF, set it to 0 and continue to next byte
		byteSlice[i] = 0
	}

	id, _ := model.ByteToId(byteSlice)
	return id
}

// IdentifierLessThan returns an identifier less than the given target.
// It decrements the target identifier by 1 by finding the rightmost byte > 0x00
// and decrementing it. If all bytes are 0x00, it wraps around to all 0xFF.
func IdentifierLessThan(target model.Identifier) model.Identifier {
	byteSlice := make([]byte, model.IdentifierSizeBytes)
	copy(byteSlice, target.Bytes())

	// Decrement from the right until we find a byte > 0x00
	for i := len(byteSlice) - 1; i >= 0; i-- {
		if byteSlice[i] > 0x00 {
			byteSlice[i]--
			break
		}
		// If byte is 0x00, set it to 0xFF and continue to next byte
		byteSlice[i] = 0xFF
	}

	id, _ := model.ByteToId(byteSlice)
	return id
}

// NeighborEntry represents a neighbor at a specific level in the lookup table.
type NeighborEntry struct {
	Level    core.Level
	Identity model.Identity
}

// LeftNeighbors returns all left neighbors from the lookup table as a slice of NeighborEntry.
// Returns an error if the lookup table access fails.
func LeftNeighbors(lt core.ImmutableLookupTable) ([]NeighborEntry, error) {
	var result []NeighborEntry
	for level := core.Level(0); level < core.MaxLookupTableLevel; level++ {
		identity, err := lt.GetEntry(core.LeftDirection, level)
		if err != nil {
			return nil, err
		}
		if identity != nil {
			result = append(result, NeighborEntry{
				Level:    level,
				Identity: *identity,
			})
		}
	}
	return result, nil
}

// RightNeighbors returns all right neighbors from the lookup table as a slice of NeighborEntry.
// Returns an error if the lookup table access fails.
func RightNeighbors(lt core.ImmutableLookupTable) ([]NeighborEntry, error) {
	var result []NeighborEntry
	for level := core.Level(0); level < core.MaxLookupTableLevel; level++ {
		identity, err := lt.GetEntry(core.RightDirection, level)
		if err != nil {
			return nil, err
		}
		if identity != nil {
			result = append(result, NeighborEntry{
				Level:    level,
				Identity: *identity,
			})
		}
	}
	return result, nil
}

// RandomLookupTableWithExtremes creates a lookup table populated with random neighbors
// at all levels, with extreme values (all zeros for left at level 0, all 0xFF for right at level 0).
// This is useful for testing edge cases.
func RandomLookupTableWithExtremes(t *testing.T) core.MutableLookupTable {
	lt := &mockLookupTable{
		leftNeighbors:  make(map[core.Level]model.Identity),
		rightNeighbors: make(map[core.Level]model.Identity),
	}

	// Add random neighbors at all levels
	for level := core.Level(0); level < core.MaxLookupTableLevel; level++ {
		leftIdentity := IdentityFixture(t)
		rightIdentity := IdentityFixture(t)
		_ = lt.AddEntry(core.LeftDirection, level, leftIdentity)
		_ = lt.AddEntry(core.RightDirection, level, rightIdentity)
	}

	// Add extreme values at level 0
	zeroBytes := make([]byte, model.IdentifierSizeBytes)
	zeroID, _ := model.ByteToId(zeroBytes)

	maxBytes := bytes.Repeat([]byte{0xFF}, model.IdentifierSizeBytes)
	maxID, _ := model.ByteToId(maxBytes)

	zeroIdentity := model.NewIdentity(zeroID, MembershipVectorFixture(t), AddressFixture(t))
	maxIdentity := model.NewIdentity(maxID, MembershipVectorFixture(t), AddressFixture(t))

	_ = lt.AddEntry(core.LeftDirection, 0, zeroIdentity)
	_ = lt.AddEntry(core.RightDirection, 0, maxIdentity)

	return lt
}

// mockLookupTable is a simple in-memory implementation of MutableLookupTable for testing.
type mockLookupTable struct {
	leftNeighbors  map[core.Level]model.Identity
	rightNeighbors map[core.Level]model.Identity
}

// GetEntry returns the neighbor at the given direction and level.
func (m *mockLookupTable) GetEntry(dir core.Direction, lev core.Level) (*model.Identity, error) {
	var identity model.Identity
	var exists bool

	switch dir {
	case core.LeftDirection:
		identity, exists = m.leftNeighbors[lev]
	case core.RightDirection:
		identity, exists = m.rightNeighbors[lev]
	}

	if !exists {
		return nil, nil
	}
	return &identity, nil
}

// AddEntry adds a neighbor at the given direction and level.
func (m *mockLookupTable) AddEntry(dir core.Direction, level core.Level, identity model.Identity) error {
	switch dir {
	case core.LeftDirection:
		m.leftNeighbors[level] = identity
	case core.RightDirection:
		m.rightNeighbors[level] = identity
	}
	return nil
}
