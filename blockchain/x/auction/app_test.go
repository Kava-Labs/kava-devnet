package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	//"testing"
)

// // TestApp is a basic integration test of creating an auction, placing a bid, and the auction closing.
// func TestApp(t *testing.T) {
// 	// setup
// 	mapp, keeper, addresses, privKeys := setUpMockApp()
// 	seller := addresses[0]
// 	//sellerKey := privKeys[0]
// 	buyer := addresses[1]
// 	buyerKey := privKeys[1]

// 	// create auction
// 	mapp.BeginBlock(abci.RequestBeginBlock{})                                                              // TODO why is this needed. Seem to get a panic about "invalid memory address or nil pointer dereference" otherwise
// 	ctx := mapp.BaseApp.NewContext(false, abci.Header{})                                                   // make sure first arg is false, otherwise no db writes
// 	keeper.StartForwardAuction(ctx, seller, sdk.NewInt64Coin("token1", 20), sdk.NewInt64Coin("token2", 0)) // lot, initBid
// 	// check seller's coins decreased
// 	mock.CheckBalance(t, mapp, seller, sdk.Coins{sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 100)})

// 	// submit bid tx
// 	msgs := []sdk.Msg{NewMsgPlaceBid(0, buyer, sdk.NewInt64Coin("token2", 10), sdk.NewInt64Coin("token1", 20))} // bid, lot
// 	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, msgs, []uint64{0}, []uint64{0}, true, true, buyerKey)
// 	// check buyer's coins decreased
// 	mock.CheckBalance(t, mapp, buyer, sdk.Coins{sdk.NewInt64Coin("token1", 100), sdk.NewInt64Coin("token2", 90)})
// 	// check seller's coins increased
// 	mock.CheckBalance(t, mapp, seller, sdk.Coins{sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 110)})

// 	// trigger endBlocker by delivering an empty block at the right height
// 	// TODO could probably just run endBlocker creating a context like this: ctx.WithBlockHeight(int64)
// 	mapp.BeginBlock((abci.RequestBeginBlock{Header: abci.Header{Height: 10}}))
// 	mapp.EndBlock((abci.RequestEndBlock{Height: 10}))
// 	mapp.Commit()
// 	// check buyer's coins increased
// 	mock.CheckBalance(t, mapp, buyer, sdk.Coins{sdk.NewInt64Coin("token1", 120), sdk.NewInt64Coin("token2", 90)})
// }

func setUpMockApp() (*mock.App, Keeper, []sdk.AccAddress, []crypto.PrivKey) {
	// Create uninitialized mock app
	mapp := mock.NewApp()

	// Register codecs
	RegisterCodec(mapp.Cdc)

	// Create keepers
	keyAuction := sdk.NewKVStoreKey("auction")
	bankKeeper := bank.NewBaseKeeper(mapp.AccountKeeper, mapp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	auctionKeeper := NewKeeper(mapp.Cdc, bankKeeper, keyAuction)

	// Register routes
	mapp.Router().AddRoute("auction", NewHandler(auctionKeeper))

	// Add endblocker
	mapp.SetEndBlocker(
		func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
			tags := EndBlocker(ctx, auctionKeeper)
			return abci.ResponseEndBlock{
				Tags: tags,
			}
		},
	)
	// Mount and load the stores
	err := mapp.CompleteSetup(keyAuction)
	if err != nil {
		panic("mock app setup failed")
	}

	// Create 2 pre-funded accounts to use for tests
	genAccs, addrs, _, privKeys := mock.CreateGenAccounts(2, sdk.Coins{sdk.NewInt64Coin("token1", 100), sdk.NewInt64Coin("token2", 100)})
	mock.SetGenesis(mapp, genAccs)

	return mapp, auctionKeeper, addrs, privKeys

}
