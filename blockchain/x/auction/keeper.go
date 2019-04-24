package auction

import sdk "github.com/cosmos/cosmos-sdk/types"

type Keeper struct {
	// TODO
}

func NewKeeper() Keeper {
	return Keeper{}
}

func (k Keeper) createAuction(seller sdk.AccAddress, amount sdk.Coins, endtime sdk.Int) sdk.Error {
	// TODO

	// subtract coins from seller
	// create auction struct
	// store auction
	// add to the "queue"
	// validation somewhere

	return nil
}

func (k Keeper) placeBid(auctionID []byte, bidder sdk.AccAddress, bid sdk.Coins) sdk.Error {
	// TODO

	// get auction from store
	// update lastest bid info if larger
	// store again

	return nil
}
