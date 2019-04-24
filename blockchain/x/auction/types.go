package auction

import sdk "github.com/cosmos/cosmos-sdk/types"

type Auction struct {
	ID           []byte // TODO type
	Seller       sdk.AccAddress
	Amount       sdk.Coins
	EndTime      sdk.Int // TODO type
	LatestBidder sdk.AccAddress
	LatestBid    sdk.Coins
}
