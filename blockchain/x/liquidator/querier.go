package liquidator

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	QueryGetOutstandingDebt = "outstanding_debt" // Get the outstanding seized debt
)

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryGetOutstandingDebt:
			return queryGetOutstandingDebt(ctx, path[1:], req, keeper)
		// case QueryGetSurplus:
		// 	return queryGetSurplus()
		default:
			return nil, sdk.ErrUnknownRequest("unknown liquidator query endpoint")
		}
	}
}

func queryGetOutstandingDebt(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	// Calculate the remaining seized debt after settling with the liquidator's stable coins.
	stableCoins := keeper.bankKeeper.GetCoins(
		ctx,
		keeper.cdpKeeper.GetLiquidatorAccountAddress(),
	).AmountOf(keeper.cdpKeeper.GetStableDenom())
	seizedDebt := keeper.GetSeizedDebt(ctx)
	settleAmount := sdk.MinInt(seizedDebt.Total, stableCoins)
	seizedDebt, err := seizedDebt.Settle(settleAmount)
	if err != nil {
		return nil, err // this shouldn't error in this context
	}

	// Get the available debt after settling
	oustandingDebt := seizedDebt.Available()

	// Encode and return
	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, oustandingDebt)
	if err2 != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}
