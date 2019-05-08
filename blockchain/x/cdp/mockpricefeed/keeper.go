package mockpricefeed

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	storeKey     sdk.StoreKey
	cdc          *codec.Codec
	CurrentPrice CurrentPrice
}

type CurrentPrice struct {
	AssetCode string
	Price     sdk.Dec
	Expiry    sdk.Int
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey) Keeper {
	return Keeper{
		storeKey: key,
		cdc:      cdc,
	}
}

func (k Keeper) GetCurrentPrice(ctx sdk.Context, assetCode string) CurrentPrice {
	// ignores asset code
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte("currentPrice"))
	var currentPrice CurrentPrice
	k.cdc.MustUnmarshalBinaryBare(bz, &currentPrice)
	return currentPrice
}

func (k Keeper) SetPrice(ctx sdk.Context, price sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(CurrentPrice{Price: price})
	store.Set([]byte("currentPrice"), bz)
}
