package liquidator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/usdx/blockchain/x/auction"
	"github.com/kava-labs/usdx/blockchain/x/cdp"
	"github.com/kava-labs/usdx/blockchain/x/pricefeed"
)

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
