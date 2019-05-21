package liquidator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	LiquidatorModuleParams LiquidatorModuleParams `json:"params"`
}

// DefaultGenesisState returns a default genesis state
// TODO pick better values
func DefaultGenesisState() GenesisState {
	return GenesisState{
		LiquidatorModuleParams{
			DebtAuctionSize: sdk.NewInt(10000),
			CollateralParams: []CollateralParams{
				{
					Denom:       "btc",
					AuctionSize: sdk.NewInt(1),
				},
				{
					Denom:       "xrp",
					AuctionSize: sdk.NewInt(10000),
				},
			},
		},
	}
}

// InitGenesis sets the genesis state in the keeper.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.setParams(ctx, data.LiquidatorModuleParams)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
// TODO write this
// func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
// 	return NewGenesisState(keeper.GetSendEnabled(ctx))
// }

// ValidateGenesis performs basic validation of genesis data returning an error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	// TODO
	// check debt auction size > 0
	// validate denoms
	// check no repeated denoms
	// check collateral auction sizes > 0
	return nil
}
