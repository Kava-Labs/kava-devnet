package peg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/bank"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keeper maintains the link to data storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	coinKeeper bank.Keeper

	storeKey sdk.StoreKey // Unexposed key to access store from sdk.Context

	cdc *codec.Codec // The wire codec for binary encoding/decoding.
}

// NewKeeper creates new instances of the nameservice Keeper
func NewKeeper(coinKeeper bank.Keeper, storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		coinKeeper: coinKeeper,
		storeKey:   storeKey,
		cdc:        cdc,
	}
}

// TODO what does this function return? - some data containing results from xrp tx
func (k Keeper) fetchXrpTransactionData(txHash string) (XrpTx, sdk.Error) {
	url := fmt.Sprintf("https://testnet.data.api.ripple.com/v2/transactions/%s", txHash)
	resp, err := http.Get(url)
	if err != nil {
		return XrpTx{}, sdk.ErrInternal("Problem fetching data from ripple api") // TODO pick more informative sdk error type
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return XrpTx{}, sdk.ErrInternal("Problem reading data from the response")
	}

	xrpTx := XrpTx{}
	json.Unmarshal(body, &xrpTx)
	// TODO validate data?
	if xrpTx.Result == "error" {
		return XrpTx{}, sdk.ErrInternal("Xrp Transaction returned error response")
	}

	return xrpTx, nil
}

func (k Keeper) isPxrpMultisgTransaction(xrpTx XrpTx) bool {
	// TODO what's the best way to 'know' the multisig address
	return true
}

func (k Keeper) hasValidMemoData(xrpTx XrpTx) bool {
	// TODO what format should the memo data actually be in?
	return true
}

func (k Keeper) mintPxrp(ctx sdk.Context, xrpTx XrpTx) (sdk.Coins, sdk.Tags, sdk.Error) {
	// txMemo := xrpTx.Transaction.Tx.Memos[0].Memo.MemoData
	// decoded, err := hex.DecodeString(txMemo)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	amount, ok := sdk.NewIntFromString(xrpTx.Transaction.Tx.Amount)
	if ok == false {
		sdk.ErrInternal("Invalid amount")
	}
	address, err := sdk.AccAddressFromHex(xrpTx.Transaction.Tx.Memos[0].Memo.MemoData)
	if err != nil {
		sdk.ErrInternal("Invalid destination address")
	}
	sendAmount := sdk.Coins{sdk.NewCoin("pxrp", amount)}

	coins, tags, err := k.coinKeeper.AddCoins(ctx, address, sendAmount)
	return coins, tags, nil

}

// cosmos1w4ekg7rpv3j8yunngagyu66nf36rxdjzg3xy6e6sg9v5k6txgemyxurg29995vn3ffms5sgz03
