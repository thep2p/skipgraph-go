package lookup

import (
	"fmt"
	"github.com/thep2p/skipgraph-go/core"
	"github.com/thep2p/skipgraph-go/core/model"
	"sync"
)

// Table corresponds to a SkipGraph node's lookup table.
type Table struct {
	lock           sync.RWMutex // used to lock the lookup table for read and write
	rightNeighbors [core.MaxLookupTableLevel]model.Identity
	leftNeighbors  [core.MaxLookupTableLevel]model.Identity
}

// AddEntry inserts the supplied Identity in the lth level of lookup table either as the left or right neighbor depending on the dir.
// lev runs from 0...MaxLookupTableLevel-1.
func (l *Table) AddEntry(dir core.Direction, level core.Level, identity model.Identity) error {
	// lock the lookup table for write access
	l.lock.Lock()
	// unlock the lookup table at the end
	defer l.lock.Unlock()

	// validate the level value
	if level >= core.MaxLookupTableLevel {
		return fmt.Errorf("level %d exceeds maximum valid level %d", level, core.MaxLookupTableLevel-1)
	}

	switch dir {
	case core.RightDirection:
		l.rightNeighbors[level] = identity
	case core.LeftDirection:
		l.leftNeighbors[level] = identity
	default:
		return fmt.Errorf("invalid direction: %s", dir)
	}

	return nil
}

// GetEntry returns the lth left/right neighbor in the lookup table depending on the dir.
// Returns nil if no neighbor exists at that position.
// lev runs from 0...MaxLookupTableLevel-1.
func (l *Table) GetEntry(dir core.Direction, lev core.Level) (*model.Identity, error) {
	// lock the lookup table for read only
	l.lock.RLock()
	// release the read-only lock at the end
	defer l.lock.RUnlock()

	// validate the level value
	if lev >= core.MaxLookupTableLevel {
		return nil, fmt.Errorf("level %d exceeds maximum valid level %d", lev, core.MaxLookupTableLevel-1)
	}

	var res model.Identity
	switch dir {
	case core.RightDirection:
		res = l.rightNeighbors[lev]
	case core.LeftDirection:
		res = l.leftNeighbors[lev]
	default:
		return nil, fmt.Errorf("invalid direction: %s", dir)
	}

	// Check if the identity is empty (all zeros)
	empty := model.Identity{}
	if res == empty {
		return nil, nil
	}

	return &res, nil
}
