package lookup_test

import (
	"github.com/stretchr/testify/require"
	"github.com/thep2p/skipgraph-go/core"
	"github.com/thep2p/skipgraph-go/core/lookup"
	"github.com/thep2p/skipgraph-go/core/model"
	"github.com/thep2p/skipgraph-go/unittest"
	"sync"
	"testing"
	"time"
)

// TestLookupTable_AddEntry test the AddEntry method of LookupTable.
func TestLookupTable_AddEntry(t *testing.T) {
	// create an empty lookup table
	lt := lookup.Table{}

	// create an empty identity
	identity := model.Identity{}

	// add the identity in a valid position
	err := lt.AddEntry(core.LeftDirection, 0, identity)
	require.NoError(t, err)

	// add the identity in a valid position
	err = lt.AddEntry(core.LeftDirection, core.MaxLookupTableLevel-1, identity)
	require.NoError(t, err)

	// add an entry with invalid level
	err = lt.AddEntry(core.LeftDirection, core.MaxLookupTableLevel, identity)
	require.Error(t, err)

	// add an entry with wrong direction
	err = lt.AddEntry(core.Direction("no where"), 0, identity)
	require.Error(t, err)
}

// TestLookupTable_OverWriteLeftEntry test the overwriting of left entry in the lookup table.
func TestLookupTable_OverWriteLeftEntry(t *testing.T) {
	// create an empty lookup table
	lt := lookup.Table{}

	// create a random identity
	identity := unittest.IdentityFixture(t)

	// add the identity in a valid position
	err := lt.AddEntry(core.LeftDirection, 0, identity)
	require.NoError(t, err)

	// create another random identity
	identity1 := unittest.IdentityFixture(t)

	// check the new identity is not equal to the previous one
	require.NotEqual(t, identity1, identity)

	// overwrite the previous entry with the new identity
	err = lt.AddEntry(core.LeftDirection, 0, identity1)
	require.NoError(t, err)

	// check that the new identity has overwritten the previous one
	retIdentity, err := lt.GetEntry(core.LeftDirection, 0)
	require.Equal(t, identity1, retIdentity)
	require.NoError(t, err)
}

// TestLookupTable_OverWriteRightEntry test the overwriting of right entry in the lookup table.
func TestLookupTable_OverWriteRightEntry(t *testing.T) {
	// create an empty lookup table
	lt := lookup.Table{}

	// create a random identity
	identity := unittest.IdentityFixture(t)

	// add the identity in a valid position
	err := lt.AddEntry(core.RightDirection, 0, identity)
	require.NoError(t, err)

	// create another random identity
	identity1 := unittest.IdentityFixture(t)

	// check the new identity is not equal to the previous one
	require.NotEqual(t, identity1, identity)

	// overwrite the previous entry with the new identity
	err = lt.AddEntry(core.RightDirection, 0, identity1)
	require.NoError(t, err)

	// check that the new identity has overwritten the previous one
	retIdentity, err := lt.GetEntry(core.RightDirection, 0)
	require.Equal(t, identity1, retIdentity)
	require.NoError(t, err)
}

// TestLookupTable_GetEntry test the GetEntry method of LookupTable.
func TestLookupTable_GetEntry(t *testing.T) {
	// create an identity
	identity := unittest.IdentityFixture(t)
	identity1 := unittest.IdentityFixture(t)
	require.NotEqual(t, identity1, identity)

	// declare an empty lookup table
	var lt lookup.Table

	// add the identity as a left neighbor into the lookup table
	err := lt.AddEntry(core.LeftDirection, 0, identity)
	require.NoError(t, err)

	// add the identity as a right neighbor into the lookup table
	err = lt.AddEntry(core.RightDirection, 0, identity1)
	require.NoError(t, err)

	// check that the inserted identity is retrievable
	retIdentity, err := lt.GetEntry(core.LeftDirection, 0)
	require.Equal(t, identity, retIdentity)
	require.NoError(t, err)

	// check that the inserted identity is retrievable
	retIdentity1, err := lt.GetEntry(core.RightDirection, 0)
	require.Equal(t, identity1, retIdentity1)
	require.NoError(t, err)

	// access a wrong level
	_, err = lt.GetEntry(core.RightDirection, core.MaxLookupTableLevel)
	require.Error(t, err)

	// access a wrong direction
	_, err = lt.GetEntry(core.Direction("no where"), 0)
	require.Error(t, err)

}

// TestLookupTable_GetEntryConcurrent test the concurrent access to the lookup table.
func TestLookupTable_Concurrency(t *testing.T) {
	// create an empty lookup table
	lt := lookup.Table{}

	// number of items to be added to the lookup table
	addCount := 2
	// number of items to be retrieved from the lookup table
	getCount := 2

	// the number of retrieved items should not exceed the number of added items
	require.LessOrEqual(t, getCount, addCount)

	wg := sync.WaitGroup{}
	wg.Add(addCount + getCount)

	for i := 0; i < addCount; i++ {
		// add some identities concurrently to the lookup table
		i := i
		go func() {
			defer wg.Done()
			identity := unittest.IdentityFixture(t)
			err := lt.AddEntry(core.LeftDirection, core.Level(i), identity)
			require.NoError(t, err)
		}()
	}
	for i := 0; i < getCount; i++ {
		// retrieve some identities concurrently from the lookup table
		i := i
		go func() {
			defer wg.Done()
			_, err := lt.GetEntry(core.LeftDirection, core.Level(i))
			require.NoError(t, err)
		}()
	}

	// check whether all the routines are finished
	// wait 2 milliseconds for each routine to finish
	unittest.CallMustReturnWithinTimeout(
		t,
		wg.Wait,
		time.Duration((getCount+addCount)*2)*time.Millisecond,
		"concurrent access to lookup table failed",
	)
}
