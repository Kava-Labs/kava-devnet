package mockpricefeed

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/usdx/blockchain/x/pricefeed"
)

const (
	StoreKey         = pricefeed.StoreKey
	DefaultCodespace = pricefeed.DefaultCodespace
)

type Keeper struct {
	storeKey     sdk.StoreKey
	cdc          *codec.Codec
	CurrentPrice pricefeed.CurrentPrice
}

func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
	}
}

func (k Keeper) GetCurrentPrice(ctx sdk.Context, assetCode string) pricefeed.CurrentPrice {
	// ignores asset code
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("currentPrice"))
	var currentPrice pricefeed.CurrentPrice
	k.cdc.MustUnmarshalBinaryBare(bz, &currentPrice)
	return currentPrice
}

func (k Keeper) SetPrice(ctx sdk.Context, price sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(pricefeed.CurrentPrice{Price: price})
	store.Set([]byte("currentPrice"), bz)
}
