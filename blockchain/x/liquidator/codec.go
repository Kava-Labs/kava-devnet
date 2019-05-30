package liquidator

import "github.com/cosmos/cosmos-sdk/codec"

var moduleCdc = codec.New()

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	moduleCdc = cdc.Seal()
}

// RegisterCodec registers concrete types on the codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSeizeAndStartCollateralAuction{}, "liquidator/MsgSeizeAndStartCollateralAuction", nil)
	cdc.RegisterConcrete(MsgStartDebtAuction{}, "liquidator/MsgStartDebtAuction", nil)
	// cdc.RegisterConcrete(MsgStartSurplusAuction{}, "liquidator/MsgStartSurplusAuction", nil)
}
