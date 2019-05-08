package cdp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/kava-labs/usdx/blockchain/x/cdp/mockpricefeed"
)

// Mock app is an ABCI app with an in memory database.
// This function creates an app, setting up the keepers, routes, begin and end blockers.
// But leaves it to the tests to call InitChain (done by calling mock.SetGenesis)
// The app works by submitting ABCI messages.
//  - InitChain sets up the app db from genesis.
//  - BeginBlock starts the delivery of a new block
//  - DeliverTx delivers a tx
//  - EndBlock signals the end of a block
//  - Commit ?
func setUpMockAppWithoutGenesis() (*mock.App, Keeper) {
	// Create uninitialized mock app
	mapp := mock.NewApp()

	// Register codecs
	RegisterCodec(mapp.Cdc)

	// Create keepers
	keyCDP := sdk.NewKVStoreKey("cdp")
	keyPriceFeed := sdk.NewKVStoreKey("pricefeed")
	priceFeedKeeper := mockpricefeed.NewKeeper(mapp.Cdc, keyPriceFeed)
	bankKeeper := bank.NewBaseKeeper(mapp.AccountKeeper, mapp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	cdpKeeper := NewKeeper(mapp.Cdc, keyCDP, mapp.ParamsKeeper.Subspace("cdpSubspace"), priceFeedKeeper, bankKeeper)

	// Register routes
	mapp.Router().AddRoute("cdp", NewHandler(cdpKeeper))

	mapp.SetInitChainer(
		func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
			res := mapp.InitChainer(ctx, req)
			InitGenesis(ctx, cdpKeeper, DefaultGenesisState()) // Create a default genesis state, then set the keeper store to it
			return res
		},
	)

	// Mount and load the stores
	err := mapp.CompleteSetup(keyPriceFeed, keyCDP)
	if err != nil {
		panic("mock app setup failed")
	}

	return mapp, cdpKeeper
}

// Generate a random account address and its corresponding pub/priv keys
func generateAccAddress() (sdk.AccAddress, crypto.PubKey, crypto.PrivKey) {
	privKey := ed25519.GenPrivKey()
	pubKey := privKey.PubKey()
	addr := sdk.AccAddress(pubKey.Address())
	return addr, pubKey, privKey
}

// Avoid cluttering test cases with long function name
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
