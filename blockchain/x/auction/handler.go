package auction

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a function to handle all "auction" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgStartAuction:
			return handleMsgStartAuction(ctx, keeper, msg)
		case MsgPlaceBid:
			return handleMsgPlaceBid(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized auction msg type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgStartAuction(ctx sdk.Context, keeper Keeper, msg MsgStartAuction) sdk.Result {
	// TODO
	return sdk.Result{}
}

func handleMsgPlaceBid(ctx sdk.Context, keeper Keeper, msg MsgPlaceBid) sdk.Result {
	// TODO
	return sdk.Result{}
}
