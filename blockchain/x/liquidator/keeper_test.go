package liquidator

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/usdx/blockchain/x/auction"
	"github.com/kava-labs/usdx/blockchain/x/cdp"
	"github.com/kava-labs/usdx/blockchain/x/pricefeed"
)
// TODO These tests get a bit messy to setup because liquidator depends on other modules that need setting up

func TestKeeper_StartCollateralAuction(t *testing.T) {
	// Setup keeper and context
	mapp, keeper := setUpMockAppWithoutGenesis()
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	// Create seized CDP
	_, addrs := mock.GeneratePrivKeyAddressPairs(1)
	testAddr := addrs[0]
	cdp := SeizedCDP{Owner: testAddr, CollateralAmount: i(10), CollateralDenom: "btc", Debt: i(500)}
	keeper.setSeizedCDP(ctx, cdp)
	keeper.bankKeeper.AddCoins(ctx, keeper.cdpKeeper.GetLiquidatorAccountAddress(), cs(sdk.NewCoin(cdp.CollateralDenom, cdp.CollateralAmount)))

	err := keeper.StartCollateralAuction(ctx, cdp.Owner, cdp.CollateralDenom)

	require.Nil(t, err)
	// Check CDP is changed correctly
	_, found := keeper.GetSeizedCDP(ctx, cdp.Owner, cdp.CollateralDenom)
	require.False(t, found)
	//require.Equal(t, SeizedCDP{Owner: testAddr, CollateralAmount: i(0), CollateralDenom: "btc", Debt: i(0)}, modifiedCDP)
	// TODO Check Auction was started by using mocked auction keeper?
}

func TestKeeper_settleDebt(t *testing.T) {
	// Setup keeper and context
	mapp, keeper := setUpMockAppWithoutGenesis()
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	// create debt
	debt := sdk.NewInt(528452456344)
	keeper.setSeizedDebt(ctx, debt)
	// create stable coins
	keeper.bankKeeper.AddCoins(ctx, keeper.cdpKeeper.GetLiquidatorAccountAddress(), cs(sdk.NewCoin(keeper.cdpKeeper.GetStableDenom(), debt)))

	// cancel out debt
	keeper.settleDebt(ctx)

	// Check there is no more debt or stable coins
	require.Equal(t, i(0), keeper.GetSeizedDebt(ctx))
	require.Equal(t, i(0), keeper.bankKeeper.GetCoins(ctx, keeper.cdpKeeper.GetLiquidatorAccountAddress()).AmountOf(keeper.cdpKeeper.GetStableDenom()))
}
func TestKeeper_GetSetSeizedDebt(t *testing.T) {
	// Setup keeper and context
	mapp, keeper := setUpMockAppWithoutGenesis()
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

func setUpMockAppWithoutGenesis() (*mock.App, Keeper) {
	// Create uninitialized mock app
	mapp := mock.NewApp()

	// Register codecs
	RegisterCodec(mapp.Cdc)

	// Create keepers
	keyPriceFeed := sdk.NewKVStoreKey(pricefeed.StoreKey)
	keyCDP := sdk.NewKVStoreKey("cdp")
	keyAuction := sdk.NewKVStoreKey("auction")
	keyLiquidator := sdk.NewKVStoreKey("liquidator")

	priceFeedKeeper := pricefeed.NewKeeper(keyPriceFeed, mapp.Cdc, pricefeed.DefaultCodespace)
	bankKeeper := bank.NewBaseKeeper(mapp.AccountKeeper, mapp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	cdpKeeper := cdp.NewKeeper(mapp.Cdc, keyCDP, mapp.ParamsKeeper.Subspace("cdpSubspace"), priceFeedKeeper, bankKeeper)
	auctionKeeper := auction.NewKeeper(mapp.Cdc, cdpKeeper, keyAuction)                         // Note: cdp keeper stands in for bank keeper
	liquidatorKeeper := NewKeeper(mapp.Cdc, keyLiquidator, cdpKeeper, auctionKeeper, cdpKeeper) // Note: cdp keeper stands in for bank keeper

	// Register routes
	mapp.Router().AddRoute("liquidator", NewHandler(liquidatorKeeper))

	mapp.SetInitChainer(
		func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
			res := mapp.InitChainer(ctx, req)
			cdp.InitGenesis(ctx, cdpKeeper, cdp.DefaultGenesisState())
			return res
		},
	)

	// Mount and load the stores
	err := mapp.CompleteSetup(keyPriceFeed, keyCDP, keyAuction, keyLiquidator)
	if err != nil {
		panic(err)
	}

	return mapp, liquidatorKeeper
}

// Avoid cluttering test cases with long function name
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
