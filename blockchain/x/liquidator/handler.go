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
		// case MsgStartSurplusAuction:
		// 	return handleMsgStartSurplusAuction(ctx, keeper)
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
	_, err = keeper.StartCollateralAuction(ctx, msg.CdpOwner, msg.CollateralDenom)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{} // TODO tags
}

func handleMsgStartDebtAuction(ctx sdk.Context, keeper Keeper) sdk.Result {
	// cancel out any debt and stable coins before trying to start auction
	keeper.settleDebt(ctx)
	// start an auction
	_, err := keeper.StartDebtAuction(ctx)
	if err != nil {
		return err.Result()
	}
	return sdk.Result{} // TODO tags
}

// With no stablity and liquidation fees, surplus auctions can never be run.
// func handleMsgStartSurplusAuction(ctx sdk.Context, keeper Keeper) sdk.Result {
//  keeper.settleDebt(ctx)
// 	_, err := keeper.StartSurplusAuction(ctx)
// 	if err != nil {
// 		return err.Result()
// 	}
// 	return sdk.Result{} // TODO tags
// }
