package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/kava-labs/usdx/blockchain/x/cdp"
	"github.com/spf13/cobra"
)

// GetCmdGetCdp queries the latest info about a particular cdp
func GetCmdGetCdp(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "getcdpinfo [ownerAddress] [collateralType]",
		Short: "get info about a cdp",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			ownerAddress := args[0]
			collateralType := args[1]
			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/getcdpinfo/%s/%s", queryRoute, ownerAddress, collateralType), nil)
			if err != nil {
				fmt.Printf("error when getting cdp info - %s", err)
				fmt.Printf("could not get current cdp info - %s %s \n", string(ownerAddress), string(collateralType))
				return nil
			}
			var out cdp.CDP
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
