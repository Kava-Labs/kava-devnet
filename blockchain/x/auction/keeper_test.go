package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"testing"
)

func TestKeeper_SetGetAuction(t *testing.T) {
	// create auction
	mapp, keeper, addresses, _ := setUpMockApp()
	mapp.BeginBlock(abci.RequestBeginBlock{}) // TODO why is this needed. Seem to get a panic about "invalid memory address or nil pointer dereference" otherwise
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	auction, _ := NewForwardAuction(addresses[0], sdk.NewInt64Coin("usdx", 100), sdk.NewInt64Coin("xrs", 0), endTime(1000))
	auction.SetID(5)

	// write and read form store
	keeper.setAuction(ctx, &auction)
	readAuction, found := keeper.getAuction(ctx, 5)
	
	// check before and after match
	require.True(t, found)
	require.Equal(t, &auction, readAuction)
}
