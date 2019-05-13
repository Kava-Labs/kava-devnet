package cdp

import "github.com/cosmos/cosmos-sdk/codec"

var msgCdc = codec.New()

func init() {
	RegisterCodec(msgCdc)
}

// RegisterCodec registers concrete types on the codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateOrModifyCDP{}, "cdp/MsgCreateOrModifyCDP", nil) // TODO what is the correct name/path for this?
	cdc.RegisterConcrete(MsgTransferCDP{}, "cdp/MsgTransferCDP", nil)
}
