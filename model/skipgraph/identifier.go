package skipgraph

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

const IdentifierSizeBytes = 32

var validate *validator.Validate

func init() {
	validate = validator.New()
}

const (
	CompareEqual   = "compare-equal"
	CompareGreater = "compare-greater"
	CompareLess    = "compare-less"
)

type ComparisonResult struct {
	// made this field unexported to ensure that only the constructor func is used
	// to create an instance of this type
	result string `validate:"oneof=compare-equal compare-greater compare-less"`
}

func (cr ComparisonResult) Result() string {
	return cr.result
}

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

// DebugInfo returns a human-readable debug info for the comparison result.
// diffIndex is the index of the first differing byte. In case of equality, diffIndex is not used.
// diffIndex is 0-indexed.
func DebugInfo(i Identifier, other Identifier, comparison ComparisonResult, diffIndex int) string {
	switch comparison {
	case CompareGreater:
		return fmt.Sprintf("%s > %s (at byte %d)", hex.EncodeToString(i[:diffIndex+1]), hex.EncodeToString(other[:diffIndex+1]), diffIndex)
	case CompareLess:
		return fmt.Sprintf("%s < %s (at byte %d)", hex.EncodeToString(i[:diffIndex+1]), hex.EncodeToString(other[:diffIndex+1]), diffIndex)
	default:
		return fmt.Sprintf("%s == %s", hex.EncodeToString(i[:diffIndex+1]), hex.EncodeToString(other[:diffIndex+1]))
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
	return Comparison{CompareEqual, DebugInfo(i, other, CompareEqual, len(i)-1), uint32(len(i) - 1)}
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
