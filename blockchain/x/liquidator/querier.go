package liquidator

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	QueryGetFreeDebt = "freedebt" // Get the free seized debt
)

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryGetFreeDebt:
			return queryGetFreeDebt(ctx, path[1:], req, keeper)
		// case QueryGetSurplus:
		// 	return queryGetSurplus()
		// case QueryGetSeizedCDPs:
		// 	return queryGetSeizedCDPs()
		default:
			return nil, sdk.ErrUnknownRequest("unknown cdp query endpoint")
		}
	}
}

func queryGetFreeDebt(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	freeDebt := keeper.GetSeizedDebt(ctx).Available()
	bz, err := codec.MarshalJSONIndent(keeper.cdc, freeDebt)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}
