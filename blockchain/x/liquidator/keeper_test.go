package liquidator

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"

	"github.com/kava-labs/usdx/blockchain/x/pricefeed"
	"github.com/kava-labs/usdx/blockchain/x/cdp"
)

func TestKeeper_StartCollateralAuction(t *testing.T) {
	// Setup keeper and context
	ctx, k := setupTestKeepers()
	// Create seized CDP
	_, addrs := mock.GeneratePrivKeyAddressPairs(1)
	cdp := SeizedCDP{Owner: addrs[0], CollateralAmount: i(10), CollateralDenom: "btc", Debt: i(500)}
	k.liquidatorKeeper.setSeizedCDP(ctx, cdp)
	k.liquidatorKeeper.bankKeeper.AddCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress(), cs(sdk.NewCoin(cdp.CollateralDenom, cdp.CollateralAmount)))

	// Start auction
	auctionID, err := k.liquidatorKeeper.StartCollateralAuction(ctx, cdp.Owner, cdp.CollateralDenom)

	// Check CDP is changed correctly
	require.NoError(t, err)
	_, found := k.liquidatorKeeper.GetSeizedCDP(ctx, cdp.Owner, cdp.CollateralDenom)
	require.False(t, found)
	_, found = k.auctionKeeper.GetAuction(ctx, auctionID)
	require.True(t, found)
}

func TestKeeper_StartDebtAuction(t *testing.T) {
	// Setup
	ctx, k := setupTestKeepers()
	initSDebt := SeizedDebt{i(2000),i(0)}
	k.liquidatorKeeper.setSeizedDebt(ctx, initSDebt)

	// Execute
	auctionID, err := k.liquidatorKeeper.StartDebtAuction(ctx)

	// Check
	require.NoError(t, err)
	require.Equal(t,
		SeizedDebt{
			initSDebt.Total,
			initSDebt.SentToAuction.Add(DebtAuctionSize),
		},
		k.liquidatorKeeper.GetSeizedDebt(ctx),
	)
	_, found := k.auctionKeeper.GetAuction(ctx, auctionID)
	require.True(t, found)
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

func TestKeeper_PartialSeizeCDP(t *testing.T) {
	// Setup
	ctx, k := setupTestKeepers()

	_, addrs := mock.GeneratePrivKeyAddressPairs(1)

	cdp.InitGenesis(ctx, k.cdpKeeper, cdp.DefaultGenesisState())
	pricefeed.InitGenesis(ctx, k.pricefeedKeeper, pricefeed.GenesisState{Assets: []pricefeed.Asset{{"btc", "a description"}}})
	k.pricefeedKeeper.SetPrice(ctx, addrs[0], "btc", sdk.MustNewDecFromStr("8000.00"), i(999999999))
	k.pricefeedKeeper.SetCurrentPrices(ctx)
	k.bankKeeper.AddCoins(ctx, addrs[0], cs(c("btc", 100)))

	k.cdpKeeper.ModifyCDP(ctx, addrs[0], "btc", i(3), i(16000))

	k.pricefeedKeeper.SetPrice(ctx, addrs[0], "btc", sdk.MustNewDecFromStr("7999.99"), i(999999999))
	k.pricefeedKeeper.SetCurrentPrices(ctx)

	// Run test function
	err := k.liquidatorKeeper.PartialSeizeCDP(ctx, addrs[0], "btc", i(2), i(10000))

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
