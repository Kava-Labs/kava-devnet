package pricefeed

import (
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// AddAsset adds an asset to the store
func (k Keeper) AddAsset(
	ctx sdk.Context,
	assetCode string,
	desc string,
) {
	assets := k.GetAssets(ctx)
	assets = append(assets, Asset{AssetCode: assetCode, Description: desc})
	store := ctx.KVStore(k.storeKey)
	store.Set(
		[]byte(AssetPrefix), k.cdc.MustMarshalBinaryBare(assets),
	)
}

// SetPrice updates the posted price for a specific oracle
func (k Keeper) SetPrice(
	ctx sdk.Context,
	oracle sdk.AccAddress,
	assetCode string,
	price sdk.Dec,
	expiry sdk.Int) (PostedPrice, sdk.Error) {
	// If the expiry is less than or equal to the current blockheight, we consider the price valid
	if expiry.GTE(sdk.NewInt(ctx.BlockHeight())) {
		store := ctx.KVStore(k.storeKey)
		prices := k.GetRawPrices(ctx, assetCode)
		var index int
		found := false
		for i := range prices {
			if prices[i].OracleAddress == oracle.String() {
				index = i
				found = true
				break
			}
		}
		// set the price for that particular oracle
		if found {
			prices[index] = PostedPrice{AssetCode: assetCode, OracleAddress: oracle.String(), Price: price, Expiry: expiry}
		} else {
			prices = append(prices, PostedPrice{
				assetCode, oracle.String(), price, expiry,
			})
			index = len(prices) - 1
		}

		store.Set(
			[]byte(RawPriceFeedPrefix+assetCode), k.cdc.MustMarshalBinaryBare(prices),
		)
		return prices[index], nil
	}
	return PostedPrice{}, ErrExpired(k.codespace)

}

// SetCurrentPrices updates the price of an asset to the meadian of all valid oracle inputs
func (k Keeper) SetCurrentPrices(ctx sdk.Context) sdk.Error {
	assets := k.GetAssets(ctx)
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
		l := len(notExpiredPrices)
		var medianPrice sdk.Dec
		var expiry sdk.Int
		// TODO make threshold for acceptance (ie. require 51% of oracles to have posted valid prices
		if l == 0 {
			// Error if there are no valid prices in the raw pricefeed
			return ErrNoValidPrice(k.codespace)
		} else if l == 1 {
			// Return immediately if there's only one price
			medianPrice = notExpiredPrices[0].Price
			expiry = notExpiredPrices[0].Expiry
		} else {
			// sort the prices
			sort.Slice(notExpiredPrices, func(i, j int) bool {
				return notExpiredPrices[i].Price.LT(notExpiredPrices[j].Price)
			})
			// If there's an even number of prices
			if l%2 == 0 {
				// TODO make sure this is safe.
				// Since it's a price and not a blance, division with precision loss is OK.
				price1 := notExpiredPrices[l/2-1].Price
				price2 := notExpiredPrices[l/2].Price
				sum := price1.Add(price2)
				divsor, _ := sdk.NewDecFromStr("2")
				medianPrice = sum.Quo(divsor)
				// TODO Check if safe, makes sense
				// Takes the average of the two expiries rounded down to the nearest Int.
				expiry = notExpiredPrices[l/2-1].Expiry.Add(notExpiredPrices[l/2].Expiry).Quo(sdk.NewInt(2))
			} else {
				// integer division, so we'll get an integer back, rounded down
				medianPrice = notExpiredPrices[l/2].Price
				expiry = notExpiredPrices[l/2].Expiry
			}
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

	return nil
}

// GetAssets returns the assets in the pricefeed system
func (k Keeper) GetAssets(ctx sdk.Context) []Asset {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get([]byte(AssetPrefix))
	var assets []Asset
	k.cdc.MustUnmarshalBinaryBare(bz, &assets)
	return assets
}

// GetAsset returns the asset if it is in the pricefeed system
func (k Keeper) GetAsset(ctx sdk.Context, assetCode string) (Asset, bool) {
	assets := k.GetAssets(ctx)

	for i := range assets {
		if assets[i].AssetCode == assetCode {
			return assets[i], true
		}
	}
	return Asset{}, false

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

// ValidatePostPrice makes sure the person posting the price is an oracle
func (k Keeper) ValidatePostPrice(ctx sdk.Context, msg MsgPostPrice) bool {
	// TODO implement this
	_, found := k.GetAsset(ctx, msg.AssetCode)
	if !found {
		return false
	}
	return true
}
