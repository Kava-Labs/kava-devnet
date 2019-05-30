package auction

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// generic sealed codec to be used throughout module
var moduleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	moduleCdc = cdc.Seal()
}

// RegisterCodec registers concrete types on the codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgPlaceBid{}, "auction/MsgPlaceBid", nil)

	// Register the Auction interface and concrete types
	cdc.RegisterInterface((*Auction)(nil), nil)
	cdc.RegisterConcrete(&ForwardAuction{}, "auction/ForwardAuction", nil)
	cdc.RegisterConcrete(&ReverseAuction{}, "auction/ReverseAuction", nil)
	cdc.RegisterConcrete(&ForwardReverseAuction{}, "auction/ForwardReverseAuction", nil)
}
