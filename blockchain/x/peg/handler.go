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
	txData, err := keeper.fetchXrpTransactionData(msg.TxHash)
	if err != nil {
		sdk.ErrInternal("TxHash was not valid").Result()
	}
	// throw if not ok
	ok := keeper.isPxrpMultisgTransaction(txData)
	if ok == false {
		sdk.ErrInternal("Tx is not to validator multisig address")
	}
	ok = keeper.hasValidMemoData(txData)
	if ok == false {
		sdk.ErrInternal("Memo is invalid")
	}
	// TODO validate something with amount?

	_, _, err = keeper.mintPxrp(ctx, txData)
	if err != nil {
		sdk.ErrInternal("Failed to mint xprp").Result()
	}
	//       k.getMemoData(Transaction.Tx.Memos[0].Memo.MemoData)
	// Check valid == true
	// Send/Mint pXRP to account for amount
	return sdk.Result{}
}
