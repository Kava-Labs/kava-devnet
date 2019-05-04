package cdp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type CDP struct {
	//ID               []byte         // removing IDs for now to make things simpler
	Owner            sdk.AccAddress // account that authorizes changes to the CDP
	CollateralDenom  string
	CollateralAmount sdk.Int
	Debt             sdk.Int
}

type CollateralStats struct { // TODO better name
	Denom     string
	TotalDebt sdk.Int
	// no fees for now // AccumulatedFees sdk.Int
}
