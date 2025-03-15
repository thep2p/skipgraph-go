package skipgraph

import (
	"fmt"
	"sync"
)

// Level is the type for the level of entries in the lookup table.
type Level int64

// MaxLookupTableLevel indicates the upper bound for the number of levels in a SkipGraph LookupTable.
const MaxLookupTableLevel Level = IdentifierSizeBytes * 8

// Direction is an enum type for the direction of a neighbor in the lookup table.
type Direction string

const (
	// RightDirection	indicates the right direction in the lookup table.
	RightDirection = Direction("right")
	// LeftDirection	indicates the left direction in the lookup table.
	LeftDirection = Direction("left")
)

// LookupTable corresponds to a SkipGraph node's lookup table.
type LookupTable struct {
	lock           sync.RWMutex // used to lock the lookup table for read and write
	rightNeighbors [MaxLookupTableLevel]Identity
	leftNeighbors  [MaxLookupTableLevel]Identity
}

// AddEntry inserts the supplied Identity in the lth level of lookup table either as the left or right neighbor depending on the dir.
// lev runs from 0...MaxLookupTableLevel-1.
func (l *LookupTable) AddEntry(dir Direction, level Level, identity Identity) error {
	// lock the lookup table for write access
	l.lock.Lock()
	// unlock the lookup table at the end
	defer l.lock.Unlock()

	// validate the level value
	if level >= MaxLookupTableLevel {
		return fmt.Errorf("position is larger than the max lookup table entry number: %d", level)
	}

	switch dir {
	case RightDirection:
		l.rightNeighbors[level] = identity
	case LeftDirection:
		l.leftNeighbors[level] = identity
	default:
		return fmt.Errorf("invalid direction: %s", dir)
	}

	return nil
}

// GetEntry returns the lth left/right neighbor in the lookup table depending on the dir.
// lev runs from 0...MaxLookupTableLevel-1.
func (l *LookupTable) GetEntry(dir Direction, lev Level) (Identity, error) {
	// lock the lookup table for read only
	l.lock.RLock()
	// release the read-only lock at the end
	defer l.lock.RUnlock()

	res := Identity{}

	// validate the level value
	if lev >= MaxLookupTableLevel {
		return res, fmt.Errorf("supplied level is larger than the max number of levels: %d", lev)
	}
	switch dir {
	case RightDirection:
		res = l.rightNeighbors[lev]
	case LeftDirection:
		res = l.leftNeighbors[lev]
	default:
		return res, fmt.Errorf("invalid direction: %s", dir)
	}
	return res, nil
}
