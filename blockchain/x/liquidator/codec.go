package liquidator

import "github.com/cosmos/cosmos-sdk/codec"

var msgCdc = codec.New()

func init() {
	RegisterCodec(msgCdc)
}

// RegisterCodec registers concrete types on the codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgStartCollateralAuction{}, "liquidator/MsgStartCollateralAuction", nil)
	cdc.RegisterConcrete(MsgStartDebtAuction{}, "liquidator/MsgStartDebtAuction", nil)
	// cdc.RegisterConcrete(MsgStartSurplusAuction{}, "liquidator/MsgStartSurplusAuction", nil)
}
