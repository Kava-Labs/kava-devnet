package mockpricefeed

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct{
	CurrentPrice Price
}

type Price struct {
	Price sdk.Dec
}

func NewKeeper(dec string) Keeper {
	v, err := sdk.NewDecFromStr(dec)
	if err != nil {
		panic("decimal initialization failed")
	}
	return Keeper{Price{v}}
}

func (k Keeper) GetPrice(ctx sdk.Context, denom string) Price {
	return k.CurrentPrice
}
