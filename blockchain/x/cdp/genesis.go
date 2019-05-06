package cdp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO add initial global debt and collateral states

// GenesisState is the state that must be provided at genesis.
type GenesisState struct {
	CdpModuleParams CdpModuleParams `json:"params"`
}

// NewGenesisState creates a new genesis state.
// TODO write this
// func NewGenesisState() GenesisState {
// 	return GenesisState{CdpModuleParams: CdpModuleParams{}}
// }

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return GenesisState{
		CdpModuleParams{
			GlobalDebtLimit: sdk.NewInt(1000000),
			CollateralParams: []CollateralParams{
				{
					Denom:            "btc",
					LiquidationRatio: sdk.MustNewDecFromStr("1.5"),
					DebtLimit:        sdk.NewInt(500000),
				},
				{
					Denom:            "xrp",
					LiquidationRatio: sdk.MustNewDecFromStr("2.0"),
					DebtLimit:        sdk.NewInt(500000),
				},
			},
		}}
}

// InitGenesis sets the genesis state in the keeper.
func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
	keeper.setParams(ctx, data.CdpModuleParams)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
// TODO write this
// func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
// 	return NewGenesisState(keeper.GetSendEnabled(ctx))
// }

// ValidateGenesis performs basic validation of bank genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	// TODO
	// validate denoms
	// check collateral debt limits sum to global limit?
	// check limits are > 0
	// check ratios are > 1
	// check no repeated denoms
	// check at least 1 collateralParams
	return nil
}
