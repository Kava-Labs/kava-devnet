package pricefeed

import "github.com/cosmos/cosmos-sdk/x/params"

// ParamStoreKeyOracleList key to get the list of oracle
var ParamStoreKeyOracleList = []byte("oraclelist")

// ParamKeyTable keytable
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeyOracleList, []string{},
	)
}

/*
Keys:								Values:
pricefeed						N/A (top level prefix)
pricefeed:raw:x 		[]PostedPrice{AssetCode: string, OracleAddress: string, Price: sdk.Dec, Expiry: sdk.Int}
pricefeed:current:x CurrentPrice{AssetCode: string, Price: sdk.Dec, Expiry: sdk.Int}
pricefeed:oracles:x []Oracle{OracleAddress: string}
pricefeed:assets 		[]Asset{AssetCode:string, Description: string}

To update the price for a particular oracle after they have made a MsgPostPrice transaction:
prices := keeper.GetPrices(AssetCode)
var index int
for i := range prices {
	if prices[i].Oracle == [OracleAddress] {
		index = i
		break
	}
}
prices[index] = PostedPrice{
	[AssetCode]
	[OracleAddress]
	[Price]
	[Expiry]
}

store.Set([]byte(RawPriceFeedPrefix+assetCode), keeper.cdc.MustMarshallBinaryBare(prices))

*/
