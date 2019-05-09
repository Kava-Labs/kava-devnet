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

func (k Keeper) StartCollateralAuction(ctx sdk.Context) { // aka Cat.flip
	// get Seized CDP (fom originalOwner and collateral denom)
	// calculate how much dai to try and raise in this auction	// daiToRaise := min(seizedCdp.Debt, params.maxDaiToBeRaised)
	// calculate how much collateral to sell (based on above)
	// subtract these values from the seizedCDP
	// start "forward reverse" auction type // auctionKeeper.StartForwardReverseAuction(moduleAccountAddress, lot, daiToRaise, seizedCDP.OriginalOwner)
	// store seizedCDP
	// result: dai in transferred to moduleAccount, collateral is transferred from module account to buyer, (and any excess collateral is transferred to original CDP owner)
}

func (k Keeper) StartDebtAuction(ctx sdk.Context) { // aka Vow.flop
	// get seizedDebt
	// check the seized debt is above a threshold
	// reduce debt
	// start reverse auction, selling minted XRS for USDX // auctionKeeper.StartReverseAuction(moduleAccountAddress, params.daiToBeRaised, sdk.Coin{"MKR", 2 ^ 256 - 1}) // infinite MKR :(
	// result: minted MKR moved to highest bidder, dai moved to moduleAccount
}
func (k Keeper) StartSurplusAuction(ctx sdk.Context) { // aka Vow.flap
	// get surplus amount
	// check there is enough surplus to be sold
	// subtract coins (maybe not do above lines?)
	// start normal auction for usdx // aK.StartForwardAuction(moduleAccountAddress, daiToBeSold, sdk.Coin{"XRS", 0})
	// end result: dai removed from module account (eventually to buyer), XRS transferred to module account
}

func (k Keeper) ConfiscateUnderCollateralizedCDP(ctx sdk.Context, owner sdk.AccAddress, collateralDenom string) sdk.Error { // aka Cat.bite
	// Get the CDP
	cdp, found := k.cdpKeeper.GetCDP(ctx, owner, collateralDenom)
	if !found {
		return sdk.ErrInternal("cdp not found")
	}
	// Check CDP is undercollateralized (?)

	// increment the total seized debt by cdp.debt. // aka Awe
	seizedDebt := k.GetSeizedDebt(ctx).Add(cdp.Debt)
	k.setSeizedDebt(ctx, seizedDebt)

	// create a SeizedCDP. This is needed because the collateral is auctioned off per CDP. // k.setSeizedCDP(SeizedCDP{cdp.Debt, cdp.Collateral}) // aka create Cat.Flip object
	seizedCDP := SeizedCDP{OriginalOwner: cdp.Owner, CollateralDenom: cdp.CollateralDenom, CollateralAmount: cdp.CollateralAmount, Debt: cdp.Debt}
	k.setSeizedCDP(ctx, seizedCDP)

	// add cdp.collateral amount of coins to the moduleAccount (so they can be transferred to the auction later)
	coins := sdk.Coins{sdk.NewCoin(cdp.CollateralDenom, cdp.CollateralAmount)}
	k.bankKeeper.AddCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress(), coins)

	// Seize the cdp in the cdp module
	k.cdpKeeper.ConfiscateCDP(ctx, owner, collateralDenom) // just empties it and updates the global debt in cdp the module
	return nil
}

func (k Keeper) settleDebt(ctx sdk.Context) { // When should this be called? At the start of every entry point here?
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
	store.Set(k.getSeizedCDPKey(cdp.OriginalOwner, cdp.CollateralDenom), bz)
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
