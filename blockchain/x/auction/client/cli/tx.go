package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/kava-labs/usdx/blockchain/x/auction"
	"github.com/spf13/cobra"
)

// GetCmdPlaceBid cli command for creating and modifying cdps.
func GetCmdPlaceBid(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "placebid [AuctionID] [Bidder] [Bid] [Lot]",
		Short: "place a bid on an auction",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}
			id := sdk.NewUintFromString(args[0])

			bid, err := sdk.ParseCoin(args[2])
			if err != nil {
				fmt.Printf("invalid bid amount - %s \n", string(args[2]))
				return err
			}

			lot, err := sdk.ParseCoin(args[3])
			if err != nil {
				fmt.Printf("invalid lot - %s \n", string(args[3]))
				return err
			}
			msg := auction.NewMsgPlaceBid(id.Uint64(), cliCtx.GetFromAddress(), bid, lot)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}
			cliCtx.PrintResponse = true
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}
