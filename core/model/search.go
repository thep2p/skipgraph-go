package model

// Level is the type for the level of entries in the lookup table.
// This is duplicated from core.Level to avoid import cycles.
type Level int64

// Direction is an enum type for the direction of a neighbor in the lookup table.
// This is duplicated from core.Direction to avoid import cycles.
type Direction string

const (
	// DirectionRight indicates the right direction in the lookup table.
	DirectionRight = Direction("right")
	// DirectionLeft indicates the left direction in the lookup table.
	DirectionLeft = Direction("left")
)

// IdSearchReq represents a request to search for an identifier in the lookup table.
// It specifies the target identifier, the maximum level to search up to, and the search direction.
type IdSearchReq struct {
	target    Identifier // The target identifier to search for
	level     Level      // Maximum level to search (inclusive, 0-indexed)
	direction Direction  // Search direction (Left or Right)
}

// NewIdSearchReq creates a new IdSearchReq instance.
// Args:
//   - target: the identifier to search for
//   - level: the maximum level to search up to (inclusive)
//   - direction: the search direction (DirectionLeft or DirectionRight)
//
// Returns:
//   - IdSearchReq: the constructed search request
func NewIdSearchReq(target Identifier, level Level, direction Direction) IdSearchReq {
	return IdSearchReq{
		target:    target,
		level:     level,
		direction: direction,
	}
}

// Target returns the target identifier being searched for.
func (r IdSearchReq) Target() Identifier {
	return r.target
}

// Level returns the maximum level to search up to (inclusive).
func (r IdSearchReq) Level() Level {
	return r.level
}

// Direction returns the search direction (Left or Right).
func (r IdSearchReq) Direction() Direction {
	return r.direction
}

// IdSearchRes represents the result of an identifier search.
// It contains the target identifier, the level where the search terminated,
// and the identifier found (or own ID as fallback).
type IdSearchRes struct {
	target           Identifier // The target identifier that was searched for
	terminationLevel Level      // The level where the search terminated
	result           Identifier // The identifier found (or own ID as fallback)
}

// NewIdSearchRes creates a new IdSearchRes instance.
// Args:
//   - target: the identifier that was searched for
//   - terminationLevel: the level where a match was found
//   - result: the matched identifier (or fallback to own ID)
//
// Returns:
//   - IdSearchRes: the constructed search result
func NewIdSearchRes(target Identifier, terminationLevel Level, result Identifier) IdSearchRes {
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
func (r IdSearchRes) TerminationLevel() Level {
	return r.terminationLevel
}

// Result returns the identifier found (or own ID as fallback).
func (r IdSearchRes) Result() Identifier {
	return r.result
}
