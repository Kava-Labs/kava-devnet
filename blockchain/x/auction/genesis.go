package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - crisis genesis state
type GenesisState struct {
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState() GenesisState {
	return GenesisState{}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}

// ValidateGenesis validates genesis state
func ValidateGenesis(data GenesisState) error {
	return nil
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	return NewGenesisState()
}
