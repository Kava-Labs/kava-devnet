package liquidator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/kava-labs/usdx/blockchain/x/cdp"
)

type cdpKeeper interface {
	GetCDP(sdk.Context, sdk.AccAddress, string) (cdp.CDP, bool)
	ConfiscateCDP(sdk.Context, sdk.AccAddress, string) sdk.Error
	GetStableDenom() string
	GetLiquidatorAccountAddress() sdk.AccAddress // TODO can this not be here? seems messy
}

type bankKeeper interface {
	GetCoins(sdk.Context, sdk.AccAddress) sdk.Coins
	AddCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
	SubtractCoins(sdk.Context, sdk.AccAddress, sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
}

type auctionKeeper interface {
	// StartForwardAuction
	// StartReverseAuction
	// StartForwardReverseAuction
}
