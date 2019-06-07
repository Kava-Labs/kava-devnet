package liquidator

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/usdx/blockchain/x/cdp"
	"github.com/kava-labs/usdx/blockchain/x/pricefeed"
)

func TestKeeper_SeizeAndStartCollateralAuction(t *testing.T) {
	// Setup
	ctx, k := setupTestKeepers()

	_, addrs := mock.GeneratePrivKeyAddressPairs(1)

	cdp.InitGenesis(ctx, k.cdpKeeper, cdp.DefaultGenesisState())
	InitGenesis(ctx, k.liquidatorKeeper, DefaultGenesisState())
	pricefeed.InitGenesis(ctx, k.pricefeedKeeper, pricefeed.GenesisState{Assets: []pricefeed.Asset{{"btc", "a description"}}})
	k.pricefeedKeeper.SetPrice(ctx, addrs[0], "btc", sdk.MustNewDecFromStr("8000.00"), i(999999999))
	k.pricefeedKeeper.SetCurrentPrices(ctx)
	k.bankKeeper.AddCoins(ctx, addrs[0], cs(c("btc", 100)))

	k.cdpKeeper.ModifyCDP(ctx, addrs[0], "btc", i(3), i(16000))

	k.pricefeedKeeper.SetPrice(ctx, addrs[0], "btc", sdk.MustNewDecFromStr("7999.99"), i(999999999))
	k.pricefeedKeeper.SetCurrentPrices(ctx)

	// Run test function
	auctionID, err := k.liquidatorKeeper.SeizeAndStartCollateralAuction(ctx, addrs[0], "btc")

	// Check CDP
	require.NoError(t, err)
	cdp, found := k.cdpKeeper.GetCDP(ctx, addrs[0], "btc")
	require.True(t, found)
	require.Equal(t, cdp.CollateralAmount, i(2)) // original amount - params.CollateralAuctionSize
	require.Equal(t, cdp.Debt, i(10667))         // original debt scaled by amount of collateral removed
	// Check auction exists
	_, found = k.auctionKeeper.GetAuction(ctx, auctionID)
	require.True(t, found)
	// TODO check auction values are correct?
}

func TestKeeper_StartDebtAuction(t *testing.T) {
	// Setup
	ctx, k := setupTestKeepers()
	InitGenesis(ctx, k.liquidatorKeeper, DefaultGenesisState())
	initSDebt := SeizedDebt{i(2000), i(0)}
	k.liquidatorKeeper.setSeizedDebt(ctx, initSDebt)

	// Execute
	auctionID, err := k.liquidatorKeeper.StartDebtAuction(ctx)

	// Check
	require.NoError(t, err)
	require.Equal(t,
		SeizedDebt{
			initSDebt.Total,
			initSDebt.SentToAuction.Add(k.liquidatorKeeper.GetParams(ctx).DebtAuctionSize),
		},
		k.liquidatorKeeper.GetSeizedDebt(ctx),
	)
	_, found := k.auctionKeeper.GetAuction(ctx, auctionID)
	require.True(t, found)
	// TODO check auction values are correct?
}

// func TestKeeper_StartSurplusAuction(t *testing.T) {
// 	// Setup
// 	ctx, k := setupTestKeepers()
// 	initSurplus := i(2000)
// 	k.liquidatorKeeper.bankKeeper.AddCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress(), cs(sdk.NewCoin(k.cdpKeeper.GetStableDenom(), initSurplus)))
// 	k.liquidatorKeeper.setSeizedDebt(ctx, i(0))

// 	// Execute
// 	auctionID, err := k.liquidatorKeeper.StartSurplusAuction(ctx)

// 	// Check
// 	require.NoError(t, err)
// 	require.Equal(t,
// 		initSurplus.Sub(SurplusAuctionSize),
// 		k.liquidatorKeeper.bankKeeper.GetCoins(ctx,
// 			k.cdpKeeper.GetLiquidatorAccountAddress(),
// 		).AmountOf(k.cdpKeeper.GetStableDenom()),
// 	)
// 	_, found := k.auctionKeeper.GetAuction(ctx, auctionID)
// 	require.True(t, found)
// }

func TestKeeper_partialSeizeCDP(t *testing.T) {
	// Setup
	ctx, k := setupTestKeepers()

	_, addrs := mock.GeneratePrivKeyAddressPairs(1)

	cdp.InitGenesis(ctx, k.cdpKeeper, cdp.DefaultGenesisState())
	InitGenesis(ctx, k.liquidatorKeeper, DefaultGenesisState())
	pricefeed.InitGenesis(ctx, k.pricefeedKeeper, pricefeed.GenesisState{Assets: []pricefeed.Asset{{"btc", "a description"}}})
	k.pricefeedKeeper.SetPrice(ctx, addrs[0], "btc", sdk.MustNewDecFromStr("8000.00"), i(999999999))
	k.pricefeedKeeper.SetCurrentPrices(ctx)
	k.bankKeeper.AddCoins(ctx, addrs[0], cs(c("btc", 100)))

	k.cdpKeeper.ModifyCDP(ctx, addrs[0], "btc", i(3), i(16000))

	k.pricefeedKeeper.SetPrice(ctx, addrs[0], "btc", sdk.MustNewDecFromStr("7999.99"), i(999999999))
	k.pricefeedKeeper.SetCurrentPrices(ctx)

	// Run test function
	err := k.liquidatorKeeper.partialSeizeCDP(ctx, addrs[0], "btc", i(2), i(10000))

	// Check
	require.NoError(t, err)
	cdp, found := k.cdpKeeper.GetCDP(ctx, addrs[0], "btc")
	require.True(t, found)
	require.Equal(t, i(1), cdp.CollateralAmount)
	require.Equal(t, i(6000), cdp.Debt)
}

func TestKeeper_GetSetSeizedDebt(t *testing.T) {
	// Setup
	ctx, k := setupTestKeepers()
	debt := SeizedDebt{i(234247645), i(2343)}

	// Run test function
	k.liquidatorKeeper.setSeizedDebt(ctx, debt)
	readDebt := k.liquidatorKeeper.GetSeizedDebt(ctx)

	// Check
	require.Equal(t, debt, readDebt)
}
