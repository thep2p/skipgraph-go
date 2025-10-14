package internal

import (
	"fmt"
	"github.com/thep2p/skipgraph-go/core"
	"github.com/thep2p/skipgraph-go/core/model"
	"sort"
)

// Entry is an internal structure used during bootstrap process
// It contains the Identity information and lookup table for a node being bootstrapped
type Entry struct {
	Identity    model.Identity
	LookupTable core.MutableLookupTable
}

// SortedEntryList is a list of entries sorted by identifier in ascending order
// It provides methods to add entries, get entries by index, and insert entries into the skip graph
type SortedEntryList struct {
	list []*Entry
}

func NewSortedEntryList() *SortedEntryList {
	return &SortedEntryList{
		list: make([]*Entry, 0),
	}
}

// Add adds an entry to the list and maintains sorted order.
func (e *SortedEntryList) Add(entry *Entry) {
	e.list = append(e.list, entry)
	e.sort()
}

// Get returns the entry at the specified index.
func (e *SortedEntryList) Get(index int) *Entry {
	return e.list[index]
}

// Len returns the number of entries in the list.
func (e *SortedEntryList) Len() int {
	return len(e.list)
}

// sort sorts the entries by identifier in ascending order.
func (e *SortedEntryList) sort() {
	sort.Slice(
		e.list, func(i, j int) bool {
			idI := e.list[i].Identity.GetIdentifier()
			idJ := e.list[j].Identity.GetIdentifier()
			comparison := idI.Compare(&idJ)
			return comparison.GetComparisonResult() == model.CompareLess
		},
	)
}

// InsertAll inserts all entries into the skip graph using the insertion algorithm.
// Returns a slice of Entry pointers representing the bootstrapped skip graph structure.
// Returns an error if any insertion fails; any error is fatal and indicates a serious bug in the bootstrap logic; crash if it occurs.
func (e *SortedEntryList) InsertAll() ([]*Entry, error) {
	for i := 0; i < e.Len(); i++ {
		if err := e.insert(i); err != nil {
			return nil, fmt.Errorf("failed to insert entry at index %d: %w", i, err)
		}
	}

	return e.list, nil
}

// Insert implements Algorithm 2 insert operation (ref. Skip Graph paper) for a single bootstrap entry
// Returns an error if insertion fails; any error is fatal and indicates a serious bug in the bootstrap logic; crash if it occurs.
func (e *SortedEntryList) insert(entryIndex int) error {
	entry := e.Get(entryIndex)
	// Start at level 0
	level := core.Level(0)

	// Link at level 0 (all entries are connected in sorted order)
	if err := e.linkLevel0(entryIndex); err != nil {
		// This should never happen; crash if it does
		return fmt.Errorf("failed to link level 0: %w", err)
	}

	// Process higher levels
	for level < core.MaxLookupTableLevel {
		level++

		// Find entries at this level with matching membership vector prefix
		leftNeighborIndex, leftNeighborExists := e.leftNeighborIndexAtLevel(entryIndex, int(level))
		rightNeighborIndex, rightNeighborExists := e.rightNeighborIndexAtLevel(entryIndex, int(level))

		// If no neighbors exist at this level, we are done
		if !leftNeighborExists && !rightNeighborExists {
			// Entry is in a singleton list at this level
			break
		}

		// Link with neighbors at this level
		if leftNeighborExists {
			leftEntry := e.Get(leftNeighborIndex) // left neighbor entry
			// Add left neighbor to this entry's lookup table
			if err := entry.LookupTable.AddEntry(core.LeftDirection, level, leftEntry.Identity); err != nil {
				return fmt.Errorf("failed to add left neighbor: %w", err)
			}

			// Update left neighbor's right pointer to this entry
			if err := leftEntry.LookupTable.AddEntry(core.RightDirection, level, entry.Identity); err != nil {
				return fmt.Errorf("failed to update left neighbor's right pointer: %w", err)
			}
		}

		if rightNeighborExists {
			rightEntry := e.Get(rightNeighborIndex) // right neighbor entry
			// Add right neighbor to this entry's lookup table
			if err := entry.LookupTable.AddEntry(core.RightDirection, level, rightEntry.Identity); err != nil {
				return fmt.Errorf("failed to add right neighbor: %w", err)
			}

			// Update right neighbor's left pointer to this entry
			if err := rightEntry.LookupTable.AddEntry(core.LeftDirection, level, entry.Identity); err != nil {
				return fmt.Errorf("failed to update right neighbor's left pointer: %w", err)
			}
		}
	}

	return nil
}

// linkLevel0 links an entry at level 0 with its immediate neighbors in sorted order.
// Any returned error is fatal and indicates a serious bug in the bootstrap logic; crash if it occurs.
func (e *SortedEntryList) linkLevel0(entryIndex int) error {
	level := core.Level(0)
	entry := e.Get(entryIndex)

	// Link with left neighbor; skip the first entry (no left neighbor)
	if entryIndex > 0 {
		leftEntry := e.Get(entryIndex - 1)
		if err := entry.LookupTable.AddEntry(core.LeftDirection, level, leftEntry.Identity); err != nil {
			return fmt.Errorf("failed to set left neighbor at level 0: %w", err)
		}
	}

	// Link with right neighbor; skip the last entry (no right neighbor)
	if entryIndex < e.Len()-1 {
		rightEntry := e.Get(entryIndex + 1)
		if err := entry.LookupTable.AddEntry(core.RightDirection, level, rightEntry.Identity); err != nil {
			return fmt.Errorf("failed to set right neighbor at level 0: %w", err)
		}
	}

	return nil
}

// leftNeighborIndexAtLevel finds the left neighbor of the entry at entryIndex at the given level.
func (e *SortedEntryList) leftNeighborIndexAtLevel(entryIndex int, level int) (int, bool) {
	entry := e.Get(entryIndex)
	entryMV := entry.Identity.GetMembershipVector()

	// Search left for the closest entry with matching prefix; looking at entries that are less than entryIndex
	// in their identifier; note that entries must be sorted by identifier in accending order.
	for i := entryIndex - 1; i >= 0; i-- {
		if entryMV.CommonPrefix(e.Get(i).Identity.GetMembershipVector()) >= level {
			return i, true
		}
	}

	return -1, false
}

func (e *SortedEntryList) rightNeighborIndexAtLevel(entryIndex int, level int) (int, bool) {
	entry := e.Get(entryIndex)
	entryMV := entry.Identity.GetMembershipVector()

	// Search right for the closest entry with matching prefix; looking at entries that are greater than entryIndex
	// in their identifier; note that entries must be sorted by identifier in accending order.
	for i := entryIndex + 1; i < e.Len(); i++ {
		if entryMV.CommonPrefix(e.Get(i).Identity.GetMembershipVector()) >= level {
			return i, true
		}
	}

	return -1, false
}
