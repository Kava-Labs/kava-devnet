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
func (k Keeper) fetchXrpTx(txHash string) (XrpTx, sdk.Error) {
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

	return xrpTx, nil
}
