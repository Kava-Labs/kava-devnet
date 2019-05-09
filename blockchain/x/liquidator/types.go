package liquidator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SeizedCDP struct {
	OriginalOwner    sdk.AccAddress
	CollateralDenom  string
	CollateralAmount sdk.Int
	Debt             sdk.Int
}

type SeizedDebt sdk.Coin // seized collateral and usdx are stored in the module account, but debt is stored here // aka Sin
