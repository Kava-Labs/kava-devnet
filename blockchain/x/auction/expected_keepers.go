package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type bankKeeper interface {
	SubtractCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Error)
	AddCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Error)
}
