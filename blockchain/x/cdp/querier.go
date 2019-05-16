package cdp

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	// QueryGetCdp command for getting the info of a particular cdp
	QueryGetCdp = "getcdpinfo"
)

// implement fmt.Stringer for CDP type
func (cdp CDP) String() string {
	return strings.TrimSpace(fmt.Sprintf(`Owner: %s
CollateralType: %s
CollateralAmount: %s
Debt: %s`, cdp.Owner, cdp.CollateralDenom, cdp.CollateralAmount, cdp.Debt))
}

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
		default:
			return nil, sdk.ErrUnknownRequest("unknown cdp query endpoint")
		}
	}
}

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
		panic("could not marshal result to JSON")
	}
	return bz, nil
}
