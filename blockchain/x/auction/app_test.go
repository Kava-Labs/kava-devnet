package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"testing"
)

// TestApp contans several basic integration tests of creating an auction, placing a bid, and the auction closing.

func TestApp_ForwardAuction(t *testing.T) {
	// Setup
	mapp, keeper, addresses, privKeys := setUpMockApp()
	seller := addresses[0]
	//sellerKey := privKeys[0]
	buyer := addresses[1]
	buyerKey := privKeys[1]

	// Create a block where an auction is started
	mapp.BeginBlock(abci.RequestBeginBlock{})                                                              
	ctx := mapp.BaseApp.NewContext(false, abci.Header{}) // make sure first arg is false, otherwise no db writes
	keeper.StartForwardAuction(ctx, seller, sdk.NewInt64Coin("token1", 20), sdk.NewInt64Coin("token2", 0)) // lot, initialBid
	mapp.EndBlock(abci.RequestEndBlock{})
	mapp.Commit()

	// Check seller's coins have decreased
	mock.CheckBalance(t, mapp, seller, sdk.Coins{sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 100)})

	// Deliver a block that contains a PlaceBid tx
	msgs := []sdk.Msg{NewMsgPlaceBid(0, buyer, sdk.NewInt64Coin("token2", 10), sdk.NewInt64Coin("token1", 20))} // bid, lot
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, msgs, []uint64{0}, []uint64{0}, true, true, buyerKey)
	
	// Check buyer's coins have decreased
	mock.CheckBalance(t, mapp, buyer, sdk.Coins{sdk.NewInt64Coin("token1", 100), sdk.NewInt64Coin("token2", 90)})
	// Check seller's coins have increased
	mock.CheckBalance(t, mapp, seller, sdk.Coins{sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 110)})

	// Deliver an empty block with high blockheight to trigger the auction to close
	h := int64(bidDuration+100)
	mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: h}})
	mapp.EndBlock(abci.RequestEndBlock{Height: h})
	mapp.Commit()

	// Check buyer's coins increased
	mock.CheckBalance(t, mapp, buyer, sdk.Coins{sdk.NewInt64Coin("token1", 120), sdk.NewInt64Coin("token2", 90)})
}

func TestApp_ReverseAuction(t *testing.T) {
	// Setup
	mapp, keeper, addresses, privKeys := setUpMockApp()
	seller := addresses[0]
	sellerKey := privKeys[0]
	buyer := addresses[1]
	//buyerKey := privKeys[1]

	// Create a block where an auction is started
	mapp.BeginBlock(abci.RequestBeginBlock{})                                                              
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	keeper.StartReverseAuction(ctx, buyer, sdk.NewInt64Coin("token1", 20), sdk.NewInt64Coin("token2", 99)) // buyer, bid, initialLot
	mapp.EndBlock(abci.RequestEndBlock{})
	mapp.Commit()

	// Check buyer's coins have decreased
	mock.CheckBalance(t, mapp, buyer, sdk.Coins{sdk.NewInt64Coin("token1", 100), sdk.NewInt64Coin("token2", 1)})

	// Deliver a block that contains a PlaceBid tx
	msgs := []sdk.Msg{NewMsgPlaceBid(0, seller, sdk.NewInt64Coin("token1", 20), sdk.NewInt64Coin("token2", 10))} // bid, lot
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, msgs, []uint64{0}, []uint64{0}, true, true, sellerKey)
	
	// Check seller's coins have decreased
	mock.CheckBalance(t, mapp, seller, sdk.Coins{sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 100)})
	// Check buyer's coins have increased
	mock.CheckBalance(t, mapp, buyer, sdk.Coins{sdk.NewInt64Coin("token1", 120), sdk.NewInt64Coin("token2", 90)})

	// Deliver an empty block with high blockheight to trigger the auction to close
	h := int64(bidDuration+100)
	mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: h}})
	mapp.EndBlock(abci.RequestEndBlock{Height: h})
	mapp.Commit()

	// Check seller's coins increased
	mock.CheckBalance(t, mapp, seller, sdk.Coins{sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 110)})
}
func TestApp_ForwardReverseAuction(t *testing.T) {
	// Setup
	mapp, keeper, addresses, privKeys := setUpMockApp()
	seller := addresses[0]
	//sellerKey := privKeys[0]
	buyer := addresses[1]
	buyerKey := privKeys[1]
	recipient := addresses[2]

	// Create a block where an auction is started
	mapp.BeginBlock(abci.RequestBeginBlock{})                                                              
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	keeper.StartForwardReverseAuction(ctx, seller, sdk.NewInt64Coin("token1", 20), sdk.NewInt64Coin("token2", 50), recipient) // seller, lot, maxBid, otherPerson
	mapp.EndBlock(abci.RequestEndBlock{})
	mapp.Commit()

	// Check seller's coins have decreased
	mock.CheckBalance(t, mapp, seller, sdk.Coins{sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 100)})

	// Deliver a block that contains a PlaceBid tx
	msgs := []sdk.Msg{NewMsgPlaceBid(0, buyer, sdk.NewInt64Coin("token2", 50), sdk.NewInt64Coin("token1", 15))} // bid, lot
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, msgs, []uint64{0}, []uint64{0}, true, true, buyerKey)
	
	// Check bidder's coins have decreased
	mock.CheckBalance(t, mapp, buyer, sdk.Coins{sdk.NewInt64Coin("token1", 100), sdk.NewInt64Coin("token2", 50)})
	// Check seller's coins have increased
	mock.CheckBalance(t, mapp, seller, sdk.Coins{sdk.NewInt64Coin("token1", 80), sdk.NewInt64Coin("token2", 150)})
	// Check "recipient" has received coins
	mock.CheckBalance(t, mapp, recipient, sdk.Coins{sdk.NewInt64Coin("token1",105), sdk.NewInt64Coin("token2", 100)})

	// Deliver an empty block with high blockheight to trigger the auction to close
	h := int64(bidDuration+100)
	mapp.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{Height: h}})
	mapp.EndBlock(abci.RequestEndBlock{Height: h})
	mapp.Commit()

	// Check buyer's coins increased
	mock.CheckBalance(t, mapp, buyer, sdk.Coins{sdk.NewInt64Coin("token1", 115), sdk.NewInt64Coin("token2", 50)})
}

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

	// Create a bunch (ie 10) of pre-funded accounts to use for tests
	genAccs, addrs, _, privKeys := mock.CreateGenAccounts(10, sdk.Coins{sdk.NewInt64Coin("token1", 100), sdk.NewInt64Coin("token2", 100)})
	mock.SetGenesis(mapp, genAccs)

	return mapp, auctionKeeper, addrs, privKeys
}
