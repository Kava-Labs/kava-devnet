package cdp

/* Notes
- Using sdk.Int as all the number types to maintain compatibility with internal type of sdk.Coin - saves type conversion when doing maths.
  Also it allows for changes to a CDP to be expressed as a +ve or -ve number.
- Only allowing one CDP per account-collateralType pair for now to keep things simple
*/
/* TODO
- what happens if a collateral type is removed from the list of allowed ones?
- Should the values used to generate a key for a stored struct be in the struct?
- standardize collateralType var name
- Should ModifyCDP be split up?
*/

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/kava-labs/usdx/blockchain/x/cdp/pricefeed" // TODO replace with real module
)

const StableDenom = "usdx" // TODO allow to be changed
const GovDenom = "xrs"

// ---------- Keeper ----------
type Keeper struct {
	storeKey       sdk.StoreKey
	pricefeed      pricefeed.Keeper
	bank           bank.Keeper
	paramsSubspace params.Subspace
	cdc            *codec.Codec
}

func NewKeeper(cdc *codec.Codec, storeKey sdk.StoreKey, subspace params.Subspace, pricefeed pricefeed.Keeper, bank bank.Keeper) Keeper {
	subspace = subspace.WithKeyTable(createParamsKeyTable())
	return Keeper{
		storeKey:       storeKey,
		pricefeed:      pricefeed,
		bank:           bank,
		paramsSubspace: subspace,
		cdc:            cdc,
	}
}

// ModifyCDP creates, changes, or deletes a CDP
// TODO add atomic error handling?
func (k Keeper) ModifyCDP(ctx sdk.Context, owner sdk.AccAddress, collateralType string, changeInCollateral sdk.Int, changeInDebt sdk.Int) sdk.Error {
	// Check collateral type ok
	p := k.GetParams(ctx)
	if !p.IsCollateralPresent(collateralType) { // maybe abstract this logic into GetCDP
		return sdk.ErrInternal("collateral type not enabled to create CDPs")
	}

	// Add/Subtract coins from owner
	// collateralType and CDP collat should always match
	var err sdk.Error
	if changeInCollateral.IsPositive() {
		_, _, err = k.bank.SubtractCoins(ctx, owner, sdk.Coins{sdk.NewCoin(collateralType, changeInCollateral)})
	} else {
		_, _, err = k.bank.AddCoins(ctx, owner, sdk.Coins{sdk.NewCoin(collateralType, changeInCollateral)})
	}
	if err != nil {
		return err
	}
	if changeInDebt.IsPositive() {
		_, _, err = k.bank.AddCoins(ctx, owner, sdk.Coins{sdk.NewCoin(StableDenom, changeInDebt)})
	} else {
		_, _, err = k.bank.SubtractCoins(ctx, owner, sdk.Coins{sdk.NewCoin(StableDenom, changeInDebt)})
	}
	if err != nil {
		return err
	}

	// Add/Subtract collateral and debt recorded in CDP
	// Get CDP (or create if not exists)
	cdp, found := k.GetCDP(ctx, owner, collateralType)
	if !found {
		cdp = CDP{Owner: owner, CollateralDenom: collateralType, CollateralAmount: sdk.ZeroInt(), Debt: sdk.ZeroInt()}
	}
	// Add/Subtract collateral and debt
	cdp.CollateralAmount = cdp.CollateralAmount.Add(changeInCollateral)
	if cdp.CollateralAmount.IsNegative() {
		return sdk.ErrInternal(" can't withdraw more collateral than present in CDP")
	}
	cdp.Debt = cdp.Debt.Add(changeInDebt)
	if cdp.Debt.IsNegative() {
		return sdk.ErrInternal("can't pay back more debt than exist in CDP")
	}
	price := k.pricefeed.GetPrice(ctx, cdp.CollateralDenom).Price // or use collateralType
	currentRatio := sdk.NewDecFromInt(cdp.CollateralAmount).Mul(price).Quo(sdk.NewDecFromInt(cdp.Debt))
	if currentRatio.LT(p.GetCollateralParams(cdp.CollateralDenom).LiquidationRatio) { // TODO LT or LTE ?
		return sdk.ErrInternal("Change to CDP would put it below liquidation ratio")
	}
	// TODO check for dust
	// Set CDP
	if cdp.CollateralAmount.IsZero() && cdp.Debt.IsZero() { // TODO maybe abstract this logic into setCDP
		k.deleteCDP(ctx, cdp)
	} else {
		k.setCDP(ctx, cdp) // if subsequent lines fail this needs to be reverted, but the sdk should do that automatically? // this should delete if it's empty
	}

	// Add/Subtract from global debt limit
	gDebt := k.GetGlobalDebt(ctx) // TODO what happens is not found?
	gDebt = gDebt.Add(changeInDebt)
	if gDebt.IsNegative() {
		sdk.ErrInternal("global debt can't be negative") // This should never happen if debt per CDP can't be negative
	}
	if gDebt.GT(p.GlobalDebtLimit) {
		sdk.ErrInternal("change to CDP would put the system over the global debt limit")
	}
	k.setGlobalDebt(ctx, gDebt)

	// Add/Subtract from collateral debt limit
	cStats := k.GetCollateralStats(ctx, cdp.CollateralDenom)
	cStats.TotalDebt = cStats.TotalDebt.Add(changeInDebt)
	if cStats.TotalDebt.IsNegative() {
		sdk.ErrInternal("total debt for this collateral type can't be negative") // This should never happen if debt per CDP can't be negative
	}
	if cStats.TotalDebt.GT(p.GetCollateralParams(cdp.CollateralDenom).DebtLimit) {
		sdk.ErrInternal("change to CDP would put the system over the debt limit for this collateral type")
	}
	k.setCollateralStats(ctx, cStats)

	return nil
}

