package auction

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var msgCdc = codec.New()

func init() {
	RegisterCodec(msgCdc)
}

// RegisterCodec registers concrete types on the codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgStartAuction{}, "usdx/MsgStartAuction", nil) // TODO what is the correct name/path for this?
	cdc.RegisterConcrete(MsgPlaceBid{}, "usdx/MsgPlaceBid", nil)
}
