package liquidator

import "github.com/cosmos/cosmos-sdk/codec"

var msgCdc = codec.New()

func init() {
	RegisterCodec(msgCdc)
}

// RegisterCodec registers concrete types on the codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSeizeAndStartCollateralAuction{}, "liquidator/MsgSeizeAndStartCollateralAuction", nil)
	cdc.RegisterConcrete(MsgStartDebtAuction{}, "liquidator/MsgStartDebtAuction", nil)
	cdc.RegisterConcrete(MsgStartSurplusAuction{}, "liquidator/MsgStartSurplusAuction", nil)
}
