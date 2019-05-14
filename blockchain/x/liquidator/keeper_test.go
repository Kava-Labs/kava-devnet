package liquidator

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)
// TODO These tests get a bit messy to setup because liquidator depends on other modules that need setting up

func TestKeeper_StartCollateralAuction(t *testing.T) {
	// Setup keeper and context
	mapp, keeper := setUpMockAppWithoutGenesis()
	mock.SetGenesis(mapp, []auth.Account{}) // TODO maybe move into mock app creation, if no gen accounts are needed
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	// Create seized CDP
	_, addrs := mock.GeneratePrivKeyAddressPairs(1)
	testAddr := addrs[0]
	cdp := SeizedCDP{Owner: testAddr, CollateralAmount: i(10), CollateralDenom: "btc", Debt: i(500)}
	keeper.setSeizedCDP(ctx, cdp)
	keeper.bankKeeper.AddCoins(ctx, keeper.cdpKeeper.GetLiquidatorAccountAddress(), cs(sdk.NewCoin(cdp.CollateralDenom, cdp.CollateralAmount)))

	// Start auction
	_, err := keeper.StartCollateralAuction(ctx, cdp.Owner, cdp.CollateralDenom)

	// Check CDP is changed correctly
	require.Nil(t, err)
	_, found := keeper.GetSeizedCDP(ctx, cdp.Owner, cdp.CollateralDenom)
	require.False(t, found)
	//require.Equal(t, SeizedCDP{Owner: testAddr, CollateralAmount: i(0), CollateralDenom: "btc", Debt: i(0)}, modifiedCDP)
	// TODO Check Auction was started by using mocked auction keeper?
}

// func TestKeeper_settleDebt(t *testing.T) {
// 	// Setup keeper and context
// 	mapp, keeper := setUpMockAppWithoutGenesis()
// 	mock.SetGenesis(mapp, []auth.Account{})
// 	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
// 	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
// 	ctx := mapp.BaseApp.NewContext(false, header)
// 	// create debt
// 	debt := sdk.NewInt(528452456344)
// 	keeper.setSeizedDebt(ctx, debt)
// 	keeper.cdpKeeper.setGlobalDebt(ctx, debt) // TODO this won't work. Setting genesis would be better.
// 	// create stable coins
// 	keeper.bankKeeper.AddCoins(ctx, keeper.cdpKeeper.GetLiquidatorAccountAddress(), cs(sdk.NewCoin(keeper.cdpKeeper.GetStableDenom(), debt)))

// 	// cancel out debt
// 	err := keeper.settleDebt(ctx)

// 	require.Nil(t, err)
// 	// Check there is no more debt or stable coins
// 	require.Equal(t, i(0), keeper.GetSeizedDebt(ctx))
// 	require.Equal(t, i(0), keeper.bankKeeper.GetCoins(ctx, keeper.cdpKeeper.GetLiquidatorAccountAddress()).AmountOf(keeper.cdpKeeper.GetStableDenom()))
// }
func TestKeeper_GetSetSeizedDebt(t *testing.T) {
	// Setup keeper and context
	mapp, keeper := setUpMockAppWithoutGenesis()
	mock.SetGenesis(mapp, []auth.Account{})
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	// create example debt
	debt := sdk.NewInt(528452456344)

	// Run test function
	keeper.setSeizedDebt(ctx, debt)
	readDebt := keeper.GetSeizedDebt(ctx)

	// check
	require.Equal(t, debt, readDebt)
}