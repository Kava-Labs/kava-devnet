package liquidator

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
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
	initSDebt := i(2000)
	k.liquidatorKeeper.setSeizedDebt(ctx, initSDebt)

	// Execute
	auctionID, err := k.liquidatorKeeper.StartDebtAuction(ctx)

	// Check
	require.Nil(t, err)
	require.Equal(t,
		initSDebt.Sub(DebtAuctionSize),
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

func TestKeeper_GetSetSeizedDebt(t *testing.T) {
	// Setup
	ctx, k := setupTestKeepers()
	debt := i(528452456344)

	// Run test function
	k.liquidatorKeeper.setSeizedDebt(ctx, debt)
	readDebt := k.liquidatorKeeper.GetSeizedDebt(ctx)

	// Check
	require.Equal(t, debt, readDebt)
}
