package mockpricefeed

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	CurrentPrice CurrentPrice
}

type CurrentPrice struct {
	AssetCode string
	Price     sdk.Dec
	Expiry    sdk.Int
}

func NewKeeper(dec string) Keeper {
	v, err := sdk.NewDecFromStr(dec)
	if err != nil {
		panic("decimal initialization failed")
	}
	return Keeper{CurrentPrice{"", v, sdk.NewInt(0)}}
}

func (k Keeper) GetPrice(ctx sdk.Context, assetCode string) CurrentPrice {
	return k.CurrentPrice
}
