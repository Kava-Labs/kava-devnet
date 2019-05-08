package cdp

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestApp_CreateModifyDeleteCDP(t *testing.T) {
	// Setup
	testAddr, _, testPrivKey := generateAccAddress()
	mapp, keeper := setUpMockAppWithoutGenesis()

	genAcc := auth.BaseAccount{
		Address: testAddr,
		Coins:   sdk.Coins{c("xrp", 100)},
	}
	mock.SetGenesis(mapp, []auth.Account{&genAcc})
	// setup pricefeed
	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	keeper.pricefeed.SetPrice(ctx, sdk.MustNewDecFromStr("1.00"))
	mapp.EndBlock(abci.RequestEndBlock{})
	mapp.Commit()

	// Create CDP
	msgs := []sdk.Msg{NewMsgCreateOrModifyCDP(testAddr, "xrp", i(10), i(5))}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, msgs, []uint64{0}, []uint64{0}, true, true, testPrivKey)

	mock.CheckBalance(t, mapp, testAddr, sdk.Coins{c(StableDenom, 5), c("xrp", 90)})

	// Modify CDP
	msgs = []sdk.Msg{NewMsgCreateOrModifyCDP(testAddr, "xrp", i(40), i(5))}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, msgs, []uint64{0}, []uint64{1}, true, true, testPrivKey)

	mock.CheckBalance(t, mapp, testAddr, sdk.Coins{c(StableDenom, 10), c("xrp", 50)})

	// Delete CDP
	msgs = []sdk.Msg{NewMsgCreateOrModifyCDP(testAddr, "xrp", i(-50), i(-10))}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, msgs, []uint64{0}, []uint64{2}, true, true, testPrivKey)

	mock.CheckBalance(t, mapp, testAddr, sdk.Coins{c("xrp", 100)})
}
