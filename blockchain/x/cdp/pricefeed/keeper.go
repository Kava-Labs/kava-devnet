package pricefeed

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct{}

type Price struct {
	Price sdk.Dec
}

func (k Keeper) GetPrice(ctx sdk.Context, denom string) Price {
	p, err := sdk.NewDecFromStr("23.452")
	if err != nil {
		panic("asfasfasdf")
	}
	return Price{
		Price: p,
	}
}
