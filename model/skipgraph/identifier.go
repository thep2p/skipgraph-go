package skipgraph

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

const IdentifierSizeBytes = 32

type ComparisonResult string

const CompareEqual ComparisonResult = "compare-equal"
const CompareGreater ComparisonResult = "compare-greater"
const CompareLess ComparisonResult = "compare-less"

// Identifier represents a 32-byte unique identifier a Skip Graph node.
type Identifier [IdentifierSizeBytes]byte

// IdentifierList is a slice of Identifier
type IdentifierList []Identifier

type Comparison struct {
	ComparisonResult ComparisonResult // one of CompareEqual, CompareGreater, CompareLess
	Left, Right      *Identifier      // the two identifiers being compared
	DiffIndex        uint32           // in case of inequality, the index of the first differing byte. 0-indexed.
}

// String converts Identifier to its hex representation.
func (i *Identifier) String() string {
	return hex.EncodeToString(i[:])
}

// Bytes returns the byte representation of an Identifier.
func (i *Identifier) Bytes() []byte {
	return i[:]
}

// DebugInfo returns a human-readable debug info for the comparison result.
func (c *Comparison) DebugInfo() string {
	switch c.ComparisonResult {
	case CompareGreater:
		return fmt.Sprintf("%s > %s (at byte %d)", hex.EncodeToString(c.Left[:c.DiffIndex+1]), hex.EncodeToString(c.Right[:c.DiffIndex+1]), c.DiffIndex)
	case CompareLess:
		return fmt.Sprintf("%s < %s (at byte %d)", hex.EncodeToString(c.Left[:c.DiffIndex+1]), hex.EncodeToString(c.Right[:c.DiffIndex+1]), c.DiffIndex)
	default:
		return fmt.Sprintf("%s == %s", hex.EncodeToString(c.Left[:c.DiffIndex+1]), hex.EncodeToString(c.Right[:c.DiffIndex+1]))
	}
}

// Compare compares two Identifiers and returns a Comparison result, including the debugging info and the first mismatching byte index, if applicable.
func (i *Identifier) Compare(other *Identifier) Comparison {
	for index := range i {
		cmp := bytes.Compare(i[index:index+1], other[index:index+1])
		switch cmp {
		case 1:
			return Comparison{CompareGreater, i, other, uint32(index)}
		case -1:

			return Comparison{CompareLess, i, other, uint32(index)}
		default:
			continue
		}
	}
	return Comparison{CompareEqual, i, other, uint32(len(i) - 1)}
}

// ByteToId converts a byte slice b to an Identifier.
// Returns error if the length of b is more than Identifier's length i.e., 32 bytes.
// If the length of b is less than 32 bytes, it is zero padded from the Left.
// It follows a big-endian representation where the 0 index of the byte slice corresponds to the most significant byte.
// Args:
//
//	b: the byte slice to be converted to an Identifier
//
// Returns:
//
//	Identifier: the converted Identifier
//	error: if the length of b is more than 32 bytes
func ByteToId(b []byte) (Identifier, error) {
	res := Identifier{0}
	if len(b) > IdentifierSizeBytes {
		return res, fmt.Errorf("input length must be at most %d bytes; found: %d", IdentifierSizeBytes, len(b))
	}
	offset := IdentifierSizeBytes - len(b)
	copy(res[offset:], b)
	return res, nil
}

// StrToId converts a string to an Identifier.
// returns error if the byte length of the string s is more than Identifier's length i.e., 32 bytes.
func StrToId(s string) (Identifier, error) {
	// converts string to byte
	b, err := hex.DecodeString(s)
	if err != nil {
		return Identifier{}, fmt.Errorf("failed to decode hex string: %s", err)
	}
	return ByteToId(b)
}
