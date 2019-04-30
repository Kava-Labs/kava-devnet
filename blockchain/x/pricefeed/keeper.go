package pricefeed

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"sort"
)

// TODO refactor constants to app.go
const (
	// ModuleKey is the name of the module
	ModuleName = "pricefeed"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName

	// Parameter store default namestore
	DefaultParamspace = ModuleName

	// Store prefix for the raw pricefeed of an asset
	RawPriceFeedPrefix = StoreKey + ":raw:"

	// Store prefix for the current price of an asset
	CurrentPricePrefix = StoreKey + ":currentprice:"

	// Store Prefix for the assets in the pricefeed system
	AssetPrefix = StoreKey + ":assets"
)

// Keeper struct for pricefeed module
type Keeper struct {
	storeKey  sdk.StoreKey
	cdc       *codec.Codec
	codespace sdk.CodespaceType
}

// NewKeeper returns a new keeper for the pricefeed modle
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey:  storeKey,
		cdc:       cdc,
		codespace: codespace,
	}
}

func (k Keeper) getAssets(ctx sdk.Context) []Asset {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(AssetPrefix))
	var assets []Asset
	k.cdc.MustUnmarshalBinaryBare(bz, &assets)
	return assets
}

// SetPrice updates the posted price for a specific oracle
func (k Keeper) setPrice(
	ctx sdk.Context,
	oracle sdk.AccAddress,
	assetCode string,
	price sdk.Dec,
	expiry sdk.Int) {
	// If the expiry is less than or equal to the current blockheight, we consider the price valid
	if expiry.GTE(sdk.NewInt(ctx.BlockHeight())) {
		store := ctx.KVStore(k.storeKey)
		prices := k.GetRawPrices(ctx, assetCode)
		var index int

		for i := range prices {
			if prices[i].OracleAddress == oracle.String() {
				index = i
				break
			}
		}
		// set the price for that particular oracle
		prices[index] = PostedPrice{AssetCode: assetCode, OracleAddress: oracle.String(), Price: price, Expiry: expiry}
		store.Set(
			[]byte(RawPriceFeedPrefix+assetCode), k.cdc.MustMarshalBinaryBare(prices),
		)
	}
	return
}

// GetCurrentPrice fetches the current median price of all oracles for a specific asset
func (k Keeper) GetCurrentPrice(ctx sdk.Context, assetCode string) CurrentPrice {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(CurrentPricePrefix + assetCode))
	var price CurrentPrice
	k.cdc.MustUnmarshalBinaryBare(bz, &price)
	return price
}

// GetRawPrices fetches the set of all prices posted by oracles for an asset
func (k Keeper) GetRawPrices(ctx sdk.Context, assetCode string) []PostedPrice {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(RawPriceFeedPrefix + assetCode))
	var prices []PostedPrice
	k.cdc.MustUnmarshalBinaryBare(bz, &prices)
	return prices
}

// SetCurrentPrices updates the price of an asset to the meadian of all valid oracle inputs
func (k Keeper) SetCurrentPrices(ctx sdk.Context) {
	assets := k.getAssets(ctx)
	for _, v := range assets {
		assetCode := v.AssetCode
		prices := k.GetRawPrices(ctx, assetCode)
		var notExpiredPrices []CurrentPrice
		// filter out expired prices
		for _, v := range prices {
			if v.Expiry.GTE(sdk.NewInt(ctx.BlockHeight())) {
				notExpiredPrices = append(notExpiredPrices, CurrentPrice{
					AssetCode: v.AssetCode,
					Price:     v.Price,
					Expiry:    v.Expiry,
				})
			}
		}
		// TODO Check if 51% of oracles have posted prices
		if len(notExpiredPrices) == 0 {
			return
		}
		// sort the prices
		sort.Slice(notExpiredPrices, func(i, j int) bool {
			return notExpiredPrices[i].Price.LT(notExpiredPrices[j].Price)
		})

		// Select the median price
		l := len(notExpiredPrices)
		var medianPrice sdk.Dec
		var expiry sdk.Int
		if l%2 == 0 {
			// TODO make sure this is safe.
			// Since it's a price and not a blance, division with precision loss is OK.
			price1 := notExpiredPrices[l/2-1].Price
			price2 := notExpiredPrices[l/2+1].Price
			sum := price1.Add(price2)
			divsor, _ := sdk.NewDecFromStr("2")
			medianPrice = sum.Quo(divsor)
			// TODO Check if safe, makes sense
			// Takes the average of the two expiries rounded down to the nearest Int.
			expiry = notExpiredPrices[l/2-1].Expiry.Add(notExpiredPrices[l/2+1].Expiry).Quo(sdk.NewInt(2))
		} else {
			medianPrice = notExpiredPrices[l/2].Price
			expiry = notExpiredPrices[l/2].Expiry
		}

		store := ctx.KVStore(k.storeKey)
		currentPrice := CurrentPrice{
			AssetCode: assetCode,
			Price:     medianPrice,
			Expiry:    expiry,
		}
		store.Set(
			[]byte(CurrentPricePrefix+assetCode), k.cdc.MustMarshalBinaryBare(currentPrice),
		)
	}

	return
}

// ValidatePostPrice makes sure the person posting the price is an oracle
func (k Keeper) ValidatePostPrice(ctx sdk.Context, msg MsgPostPrice) bool {
	// TODO implement this
	return true
}
