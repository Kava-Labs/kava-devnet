package cdp

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestAddSubtractGetCoins(t *testing.T) {
	normalAddr, _, _ := generateAccAddress()

	tests := []struct {
		name          string
		address       sdk.AccAddress
		add           bool
		amount        sdk.Coins
		expectedCoins sdk.Coins
	}{
		{"addNormalAddress", normalAddr, true, sdk.Coins{c(StableDenom, 53)}, sdk.Coins{c(StableDenom, 153), c(GovDenom, 100)}},
		{"subNormalAddress", normalAddr, false, sdk.Coins{c(StableDenom, 53)}, sdk.Coins{c(StableDenom, 47), c(GovDenom, 100)}},
		{"addLiquidatorStable", LiquidatorAccountAddress, true, sdk.Coins{c(StableDenom, 53)}, sdk.Coins{c(StableDenom, 153)}},
		{"subLiquidatorStable", LiquidatorAccountAddress, false, sdk.Coins{c(StableDenom, 53)}, sdk.Coins{c(StableDenom, 47)}},
		{"addLiquidatorGov", LiquidatorAccountAddress, true, sdk.Coins{c(GovDenom, 53)}, sdk.Coins{c(StableDenom, 100)}},  // no change to balance
		{"subLiquidatorGov", LiquidatorAccountAddress, false, sdk.Coins{c(GovDenom, 53)}, sdk.Coins{c(StableDenom, 100)}}, // no change to balance
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// setup keeper
			mapp, keeper := setUpMockAppWithoutGenAccounts()
			// initialize an account with coins
			genAcc := auth.BaseAccount{
				Address: normalAddr,
				Coins:   sdk.Coins{c(StableDenom, 100), c(GovDenom, 100)},
			}
			mock.SetGenesis(mapp, []auth.Account{&genAcc})

			// create a new context
			mapp.BeginBlock(abci.RequestBeginBlock{})
			ctx := mapp.BaseApp.NewContext(false, abci.Header{})
			keeper.setLiquidatorModuleAccount(ctx, LiquidatorModuleAccount{sdk.Coins{c(StableDenom, 100)}}) // set gov coin "balance" to zero

			if tc.add {
				keeper.AddCoins(ctx, tc.address, tc.amount)
			} else {
				keeper.SubtractCoins(ctx, tc.address, tc.amount)
			}

			mapp.EndBlock(abci.RequestEndBlock{})
			mapp.Commit()

			// check balances are as expected
			require.Equal(t, tc.expectedCoins, keeper.GetCoins(ctx, tc.address))
		})
	}
}
