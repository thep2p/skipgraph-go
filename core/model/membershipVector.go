package model

import (
	"encoding/hex"
	"fmt"
)

// MembershipVectorSize is the size of MembershipVector.
const MembershipVectorSize = 32

// MembershipVector represents a SkipGraph node's name id which is a 32 byte array.
type MembershipVector [MembershipVectorSize]byte

// String returns hex encoding of a MembershipVector.
func (m MembershipVector) String() string {
	return hex.EncodeToString(m[:])
}

// ToBinaryString returns binary representation of a MembershipVector.
func (m MembershipVector) ToBinaryString() string {
	var s string
	for i := 0; i < len(m); i++ {
		s = s + ToBinaryString(m[i])
	}
	return s
}

// ToBinaryString returns binary representation of a byte value.
func ToBinaryString(b byte) string {
	var s string
	for j := 0; j < 8; j++ {
		// m[i] is an 8 bit value i.e., x0 x1 x2 ... x7
		v := b >> (7 - j)   // v = 0 ... x0 x1 ... xj-1 xj (cuts the last 7-j bits to shift the jth bit to the least significant bit)
		b := v & 0b00000001 // get the value of the least significant bit (which is xj)
		if b == 1 {
			s = s + "1"
		} else {
			s = s + "0"
		}
	}
	return s
}

// GetPrefixBits returns the first numBits bits as a string representation.
// If numBits exceeds the length of the binary representation, returns the full string.
func (m MembershipVector) GetPrefixBits(numBits int) string {
	binaryStr := m.ToBinaryString()
	if numBits > len(binaryStr) {
		return binaryStr
	}
	return binaryStr[:numBits]
}

// CommonPrefix returns the longest common bit prefix of the supplied MembershipVectors.
func (m MembershipVector) CommonPrefix(other MembershipVector) int {
	// convert to bit string
	s1 := m.ToBinaryString()
	s2 := other.ToBinaryString()

	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return i
		}
	}
	// TODO: comment
	return MembershipVectorSize * 8 // m and other are identical
}

// ToMembershipVector converts a byte slice to a MembershipVector.
// returns error if length of s is more than MembershipVector's length i.e., MembershipVectorSize bytes.
func ToMembershipVector(s []byte) (MembershipVector, error) {
	res := MembershipVector{0}
	if len(s) > MembershipVectorSize {
		return res, fmt.Errorf("input length must be at most %d bytes; found: %d", MembershipVectorSize, len(s))
	}
	index := MembershipVectorSize - 1
	for i := len(s) - 1; i >= 0; i-- {
		res[index] = s[i]
		index--
	}
	return res, nil
}

// StringToMembershipVector converts a string to a MembershipVector.
// returns error if the byte length of the string is more than MembershipVector's length i.e., MembershipVectorSize bytes.
func StringToMembershipVector(s string) (MembershipVector, error) {
	b := []byte(s)
	res := MembershipVector{0}
	if len(b) > MembershipVectorSize {
		return res, fmt.Errorf("input length must be at most %d bytes; found: %d", MembershipVectorSize, len(b))
	}
	index := MembershipVectorSize - 1
	for i := len(b) - 1; i >= 0; i-- {
		res[index] = b[i]
		index--
	}
	return res, nil
}
