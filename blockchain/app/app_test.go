package app_test // Place tests outside of app package to force usage of public methods

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/kava-labs/usdx/blockchain/app"
	"github.com/kava-labs/usdx/blockchain/x/auction"
	"github.com/kava-labs/usdx/blockchain/x/cdp"
	"github.com/kava-labs/usdx/blockchain/x/liquidator"
	"github.com/kava-labs/usdx/blockchain/x/pricefeed"
)

func TestApp_Basic(t *testing.T) {
	// Setup mock app with 4 accounts, 30 BTC each
	//  0. risky CDP owner
	//  1. liquidizer and initial bider
	//  2. winning bidder
	//  3. price oracle
	mapp := setupApp()
	genAccs, addrs, _, privKeys := mock.CreateGenAccounts(4, cs(c("btc", 30)))
	pricefeedGenState := pricefeed.GenesisState{
		Assets: []pricefeed.Asset{{AssetCode: "btc", Description: ""}},
	}
	genState := app.GenesisState{
		AuthData:      auth.DefaultGenesisState(),
		BankData:      bank.DefaultGenesisState(),
		CdpData:       cdp.DefaultGenesisState(),
		PricefeedData: pricefeedGenState,
		Accounts:      genAccs,
	}
	cdc := app.MakeCodec() // TODO does this need to be `app.cdc` or will this do? Should we export mapp.Cdc ?
	setGenesis(cdc, mapp, genState)

	// Set max bid high to make this test easier
	// TODO remove with more advanced tests
	originalMaxBid := liquidator.CollateralAuctionMaxBid
	liquidator.CollateralAuctionMaxBid = i(100000)
	defer (func() {
		liquidator.CollateralAuctionMaxBid = originalMaxBid // reset to avoid messing up any other tests
	})()

	// Set the current price @ 8k $/BTC
	// pricefeed assets were added in genesis, post price message, leave to endblocker for price to be set
	msgs := []sdk.Msg{pricefeed.NewMsgPostPrice(addrs[3], "btc", sdk.MustNewDecFromStr("8000.00"), i(9999999999999))} // long expiry
	mock.SignCheckDeliver(t, cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{3}, []uint64{0}, true, true, privKeys[3])

	// Create CDP and withdraw maximum stable coin (liquidation ratio 1.5)
	msgs = []sdk.Msg{cdp.NewMsgCreateOrModifyCDP(addrs[0], "btc", i(9), i(48000))}
	mock.SignCheckDeliver(t, cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{0}, []uint64{0}, true, true, privKeys[0])
	// Check balance
	checkBalance(t, cdc, mapp, addrs[0], cs(c("btc", 21), c(cdp.StableDenom, 48000)))

	// Give other accounts some stable coin
	msgs = []sdk.Msg{
		cdp.NewMsgCreateOrModifyCDP(addrs[1], "btc", i(30), i(100000)),     // addrs[1] creates a very well collateralized CDP
		bank.NewMsgSend(addrs[1], addrs[2], cs(c(cdp.StableDenom, 50000))), // transfer some stable coin to addrs[2]
	}
	mock.SignCheckDeliver(t, cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{1}, []uint64{0}, true, true, privKeys[1])

	// Reduce price by 1/4 to 6k $/BTC
	msgs = []sdk.Msg{pricefeed.NewMsgPostPrice(addrs[3], "btc", sdk.MustNewDecFromStr("6000.00"), i(9999999999999))}
	mock.SignCheckDeliver(t, cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{3}, []uint64{1}, true, true, privKeys[3])

	// Liquidate CDP and start collateral auction
	msgs = []sdk.Msg{liquidator.MsgSeizeAndStartCollateralAuction{addrs[1], addrs[0], "btc"}}
	mock.SignCheckDeliver(t, cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{1}, []uint64{1}, true, true, privKeys[1])
	// TODO check auction exists?

	// Place couple of bids, auction end
	msgs = []sdk.Msg{auction.NewMsgPlaceBid(auction.ID(0), addrs[1], c(cdp.StableDenom, 10), c("btc", 9))}
	mock.SignCheckDeliver(t, cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{1}, []uint64{2}, true, true, privKeys[1])

	msgs = []sdk.Msg{auction.NewMsgPlaceBid(auction.ID(0), addrs[2], c(cdp.StableDenom, 48000), c("btc", 8))}
	mock.SignCheckDeliver(t, cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{2}, []uint64{0}, true, true, privKeys[2])

	// wait for end
	auctionEndTime := mapp.LastBlockHeight() + int64(auction.BidDuration) // TODO get from auction itself
	for i := int64(0); i < auctionEndTime; i++ {
		deliverEmptyBlock(mapp) // TODO maybe be away of loading a particular height, see gaiaApp.LoadHeight()
	}

	// check balances
	checkBalance(t, cdc, mapp, addrs[0], cs(c("btc", 22), c(cdp.StableDenom, 48000))) // risky cdp owner
	checkBalance(t, cdc, mapp, addrs[1], cs(c("btc", 0), c(cdp.StableDenom, 50000)))  // initial bidder
	checkBalance(t, cdc, mapp, addrs[2], cs(c("btc", 38), c(cdp.StableDenom, 2000)))  // winning bidder
	// check debt and collateral
	// TODO

}

// ---------- Test Helpers ----------

// setupApp creates a new usdx app, but with a in memory database and no logging
func setupApp() *app.UsdxApp {
	logger := log.NewNopLogger() // set to not print out any logs
	// logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app") // This will print out logs
	db := dbm.NewMemDB()
	return app.NewUsdxApp(logger, db)
}

// setGenesis sends a genesis state to the app through an abci.InitChain request
func setGenesis(cdc *codec.Codec, mapp *app.UsdxApp, genState app.GenesisState) {
	req := abci.RequestInitChain{
		AppStateBytes: cdc.MustMarshalJSON(genState),
	}
	mapp.InitChain(req)
	mapp.Commit()
}

// checkBalance checks if an account has an amount of coins
func checkBalance(t *testing.T, cdc *codec.Codec, app *app.UsdxApp, addr sdk.AccAddress, exp sdk.Coins) {
	route := fmt.Sprintf("custom/%s/%s", auth.QuerierRoute, auth.QueryAccount)
	res := app.Query(abci.RequestQuery{
		Path: route,
		Data: cdc.MustMarshalJSON(auth.QueryAccountParams{Address: addr}),
		// Height: ,
		// Prove: ,
	})
	// res is of type abci.ResponseQuery, use Value method to get response from queriers
	var acc auth.Account
	cdc.MustUnmarshalJSON(res.Value, &acc)

	require.Equal(t, exp, acc.GetCoins())
}

// deliverEmptyBlock sends an empty block to the app to advance the blockheight
// TODO is there a better way of advancing blockheight?
func deliverEmptyBlock(app *app.UsdxApp) { // Ideally app type would be abci.Application, but it doesn't have LastBlockHeight method
	app.BeginBlock(abci.RequestBeginBlock{Header: nextHeader(app)})
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

func nextHeader(app *app.UsdxApp) abci.Header {
	return abci.Header{Height: app.LastBlockHeight() + 1}
}

// Helpers functions to avoid long function names
func i(in int64) sdk.Int                    { return sdk.NewInt(in) }
func c(denom string, amount int64) sdk.Coin { return sdk.NewInt64Coin(denom, amount) }
func cs(coins ...sdk.Coin) sdk.Coins        { return sdk.NewCoins(coins...) }
