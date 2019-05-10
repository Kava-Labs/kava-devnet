package liquidator

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	//"github.com/cosmos/cosmos-sdk/x/bank"
	//"github.com/kava-labs/usdx/blockchain/x/auction"
	//"github.com/kava-labs/usdx/blockchain/x/cdp"
)

type Keeper struct {
	cdc           *codec.Codec
	storeKey      sdk.StoreKey
	cdpKeeper     cdpKeeper
	auctionKeeper auctionKeeper
	bankKeeper    bankKeeper
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, cdpKeeper cdpKeeper, auctionKeeper auctionKeeper, bankKeeper bankKeeper) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		cdpKeeper:     cdpKeeper,
		auctionKeeper: auctionKeeper,
		bankKeeper:    bankKeeper,
	}
}

var ( // TODO move into params, pick defaults
	CollateralAuctionMaxBid = sdk.NewInt(1000)
	DebtAuctionSize         = sdk.NewInt(1000)
	SurplusAuctionSize      = sdk.NewInt(1000)
)

// result: stable coin is transferred to moduleAccount, collateral is transferred from module account to buyer, (and any excess collateral is transferred to original CDP owner)
func (k Keeper) StartCollateralAuction(ctx sdk.Context, originalOwner sdk.AccAddress, collateralDenom string) sdk.Error { // aka Cat.flip
	// Get Seized CDP
	seizedCDP, found := k.GetSeizedCDP(ctx, originalOwner, collateralDenom)
	if !found {
		return sdk.ErrInternal("CDP not found")
	}
	// Calculate how much stable coin to try and raise in this auction
	//params := GetParams(ctx) // TODO write this and fix
	stableToRaise := sdk.MinInt(seizedCDP.Debt, CollateralAuctionMaxBid) // TODO better name
	// calculate how much collateral to sell: collateralToSell/collateral = stableToRaise/debt
	// TODO test the maths here
	collateralToSell := sdk.NewDecFromInt(stableToRaise).Quo(sdk.NewDecFromInt(seizedCDP.Debt)).Mul(sdk.NewDecFromInt(seizedCDP.CollateralAmount)).RoundInt()

	// Subtract these values from the seizedCDP
	seizedCDP.Debt = seizedCDP.Debt.Sub(stableToRaise)
	seizedCDP.CollateralAmount = seizedCDP.CollateralAmount.Sub(collateralToSell)
	// Start "forward reverse" auction type
	lot := sdk.NewCoin(seizedCDP.CollateralDenom, collateralToSell)
	maxBid := sdk.NewCoin(k.cdpKeeper.GetStableDenom(), stableToRaise)
	err := k.auctionKeeper.StartForwardReverseAuction(ctx, k.cdpKeeper.GetLiquidatorAccountAddress(), lot, maxBid, seizedCDP.Owner)
	if err != nil {
		return err
	}
	// Store seizedCDP
	k.setSeizedCDP(ctx, seizedCDP) // TODO delete if empty
	return nil
}

// result: minted gov coin moved to highest bidder, stable coin moved to moduleAccount
func (k Keeper) StartDebtAuction(ctx sdk.Context) sdk.Error { // aka Vow.flop
	// TODO call k.settleDebt(ctx) ?
	// get seizedDebt
	seizedDebt := k.GetSeizedDebt(ctx)
	// check the seized debt is above a threshold
	if seizedDebt.LT(DebtAuctionSize) {
		return sdk.ErrInternal("not enough seized debt to start an auction")
	}
	// start reverse auction, selling minted gov coin for stable coin
	err := k.auctionKeeper.StartReverseAuction(
		ctx,
		k.cdpKeeper.GetLiquidatorAccountAddress(),
		sdk.NewCoin(k.cdpKeeper.GetStableDenom(), DebtAuctionSize),
		sdk.NewInt64Coin(k.cdpKeeper.GetGovDenom(), 2^255-1), // TODO is there a way to avoid potentially minting infinite gov coin?
	)
	if err != nil {
		return err
	}
	// reduce debt
	k.setSeizedDebt(ctx, seizedDebt.Sub(DebtAuctionSize))
	return nil
}

