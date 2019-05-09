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
	cdc.RegisterConcrete(MsgPlaceBid{}, "usdx/MsgPlaceBid", nil) // TODO what is the correct name/path for this?

	// Register the Auction interface and concrete types
	cdc.RegisterInterface((*Auction)(nil), nil)
	cdc.RegisterConcrete(&ForwardAuction{}, "auction/ForwardAuction", nil)
	cdc.RegisterConcrete(&ReverseAuction{}, "auction/ReverseAuction", nil)
	cdc.RegisterConcrete(&ForwardReverseAuction{}, "auction/ForwardReverseAuction", nil)
}
