package client


import (
	"github.com/cosmos/cosmos-sdk/client"
	auctioncmd "github.com/kava-labs/usdx/blockchain/x/auction/client/cli"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"
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
	// Group nameservice queries under a subcommand
	auctionQueryCmd := &cobra.Command{
		Use:   "auction",
		Short: "Querying commands for the auction module",
	}

	auctionQueryCmd.AddCommand(client.GetCommands(
		auctioncmd.GetCmdGetAuctions(mc.storeKey, mc.cdc),
	)...)

	return auctionQueryCmd
}

// GetTxCmd returns the transaction commands for this module
func (mc ModuleClient) GetTxCmd() *cobra.Command {
	auctionTxCmd := &cobra.Command{
		Use:   "auction",
		Short: "auction transactions subcommands",
	}

	auctionTxCmd.AddCommand(client.PostCommands(
		auctioncmd.GetCmdPlaceBid(mc.cdc),
	)...)

	return auctionTxCmd
}