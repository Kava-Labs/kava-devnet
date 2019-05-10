package liquidator

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Handle all liquidator messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSeizeAndStartCollateralAuction:
			return handleMsgSeizeAndStartCollateralAuction(ctx, keeper, msg)
		case MsgStartDebtAuction:
			return handleMsgStartDebtAuction(ctx, keeper)
		case MsgStartSurplusAuction:
			return handleMsgStartSurplusAuction(ctx, keeper)
		default:
			errMsg := fmt.Sprintf("Unrecognized liquidator msg type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSeizeAndStartCollateralAuction(ctx sdk.Context, keeper Keeper, msg MsgSeizeAndStartCollateralAuction) sdk.Result {
	err := keeper.SeizeUnderCollateralizedCDP(ctx, msg.CdpOwner, msg.CollateralDenom)
	if err != nil {
		return err.Result()
	}
	err = keeper.StartCollateralAuction(ctx, msg.CdpOwner, msg.CollateralDenom)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{} // TODO tags
}

func handleMsgStartDebtAuction(ctx sdk.Context, keeper Keeper) sdk.Result {
	err := keeper.StartDebtAuction(ctx)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{} // TODO tags
}

func handleMsgStartSurplusAuction(ctx sdk.Context, keeper Keeper) sdk.Result {
	err := keeper.StartSurplusAuction(ctx)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{} // TODO tags
}
