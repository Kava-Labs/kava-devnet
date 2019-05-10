package liquidator

import sdk "github.com/cosmos/cosmos-sdk/types"

// Message types for starting various auctions. TODO Should they place an initial bid as well?

type MsgSeizeAndStartCollateralAuction struct {
	CdpOwner        sdk.AccAddress
	CollateralDenom string
}

func NewMsgSeizeAndStartCollateralAuction(cdpOwner sdk.AccAddress, collateralDenom string) MsgSeizeAndStartCollateralAuction {
	return MsgSeizeAndStartCollateralAuction{
		CdpOwner:        cdpOwner,
		CollateralDenom: collateralDenom,
	}
}

// Route return the message type used for routing the message.
func (msg MsgSeizeAndStartCollateralAuction) Route() string { return "liquidator" }

// Type returns a human-readable string for the message, intended for utilization within tags.
func (msg MsgSeizeAndStartCollateralAuction) Type() string { return "seize_and_start_auction" } // TODO snake case?

// ValidateBasic does a simple validation check that doesn't require access to any other information.
func (msg MsgSeizeAndStartCollateralAuction) ValidateBasic() sdk.Error {
	if msg.CdpOwner.Empty() {
		return sdk.ErrInternal("invalid (empty) CDP owner address")
	}
	// TODO check coin denoms
	return nil
}

// GetSignBytes gets the canonical byte representation of the Msg.
func (msg MsgSeizeAndStartCollateralAuction) GetSignBytes() []byte {
	bz := msgCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners returns the addresses of signers that must sign.
func (msg MsgSeizeAndStartCollateralAuction) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{} // TODO does leaving this blank work?
}

type MsgStartDebtAuction struct{}

func NewMsgStartDebtAuction() MsgStartDebtAuction        { return MsgStartDebtAuction{} }
func (msg MsgStartDebtAuction) Route() string            { return "liquidator" }
func (msg MsgStartDebtAuction) Type() string             { return "start_debt_auction" } // TODO snake case?
func (msg MsgStartDebtAuction) ValidateBasic() sdk.Error { return nil }
func (msg MsgStartDebtAuction) GetSignBytes() []byte {
	return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}
func (msg MsgStartDebtAuction) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }

type MsgStartSurplusAuction struct{}

func NewMsgStartSurplusAuction() MsgStartSurplusAuction     { return MsgStartSurplusAuction{} }
func (msg MsgStartSurplusAuction) Route() string            { return "liquidator" }
func (msg MsgStartSurplusAuction) Type() string             { return "start_debt_auction" } // TODO snake case?
func (msg MsgStartSurplusAuction) ValidateBasic() sdk.Error { return nil }
func (msg MsgStartSurplusAuction) GetSignBytes() []byte {
	return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}
func (msg MsgStartSurplusAuction) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }
