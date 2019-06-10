package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/kava-labs/kava-devnet/blockchain/x/cdp"
	"github.com/spf13/cobra"
)

// GetCmdModifyCdp cli command for creating and modifying cdps.
func GetCmdModifyCdp(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "modifycdp [ownerAddress] [collateralType] [collateralChange] [debtChange]",
		Short: "create or modify a cdp",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}
			collateralChange, ok := sdk.NewIntFromString(args[2])
			if !ok {
				fmt.Printf("invalid collateral amount - %s \n", string(args[2]))
				return nil
			}
			debtChange, ok := sdk.NewIntFromString(args[3])
			if !ok {
				fmt.Printf("invalid debt amount - %s \n", string(args[3]))
				return nil
			}
			msg := cdp.NewMsgCreateOrModifyCDP(cliCtx.GetFromAddress(), args[1], collateralChange, debtChange)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}
			cliCtx.PrintResponse = true
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
