package cdp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

/*
How this uses the sdk params module:
 - Put all the params for this module in one struct `CDPModuleParams`
 - Store this in the keeper's paramSubspace under one key
 - Provide a function to load the param struct all at once `keeper.GetParams(ctx)`
It's possible to set individual key value pairs within a paramSubspace, but reading and setting them is awkward (an empty variable needs to be created, then Get writes the value into it)
This approach will be awkward if we ever need to write individual parameters (because they're stored all together). If this happens do as the sdk modules do - store parameters separately with custom get/set func for each.
*/

type CdpModuleParams struct {
	GlobalDebtLimit  sdk.Int
	CollateralParams []CollateralParams
}

type CollateralParams struct {
	Denom            string  // Coin name of collateral type
	LiquidationRatio sdk.Dec // The ratio (Collateral (priced in stable coin) / Debt) under which a CDP will be liquidated
	DebtLimit        sdk.Int // Maximum amount of debt allowed to be drawn from this collateral type
	//DebtFloor        sdk.Int // used to prevent dust
}

var cdpModuleParamsKey = []byte("CdpModuleParams")

func createParamsKeyTable() params.KeyTable {
	return params.NewKeyTable(
		cdpModuleParamsKey, CdpModuleParams{},
	)
}

// Helper methods to search the list of collateral params for a particular denom. Wouldn't be needed if amino supported maps.

func (p CdpModuleParams) GetCollateralParams(collateralDenom string) CollateralParams {
	// search for matching denom, return
	for _, cp := range p.CollateralParams {
		if cp.Denom == collateralDenom {
			return cp
		}
	}
	// panic if not found, to be safe
	panic("collateral params not found in module params")
}
func (p CdpModuleParams) IsCollateralPresent(collateralDenom string) bool {
	// search for matching denom, return
	for _, cp := range p.CollateralParams {
		if cp.Denom == collateralDenom {
			return true
		}
	}
	return false
}
