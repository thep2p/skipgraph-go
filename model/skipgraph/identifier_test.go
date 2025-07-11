package skipgraph_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/require"
	"github/thep2p/skipgraph-go/model/skipgraph"
	"github/thep2p/skipgraph-go/unittest"
	"testing"
)

func TestNewComparisonResult(t *testing.T) {
	cr, err := skipgraph.NewComparisonResult(skipgraph.CompareEqual)
	require.NoError(t, err)
	require.Equal(t, skipgraph.CompareEqual, cr.Result())

	cr, err = skipgraph.NewComparisonResult(skipgraph.CompareGreater)
	require.NoError(t, err)
	require.Equal(t, skipgraph.CompareGreater, cr.Result())

	cr, err = skipgraph.NewComparisonResult(skipgraph.CompareLess)
	require.NoError(t, err)
	require.Equal(t, skipgraph.CompareLess, cr.Result())

	_, err = skipgraph.NewComparisonResult("invalid")
	require.Error(t, err)
}
func TestDebugInfo(t *testing.T) {
	id1, err := skipgraph.ByteToId([]byte{0x00, 0x01, 0x02, 0x03})
	require.NoError(t, err)
	id2, err := skipgraph.ByteToId([]byte{0x00, 0x01, 0x02, 0x04})
	require.NoError(t, err)

	// Test CompareGreater
	crGreater, err := skipgraph.NewComparisonResult(skipgraph.CompareGreater)
	require.NoError(t, err)
	c := skipgraph.NewComparison(*crGreater, &id2, &id1, uint32(31))
	debugInfo := c.DebugInfo()
	expected := "0000000000000000000000000000000000000000000000000000000000010204 > 0000000000000000000000000000000000000000000000000000000000010203 (at byte 31)"
	require.Equal(t, expected, debugInfo)

	// Test CompareLess
	crLess, err := skipgraph.NewComparisonResult(skipgraph.CompareLess)
	require.NoError(t, err)
	c = skipgraph.NewComparison(*crLess, &id1, &id2, uint32(31))
	debugInfo = c.DebugInfo()
	expected = "0000000000000000000000000000000000000000000000000000000000010203 < 0000000000000000000000000000000000000000000000000000000000010204 (at byte 31)"
	require.Equal(t, expected, debugInfo)

	// Test CompareEqual
	crEqual, err := skipgraph.NewComparisonResult(skipgraph.CompareEqual)
	require.NoError(t, err)
	c = skipgraph.NewComparison(*crEqual, &id1, &id1, uint32(len(id1)-1))
	debugInfo = c.DebugInfo()
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
	crEqual, err := skipgraph.NewComparisonResult(skipgraph.CompareEqual)
	require.NoError(t, err)
	exp := skipgraph.NewComparison(*crEqual, &id0, &id0, uint32(len(id0)-1))
	res := id0.Compare(&id0)
	require.Equal(t, exp.DebugInfo(), res.DebugInfo())

	exp = skipgraph.NewComparison(*crEqual, &id1, &id1, uint32(len(id1)-1))
	res = id1.Compare(&id1)
	require.Equal(t, exp.DebugInfo(), res.DebugInfo())

	exp = skipgraph.NewComparison(*crEqual, &id2, &id2, uint32(len(id2)-1))
	res = id2.Compare(&id2)
	require.Equal(t, exp.DebugInfo(), res.DebugInfo())

	// id0 < id1
	comp := id0.Compare(&id1)
	require.Equal(t, skipgraph.CompareLess, comp.GetComparisonResult())
	require.Equal(t, "00 < 7f (at byte 0)", comp.DebugInfo())
	require.Equal(t, uint32(0), comp.GetDiffIndex())

	comp = id1.Compare(&id0)
	require.Equal(t, skipgraph.CompareGreater, comp.GetComparisonResult())
	require.Equal(t, "7f > 00 (at byte 0)", comp.DebugInfo())
	require.Equal(t, uint32(0), comp.GetDiffIndex())

	// id1 < id2
	comp = id1.Compare(&id2)
	require.Equal(t, skipgraph.CompareLess, comp.GetComparisonResult())
	require.Equal(t, "7f < ff (at byte 0)", comp.DebugInfo())
	require.Equal(t, uint32(0), comp.GetDiffIndex())

	comp = id2.Compare(&id1)
	require.Equal(t, skipgraph.CompareGreater, comp.GetComparisonResult())
	require.Equal(t, "ff > 7f (at byte 0)", comp.DebugInfo())
	require.Equal(t, uint32(0), comp.GetDiffIndex())

	// id0 < id2
	comp = id0.Compare(&id2)
	require.Equal(t, skipgraph.CompareLess, comp.GetComparisonResult())
	require.Equal(t, "00 < ff (at byte 0)", comp.DebugInfo())
	require.Equal(t, uint32(0), comp.GetDiffIndex())

	comp = id2.Compare(&id0)
	require.Equal(t, skipgraph.CompareGreater, comp.GetComparisonResult())
	require.Equal(t, "ff > 00 (at byte 0)", comp.DebugInfo())
	require.Equal(t, uint32(0), comp.GetDiffIndex())

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
	c := idRandomGreater.Compare(&idRandomGreater)
	require.Equal(t, skipgraph.CompareEqual, c.GetComparisonResult())
	c = idRandomLess.Compare(&idRandomLess)
	require.Equal(t, skipgraph.CompareEqual, c.GetComparisonResult())

	comp = idRandomGreater.Compare(&idRandomLess)
	require.Equal(t, skipgraph.CompareGreater, comp.GetComparisonResult())
	require.Equal(t, uint32(differingByteIndex), comp.GetDiffIndex())
	expectedDebugInfo := fmt.Sprintf("%s > %s (at byte %d)", hex.EncodeToString(idRandomGreater[:differingByteIndex+1]), hex.EncodeToString(idRandomLess[:differingByteIndex+1]), differingByteIndex)
	require.Equal(t, expectedDebugInfo, comp.DebugInfo())

	comp = idRandomLess.Compare(&idRandomGreater)
	require.Equal(t, skipgraph.CompareLess, comp.GetComparisonResult())
	require.Equal(t, uint32(differingByteIndex), comp.GetDiffIndex())
	expectedDebugInfo = fmt.Sprintf("%s < %s (at byte %d)", hex.EncodeToString(idRandomLess[:differingByteIndex+1]), hex.EncodeToString(idRandomGreater[:differingByteIndex+1]), differingByteIndex)
	require.Equal(t, expectedDebugInfo, comp.DebugInfo())
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
