package liquidator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/usdx/blockchain/x/cdp"
)

type cdpKeeper interface {
	SeizeCDP(sdk.Context, sdk.AccAddress, string) (cdp.CDP, sdk.Error) // TODO is the dependence on a cdp type here bad?
	GetStableDenom() string                                            // TODO can this be removed somehow?
	GetGovDenom() string
	GetLiquidatorAccountAddress() sdk.AccAddress // This won't need to exist once the module account is defined in this module (instead of in the cdp module)
}

type bankKeeper interface {
	GetCoins(sdk.Context, sdk.AccAddress) sdk.Coins
	AddCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
	SubtractCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
}

type auctionKeeper interface {
	StartForwardAuction(sdk.Context, sdk.AccAddress, sdk.Coin, sdk.Coin) sdk.Error
	StartReverseAuction(sdk.Context, sdk.AccAddress, sdk.Coin, sdk.Coin) sdk.Error
	StartForwardReverseAuction(sdk.Context, sdk.AccAddress, sdk.Coin, sdk.Coin, sdk.AccAddress) sdk.Error
}
