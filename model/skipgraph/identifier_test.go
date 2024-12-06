package skipgraph_test

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"github/yhassanzadeh13/skipgraph-go/model/skipgraph"
	"github/yhassanzadeh13/skipgraph-go/unittest"
	"testing"
)

// TestCompare checks the correctness of the Identifier comparison function
func TestCompare(t *testing.T) {
	s1 := []byte("12")
	s2 := []byte("22")
	s3 := []byte("12")
	i1, err := skipgraph.ByteToId(s1)
	require.NoError(t, err)
	i2, err := skipgraph.ByteToId(s2)
	require.NoError(t, err)
	i3, err := skipgraph.ByteToId(s3)
	require.NoError(t, err)

	require.Equal(t, skipgraph.CompareLess, i1.Compare(i2))
	require.Equal(t, skipgraph.CompareGreater, i2.Compare(i1))
	require.Equal(t, skipgraph.CompareEqual, i1.Compare(i3))
}

func TestIdentifier_Bytes(t *testing.T) {
	// 32 bytes of zero
	b := bytes.Repeat([]byte{0}, skipgraph.IdentifierSize)
	id, err := skipgraph.ByteToId(b)
	require.NoError(t, err)
	require.Equal(t, b, id.Bytes())

	// 32 bytes of 1
	b = bytes.Repeat([]byte{255}, skipgraph.IdentifierSize)
	id, err = skipgraph.ByteToId(b)
	require.NoError(t, err)
	require.Equal(t, b, id.Bytes())

	// 32 random bytes
	b = unittest.RandomBytesFixture(t, skipgraph.IdentifierSize)
	id, err = skipgraph.ByteToId(b)
	require.NoError(t, err)
	require.Equal(t, b, id.Bytes())

	// 31 random bytes should be zero padded
	b = unittest.RandomBytesFixture(t, skipgraph.IdentifierSize-1)
	id, err = skipgraph.ByteToId(b)
	require.NoError(t, err)
	unittest.MustHaveZeroPrefixBytes(t, id.Bytes(), 1, b...)

	// 33 random bytes should return an error
	b = unittest.RandomBytesFixture(t, skipgraph.IdentifierSize+1)
	_, err = skipgraph.ByteToId(b)
	require.Error(t, err)

}
