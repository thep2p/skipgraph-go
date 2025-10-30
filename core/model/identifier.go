package model

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/go-playground/validator/v10"
)

const IdentifierSizeBytes = 32

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithPrivateFieldValidation())
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
func NewComparisonResult(s string) (*ComparisonResult, error) {
	cr := ComparisonResult{s}
	err := validate.Struct(cr)
	if err != nil {
		return nil, fmt.Errorf("failed to validate the comparison result upon instantiation: %w", err)
	}
	return &cr, nil
}

// Identifier represents a 32-byte unique identifier a Skip Graph node.
type Identifier [IdentifierSizeBytes]byte

// IdentifierList is a slice of Identifier
type IdentifierList []Identifier

type Comparison struct {
	comparisonResult ComparisonResult // one of CompareEqual, CompareGreater, CompareLess
	left, right      *Identifier      // the two identifiers being compared
	diffIndex        uint32           // in case of inequality, the index of the first differing byte. 0-indexed.
}

// NewComparison creates a new Comparison instance.
func NewComparison(result ComparisonResult, left, right *Identifier, diffIndex uint32) *Comparison {
	return &Comparison{
		comparisonResult: result,
		left:             left,
		right:            right,
		diffIndex:        diffIndex,
	}
}

// GetComparisonResult returns the comparison result.
func (c *Comparison) GetComparisonResult() string {
	return c.comparisonResult.Result()
}

// GetLeft returns the left identifier.
func (c *Comparison) GetLeft() *Identifier {
	return c.left
}

// GetRight returns the right identifier.
func (c *Comparison) GetRight() *Identifier {
	return c.right
}

// GetDiffIndex returns the index of the first differing byte.
func (c *Comparison) GetDiffIndex() uint32 {
	return c.diffIndex
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
	switch c.GetComparisonResult() {
	case CompareGreater:
		return fmt.Sprintf(
			"%s > %s (at byte %d)",
			hex.EncodeToString(c.GetLeft()[:c.GetDiffIndex()+1]),
			hex.EncodeToString(c.GetRight()[:c.GetDiffIndex()+1]),
			c.GetDiffIndex(),
		)
	case CompareLess:
		return fmt.Sprintf(
			"%s < %s (at byte %d)",
			hex.EncodeToString(c.GetLeft()[:c.GetDiffIndex()+1]),
			hex.EncodeToString(c.GetRight()[:c.GetDiffIndex()+1]),
			c.GetDiffIndex(),
		)
	default:
		return fmt.Sprintf("%s == %s", hex.EncodeToString(c.GetLeft()[:c.GetDiffIndex()+1]), hex.EncodeToString(c.GetRight()[:c.GetDiffIndex()+1]))
	}
}

// Compare compares two Identifiers and returns a Comparison result, including the debugging info and the first mismatching byte index, if applicable.
func (i *Identifier) Compare(other *Identifier) Comparison {
	for index := range i {
		cmp := bytes.Compare(i[index:index+1], other[index:index+1])
		switch cmp {
		case 1:
			cr, err := NewComparisonResult(CompareGreater)
			if err != nil {
				panic(err)
			}
			return Comparison{*cr, i, other, uint32(index)}
		case -1:
			cr, err := NewComparisonResult(CompareLess)
			if err != nil {
				panic(err)
			}
			return Comparison{*cr, i, other, uint32(index)}
		default:
			continue
		}
	}
	cr, err := NewComparisonResult(CompareEqual)
	if err != nil {
		panic(err)
	}
	return Comparison{*cr, i, other, uint32(len(i) - 1)}
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
		return res, fmt.Errorf("%w: must be at most %d bytes, found %d", ErrIdentifierTooLarge, IdentifierSizeBytes, len(b))
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
		return Identifier{}, fmt.Errorf("%w: %s", ErrInvalidHexString, err)
	}
	return ByteToId(b)
}
