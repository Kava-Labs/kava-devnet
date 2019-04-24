package auction

import sdk "github.com/cosmos/cosmos-sdk/types"

type Auction struct {
	ID           auctionID
	Seller       sdk.AccAddress
	Amount       sdk.Coins
	EndTime      sdk.Int // TODO type
	LatestBidder sdk.AccAddress
	LatestBid    sdk.Coins
}

type auctionID uint64 // copyied from how the gov module IDs its proposals
