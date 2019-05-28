package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"

	"github.com/kava-labs/usdx/blockchain/x/liquidator/client/cli"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

// NewModuleClient creates client for the module
func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:   "liquidator",
		Short: "Querying commands for the liquidator module",
	}

	queryCmd.AddCommand(client.GetCommands(
		cli.GetCmd_GetOutstandingDebt(mc.storeKey, mc.cdc),
	)...)

	return queryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:   "liquidator",
		Short: "Liquidator transactions subcommands",
	}

	txCmd.AddCommand(client.PostCommands(
		cli.GetCmd_SeizeAndStartCollateralAuction(mc.cdc),
		cli.GetCmd_StartDebtAuction(mc.cdc),
	)...)

	return txCmd
}
