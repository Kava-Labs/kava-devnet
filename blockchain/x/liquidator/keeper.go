package liquidator

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/kava-labs/usdx/blockchain/x/auction"
)

type Keeper struct {
	cdc            *codec.Codec
	paramsSubspace params.Subspace
	storeKey       sdk.StoreKey
	cdpKeeper      cdpKeeper
	auctionKeeper  auctionKeeper
	bankKeeper     bankKeeper
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, subspace params.Subspace, cdpKeeper cdpKeeper, auctionKeeper auctionKeeper, bankKeeper bankKeeper) Keeper {
	subspace = subspace.WithKeyTable(createParamsKeyTable())
	return Keeper{
		cdc:            cdc,
		paramsSubspace: subspace,
		storeKey:       storeKey,
		cdpKeeper:      cdpKeeper,
		auctionKeeper:  auctionKeeper,
		bankKeeper:     bankKeeper,
	}
}

var ( // TODO move into params, pick good defaults
	CollateralAuctionSize = sdk.NewInt(1000) // known as Cat.ilk[n].lump in maker
	DebtAuctionSize       = sdk.NewInt(1000) // known as Vow.sump in maker
	SurplusAuctionSize    = sdk.NewInt(1000) // known as Voe.bump in maker
)

// SeizeAndStartCollateralAuction pulls collateral out of a CDP and sells it in an auction for stable coin. Excess collateral goes to the original CDP owner.
// Known as Cat.bite in maker
// result: stable coin is transferred to module account, collateral is transferred from module account to buyer, (and any excess collateral is transferred to original CDP owner)
func (k Keeper) SeizeAndStartCollateralAuction(ctx sdk.Context, owner sdk.AccAddress, collateralDenom string) (auction.ID, sdk.Error) {
	// Get CDP
	cdp, found := k.cdpKeeper.GetCDP(ctx, owner, collateralDenom)
	if !found {
		return 0, sdk.ErrInternal("")
	}

	// Calculate amount of collateral to sell in this auction
	collateralToSell := sdk.MinInt(cdp.CollateralAmount, CollateralAuctionSize)
	// Calculate the corresponding maximum amount of stable coin to raise TODO test maths
	stableToRaise := sdk.NewDecFromInt(collateralToSell).Quo(sdk.NewDecFromInt(cdp.CollateralAmount)).Mul(sdk.NewDecFromInt(cdp.Debt)).RoundInt()

	// Seize the collateral and debt from the CDP
	err := k.partialSeizeCDP(ctx, owner, collateralDenom, collateralToSell, stableToRaise)
	if err != nil {
		return 0, err
	}

	// Start "forward reverse" auction type
	lot := sdk.NewCoin(cdp.CollateralDenom, collateralToSell)
	maxBid := sdk.NewCoin(k.cdpKeeper.GetStableDenom(), stableToRaise)
	auctionID, err := k.auctionKeeper.StartForwardReverseAuction(ctx, k.cdpKeeper.GetLiquidatorAccountAddress(), lot, maxBid, owner)
	if err != nil {
		panic(err) // TODO how can errors here be handled to be safe with the state update in PartialSeizeCDP?
	}
	return auctionID, nil
}

// StartDebtAuction sells off minted gov coin to raise set amounts of stable coin.
// Known as Vow.flop in maker
// result: minted gov coin moved to highest bidder, stable coin moved to moduleAccount
func (k Keeper) StartDebtAuction(ctx sdk.Context) (auction.ID, sdk.Error) {

	// Ensure amount of seized stable coin is 0 (ie Joy = 0)
	stableCoins := k.bankKeeper.GetCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress()).AmountOf(k.cdpKeeper.GetStableDenom())
	if !stableCoins.IsZero() {
		return 0, sdk.ErrInternal("debt auction cannot be started as there is outstanding stable coins")
	}

	// check the seized debt is above a threshold
	seizedDebt := k.GetSeizedDebt(ctx)
	if seizedDebt.Available().LT(DebtAuctionSize) {
		return 0, sdk.ErrInternal("not enough seized debt to start an auction")
	}
	// start reverse auction, selling minted gov coin for stable coin
	auctionID, err := k.auctionKeeper.StartReverseAuction(
		ctx,
		k.cdpKeeper.GetLiquidatorAccountAddress(),
		sdk.NewCoin(k.cdpKeeper.GetStableDenom(), DebtAuctionSize),
		sdk.NewInt64Coin(k.cdpKeeper.GetGovDenom(), 2^255-1), // TODO is there a way to avoid potentially minting infinite gov coin?
	)
	if err != nil {
		return 0, err
	}
	// Record amount of debt sent for auction. Debt can only be reduced in lock step with reducing stable coin
	seizedDebt.SentToAuction = seizedDebt.SentToAuction.Add(DebtAuctionSize)
	k.setSeizedDebt(ctx, seizedDebt)
	return auctionID, nil
}

