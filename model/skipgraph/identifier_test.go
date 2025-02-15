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

func TestDebugInfo(t *testing.T) {
	id1, err := skipgraph.ByteToId([]byte{0x00, 0x01, 0x02, 0x03})
	require.NoError(t, err)
	id2, err := skipgraph.ByteToId([]byte{0x00, 0x01, 0x02, 0x04})
	require.NoError(t, err)

	// Test CompareGreater
	debugInfo := skipgraph.DebugInfo(id2, id1, skipgraph.CompareGreater, 31)
	expected := "0000000000000000000000000000000000000000000000000000000000010204 > 0000000000000000000000000000000000000000000000000000000000010203 (at byte 31)"
	require.Equal(t, expected, debugInfo)

	// Test CompareLess
	debugInfo = skipgraph.DebugInfo(id1, id2, skipgraph.CompareLess, 31)
	expected = "0000000000000000000000000000000000000000000000000000000000010203 < 0000000000000000000000000000000000000000000000000000000000010204 (at byte 31)"
	require.Equal(t, expected, debugInfo)

	// Test CompareEqual
	debugInfo = skipgraph.DebugInfo(id1, id1, skipgraph.CompareEqual, len(id1)-1)
	expected = "0000000000000000000000000000000000000000000000000000000000010203 == 0000000000000000000000000000000000000000000000000000000000010203"
	require.Equal(t, expected, debugInfo)
}

func TestIdentifierCompare(t *testing.T) {
	id0, err := skipgraph.ByteToId(bytes.Repeat([]byte{0}, skipgraph.IdentifierSizeBytes))
	require.NoError(t, err)
	id1, err := skipgraph.ByteToId(bytes.Repeat([]byte{127}, skipgraph.IdentifierSizeBytes))
	require.NoError(t, err)
	id2, err := skipgraph.ByteToId(bytes.Repeat([]byte{255}, skipgraph.IdentifierSizeBytes))
	require.NoError(t, err)

	// each id is equal to itself
	require.Equal(t, skipgraph.DebugInfo(id0, id0, skipgraph.CompareEqual, len(id0)-1), id0.Compare(id0).DebugInfo)
	require.Equal(t, skipgraph.DebugInfo(id1, id1, skipgraph.CompareEqual, len(id1)-1), id1.Compare(id1).DebugInfo)
	require.Equal(t, skipgraph.DebugInfo(id2, id2, skipgraph.CompareEqual, len(id2)-1), id2.Compare(id2).DebugInfo)

	// id0 < id1
	comp := id0.Compare(id1)
	require.Equal(t, skipgraph.CompareLess, comp.ComparisonResult)
	require.Equal(t, "00 < 7f (at byte 0)", comp.DebugInfo)
	require.Equal(t, uint32(0), comp.DiffIndex)

	comp = id1.Compare(id0)
	require.Equal(t, skipgraph.CompareGreater, comp.ComparisonResult)
	require.Equal(t, "7f > 00 (at byte 0)", comp.DebugInfo)
	require.Equal(t, uint32(0), comp.DiffIndex)

	// id1 < id2
	comp = id1.Compare(id2)
	require.Equal(t, skipgraph.CompareLess, comp.ComparisonResult)
	require.Equal(t, "7f < ff (at byte 0)", comp.DebugInfo)
	require.Equal(t, uint32(0), comp.DiffIndex)

	comp = id2.Compare(id1)
	require.Equal(t, skipgraph.CompareGreater, comp.ComparisonResult)
	require.Equal(t, "ff > 7f (at byte 0)", comp.DebugInfo)
	require.Equal(t, uint32(0), comp.DiffIndex)

	// id0 < id2
	comp = id0.Compare(id2)
	require.Equal(t, skipgraph.CompareLess, comp.ComparisonResult)
	require.Equal(t, "00 < ff (at byte 0)", comp.DebugInfo)
	require.Equal(t, uint32(0), comp.DiffIndex)

	comp = id2.Compare(id0)
	require.Equal(t, skipgraph.CompareGreater, comp.ComparisonResult)
	require.Equal(t, "ff > 00 (at byte 0)", comp.DebugInfo)
	require.Equal(t, uint32(0), comp.DiffIndex)

	// construct two random identifiers that differ only at index 16th (17th byte)
	differingByteIndex := skipgraph.IdentifierSizeBytes / 2
	leftBytes := unittest.RandomBytesFixture(t, differingByteIndex)
	rightBytes := unittest.RandomBytesFixture(t, skipgraph.IdentifierSizeBytes-differingByteIndex-1)

	// randomGreater = leftBytes<16>|1|rightBytes<15>
	randomGreater := append(append(leftBytes, 1), rightBytes...)
	idRandomGreater, err := skipgraph.ByteToId(randomGreater)
	require.NoError(t, err)

	// randomLess = leftBytes<16>|0|rightBytes<15>
	randomLess := append(append(leftBytes, 0), rightBytes...)
	idRandomLess, err := skipgraph.ByteToId(randomLess)
	require.NoError(t, err)

	// each identifier is equal to itself
	require.Equal(t, skipgraph.CompareEqual, idRandomGreater.Compare(idRandomGreater).ComparisonResult)
	require.Equal(t, skipgraph.CompareEqual, idRandomLess.Compare(idRandomLess).ComparisonResult)

	comp = idRandomGreater.Compare(idRandomLess)
	require.Equal(t, skipgraph.CompareGreater, comp.ComparisonResult)
	require.Equal(t, uint32(differingByteIndex), comp.DiffIndex)
	expectedDebugInfo := fmt.Sprintf("%s > %s (at byte %d)", hex.EncodeToString(idRandomGreater[:differingByteIndex+1]), hex.EncodeToString(idRandomLess[:differingByteIndex+1]), differingByteIndex)
	require.Equal(t, expectedDebugInfo, comp.DebugInfo)

	comp = idRandomLess.Compare(idRandomGreater)
	require.Equal(t, skipgraph.CompareLess, comp.ComparisonResult)
	require.Equal(t, uint32(differingByteIndex), comp.DiffIndex)
	expectedDebugInfo = fmt.Sprintf("%s < %s (at byte %d)", hex.EncodeToString(idRandomLess[:differingByteIndex+1]), hex.EncodeToString(idRandomGreater[:differingByteIndex+1]), differingByteIndex)
	require.Equal(t, expectedDebugInfo, comp.DebugInfo)
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
