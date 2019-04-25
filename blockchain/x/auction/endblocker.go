package auction

import sdk "github.com/cosmos/cosmos-sdk/types"

// EndBlocker runs at the end of every block.
func EndBlocker(ctx sdk.Context, k Keeper) sdk.Tags {

	// get an iterator of expired auctions
	expiredAuctions := k.getQueueIterator(ctx, endTime(ctx.BlockHeight()))
	defer expiredAuctions.Close()

	// loop through and close them - distribute funds, delete from store (and queue)
	for ; expiredAuctions.Valid(); expiredAuctions.Next() {
		var auctionID auctionID
		k.cdc.MustUnmarshalBinaryLengthPrefixed(expiredAuctions.Value(), &auctionID)

		k.CloseAuction(ctx, auctionID)
	}

	return sdk.Tags{}
}
