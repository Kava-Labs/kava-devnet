package auction

import sdk "github.com/cosmos/cosmos-sdk/types"

type Auction struct {
	ID           auctionID
	Seller       sdk.AccAddress
	Amount       sdk.Coins // TODO limit an auction to only one type of coin?
	EndTime      endTime
	LatestBidder sdk.AccAddress
	LatestBid    sdk.Coins
}

type auctionID uint64 // copyied from how the gov module IDs its proposals
type endTime int64    // type of BlockHeight
