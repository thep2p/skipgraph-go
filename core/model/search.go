package model

import (
	"fmt"
	"github.com/thep2p/skipgraph-go/core/types"
)

// TODO: Technical Debt - Type Import from core/types
// This file imports Level and Direction types from core/types package.
// These types were moved to core/types to break an import cycle between core and core/model.
// This is the correct architectural pattern per CLAUDE.md guidelines:
// - NEVER duplicate types to avoid import cycles
// - ALWAYS create a shared types package for primitive types shared across packages
// - Establish clear dependency hierarchy: core → core/types ← core/model
//
// The current design is intentional and follows best practices. This comment documents
// the architectural decision for future maintainers.

// IdSearchReq represents a request to search for an identifier in the lookup table.
// It specifies the target identifier, the maximum level to search up to, and the search direction.
type IdSearchReq struct {
	target    Identifier      // The target identifier to search for
	level     types.Level     // Maximum level to search (inclusive, 0-indexed)
	direction types.Direction // Search direction (Left or Right)
}

// NewIdSearchReq creates a new IdSearchReq instance with input validation.
// Args:
//   - target: the identifier to search for
//   - level: the maximum level to search up to (inclusive)
//   - direction: the search direction (types.DirectionLeft or types.DirectionRight)
//
// Returns:
//   - IdSearchReq: the constructed search request
//   - error: validation error if inputs are invalid
//
// Validation rules:
//   - level must be >= 0
//   - level must be < IdentifierSizeBytes * 8 (MaxLookupTableLevel)
//   - direction must be either DirectionLeft or DirectionRight
func NewIdSearchReq(target Identifier, level types.Level, direction types.Direction) (IdSearchReq, error) {
	// Validate level bounds
	const maxLookupTableLevel = IdentifierSizeBytes * 8
	if level < 0 {
		return IdSearchReq{}, fmt.Errorf("level must be non-negative, got: %d", level)
	}
	if level >= maxLookupTableLevel {
		return IdSearchReq{}, fmt.Errorf("level must be less than %d, got: %d", maxLookupTableLevel, level)
	}

	// Validate direction
	if direction != types.DirectionLeft && direction != types.DirectionRight {
		return IdSearchReq{}, fmt.Errorf("direction must be either DirectionLeft or DirectionRight, got: %s", direction)
	}

	return IdSearchReq{
		target:    target,
		level:     level,
		direction: direction,
	}, nil
}

// Target returns the target identifier being searched for.
func (r IdSearchReq) Target() Identifier {
	return r.target
}

// Level returns the maximum level to search up to (inclusive).
func (r IdSearchReq) Level() types.Level {
	return r.level
}

// Direction returns the search direction (Left or Right).
func (r IdSearchReq) Direction() types.Direction {
	return r.direction
}

// IdSearchRes represents the result of an identifier search.
// It contains the target identifier, the level where the search terminated,
// and the identifier found (or own ID as fallback).
type IdSearchRes struct {
	target           Identifier  // The target identifier that was searched for
	terminationLevel types.Level // The level where the search terminated
	result           Identifier  // The identifier found (or own ID as fallback)
}

// NewIdSearchRes creates a new IdSearchRes instance.
// Args:
//   - target: the identifier that was searched for
//   - terminationLevel: the level where a match was found
//   - result: the matched identifier (or fallback to own ID)
//
// Returns:
//   - IdSearchRes: the constructed search result
func NewIdSearchRes(target Identifier, terminationLevel types.Level, result Identifier) IdSearchRes {
	return IdSearchRes{
		target:           target,
		terminationLevel: terminationLevel,
		result:           result,
	}
}

// Target returns the target identifier that was searched for.
func (r IdSearchRes) Target() Identifier {
	return r.target
}

// TerminationLevel returns the level where the search terminated.
func (r IdSearchRes) TerminationLevel() types.Level {
	return r.terminationLevel
}

// Result returns the identifier found (or own ID as fallback).
func (r IdSearchRes) Result() Identifier {
	return r.result
}
