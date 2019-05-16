package liquidator

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/usdx/blockchain/x/auction"
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

var ( // TODO move into params, pick good defaults
	CollateralAuctionMaxBid = sdk.NewInt(1000) // known as Cat.ilk[n].lump in maker
	DebtAuctionSize         = sdk.NewInt(1000) // known as Vow.sump in maker
	SurplusAuctionSize      = sdk.NewInt(1000) // known as Voe.bump in maker
)

// StartCollateralAuction pulls collateral out of a (seized) CDP and sells it in an auction for stable coin. Excess collateral goes to the original CDP owner
// Known as Cat.flip in maker
// result: stable coin is transferred to moduleAccount, collateral is transferred from module account to buyer, (and any excess collateral is transferred to original CDP owner)
func (k Keeper) StartCollateralAuction(ctx sdk.Context, originalOwner sdk.AccAddress, collateralDenom string) (auction.ID, sdk.Error) {
	// Get Seized CDP
	seizedCDP, found := k.GetSeizedCDP(ctx, originalOwner, collateralDenom)
	if !found {
		return 0, sdk.ErrInternal("CDP not found")
	}
	// Calculate how much stable coin to try and raise in this auction
	stableToRaise := sdk.MinInt(seizedCDP.Debt, CollateralAuctionMaxBid)
	// calculate how much collateral to sell: collateralToSell/collateral = stableToRaise/debt
	// TODO test the maths here
	collateralToSell := sdk.NewDecFromInt(stableToRaise).Quo(sdk.NewDecFromInt(seizedCDP.Debt)).Mul(sdk.NewDecFromInt(seizedCDP.CollateralAmount)).RoundInt()

	// Subtract these values from the seizedCDP
	seizedCDP.Debt = seizedCDP.Debt.Sub(stableToRaise)
	seizedCDP.CollateralAmount = seizedCDP.CollateralAmount.Sub(collateralToSell)
	// Start "forward reverse" auction type
	lot := sdk.NewCoin(seizedCDP.CollateralDenom, collateralToSell)
	maxBid := sdk.NewCoin(k.cdpKeeper.GetStableDenom(), stableToRaise)
	auctionID, err := k.auctionKeeper.StartForwardReverseAuction(ctx, k.cdpKeeper.GetLiquidatorAccountAddress(), lot, maxBid, seizedCDP.Owner)
	if err != nil {
		return 0, err
	}
	// Store seizedCDP
	if seizedCDP.CollateralAmount.IsZero() && seizedCDP.Debt.IsZero() { // TODO maybe abstract this logic into setCDP
		k.deleteSeizedCDP(ctx, seizedCDP)
	} else {
		k.setSeizedCDP(ctx, seizedCDP)
	}
	return auctionID, nil
}

// StartDebtAuction sells off minted gov coin to raise set amounts of stable coin.
// Known as Vow.flop in maker
// result: minted gov coin moved to highest bidder, stable coin moved to moduleAccount
func (k Keeper) StartDebtAuction(ctx sdk.Context) (auction.ID, sdk.Error) {
	// TODO where is the best place for settleDebt to be called? Should it be a message type?
	k.settleDebt(ctx)

	// check the seized debt is above a threshold
	seizedDebt := k.GetSeizedDebt(ctx)
	if seizedDebt.LT(DebtAuctionSize) {
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
	// reduce debt
	k.setSeizedDebt(ctx, seizedDebt.Sub(DebtAuctionSize))
	return auctionID, nil
}

// StartSurplusAuction sells off excess stable coin in exchange for gov coin, which is burned
// Known as Vow.flap in maker
// result: stable coin removed from module account (eventually to buyer), gov coin transferred to module account
func (k Keeper) StartSurplusAuction(ctx sdk.Context) (auction.ID, sdk.Error) {
	// TODO where is the best place for settleDebt to be called? Should it be a message type?
	k.settleDebt(ctx)

	// check there is enough surplus to be sold
	surplus := k.bankKeeper.GetCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress()).AmountOf(k.cdpKeeper.GetStableDenom())
	if surplus.LT(SurplusAuctionSize) {
		return 0, sdk.ErrInternal("not enough surplus stable coin to start an auction")
	}
	// start normal auction, selling stable coin
	auctionID, err := k.auctionKeeper.StartForwardAuction(
		ctx,
		k.cdpKeeper.GetLiquidatorAccountAddress(),
		sdk.NewCoin(k.cdpKeeper.GetStableDenom(), SurplusAuctionSize),
		sdk.NewInt64Coin(k.cdpKeeper.GetGovDenom(), 0),
	)
	if err != nil {
		return 0, err
	}
	// Starting the auction will remove coins from the account, so they don't need modified here.
	return auctionID, nil
}