// With no stability and liquidation fees, surplus auctions can never be run.
// StartSurplusAuction sells off excess stable coin in exchange for gov coin, which is burned
// Known as Vow.flap in maker
// result: stable coin removed from module account (eventually to buyer), gov coin transferred to module account
// func (k Keeper) StartSurplusAuction(ctx sdk.Context) (auction.ID, sdk.Error) {

// 	// TODO ensure seized debt is 0

// 	// check there is enough surplus to be sold
// 	surplus := k.bankKeeper.GetCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress()).AmountOf(k.cdpKeeper.GetStableDenom())
// 	if surplus.LT(SurplusAuctionSize) {
// 		return 0, sdk.ErrInternal("not enough surplus stable coin to start an auction")
// 	}
// 	// start normal auction, selling stable coin
// 	auctionID, err := k.auctionKeeper.StartForwardAuction(
// 		ctx,
// 		k.cdpKeeper.GetLiquidatorAccountAddress(),
// 		sdk.NewCoin(k.cdpKeeper.GetStableDenom(), SurplusAuctionSize),
// 		sdk.NewInt64Coin(k.cdpKeeper.GetGovDenom(), 0),
// 	)
// 	if err != nil {
// 		return 0, err
// 	}
// 	// Starting the auction will remove coins from the account, so they don't need modified here.
// 	return auctionID, nil
// }

// PartialSeizeCDP seizes some collateral and debt from an under-collateralized CDP.
func (k Keeper) partialSeizeCDP(ctx sdk.Context, owner sdk.AccAddress, collateralDenom string, collateralToSeize sdk.Int, debtToSeize sdk.Int) sdk.Error { // aka Cat.bite
	// Seize debt and collateral in the cdp module. This also validates the inputs.
	err := k.cdpKeeper.PartialSeizeCDP(ctx, owner, collateralDenom, collateralToSeize, debtToSeize)
	if err != nil {
		return err // cdp could be not found, or not under collateralized, or inputs invalid
	}

	// increment the total seized debt (Awe) by cdp.debt
	seizedDebt := k.GetSeizedDebt(ctx)
	seizedDebt.Total = seizedDebt.Total.Add(debtToSeize)
	k.setSeizedDebt(ctx, seizedDebt)

	// add cdp.collateral amount of coins to the moduleAccount (so they can be transferred to the auction later)
	coins := sdk.NewCoins(sdk.NewCoin(collateralDenom, collateralToSeize))
	_, _, err = k.bankKeeper.AddCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress(), coins)
	if err != nil {
		panic(err) // TODO this shouldn't happen?
	}
	return nil
}

// SettleDebt removes equal amounts of debt and stable coin from the liquidator's reserves (and also updates the global debt in the cdp module).
// This is called in the handler when a debt or surplus auction is started
// TODO Should this be called with an amount, rather than annihilating the maximum?
func (k Keeper) settleDebt(ctx sdk.Context) sdk.Error {
	// Calculate max amount of debt and stable coins that can be settled (ie annihilated)
	debt := k.GetSeizedDebt(ctx)
	stableCoins := k.bankKeeper.GetCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress()).AmountOf(k.cdpKeeper.GetStableDenom())
	settleAmount := sdk.MinInt(debt.Total, stableCoins)

	// Call cdp module to reduce GlobalDebt. This can fail if genesis not set
	err := k.cdpKeeper.ReduceGlobalDebt(ctx, settleAmount)
	if err != nil {
		return err
	}

	// Decrement total seized debt (also decrement from SentToAuction debt)
	updatedDebt, err := debt.Settle(settleAmount)
	if err != nil {
		return err // this should not error in this context
	}
	k.setSeizedDebt(ctx, updatedDebt)

	// Subtract stable coin from moduleAccout
	k.bankKeeper.SubtractCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress(), sdk.Coins{sdk.NewCoin(k.cdpKeeper.GetStableDenom(), settleAmount)})
	return nil
}

// ---------- Module Parameters ----------

func (k Keeper) GetParams(ctx sdk.Context) LiquidatorModuleParams {
	var params LiquidatorModuleParams
	k.paramsSubspace.Get(ctx, moduleParamsKey, &params)
	return params
}

// This is only needed to be able to setup the store from the genesis file. The keeper should not change any of the params itself.
func (k Keeper) setParams(ctx sdk.Context, params LiquidatorModuleParams) {
	k.paramsSubspace.Set(ctx, moduleParamsKey, &params)
}

// ---------- Store Wrappers ----------

func (k Keeper) getSeizedDebtKey() []byte {
	return []byte("seizedDebt")
}
func (k Keeper) GetSeizedDebt(ctx sdk.Context) SeizedDebt {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getSeizedDebtKey())
	if bz == nil {
		// TODO make initial seized debt and CDPs configurable at genesis, then panic here if not found
		bz = k.cdc.MustMarshalBinaryLengthPrefixed(SeizedDebt{sdk.ZeroInt(), sdk.ZeroInt()})
	}
	var seizedDebt SeizedDebt
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &seizedDebt)
	return seizedDebt
}
func (k Keeper) setSeizedDebt(ctx sdk.Context, debt SeizedDebt) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(debt)
	store.Set(k.getSeizedDebtKey(), bz)
}
