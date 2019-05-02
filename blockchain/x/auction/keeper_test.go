package auction

import (
	"testing"
	"github.com/stretchr/testify/require"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"	
)

func TestKeeper_SetGetDeleteAuction(t *testing.T) {
	// setup keeper, create auction
	mapp, keeper, addresses, _ := setUpMockApp()
	mapp.BeginBlock(abci.RequestBeginBlock{}) // Without this it panics about "invalid memory address or nil pointer dereference"
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	auction, _ := NewForwardAuction(addresses[0], sdk.NewInt64Coin("usdx", 100), sdk.NewInt64Coin("xrs", 0), endTime(1000))
	id := auctionID(5)
	auction.SetID(id)

	// write and read from store
	keeper.setAuction(ctx, &auction)
	readAuction, found := keeper.getAuction(ctx, id)

	// check before and after match
	require.True(t, found)
	require.Equal(t, &auction, readAuction)
	// check auction is in queue
	iter := keeper.getQueueIterator(ctx, 100000)
	require.Equal(t, 1, len(convertIteratorToSlice(keeper, iter)))
	iter.Close()

	// delete auction
	keeper.deleteAuction(ctx, id)
	
	// check auction does not exist
	_, found = keeper.getAuction(ctx, id)
	require.False(t, found)
	// check auction not in queue
	iter = keeper.getQueueIterator(ctx, 100000)
	require.Equal(t, 0, len(convertIteratorToSlice(keeper, iter)))
	iter.Close()

}

// TODO convert to table driven test with more test cases
func TestKeeper_ExpiredAuctionQueue(t *testing.T) {
	// setup keeper
	mapp, keeper, _, _ := setUpMockApp()
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	// create an example queue
	type queue []struct {
		endTime   endTime
		auctionID auctionID
	}
	q := queue{{1000, 0}, {1300, 2}, {5200, 1}}

	// write and read queue
	for _, v := range q {
		keeper.insertIntoQueue(ctx, v.endTime, v.auctionID)
	}
	iter := keeper.getQueueIterator(ctx, 1000)

	// check before and after match
	i := 0
	for ; iter.Valid(); iter.Next() {
		var auctionID auctionID
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(iter.Value(), &auctionID)
		require.Equal(t, q[i].auctionID, auctionID)
		i++
	}

}

func convertIteratorToSlice(keeper Keeper, iterator sdk.Iterator) []auctionID {
	var queue []auctionID
	for ; iterator.Valid(); iterator.Next() {
		var auctionID auctionID
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &auctionID)
		queue = append(queue, auctionID)
	}
	return queue
}