func (k Keeper) SeizeUnderCollateralizedCDP(ctx sdk.Context, owner sdk.AccAddress, collateralDenom string) sdk.Error { // aka Cat.bite
	// Seize the cdp in the cdp module
	cdp, err := k.cdpKeeper.SeizeCDP(ctx, owner, collateralDenom) // just empties it and updates the global debt in cdp the module
	if err != nil {
		return err // cdp could be not found, or not under collateralized
	}

	// increment the total seized debt by cdp.debt. // aka Awe
	seizedDebt := k.GetSeizedDebt(ctx).Add(cdp.Debt)
	k.setSeizedDebt(ctx, seizedDebt)

	// create a SeizedCDP. This is needed because the collateral is auctioned off per CDP. // aka create Cat.Flip object
	k.setSeizedCDP(ctx, cdp)

	// add cdp.collateral amount of coins to the moduleAccount (so they can be transferred to the auction later)
	coins := sdk.NewCoins(sdk.NewCoin(cdp.CollateralDenom, cdp.CollateralAmount))
	_, _, err = k.bankKeeper.AddCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress(), coins)
	if err != nil {
		panic(err) // TODO this shouldn't happen?
	}
	return nil
}

// SettleDebt removes equal amounts of debt and stable coin from the liquidator's reserves (and also updates the total debt counter)
// Debt and stable coin balances decreased by this function and by starting surplus/debt auctions
// Debt is incremented only by SeizeUnderCollateralizedCDP
// Stable coins are incremented only by auction.PlaceBid and auction.Close
// Start Debt/Surplus Auction is only function that depends on debt/stableCoin balances
// TODO When should this be called? Should it be called with an amount, rather than annihilating the maximum? Currently called before starting the surplus/debt auctions
func (k Keeper) settleDebt(ctx sdk.Context) sdk.Error {
	// calculate max amount of debt and stable coins that can be settled (ie annihilated)
	debt := k.GetSeizedDebt(ctx)
	stableCoins := k.bankKeeper.GetCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress()).AmountOf(k.cdpKeeper.GetStableDenom())
	settleAmount := sdk.MinInt(debt, stableCoins)

	// Call cdp module to reduce GlobalDebt. This can fail if genesis not set
	err := k.cdpKeeper.ReduceGlobalDebt(ctx, settleAmount)
	if err != nil {
		return err
	}

	// decrement total seized debt by above amount
	k.setSeizedDebt(ctx, debt.Sub(settleAmount))

	// subtract stable coin from moduleAccout
	k.bankKeeper.SubtractCoins(ctx, k.cdpKeeper.GetLiquidatorAccountAddress(), sdk.Coins{sdk.NewCoin(k.cdpKeeper.GetStableDenom(), settleAmount)})
	return nil
}

// ---------- Store Wrappers ----------

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
func (k Keeper) deleteSeizedCDP(ctx sdk.Context, cdp SeizedCDP) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(k.getSeizedCDPKey(cdp.Owner, cdp.CollateralDenom))
}

// TODO setting and getting seized debt could be abstracted into add/subtract
// seized debt is known as Awe in maker
func (k Keeper) getSeizedDebtKey() []byte {
	return []byte("seizedDebt")
}
func (k Keeper) GetSeizedDebt(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getSeizedDebtKey())
	if bz == nil {
		// TODO make initial seized debt and CDPs configurable at genesis, then panic here if not found
		bz = k.cdc.MustMarshalBinaryLengthPrefixed(sdk.NewInt(0))
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