// end result: stable coin removed from module account (eventually to buyer), gov coin transferred to module account
func (k Keeper) StartSurplusAuction(ctx sdk.Context) sdk.Error { // aka Vow.flap
	// TODO call k.settleDebt(ctx) ?

	// check there is enough surplus to be sold
	surplus := k.bankKeeper.GetCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress()).AmountOf(k.cdpKeeper.GetStableDenom())
	if surplus.LT(SurplusAuctionSize) {
		return sdk.ErrInternal("not enough surplus stable coin to start an auction")
	}
	// start normal auction, selling stable coin
	err := k.auctionKeeper.StartForwardAuction(
		ctx,
		k.cdpKeeper.GetLiquidatorAccountAddress(),
		sdk.NewCoin(k.cdpKeeper.GetStableDenom(), SurplusAuctionSize),
		sdk.NewInt64Coin(k.cdpKeeper.GetGovDenom(), 0),
	)
	if err != nil {
		return err
	}
	// Starting the auction will remove coins from the account, so they don't need modified here.
	return nil
}

func (k Keeper) SeizeUnderCollateralizedCDP(ctx sdk.Context, owner sdk.AccAddress, collateralDenom string) sdk.Error { // aka Cat.bite
	// Seize the cdp in the cdp module
	cdp, err := k.cdpKeeper.SeizeCDP(ctx, owner, collateralDenom) // just empties it and updates the global debt in cdp the module
	if err != nil {
		return err // cdp could be not found, or not undercollateralized
	}

	// increment the total seized debt by cdp.debt. // aka Awe
	seizedDebt := k.GetSeizedDebt(ctx).Add(cdp.Debt)
	k.setSeizedDebt(ctx, seizedDebt)

	// create a SeizedCDP. This is needed because the collateral is auctioned off per CDP. // k.setSeizedCDP(SeizedCDP{cdp.Debt, cdp.Collateral}) // aka create Cat.Flip object
	//seizedCDP := SeizedCDP{OriginalOwner: cdp.Owner, CollateralDenom: cdp.CollateralDenom, CollateralAmount: cdp.CollateralAmount, Debt: cdp.Debt}
	k.setSeizedCDP(ctx, cdp)

	// add cdp.collateral amount of coins to the moduleAccount (so they can be transferred to the auction later)
	coins := sdk.NewCoins(sdk.NewCoin(cdp.CollateralDenom, cdp.CollateralAmount))
	_, _, err = k.bankKeeper.AddCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress(), coins)
	if err != nil {
		panic(err) // TODO this shouldn't happen?
	}
	return nil
}

// Only this function decrements debt and stable coin balances
// Debt is incremented only by SeizeUnderCollateralizedCDP
// Stable coins are incremented only by auction.PlaceBid and auction.Close
// Start Debt/Surplus Auction is only function that depends on debt/stableCoin balances
// When should this be called?
// TODO Fix Bug - this does not reduce the total debt counter in the CDP module
func (k Keeper) settleDebt(ctx sdk.Context) {
	// calculate max amount of debt and stable coins that can be settled (ie annihilated)
	debt := k.GetSeizedDebt(ctx)
	stableCoins := k.bankKeeper.GetCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress()).AmountOf(k.cdpKeeper.GetStableDenom())
	settleAmount := sdk.MinInt(debt, stableCoins)

	// decrement total seized debt by above amount
	k.setSeizedDebt(ctx, debt.Sub(settleAmount))

	// subtract stable coin from moduleAccout
	k.bankKeeper.SubtractCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress(), sdk.Coins{sdk.NewCoin(k.cdpKeeper.GetStableDenom(), settleAmount)})
}

func (k Keeper) getSeizedCDPKey(owner sdk.AccAddress, collateralDenom string) []byte {
	return []byte(owner.String() + collateralDenom)
}
func (k Keeper) GetSeizedCDP(ctx sdk.Context, originalOwner sdk.AccAddress, collateralDenom string) (SeizedCDP, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getSeizedCDPKey(originalOwner, collateralDenom))
	if bz == nil {
		return SeizedCDP{}, false
	}
	var seizedCdp SeizedCDP
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &seizedCdp)
	return seizedCdp, true
}
func (k Keeper) setSeizedCDP(ctx sdk.Context, cdp SeizedCDP) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(cdp)
	store.Set(k.getSeizedCDPKey(cdp.Owner, cdp.CollateralDenom), bz)
}

// TODO could abstract setting and getting seized debt into add/subtract
func (k Keeper) getSeizedDebtKey() []byte {
	return []byte("seizedDebt")
}
func (k Keeper) GetSeizedDebt(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getSeizedDebtKey())
	if bz == nil {
		panic("seized debt not set")
	}
	var seizedDebt sdk.Int
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &seizedDebt)
	return seizedDebt
}
func (k Keeper) setSeizedDebt(ctx sdk.Context, debt sdk.Int) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(debt)
	store.Set(k.getSeizedDebtKey(), bz)
}
