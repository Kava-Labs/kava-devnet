package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/kava-labs/kava-devnet/blockchain/x/liquidator"
)

// GetCmd_GetOutstandingDebt queries for the remaining available debt in the liquidator module after settlement with the module's stablecoin balance.
func GetCmd_GetOutstandingDebt(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "debt",
		Short: "get the outstanding seized debt",
		Long:  "Get the remaining available debt after settlement with the liquidator's stable coin balance.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, liquidator.QueryGetOutstandingDebt), nil)
			if err != nil {
				return err
			}
			var outstandingDebt sdk.Int
			cdc.MustUnmarshalJSON(res, &outstandingDebt)
			return cliCtx.PrintOutput(outstandingDebt)
		},
	}
}
