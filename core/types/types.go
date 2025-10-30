// Package types defines common primitive types used across the Skip Graph implementation.
// This package exists to break import cycles between core and core/model packages.
package types

// Level is the type for the level of entries in the lookup table.
// Levels are 0-indexed and range from 0 to MaxLookupTableLevel-1.
type Level int64

// Direction is an enum type for the direction of a neighbor in the lookup table.
// Valid values are DirectionRight and DirectionLeft.
type Direction string

const (
	// DirectionRight indicates the right direction in the lookup table.
	DirectionRight = Direction("right")
	// DirectionLeft indicates the left direction in the lookup table.
	DirectionLeft = Direction("left")
)
