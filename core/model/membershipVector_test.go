package model_test

import (
	"errors"
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/core/model"
	"testing"
)

// TestMembershipVectorCompare tests CommonPrefix method.
func TestMembershipVector_CommonPrefix(t *testing.T) {
	// create two membershipVectors with 32 * 8 common prefix
	v1 := model.MembershipVector{0}
	res := v1.CommonPrefix(v1)
	require.Equal(t, 256, res)

	// create two membershipVectors with no common prefix
	v2 := model.MembershipVector{0}
	v2[0] = 255
	res = v1.CommonPrefix(v2)
	require.Equal(t, 0, res)

	// create two membershipVectors with non-zero common prefix
	v1[0] = 253
	res = v1.CommonPrefix(v2)
	require.Equal(t, 6, res)
}

// TestToBinaryString tests correctness of ToBinaryString.
func TestToBinaryString(t *testing.T) {
	v1 := byte(1) // 00000001
	s1 := model.ToBinaryString(v1)
	require.Equal(t, "00000001", s1)

	v2 := byte(2) // 00000010
	s2 := model.ToBinaryString(v2)
	require.Equal(t, "00000010", s2)

	v3 := byte(128) // 10000000
	s3 := model.ToBinaryString(v3)
	require.Equal(t, "10000000", s3)

	v4 := byte(65) // 01000001
	s4 := model.ToBinaryString(v4)
	require.Equal(t, "01000001", s4)
}

func TestToMembershipVector(t *testing.T) {
	// check a valid conversation
	v1 := []byte{0}
	v1[0] = 255
	m, err := model.ToMembershipVector(v1)
	require.NoError(t, err)
	require.Equal(t, model.MembershipVectorSize, len(m))
	// check zero leading is added and the last byte is equal to 255
	require.Equal(t, uint8(255), m[model.MembershipVectorSize-1])

	// check invalid input
	v2 := [33]byte{1}
	_, err2 := model.ToMembershipVector(v2[:])
	require.Error(t, err2)

}

// TestMembershipVector_GetPrefixBits tests the GetPrefixBits method.
func TestMembershipVector_GetPrefixBits(t *testing.T) {
	// Create a membership vector with known binary representation
	// Using byte 170 (10101010) and 85 (01010101) for easy pattern verification
	mv := model.MembershipVector{}
	mv[0] = 170 // 10101010
	mv[1] = 85  // 01010101

	// Test getting first 8 bits
	result, err := mv.GetPrefixBits(8)
	require.NoError(t, err)
	require.Equal(t, "10101010", result)

	// Test getting first 16 bits
	result, err = mv.GetPrefixBits(16)
	require.NoError(t, err)
	require.Equal(t, "1010101001010101", result)

	// Test getting first 4 bits
	result, err = mv.GetPrefixBits(4)
	require.NoError(t, err)
	require.Equal(t, "1010", result)

	// Test getting 0 bits
	result, err = mv.GetPrefixBits(0)
	require.NoError(t, err)
	require.Equal(t, "", result)

	// Test getting exactly 256 bits (full size)
	fullBinary := mv.ToBinaryString()
	result, err = mv.GetPrefixBits(256)
	require.NoError(t, err)
	require.Equal(t, fullBinary, result)

	// Test getting more bits than available (should return error)
	result, err = mv.GetPrefixBits(300)
	require.Error(t, err)
	require.Equal(t, "", result)
	require.True(t, errors.Is(err, model.ErrNumBitsExceedsMax), "expected ErrNumBitsExceedsMax")

	// Test getting negative bits (should return error)
	result, err = mv.GetPrefixBits(-1)
	require.Error(t, err)
	require.Equal(t, "", result)
	require.True(t, errors.Is(err, model.ErrNegativeNumBits), "expected ErrNegativeNumBits")

	// Test with all zeros
	mvZero := model.MembershipVector{}
	result, err = mvZero.GetPrefixBits(8)
	require.NoError(t, err)
	require.Equal(t, "00000000", result)

	// Test with all ones
	mvOnes := model.MembershipVector{}
	for i := 0; i < model.MembershipVectorSize; i++ {
		mvOnes[i] = 255
	}
	result, err = mvOnes.GetPrefixBits(8)
	require.NoError(t, err)
	require.Equal(t, "11111111", result)
}
