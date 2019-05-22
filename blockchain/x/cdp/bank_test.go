package cdp

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Test the bank functionality of the CDP keeper
func TestKeeper_AddSubtractGetCoins(t *testing.T) {
	_, addrs := mock.GeneratePrivKeyAddressPairs(1)
	normalAddr := addrs[0]

	tests := []struct {
		name          string
		address       sdk.AccAddress
		shouldAdd     bool
		amount        sdk.Coins
		expectedCoins sdk.Coins
	}{
		{"addNormalAddress", normalAddr, true, cs(c(StableDenom, 53)), cs(c(StableDenom, 153), c(GovDenom, 100))},
		{"subNormalAddress", normalAddr, false, cs(c(StableDenom, 53)), cs(c(StableDenom, 47), c(GovDenom, 100))},
		{"addLiquidatorStable", LiquidatorAccountAddress, true, cs(c(StableDenom, 53)), cs(c(StableDenom, 153))},
		{"subLiquidatorStable", LiquidatorAccountAddress, false, cs(c(StableDenom, 53)), cs(c(StableDenom, 47))},
		{"addLiquidatorGov", LiquidatorAccountAddress, true, cs(c(GovDenom, 53)), cs(c(StableDenom, 100))},  // no change to balance
		{"subLiquidatorGov", LiquidatorAccountAddress, false, cs(c(GovDenom, 53)), cs(c(StableDenom, 100))}, // no change to balance
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup keeper
			mapp, keeper := setUpMockAppWithoutGenesis()
			// initialize an account with coins
			genAcc := auth.BaseAccount{
				Address: normalAddr,
				Coins:   cs(c(StableDenom, 100), c(GovDenom, 100)),
			}
			mock.SetGenesis(mapp, []auth.Account{&genAcc})

			// create a new context and setup the liquidator account
			header := abci.Header{Height: mapp.LastBlockHeight() + 1}
			mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
			ctx := mapp.BaseApp.NewContext(false, header)
			keeper.setLiquidatorModuleAccount(ctx, LiquidatorModuleAccount{cs(c(StableDenom, 100))}) // set gov coin "balance" to zero

			// perform the test action
			var err sdk.Error
			if tc.shouldAdd {
				_, _, err = keeper.AddCoins(ctx, tc.address, tc.amount)
			} else {
				_, _, err = keeper.SubtractCoins(ctx, tc.address, tc.amount)
			}

			mapp.EndBlock(abci.RequestEndBlock{})
			mapp.Commit()

			// check balances are as expected
			require.NoError(t, err)
			require.Equal(t, tc.expectedCoins, keeper.GetCoins(ctx, tc.address))
		})
	}
}
