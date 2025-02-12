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

// String converts Identifier to its hex representation.
func (i Identifier) String() string {
	return hex.EncodeToString(i[:])
}

// Bytes returns the byte representation of an Identifier.
func (i Identifier) Bytes() []byte {
	return i[:]
}

type Comparison struct {
	ComparisonResult ComparisonResult // one of CompareEqual, CompareGreater, CompareLess
	DebugInfo        string           // in case of inequality, a human-readable debug info with the index of the first differing byte
	DiffIndex        uint32           // in case of inequality, the index of the first differing byte
}

func DebugInfo(i Identifier, other Identifier, comparison ComparisonResult, index ...int) string {
	switch comparison {
	case CompareGreater:
		return fmt.Sprintf("%s > %s (at byte %d)", hex.EncodeToString(i[index[0]:index[0]+1]), hex.EncodeToString(other[index[0]:index[0]+1]), index[0])
	case CompareLess:
		return fmt.Sprintf("%s < %s (at byte %d)", hex.EncodeToString(i[index[0]:index[0]+1]), hex.EncodeToString(other[index[0]:index[0]+1]), index[0])
	default:
		return ""
	}
}

// Compare compares two Identifiers and returns a Comparison result, including the debugging info and the first mismatching byte index, if applicable.
func (i Identifier) Compare(other Identifier) Comparison {
	for index := range i {
		cmp := bytes.Compare(i[index:index+1], other[index:index+1])
		switch cmp {
		case 1:
			return Comparison{CompareGreater, DebugInfo(i, other, CompareGreater, index), uint32(index)}
		case -1:

			return Comparison{CompareLess, DebugInfo(i, other, CompareLess, index), uint32(index)}
		default:
			continue
		}
	}
	return Comparison{ComparisonResult: CompareEqual}
}

// ByteToId converts a byte slice b to an Identifier.
// Returns error if the length of b is more than Identifier's length i.e., 32 bytes.
// If the length of b is less than 32 bytes, it is zero padded from the left.
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
