package cdp

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"

	pricefeed "github.com/kava-labs/usdx/blockchain/x/cdp/mockpricefeed" // TODO replace with real module
)

const StableDenom = "usdx" // TODO allow to be changed
const GovDenom = "xrs"

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
// TODO can/should this function be split up?
func (k Keeper) ModifyCDP(ctx sdk.Context, owner sdk.AccAddress, collateralDenom string, changeInCollateral sdk.Int, changeInDebt sdk.Int) sdk.Error {

	// Phase 1: Get state, make changes in memory and check if they're ok.

	// Check collateral type ok
	p := k.GetParams(ctx)
	if !p.IsCollateralPresent(collateralDenom) { // maybe abstract this logic into GetCDP
		return sdk.ErrInternal("collateral type not enabled to create CDPs")
	}

	// Check the owner has enough collateral and stable coins
	if changeInCollateral.IsPositive() { // adding collateral to CDP
		ok := k.bank.HasCoins(ctx, owner, sdk.Coins{sdk.NewCoin(collateralDenom, changeInCollateral)})
		if !ok {
			return sdk.ErrInsufficientCoins("not enough collateral in sender's account")
		}
	}
	if changeInDebt.IsNegative() { // reducing debt, by adding stable coin to CDP
		ok := k.bank.HasCoins(ctx, owner, sdk.Coins{sdk.NewCoin(StableDenom, changeInDebt.Neg())})
		if !ok {
			return sdk.ErrInsufficientCoins("not enough stable coin in sender's account")
		}
	}

	// Add/Subtract collateral and debt recorded in CDP
	// Get CDP (or create if not exists)
	cdp, found := k.GetCDP(ctx, owner, collateralDenom)
	if !found {
		cdp = CDP{Owner: owner, CollateralDenom: collateralDenom, CollateralAmount: sdk.ZeroInt(), Debt: sdk.ZeroInt()}
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
	price := k.pricefeed.GetCurrentPrice(ctx, cdp.CollateralDenom).Price  // or use collateralDenom
	collateralValue := sdk.NewDecFromInt(cdp.CollateralAmount).Mul(price) // split up the calculation to avoid using division to avoid div by 0 errors.
	minCollateralValue := p.GetCollateralParams(cdp.CollateralDenom).LiquidationRatio.Mul(sdk.NewDecFromInt(cdp.Debt))
	if collateralValue.LT(minCollateralValue) { // TODO LT or LTE ?
		return sdk.ErrInternal("Change to CDP would put it below liquidation ratio")
	}
	// TODO check for dust

	// Add/Subtract from global debt limit
	gDebt := k.GetGlobalDebt(ctx)
	gDebt = gDebt.Add(changeInDebt)
	if gDebt.IsNegative() {
		return sdk.ErrInternal("global debt can't be negative") // This should never happen if debt per CDP can't be negative
	}
	if gDebt.GT(p.GlobalDebtLimit) {
		return sdk.ErrInternal("change to CDP would put the system over the global debt limit")
	}

	// Add/Subtract from collateral debt limit
	cState, found := k.GetCollateralState(ctx, cdp.CollateralDenom)
	if !found {
		cState = CollateralState{Denom: cdp.CollateralDenom, TotalDebt: sdk.ZeroInt()} // Already checked that this denom is authorized, so ok to create new CollateralState
	}
	cState.TotalDebt = cState.TotalDebt.Add(changeInDebt)
	if cState.TotalDebt.IsNegative() {
		return sdk.ErrInternal("total debt for this collateral type can't be negative") // This should never happen if debt per CDP can't be negative
	}
	if cState.TotalDebt.GT(p.GetCollateralParams(cdp.CollateralDenom).DebtLimit) {
		return sdk.ErrInternal("change to CDP would put the system over the debt limit for this collateral type")
	}

	// Phase 2: Update all the state

	// change owner's coins (increase or decrease)
	_, _, err := k.bank.AddCoins(ctx, owner, sdk.Coins{sdk.Coin{collateralDenom, changeInCollateral.Neg()}})
	if err != nil {
		panic("error in adding coins") // this shouldn't happen
	}
	_, _, err = k.bank.AddCoins(ctx, owner, sdk.Coins{sdk.Coin{StableDenom, changeInDebt}})
	if err != nil {
		panic("error in adding coins") // this shouldn't happen
	}
	// Set CDP
	if cdp.CollateralAmount.IsZero() && cdp.Debt.IsZero() { // TODO maybe abstract this logic into setCDP
		k.deleteCDP(ctx, cdp)
	} else {
		k.setCDP(ctx, cdp) // if subsequent lines fail this needs to be reverted, but the sdk should do that automatically?
	}
	// set total debts
	k.setGlobalDebt(ctx, gDebt)
	k.setCollateralState(ctx, cState)

	return nil
}

// TransferCDP allows people to transfer ownership of their CDPs to others
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

// This is only needed to be able to setup the store from the genesis file. The keeper should not change any of the params itself.
func (k Keeper) setParams(ctx sdk.Context, cdpModuleParams CdpModuleParams) {
	k.paramsSubspace.Set(ctx, cdpModuleParamsKey, &cdpModuleParams)
}

// ---------- Store Wrappers ----------

func (k Keeper) getCDPKey(owner sdk.AccAddress, collateralDenom string) []byte {
	return []byte(owner.String() + collateralDenom)
}
func (k Keeper) GetCDP(ctx sdk.Context, owner sdk.AccAddress, collateralDenom string) (CDP, bool) {
	// get store
	store := ctx.KVStore(k.storeKey)
	// get CDP
	bz := store.Get(k.getCDPKey(owner, collateralDenom))
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
func (k Keeper) deleteCDP(ctx sdk.Context, cdp CDP) { // TODO should this id the cdp by passing in owner,collateralDenom pair?
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
		panic("global debt not found")
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

func (k Keeper) getCollateralStateKey(collateralDenom string) []byte {
	return []byte(collateralDenom)
}
func (k Keeper) GetCollateralState(ctx sdk.Context, collateralDenom string) (CollateralState, bool) {
	// get store
	store := ctx.KVStore(k.storeKey)
	// get bytes
	bz := store.Get(k.getCollateralStateKey(collateralDenom))
	// unmarshal
	if bz == nil {
		return CollateralState{}, false
	}
	var cState CollateralState
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &cState)
	return cState, true
}
func (k Keeper) setCollateralState(ctx sdk.Context, collateralstate CollateralState) {
	// get store
	store := ctx.KVStore(k.storeKey)
	// marshal and set
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(collateralstate)
	store.Set(k.getCollateralStateKey(collateralstate.Denom), bz)
}

// ---------- Weird Bank Stuff ----------
// This only exists because module accounts aren't really a thing yet.
// Also because we need module accounts that allow for burning/minting.

// These functions make the CDP module act as a bank keeper, ie it fulfills the bank.Keeper interface.
// It intercepts calls to send coins to/from the liquidator module account, otherwise passing the calls onto the normal bank keeper.

// Not sure if module accounts are good, but they make the auction module more general:
// - startAuction would just "mints" coins, relying on calling function to decrement them somewhere
// - closeAuction would have to call something specific for the receiver module to accept coins (like liquidationKeeper.AddStableCoins)

// The auction and liquidator modules can probably just use SendCoins to keep things safe (instead of AddCoins and SubtractCoins).
// So they should define their own interfaces which this module should fulfill, rather than this fulfilling the entire bank.Keeper interface.

// bank.Keeper interfaces:
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
// type Keeper interface {
// 	SendKeeper
// 	SetCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) sdk.Error
// 	SubtractCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
// 	AddCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Coins, sdk.Tags, sdk.Error)
// 	InputOutputCoins(ctx sdk.Context, inputs []Input, outputs []Output) (sdk.Tags, sdk.Error)
// 	DelegateCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)
// 	UndelegateCoins(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coins) (sdk.Tags, sdk.Error)

var LiquidatorAccountAddress = sdk.AccAddress([]byte("whatever"))
var liquidatorAccountKey = []byte("liquidatorAccount")

type LiquidatorModuleAccount struct {
	Coins sdk.Coins // keeps track of seized collateral, surplus usdx, and mints/burns gov coins
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

// TODO Should this return anything for the gov coin balance? Currently returns nothing.
func (k Keeper) GetCoins(ctx sdk.Context, address sdk.AccAddress) sdk.Coins {
	if address.Equals(LiquidatorAccountAddress) {
		return k.getLiquidatorModuleAccount(ctx).Coins
	} else {
		return k.bank.GetCoins(ctx, address)
	}
}

// TODO test this with unsorted coins
func (k Keeper) HasCoins(ctx sdk.Context, address sdk.AccAddress, amount sdk.Coins) bool {
	if address.Equals(LiquidatorAccountAddress) {
		return true
	} else {
		return k.getLiquidatorModuleAccount(ctx).Coins.IsAllGTE(stripGovCoin(amount))
	}
}

func (k Keeper) getLiquidatorModuleAccount(ctx sdk.Context) LiquidatorModuleAccount {
	// get store
	store := ctx.KVStore(k.storeKey)
	// get bytes
	bz := store.Get(liquidatorAccountKey)
	if bz == nil {
		return LiquidatorModuleAccount{} // TODO is it safe to do this, or better to initialize the account explicitly
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
