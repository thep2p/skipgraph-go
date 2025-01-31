package skipgraph

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

const IdentifierSizeBytes = 32
const CompareEqual = "compare-equal"
const CompareGreater = "compare-greater"
const CompareLess = "compare-less"

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
	DebugInfo string
	DiffIndex uint32
}

// Compare compares two Identifiers and returns 0 if equal, 1 if other > i and -1 if other < i.
func (i Identifier) Compare(other Identifier) Comparison {
	for index, _ := range i {
		cmp := bytes.Compare(i[index:index], other[index:index])
		switch cmp {
		case 1:
			return Comparison{CompareGreater, uint32(index)}
		case -1:

			return Comparison{CompareLess, uint32(index)}
		default:
			continue
		}
	}
	return Comparison{DebugInfo: CompareEqual}
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
