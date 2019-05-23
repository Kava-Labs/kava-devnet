package cdp

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	QueryGetCdp                     = "cdp"
	QueryGetCdps                    = "cdps"
	QueryGetUnderCollateralizedCdps = "under-collateralized-cdps"
	QueryGetParams                  = "params"
)

// QueryGetCdpResp response to a getcdpinfo query
type QueryGetCdpResp []string

// implement fmt.Stringer for QueryGetCdpResp type
func (result QueryGetCdpResp) String() string {
	return strings.Join(result[:], "\n")
}

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryGetCdp:
			return queryGetCdp(ctx, path[1:], req, keeper)
		case QueryGetCdps:
			return queryGetCdps(ctx, req, keeper)
		case QueryGetUnderCollateralizedCdps:
			return queryGetUnderCollateralizedCdps(ctx, req, keeper)
		case QueryGetParams:
			return queryGetParams(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown cdp query endpoint")
		}
	}
}

// queryGetCdp fetches a single CDP
func queryGetCdp(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	ownerAddress := path[0]
	collateralDenom := path[1]
	addr, err2 := sdk.AccAddressFromBech32(ownerAddress)
	if err2 != nil {
		return []byte{}, sdk.ErrUnknownRequest("invalid address")
	}
	cdp, found := keeper.GetCDP(ctx, addr, collateralDenom)
	if !found {
		return []byte{}, sdk.ErrUnknownRequest("cdp not found")
	}
	bz, err3 := codec.MarshalJSONIndent(keeper.cdc, cdp)
	if err3 != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err3.Error()))
	}
	return bz, nil
}

// TODO Can these structs be renamed or grouped together into something less confusing?
type QueryCdpsParams struct {
	CollateralDenom string // If this is "" then all CDPs will be returned
}

// queryGetCdps fetches all the CDPs, or all CDPS of a particular collateral type
func queryGetCdps(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	// Decode request
	var requestParams QueryCdpsParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	// Get CDPs
	cdps := keeper.GetCDPs(ctx, requestParams.CollateralDenom)

	// Encode results
	bz, err := codec.MarshalJSONIndent(keeper.cdc, cdps)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}

type QueryUnderCollateralizedCdpsParams struct {
	CollateralDenom string
	Price           sdk.Dec
}

// queryGetUnderCollateralizedCdps fetches all the CDPs (of a collateral type) that would be under the liquidation ratio at the specified price
func queryGetUnderCollateralizedCdps(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	// Decode request
	var requestParams QueryUnderCollateralizedCdpsParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &requestParams)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	// Get CDPs
	cdps, err2 := keeper.GetUnderCollateralizedCDPs(ctx, requestParams.CollateralDenom, requestParams.Price)
	if err2 != nil {
		return nil, err2
	}

	// Encode results
	bz, err3 := codec.MarshalJSONIndent(keeper.cdc, cdps)
	if err3 != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err3.Error()))
	}
	return bz, nil
}

// queryGetParams fetches the cdp module parameters
// TODO does this need to exist? Can you use cliCtx.QueryStore instead?
func queryGetParams(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	// Get params
	params := keeper.GetParams(ctx)

	// Encode results
	bz, err := codec.MarshalJSONIndent(keeper.cdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return bz, nil
}
