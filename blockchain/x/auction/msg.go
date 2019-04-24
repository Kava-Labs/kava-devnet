package auction

import sdk "github.com/cosmos/cosmos-sdk/types"

type MsgStartAuction struct {
	Seller  sdk.AccAddress
	Amount  sdk.Coins
	EndTime sdk.Int // TODO find best type
}

// NewMsgStartAuction returns a new MsgStartAuction.
func NewMsgStartAuction(seller sdk.AccAddress, amount sdk.Coins, endtime sdk.Int) MsgStartAuction {
	return MsgStartAuction{
		Seller:  seller,
		Amount:  amount,
		EndTime: endtime,
	}
}

// Route return the message type used for routing the message.
func (msg MsgStartAuction) Route() string { return "auction" }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgStartAuction) Type() string { return "start_auction" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgStartAuction) ValidateBasic() sdk.Error {
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgStartAuction) GetSignBytes() []byte {
	bz := msgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgStartAuction) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Seller}
}

type MsgPlaceBid struct {
	AuctionID []byte // TODO type
	Bidder    sdk.AccAddress
	Bid       sdk.Coins
}

// NewMsgPlaceBid returns a new MsgPlaceBid.
func NewMsgPlaceBid(auctionID []byte, bidder sdk.AccAddress, bid sdk.Coins) MsgPlaceBid {
	return MsgPlaceBid{
		AuctionID: auctionID,
		Bidder:    bidder,
		Bid:       bid,
	}
}

// Route return the message type used for routing the message.
func (msg MsgPlaceBid) Route() string { return "auction" }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgPlaceBid) Type() string { return "place_bid" }

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgPlaceBid) ValidateBasic() sdk.Error {
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgPlaceBid) GetSignBytes() []byte {
	bz := msgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgPlaceBid) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Bidder}
}
