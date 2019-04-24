package auction

import sdk "github.com/cosmos/cosmos-sdk/types"

// EndBlocker runs at the end of every block.
func EndBlocker(ctx sdk.Context, keeper Keeper) sdk.Tags {
	// TODO

	// get an iterator for expired auctions
	// loop through and close them - distribute funds, delete from store (and queue)

	return sdk.Tags{}
}
