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
