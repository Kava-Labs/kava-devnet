package auction

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	// QueryGetAuction command for getting the information about a particular auction
	QueryGetAuction = "getauctions"
)

// NewQuerier is the module level router for state queries
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryGetAuction:
			return queryAuctions(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown auction query endpoint")
		}
	}
}

func queryAuctions(ctx sdk.Context, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var AuctionsList QueryResAuctions

	iterator := keeper.GetAuctionIterator(ctx)

	for ; iterator.Valid(); iterator.Next() {

		var auction Auction
		keeper.cdc.MustUnmarshalBinaryBare(iterator.Value(), &auction)
		AuctionsList = append(AuctionsList, auction.String())
	}

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, AuctionsList)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}

	return bz, nil
}

// QueryResAuctions Result Payload for an auctions query
type QueryResAuctions []string

// implement fmt.Stringer
func (n QueryResAuctions) String() string {
	return strings.Join(n[:], "\n")
}
