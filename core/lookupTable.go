package core

import "github/thep2p/skipgraph-go/core/model"

// Level is the type for the level of entries in the lookup table.
type Level int64

// MaxLookupTableLevel indicates the upper bound for the number of levels in a SkipGraph LookupTable.
const MaxLookupTableLevel Level = model.IdentifierSizeBytes * 8

// Direction is an enum type for the direction of a neighbor in the lookup table.
type Direction string

const (
	// RightDirection	indicates the right direction in the lookup table.
	RightDirection = Direction("right")
	// LeftDirection	indicates the left direction in the lookup table.
	LeftDirection = Direction("left")
)

// ImmutableLookupTable represents a read-only view of a LookupTable.
// It is meant to apply the principle of least privilege by exposing only the methods needed for read-only access.
// e.g., in search operations where the lookup table is not supposed to be modified.
type ImmutableLookupTable interface {
	// GetEntry returns the lth left/right neighbor in the lookup table depending on the dir.
	// lev runs from 0...MaxLookupTableLevel-1.
	GetEntry(dir Direction, lev Level) (model.Identity, error)
}

// MutableLookupTable represents a read-write view of a LookupTable.
// It extends ImmutableLookupTable by adding methods for modifying the lookup table.
// e.g., in join operations where the lookup table needs to be updated.
type MutableLookupTable interface {
	ImmutableLookupTable
	// AddEntry inserts the supplied Identity in the lth level of lookup table either as the left or right neighbor depending on the dir.
	// lev runs from 0...MaxLookupTableLevel-1.
	AddEntry(dir Direction, level Level, identity model.Identity) error
}

ss