// convenience function to allow people to give their CDPs to others
func (k Keeper) TransferCDP(ctx sdk.Context, from sdk.AccAddress, to sdk.AccAddress, cdpID string) sdk.Error {
	// TODO
	return nil
}

// Not sure if this is really needed. But it allows the cdp module to track total debt and debt per asset.
func (k Keeper) ConfiscateCDP(ctx sdk.Context, cdpID string) sdk.Error {
	// get CDP
	// empty CDP of collateral and debt, ie set values to zero
	// store CDP

	// update debt per collateral type
	// update global debt

	// update global seized debt ? this is what maker does (Vat.grab) but it's not used anywhere
	return nil
}

func (k Keeper) GetUnderCollateralizedCDPs() sdk.Error {
	// get current prices of assets // priceFeedKeeper.GetCurrentPrice(denom)

	// get an iterator over the CDPs that only includes undercollateralized CDPs
	//    should be possible to store cdps by a key that is their collateral/debt ratio, then the iterator thing can be used to get only the undercollaterized ones (for a given price)

	// combine all the iterators for the different assets?

	// return iterator
	return nil
}

// ---------- Parameter Fetching ----------

func (k Keeper) GetParams(ctx sdk.Context) CdpModuleParams {
	var p CdpModuleParams
	k.paramsSubspace.Get(ctx, cdpModuleParamsKey, &p)
	return p
}

// ---------- Keeper Store Wrappers ----------
// func (k Keeper) getCDPID(owner sdk.AccAddress, collateralType string) string {
// 	return owner.String() + collateralType
// }
// TODO should this be attached to the keeper or not?
func (k Keeper) getCDPKey(owner sdk.AccAddress, collateralType string) []byte {
	return []byte(owner.String() + collateralType)
}
func (k Keeper) GetCDP(ctx sdk.Context, owner sdk.AccAddress, collateralType string) (CDP, bool) {
	// get store
	store := ctx.KVStore(k.storeKey)
	// get CDP
	bz := store.Get(k.getCDPKey(owner, collateralType))
	// unmarshal
	if bz == nil {
		return CDP{}, false
	}
	var cdp CDP
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &cdp)
	return cdp, true
}
func (k Keeper) setCDP(ctx sdk.Context, cdp CDP) {
	// get store
	store := ctx.KVStore(k.storeKey)
	// marshal and set
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(cdp)
	store.Set(k.getCDPKey(cdp.Owner, cdp.CollateralDenom), bz)
	// TODO add to iterator
}
func (k Keeper) deleteCDP(ctx sdk.Context, cdp CDP) {
	// get store
	store := ctx.KVStore(k.storeKey)
	// delete key
	store.Delete(k.getCDPKey(cdp.Owner, cdp.CollateralDenom))
	// TODO remove from iterator
}

var globalDebtKey = []byte("globalDebt")

func (k Keeper) GetGlobalDebt(ctx sdk.Context) sdk.Int {
	// get store
	store := ctx.KVStore(k.storeKey)
	// get bytes
	bz := store.Get(globalDebtKey)
	// unmarshal
	if bz == nil {
		panic("global debt not found") // TODO what is the correct behavior if not found?
	}
	var globalDebt sdk.Int
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &globalDebt)
	return globalDebt
}
func (k Keeper) setGlobalDebt(ctx sdk.Context, globalDebt sdk.Int) {
	// get store
	store := ctx.KVStore(k.storeKey)
	// marshal and set
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(globalDebt)
	store.Set(globalDebtKey, bz)
}

