package cli

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/spf13/cobra"

	"github.com/kava-labs/usdx/blockchain/x/liquidator"
)

func GetCmd_SeizeAndStartCollateralAuction(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seize [cdp-owner] [collateral-denom]",
		Short: "seize funds from a CDP and send to auction",
		Long:  "Seize a fixed amount of collateral and debt from a CDP and start a 'forward-reverse' auction with the collateral to cover the debt.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Setup
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithAccountDecoder(cdc)

			// Validate inputs
			sender := cliCtx.GetFromAddress()
			cdpOwner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			denom := args[1]
			// TODO validate denom?

			// Prepare and send message
			msgs := []sdk.Msg{liquidator.MsgSeizeAndStartCollateralAuction{
				Sender:          sender,
				CdpOwner:        cdpOwner,
				CollateralDenom: denom,
			}}
			// TODO print out results like auction ID?
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, msgs, false)
		},
	}
	return cmd
}

func GetCmd_StartDebtAuction(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint", // TODO is this a reasonable name?
		Short: "start a debt auction, minting gov coin to cover debt",
		Long:  "Start a reverse auction, selling off minted gov coin to raise a fixed amount of stable coin.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Setup
			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			sender := cliCtx.GetFromAddress()

			// Prepare and send message
			msgs := []sdk.Msg{liquidator.MsgStartDebtAuction{
				Sender: sender,
			}}
			// TODO print out results like auction ID?
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, msgs, false)
		},
	}
	return cmd
}
