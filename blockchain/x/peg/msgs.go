package peg

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgXrpTx defines a XRP transaction message
type MsgXrpTx struct {
	From   sdk.AccAddress // client that sent in this address
	TxHash string
}

// NewMsgXrpTx is a constructor function for MsgSetName
func NewMsgXrpTx(txHash string, from sdk.AccAddress) MsgXrpTx {
	return MsgXrpTx{
		From:   from,
		TxHash: txHash,
	}
}

// Route should return the name of the module
func (msg MsgXrpTx) Route() string { return "peg" }

// Type should return the action
func (msg MsgXrpTx) Type() string { return "xrp_tx" }

// ValidateBasic runs stateless checks on the message
func (msg MsgXrpTx) ValidateBasic() sdk.Error {
	// TODO do some validation
	return nil
}

// GetSignBytes encodes the message for signing
func (msg MsgXrpTx) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return sdk.MustSortJSON(b)
}

// GetSigners defines whose signature is required
func (msg MsgXrpTx) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.From}
}
