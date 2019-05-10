package cdp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/usdx/blockchain/x/pricefeed" // TODO What's the go pattern for avoiding returning specific types from interfaces?
)

type bankKeeper interface {
	GetCoins(sdk.Context, sdk.AccAddress) sdk.Coins
	HasCoins(sdk.Context, sdk.AccAddress, sdk.Coins) bool
	AddCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
	SubtractCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
}

type pricefeedKeeper interface {
	GetCurrentPrice(sdk.Context, string) pricefeed.CurrentPrice
	// SetPrice(sdk.Context, sdk.Dec)
}
