package unittest

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// MustHaveZeroPrefixBytes checks if the given byte slice has zero prefix of the given length and the rest of the bytes are equal to the given bytes.
// Args:
//
//	t: the testing.T instance
//	b: the byte slice to be checked
//	zeroPrefix: the length of the zero prefix
//	rest: the rest of the bytes to be checked (optional)
func MustHaveZeroPrefixBytes(t *testing.T, b []byte, zeroPrefix int, rest ...byte) {
	for i := 0; i < zeroPrefix; i++ {
		require.Equal(t, byte(0), b[i], "byte %d is not zero", i)
	}
	for i, r := range rest {
		require.Equal(t, r, b[zeroPrefix+i], "byte %d is not equal to %d", zeroPrefix+i, r)
	}
}