func (k Keeper) getCollateralStatsKey(collateralType string) []byte {
	return []byte(collateralType)
}
func (k Keeper) GetCollateralStats(ctx sdk.Context, collateralType string) CollateralStats {
	// get store
	store := ctx.KVStore(k.storeKey)
	// get bytes
	bz := store.Get(k.getCollateralStatsKey(collateralType))
	// unmarshal
	if bz == nil {
		panic("collateral stats not found") // TODO what is the correct behavior if not found?
	}
	var cStats CollateralStats
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &cStats)
	return cStats
}
func (k Keeper) setCollateralStats(ctx sdk.Context, collateralStats CollateralStats) {
	// get store
	store := ctx.KVStore(k.storeKey)
	// marshal and set
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(collateralStats)
	store.Set(k.getCollateralStatsKey(collateralStats.Denom), bz)
}

// ---------- Weird Bank Stuff ----------
// This only exists because module accounts aren't a thing yet.
// Also because we need module accounts that allow for burning/minting.

// These functions make the CDP module act as a bank keeper, intercepting calls to move coins from the liquidator module account.

// Not sure if module accounts are good, but they make the auction module more general:
// - startAuction would just "mints" coins, relying on calling function to decrement them somewhere
// - closeAuction would have to call something specific for the receiver module to accept coins (like liquidationKeeper.AddStableCoins)

// With account modules all CDP functions can probably use just SendCoins to keep things safe (instead of AddCoins and SubtractCoins)
// type SendKeeper interface {
// 	type ViewKeeper interface {
// 		GetCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
// 		HasCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) bool
// 		Codespace() sdk.CodespaceType
// 	}
// 	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
// 	GetSendEnabled(ctx sdk.Context) bool
// 	SetSendEnabled(ctx sdk.Context, enabled bool)
// }

var LiquidatorAccountAddress = sdk.AccAddress([]byte("whatever"))
var liquidatorAccountKey = []byte("liquidatorAccount")

type LiquidatorModuleAccount struct {
	Coins sdk.Coins // keeps track of seized collateral and surplus usdx
}

func (k Keeper) AddCoins(ctx sdk.Context, address sdk.AccAddress, amount sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error) {
	// intercept module account
	if address.Equals(LiquidatorAccountAddress) {
		// remove gov token from list
		filteredCoins := stripGovCoin(amount)
		// add coins to module account
		lma := k.getLiquidatorModuleAccount(ctx)
		updatedCoins := lma.Coins.Plus(filteredCoins)
		if updatedCoins.IsAnyNegative() {
			panic("") // TODO return error, follow how bank does it
		}
		lma.Coins = updatedCoins
		k.setLiquidatorModuleAccount(ctx, lma)
		return updatedCoins, sdk.Tags{}, nil // TODO add tags
	} else {
		return k.bank.AddCoins(ctx, address, amount)
	}
}

// TODO abstract stuff better
func (k Keeper) SubtractCoins(ctx sdk.Context, address sdk.AccAddress, amount sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error) {
	// intercept module account
	if address.Equals(LiquidatorAccountAddress) {
		// remove gov token from list
		filteredCoins := stripGovCoin(amount)
		// subtract coins from module account
		lma := k.getLiquidatorModuleAccount(ctx)
		updatedCoins := lma.Coins.Minus(filteredCoins)
		if updatedCoins.IsAnyNegative() {
			panic("") // TODO return error, follow how bank does it
		}
		lma.Coins = updatedCoins
		k.setLiquidatorModuleAccount(ctx, lma)
		return updatedCoins, sdk.Tags{}, nil // TODO add tags
	} else {
		return k.bank.SubtractCoins(ctx, address, amount)
	}
}
func (k Keeper) GetCoins(ctx sdk.Context, address sdk.AccAddress) sdk.Coins {
	return k.getLiquidatorModuleAccount(ctx).Coins
	// TODO Should this return anything for the XRS balance? Currently returns zero
}
func (k Keeper) HasCoins(ctx sdk.Context, address sdk.AccAddress, amount sdk.Coins) bool {
	return k.getLiquidatorModuleAccount(ctx).Coins.IsAllGTE(stripGovCoin(amount))
	// TODO test
}

func (k Keeper) getLiquidatorModuleAccount(ctx sdk.Context) LiquidatorModuleAccount {
	// get store
	store := ctx.KVStore(k.storeKey)
	// get bytes
	bz := store.Get(liquidatorAccountKey)
	if bz == nil {
		panic("liquidator account not found")
	}
	// unmarshal
	var lma LiquidatorModuleAccount
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &lma)
	return lma
}
func (k Keeper) setLiquidatorModuleAccount(ctx sdk.Context, lma LiquidatorModuleAccount) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(lma)
	store.Set(liquidatorAccountKey, bz)
}
func stripGovCoin(coins sdk.Coins) sdk.Coins {
	filteredCoins := sdk.Coins{}
	for _, c := range coins {
		if c.Denom != GovDenom {
			filteredCoins = append(filteredCoins, c)
		}
	}
	return filteredCoins
}
