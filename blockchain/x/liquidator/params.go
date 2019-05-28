package liquidator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

/*
How this uses the sdk params module:
 - Put all the params for this module in one struct `LiquidatorModuleParams`
 - Store this in the keeper's paramSubspace under one key
 - Provide a function to load the param struct all at once `keeper.GetParams(ctx)`
It's possible to set individual key value pairs within a paramSubspace, but reading and setting them is awkward (an empty variable needs to be created, then Get writes the value into it)
This approach will be awkward if we ever need to write individual parameters (because they're stored all together). If this happens do as the sdk modules do - store parameters separately with custom get/set func for each.
*/

type LiquidatorModuleParams struct {
	DebtAuctionSize sdk.Int
	//SurplusAuctionSize sdk.Int
	CollateralParams []CollateralParams
}

type CollateralParams struct {
	Denom       string  // Coin name of collateral type
	AuctionSize sdk.Int // Max amount of collateral to sell off in any one auction. Known as lump in Maker.
	// LiquidationPenalty
}

var moduleParamsKey = []byte("LiquidatorModuleParams")

func createParamsKeyTable() params.KeyTable {
	return params.NewKeyTable(
		moduleParamsKey, LiquidatorModuleParams{},
	)
}

// Helper methods to search the list of collateral params for a particular denom. Wouldn't be needed if amino supported maps.

func (p LiquidatorModuleParams) GetCollateralParams(collateralDenom string) CollateralParams {
	// search for matching denom, return
	for _, cp := range p.CollateralParams {
		if cp.Denom == collateralDenom {
			return cp
		}
	}
	// panic if not found, to be safe
	panic("collateral params not found in module params")
}
