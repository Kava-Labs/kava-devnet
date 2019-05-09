package pricefeed

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

type testHelper struct {
	mApp     *mock.App
	keeper   Keeper
	addrs    []sdk.AccAddress
	pubKeys  []crypto.PubKey
	privKeys []crypto.PrivKey
}

func getMockApp(t *testing.T, numGenAccs int, genState GenesisState, genAccs []auth.Account) testHelper {
	mApp := mock.NewApp()
	RegisterCodec(mApp.Cdc)
	keyPricefeed := sdk.NewKVStoreKey("pricefeed")
	keeper := NewKeeper(keyPricefeed, mApp.Cdc, DefaultCodespace)

	// Register routes
	mApp.Router().AddRoute(RouterKey, NewHandler(keeper))

	// Add endblocker
	mApp.SetEndBlocker(
		func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
			tags := EndBlocker(ctx, keeper)
			return abci.ResponseEndBlock{
				Tags: tags,
			}
		},
	)

	require.NoError(t, mApp.CompleteSetup(keyPricefeed))

	valTokens := sdk.TokensFromTendermintPower(42)
	var (
		addrs    []sdk.AccAddress
		pubKeys  []crypto.PubKey
		privKeys []crypto.PrivKey
	)

	if genAccs == nil || len(genAccs) == 0 {
		genAccs, addrs, pubKeys, privKeys = mock.CreateGenAccounts(numGenAccs,
			sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, valTokens)})
	}

	mock.SetGenesis(mApp, genAccs)
	return testHelper{mApp, keeper, addrs, pubKeys, privKeys}
}
