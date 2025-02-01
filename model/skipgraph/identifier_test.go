package skipgraph_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
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

func TestIdentifierCompare(t *testing.T) {
	id0, err := skipgraph.ByteToId(bytes.Repeat([]byte{0}, skipgraph.IdentifierSizeBytes))
	require.NoError(t, err)
	id1, err := skipgraph.ByteToId(bytes.Repeat([]byte{127}, skipgraph.IdentifierSizeBytes))
	require.NoError(t, err)
	id2, err := skipgraph.ByteToId(bytes.Repeat([]byte{255}, skipgraph.IdentifierSizeBytes))
	require.NoError(t, err)

	// each id is equal to itself
	require.Equal(t, skipgraph.CompareEqual, id0.Compare(id0).DebugInfo)
	require.Equal(t, skipgraph.CompareEqual, id1.Compare(id1).DebugInfo)
	require.Equal(t, skipgraph.CompareEqual, id2.Compare(id2).DebugInfo)

	// id0 < id1
	comp := id0.Compare(id1)
	require.Equal(t, skipgraph.CompareLess, comp.DebugInfo)
	require.Equal(t, uint32(0), comp.DiffIndex)
	require.Equal(t, "00 < 7f (at byte 0)", fmt.Sprintf("%02x < %02x (at byte %d)", id0[0], id1[0], comp.DiffIndex))

	comp = id1.Compare(id0)
	require.Equal(t, skipgraph.CompareGreater, comp.DebugInfo)
	require.Equal(t, uint32(0), comp.DiffIndex)
	require.Equal(t, "7f > 00 (at byte 0)", fmt.Sprintf("%02x > %02x (at byte %d)", id1[0], id0[0], comp.DiffIndex))

	// id1 < id2
	comp = id1.Compare(id2)
	require.Equal(t, skipgraph.CompareLess, comp.DebugInfo)
	require.Equal(t, uint32(0), comp.DiffIndex)
	require.Equal(t, "7f < ff (at byte 0)", fmt.Sprintf("%02x < %02x (at byte %d)", id1[0], id2[0], comp.DiffIndex))

	comp = id2.Compare(id1)
	require.Equal(t, skipgraph.CompareGreater, comp.DebugInfo)
	require.Equal(t, uint32(0), comp.DiffIndex)
	require.Equal(t, "ff > 7f (at byte 0)", fmt.Sprintf("%02x > %02x (at byte %d)", id2[0], id1[0], comp.DiffIndex))

	// id0 < id2
	comp = id0.Compare(id2)
	require.Equal(t, skipgraph.CompareLess, comp.DebugInfo)
	require.Equal(t, uint32(0), comp.DiffIndex)
	require.Equal(t, "00 < ff (at byte 0)", fmt.Sprintf("%02x < %02x (at byte %d)", id0[0], id2[0], comp.DiffIndex))

	comp = id2.Compare(id0)
	require.Equal(t, skipgraph.CompareGreater, comp.DebugInfo)
	require.Equal(t, uint32(0), comp.DiffIndex)
	require.Equal(t, "ff > 00 (at byte 0)", fmt.Sprintf("%02x > %02x (at byte %d)", id2[0], id0[0], comp.DiffIndex))

	// two random identifiers composed that differ only at index 16th (17th byte)
	differingByteIndex := skipgraph.IdentifierSizeBytes / 2
	leftBytes := unittest.RandomBytesFixture(t, differingByteIndex)
	rightBytes := unittest.RandomBytesFixture(t, skipgraph.IdentifierSizeBytes-differingByteIndex-1)

	// randomGreater = random<16>|1|random<15>
	randomGreater := append(append(leftBytes, 1), rightBytes...)
	idRandomGreater, err := skipgraph.ByteToId(randomGreater)
	require.NoError(t, err)

	// randomLess = random<16>|0|random<15>
	randomLess := append(append(leftBytes, 0), rightBytes...)
	idRandomLess, err := skipgraph.ByteToId(randomLess)
	require.NoError(t, err)

	// each identifier is equal to itself
	require.Equal(t, skipgraph.CompareEqual, idRandomGreater.Compare(idRandomGreater).DebugInfo)
	require.Equal(t, skipgraph.CompareEqual, idRandomLess.Compare(idRandomLess).DebugInfo)

	comp = idRandomGreater.Compare(idRandomLess)
	require.Equal(t, skipgraph.CompareGreater, comp.DebugInfo)
	require.Equal(t, uint32(differingByteIndex), comp.DiffIndex)
	require.Equal(t, fmt.Sprintf("%s > %s (at byte %d)", hex.EncodeToString(idRandomGreater[:differingByteIndex+1]), hex.EncodeToString(idRandomLess[:differingByteIndex+1]), differingByteIndex), fmt.Sprintf("%s > %s (at byte %d)", hex.EncodeToString(idRandomGreater[:differingByteIndex+1]), hex.EncodeToString(idRandomLess[:differingByteIndex+1]), differingByteIndex))

	comp = idRandomLess.Compare(idRandomGreater)
	require.Equal(t, skipgraph.CompareLess, comp.DebugInfo)
	require.Equal(t, uint32(differingByteIndex), comp.DiffIndex)
	require.Equal(t, fmt.Sprintf("%s < %s (at byte %d)", hex.EncodeToString(idRandomLess[:differingByteIndex+1]), hex.EncodeToString(idRandomGreater[:differingByteIndex+1]), differingByteIndex), fmt.Sprintf("%s < %s (at byte %d)", hex.EncodeToString(idRandomLess[:differingByteIndex+1]), hex.EncodeToString(idRandomGreater[:differingByteIndex+1]), differingByteIndex))
}
func TestIdentifier_Bytes(t *testing.T) {
	// 32 bytes of zero
	b := bytes.Repeat([]byte{0}, skipgraph.IdentifierSizeBytes)
	id, err := skipgraph.ByteToId(b)
	require.NoError(t, err)
	require.Equal(t, b, id.Bytes())

	// 32 bytes of 1
	b = bytes.Repeat([]byte{255}, skipgraph.IdentifierSizeBytes)
	id, err = skipgraph.ByteToId(b)
	require.NoError(t, err)
	require.Equal(t, b, id.Bytes())

	// 32 random bytes
	b = unittest.RandomBytesFixture(t, skipgraph.IdentifierSizeBytes)
	id, err = skipgraph.ByteToId(b)
	require.NoError(t, err)
	require.Equal(t, b, id.Bytes())

	// 31 random bytes should be zero padded
	b = unittest.RandomBytesFixture(t, skipgraph.IdentifierSizeBytes-1)
	id, err = skipgraph.ByteToId(b)
	require.NoError(t, err)
	unittest.MustHaveZeroPrefixBytes(t, id.Bytes(), 1, b...)

	// 33 random bytes should return an error
	b = unittest.RandomBytesFixture(t, skipgraph.IdentifierSizeBytes+1)
	_, err = skipgraph.ByteToId(b)
	require.Error(t, err)

}
