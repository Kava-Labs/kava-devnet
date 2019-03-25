package peg

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "nameservice" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgXrpTx:
			return handleMsgXrpTx(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized nameservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle a message to set name
func handleMsgXrpTx(ctx sdk.Context, keeper Keeper, msg MsgXrpTx) sdk.Result {

	// Request tx hash from ripple api
	_, err := keeper.fetchXrpTx(msg.TxHash)
	if err != nil {
		return err.Result() // fetchXrpTx returns a sdk.Error type
	}
	// Parse out data (valid, memo, amount)
	//       k.getMemoData(Transaction.Tx.Memos[0].Memo.MemoData)
	// Check valid == true
	// Send/Mint pXRP to account for amount

	return sdk.Result{}
}
