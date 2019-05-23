package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/kava-labs/usdx/blockchain/x/cdp"
)

// GetCmd_GetCdp queries the latest info about a particular cdp
func GetCmd_GetCdp(queryRoute string, cdc *codec.Codec) *cobra.Command {
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

func GetCmd_GetCdps(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cdps [collateralType]",
		Short: "get info about many cdps",
		Long:  "Get all CDPS or specify a collateral type to get only CDPs with that collateral type.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// prepare params for querier
			denom := args[0] // TODO will this fail if there are no args?
			bz, err := cdc.MarshalJSON(cdp.QueryCdpsParams{CollateralDenom: denom})
			if err != nil {
				return err
			}

			// query
			route := fmt.Sprintf("custom/%s/%s", storeKey, cdp.QueryGetCdps)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			// decode and print results
			var out cdp.CDPs
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmd_GetUnderCollateralizedCdps(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "bad-cdps [collateralType] [price]",
		Short: "get under collateralized CDPs",
		Long:  "Get all CDPS of a particular collateral type that will be under collateralized at the specified price. Pass in the current price to get currently under collateralized CDPs.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// prepare params for querier
			price, errSdk := sdk.NewDecFromStr(args[1])
			if errSdk != nil {
				return fmt.Errorf(errSdk.Error()) // TODO check this returns useful output
			}
			bz, err := cdc.MarshalJSON(cdp.QueryUnderCollateralizedCdpsParams{
				CollateralDenom: args[0],
				Price:           price,
			})
			if err != nil {
				return err
			}

			// query
			route := fmt.Sprintf("custom/%s/%s", storeKey, cdp.QueryGetUnderCollateralizedCdps)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			// decode and print results
			var out cdp.CDPs
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmd_GetParams(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the cdp module parameters",
		Long:  "Get the current global cdp module parameters.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// query
			route := fmt.Sprintf("custom/%s/%s", storeKey, cdp.QueryGetParams)
			res, err := cliCtx.QueryWithData(route, nil) // TODO use cliCtx.QueryStore?
			if err != nil {
				return err
			}

			// decode and print results
			var out cdp.CdpModuleParams
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
