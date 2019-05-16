package liquidator // TODO change to liquidator_test package?

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/usdx/blockchain/x/auction"
	"github.com/kava-labs/usdx/blockchain/x/cdp"
	"github.com/kava-labs/usdx/blockchain/x/pricefeed"
)

func TestApp(t *testing.T) {
	// Setup mock app with 4 accounts, 30 BTC each
	// 0 - risky CDP owner, 1 - liquidizer and initial bider, 2 - final bidder, 3 - oracle
	mapp, _ := setUpMockAppWithoutGenesis()
	genAccs, addrs, _, privKeys := mock.CreateGenAccounts(4, cs(c("btc", 30)))
	mock.SetGenesis(mapp, genAccs)
	// Set max bid high to make this test easier TODO remove with more advanced tests
	originalMaxBid := CollateralAuctionMaxBid
	CollateralAuctionMaxBid = i(100000)
	defer (func() {
		CollateralAuctionMaxBid = originalMaxBid // reset to avoid messing up other tests
	})()

	// Set the current price @ 8k $/BTC
	// pricefeed assets were added in genesis, post price message, leave to endblocker for price to be set
	msgs := []sdk.Msg{pricefeed.NewMsgPostPrice(addrs[3], "btc", sdk.MustNewDecFromStr("8000.00"), i(9999999999999))} // long expiry
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{3}, []uint64{0}, true, true, privKeys[3])

	// Create CDP and withdraw maximum stable coin (liquidation ratio 1.5)
	msgs = []sdk.Msg{cdp.NewMsgCreateOrModifyCDP(addrs[0], "btc", i(9), i(48000))}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{0}, []uint64{0}, true, true, privKeys[0])
	// Check balance
	mock.CheckBalance(t, mapp, addrs[0], cs(c("btc", 21), c(cdp.StableDenom, 48000)))

	// Give other accounts some stable coin
	msgs = []sdk.Msg{
		cdp.NewMsgCreateOrModifyCDP(addrs[1], "btc", i(30), i(100000)), // addrs[1] creates a very well collateralized CDP
		bank.NewMsgSend(addrs[1], addrs[2], cs(c(cdp.StableDenom, 50000))), // transfer some stable coin to addrs[2]
	}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{1}, []uint64{0}, true, true, privKeys[1])

	// Reduce price by 1/4 to 6k $/BTC
	msgs = []sdk.Msg{pricefeed.NewMsgPostPrice(addrs[3], "btc", sdk.MustNewDecFromStr("6000.00"), i(9999999999999))}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{3}, []uint64{1}, true, true, privKeys[3])

	// Liquidate CDP and start collateral auction
	msgs = []sdk.Msg{MsgSeizeAndStartCollateralAuction{addrs[1], addrs[0], "btc"}}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{1}, []uint64{1}, true, true, privKeys[1])
	// TODO check

	// Place couple of bids, auction end
	msgs = []sdk.Msg{auction.NewMsgPlaceBid(auction.ID(0), addrs[1], c(cdp.StableDenom, 10), c("btc", 9))}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{1}, []uint64{2}, true, true, privKeys[1])

	msgs = []sdk.Msg{auction.NewMsgPlaceBid(auction.ID(0), addrs[2], c(cdp.StableDenom, 48000), c("btc", 8))}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, nextHeader(mapp), msgs, []uint64{2}, []uint64{0}, true, true, privKeys[2])

	// wait for end
	auctionEndTime := mapp.LastBlockHeight() + int64(auction.BidDuration) // TODO get from auction itself
	for i := int64(0); i < auctionEndTime; i++ {
		deliverEmptyBlock(mapp)	
	}

	// check balances
	mock.CheckBalance(t, mapp, addrs[0], cs(c("btc", 22), c(cdp.StableDenom, 48000))) // risky cdp owner
	mock.CheckBalance(t, mapp, addrs[1], cs(c("btc", 0), c(cdp.StableDenom, 50000))) // initial bidder
	mock.CheckBalance(t, mapp, addrs[2], cs(c("btc", 38), c(cdp.StableDenom, 2000))) // winning bidder
	// check debt and collateral

}

func nextHeader(mapp *mock.App) abci.Header {
	return abci.Header{Height: mapp.LastBlockHeight() + 1}
}

func deliverEmptyBlock(mapp *mock.App) {
	mapp.BeginBlock(abci.RequestBeginBlock{Header: nextHeader(mapp)})
	mapp.EndBlock(abci.RequestEndBlock{})
	mapp.Commit()
}
