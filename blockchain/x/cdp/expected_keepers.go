package cdp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/usdx/blockchain/x/pricefeed"
)

type bankKeeper interface {
	GetCoins(sdk.Context, sdk.AccAddress) sdk.Coins
	HasCoins(sdk.Context, sdk.AccAddress, sdk.Coins) bool
	AddCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Error)
	SubtractCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Error)
}

type pricefeedKeeper interface {
	GetCurrentPrice(sdk.Context, string) pricefeed.CurrentPrice
	// These are used for testing TODO replace mockApp with keeper in tests to remove these
	AddAsset(sdk.Context, string, string)
	SetPrice(sdk.Context, sdk.AccAddress, string, sdk.Dec, sdk.Int) (pricefeed.PostedPrice, sdk.Error)
	SetCurrentPrices(sdk.Context) sdk.Error
}
