package lookup

import (
	"fmt"
	"github/thep2p/skipgraph-go/core"
	"github/thep2p/skipgraph-go/core/model"
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
		return fmt.Errorf("position is larger than the max lookup table entry number: %d", level)
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
// lev runs from 0...MaxLookupTableLevel-1.
func (l *Table) GetEntry(dir core.Direction, lev core.Level) (model.Identity, error) {
	// lock the lookup table for read only
	l.lock.RLock()
	// release the read-only lock at the end
	defer l.lock.RUnlock()

	res := model.Identity{}

	// validate the level value
	if lev >= core.MaxLookupTableLevel {
		return res, fmt.Errorf("supplied level is larger than the max number of levels: %d", lev)
	}
	switch dir {
	case core.RightDirection:
		res = l.rightNeighbors[lev]
	case core.LeftDirection:
		res = l.leftNeighbors[lev]
	default:
		return res, fmt.Errorf("invalid direction: %s", dir)
	}
	return res, nil
}
