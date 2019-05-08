package pricefeed

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var msgCdc = codec.New()

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgPostPrice{}, "pricefeed/MsgPostPrice", nil)
}

func init() {
	RegisterCodec(msgCdc)
}
