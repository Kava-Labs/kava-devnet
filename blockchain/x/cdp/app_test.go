package cdp

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/kava-labs/kava-devnet/blockchain/x/pricefeed"
)

func TestApp_CreateModifyDeleteCDP(t *testing.T) {
	// Setup
	mapp, keeper, pfKeeper := setUpMockAppWithoutGenesis()
	genAccs, addrs, _, privKeys := mock.CreateGenAccounts(1, cs(c("xrp", 100)))
	testAddr := addrs[0]
	testPrivKey := privKeys[0]
	mock.SetGenesis(mapp, genAccs)
	mock.CheckBalance(t, mapp, testAddr, cs(c("xrp", 100)))
	// setup pricefeed, TODO can this be shortened a bit?
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, header)
	params := CdpParams{
		GlobalDebtLimit: sdk.NewInt(100000),
		CollateralParams: []CollateralParams{
			{
				Denom:            "xrp",
				LiquidationRatio: sdk.MustNewDecFromStr("1.5"),
				DebtLimit:        sdk.NewInt(10000),
			},
		},
		StableDenoms: []string{"usdx"},
	}
	keeper.SetParams(ctx, params)
	keeper.SetGlobalDebt(ctx, sdk.NewInt(0))
	ap := pricefeed.AssetParams{
		Assets: []pricefeed.Asset{pricefeed.Asset{AssetCode: "xrp", Description: ""}},
	}
	pfKeeper.SetAssetParams(ctx, ap)
	pfKeeper.SetPrice(
		ctx, sdk.AccAddress{}, "xrp",
		sdk.MustNewDecFromStr("1.00"),
		sdk.NewInt(10))
	pfKeeper.SetCurrentPrices(ctx)
	mapp.EndBlock(abci.RequestEndBlock{})
	mapp.Commit()

	// Create CDP
	msgs := []sdk.Msg{NewMsgCreateOrModifyCDP(testAddr, "xrp", i(10), i(5))}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, abci.Header{Height: mapp.LastBlockHeight() + 1}, msgs, []uint64{0}, []uint64{0}, true, true, testPrivKey)

	mock.CheckBalance(t, mapp, testAddr, cs(c("usdx", 5), c("xrp", 90)))

	// Modify CDP
	msgs = []sdk.Msg{NewMsgCreateOrModifyCDP(testAddr, "xrp", i(40), i(5))}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, abci.Header{Height: mapp.LastBlockHeight() + 1}, msgs, []uint64{0}, []uint64{1}, true, true, testPrivKey)

	mock.CheckBalance(t, mapp, testAddr, cs(c("usdx", 10), c("xrp", 50)))

	// Delete CDP
	msgs = []sdk.Msg{NewMsgCreateOrModifyCDP(testAddr, "xrp", i(-50), i(-10))}
	mock.SignCheckDeliver(t, mapp.Cdc, mapp.BaseApp, abci.Header{Height: mapp.LastBlockHeight() + 1}, msgs, []uint64{0}, []uint64{2}, true, true, testPrivKey)

	mock.CheckBalance(t, mapp, testAddr, cs(c("xrp", 100)))
}
