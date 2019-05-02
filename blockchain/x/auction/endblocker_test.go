package auction

import (
	"testing"
	"github.com/stretchr/testify/require"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"	
)

func TestKeeper_EndBlocker(t *testing.T) {
	// setup keeper and auction
	mapp, keeper, addresses, _ := setUpMockApp()
	mapp.BeginBlock(abci.RequestBeginBlock{}) 
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	seller := addresses[0]
	keeper.StartForwardAuction(ctx, seller, sdk.NewInt64Coin("token1", 20), sdk.NewInt64Coin("token2", 0))
	
	// run the endblocker, simulating a block height after auction expiry
	EndBlocker(ctx.WithBlockHeight(int64(maxAuctionDuration)+1), keeper)

	// check auction has been closed
	_, found := keeper.getAuction(ctx, 0)
	require.False(t, found)

}