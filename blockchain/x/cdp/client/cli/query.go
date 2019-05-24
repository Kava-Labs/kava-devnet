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
		Use:   "cdp [ownerAddress] [collateralType]",
		Short: "get info about a cdp",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			ownerAddress, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			collateralType := args[1] // TODO validation?
			bz, err := cdc.MarshalJSON(cdp.QueryCdpsParams{
				Owner:           ownerAddress,
				CollateralDenom: collateralType,
			})
			if err != nil {
				return err
			}

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, cdp.QueryGetCdps)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				fmt.Printf("error when getting cdp info - %s", err)
				fmt.Printf("could not get current cdp info - %s %s \n", string(ownerAddress), string(collateralType))
				return err
			}

			// Decode and print results
			var cdps cdp.CDPs
			cdc.MustUnmarshalJSON(res, &cdps)
			if len(cdps) != 1 {
				panic("Unexpected number of CDPs returned from querier. This shouldn't happen.")
			}
			return cliCtx.PrintOutput(cdps[0])
		},
	}
}

func GetCmd_GetCdps(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cdps [collateralType]",
		Short: "get info about many cdps",
		Long:  "Get all CDPs or specify a collateral type to get only CDPs with that collateral type.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			bz, err := cdc.MarshalJSON(cdp.QueryCdpsParams{CollateralDenom: args[0]}) // denom="" returns all CDPs // TODO will this fail if there are no args?
			if err != nil {
				return err
			}

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, cdp.QueryGetCdps)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			// Decode and print results
			var out cdp.CDPs
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmd_GetUnderCollateralizedCdps(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "bad-cdps [collateralType] [price]",
		Short: "get under collateralized CDPs",
		Long:  "Get all CDPS of a particular collateral type that will be under collateralized at the specified price. Pass in the current price to get currently under collateralized CDPs.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Prepare params for querier
			price, errSdk := sdk.NewDecFromStr(args[1])
			if errSdk != nil {
				return fmt.Errorf(errSdk.Error()) // TODO check this returns useful output
			}
			bz, err := cdc.MarshalJSON(cdp.QueryCdpsParams{
				CollateralDenom:       args[0],
				UnderCollateralizedAt: price,
			})
			if err != nil {
				return err
			}

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, cdp.QueryGetCdps)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			// Decode and print results
			var out cdp.CDPs
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmd_GetParams(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "get the cdp module parameters",
		Long:  "Get the current global cdp module parameters.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Query
			route := fmt.Sprintf("custom/%s/%s", queryRoute, cdp.QueryGetParams)
			res, err := cliCtx.QueryWithData(route, nil) // TODO use cliCtx.QueryStore?
			if err != nil {
				return err
			}

			// Decode and print results
			var out cdp.CdpModuleParams
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
