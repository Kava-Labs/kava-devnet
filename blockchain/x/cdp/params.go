package cdp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// put all CDPconfig in one struct
// store this in the paramSubSpace under one key
// provide function to load the CDPconfig all at once
// provide function to check denom in list of authed denoms, and to get individual config (both wouldn't be needed if amino supported maps)

// loading all at once won't work if the keeper ever needs to modify the config - then switch to what some of the sdk modules do and have a custom get/set func for each parameter

type CdpModuleParams struct {
	GlobalDebtLimit  sdk.Int // TODO which is a better term: "debt limit" or "debt ceiling"?
	CollateralParams []CollateralParams
}

// TODO are these types ok?
type CollateralParams struct {
	Denom            string  // Coin name of collateral type
	LiquidationRatio sdk.Dec // The ratio (Collateral (priced in stable coin) / Debt) under which a CDP will be liquidated
	DebtLimit        sdk.Int // Maximum amount of debt allowed to be drawn from this collateral type
	DebtFloor        sdk.Int // used to prevent dust
}

var cdpModuleParamsKey = []byte("CdpModuleParams")

func createParamsKeyTable() params.KeyTable {
	return params.NewKeyTable(
		cdpModuleParamsKey, CdpModuleParams{},
	)
}

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
