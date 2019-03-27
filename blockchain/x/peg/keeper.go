package peg

import (
	"encoding/hex"
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
	storeKey   sdk.StoreKey // Unexposed key to access store from sdk.Context
	cdc        *codec.Codec // The wire codec for binary encoding/decoding.
}

// NewKeeper creates new instances of the nameservice Keeper
func NewKeeper(coinKeeper bank.Keeper, storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		coinKeeper: coinKeeper,
		storeKey:   storeKey,
		cdc:        cdc,
	}
}

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

func (k Keeper) mintPxrp(ctx sdk.Context, xrpTx XrpTx) (sdk.Tags, sdk.Error) {
	amount, ok := sdk.NewIntFromString(xrpTx.Transaction.Tx.Amount)
	if ok == false {
		return nil, sdk.ErrInternal("Invalid amount")
	}
	destAddressBech32, err := decodeXrpTxMemoData(xrpTx.Transaction.Tx.Memos[0].Memo.MemoData)
	if err != nil {
		return nil, sdk.ErrInternal("Invalid memo")
	}
	destAddress, err := sdk.AccAddressFromBech32(destAddressBech32)
	if err != nil {
		fmt.Println(err)
		return nil, sdk.ErrInternal("Invalid destination address")
	}

	sendAmount := sdk.Coins{sdk.NewCoin("pxrp", amount)}
	_, tags, errSdk := k.coinKeeper.AddCoins(ctx, destAddress, sendAmount)
	if errSdk != nil {
		return nil, errSdk
	}
	return tags, nil

}

func decodeXrpTxMemoData(memoData string) (string, error) {
	bz, err := hex.DecodeString(memoData)
	if err != nil {
		return "", err
	}
	return string(bz), err
}
