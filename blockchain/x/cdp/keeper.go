package cdp

// ---------- Keeper ----------
type Keeper struct {
	// pricefeed keeper
	// bank keeper
}

// ModifyCDP creates, changes, or deletes a CDP
func (k Keeper) ModifyCDP(owner, collateralType, collateral, debt) sdk.Error {
	// try getting cdp
	cdp, found := GetCDP(owner, collateralType)
	// if none create a blank one (in memory)
	if !found {
		var cdp CDP
	}
	// add/subtract coins from owner
	// add/subtract collateral and debt recorded in CDP
	
	// check CDP is OK (call pricefeed.GetCurrentPrice) - collateral ratio, collateral and total debt ceilings, dust // cdp.Collateral.Mul(k.priceFeedKeeper.GetPrice()).Div(cdp.Debt).GTE(ilk.CollateralRatio)

	// update globalDebt and debt for the particular collateral type

	// store / delete updated CDP
}

// convience function to allow people to give their CDPs to others
func (k Keeper) TransferCDP(owner, to) {
	// TODO
}

// Not sure if this is really needed. But it allows the cdp module to track total debt and debt per asset.
func (k Keeper) ConfiscateCDP(owner, collateralType) {
	// empty CDP of collateral and debt, ie set values to zero

	// update debt per collateral type
	// update global seized debt ? this is what maker does (Vat.grab) but it's not used anywhere
}

func (k Keeper) GetUnderCollateralizedCDPs() {
	// get current prices of assets // priceFeedKeeper.GetCurrentPrice(denom)

	// get an iterator over the CDPs that only includes undercollateralized CDPs
	//    should be possible to store cdps by a key that is their collateral/debt ratio, then the iterator thing can be used to get only the undercollaterized ones (for a given price)

	// combine all the iterators for the different assets?

	// return iterator
}

// Store wrappers:

func (k Keeper) GetCDP(owner, collateralType) (CDP, bool) {
	// get store
	// get CDP
	// unmarshal
}
func (k Keeper) setCDP(owner, collateralType, CDP) {
	// get store
	// marshal and set
	// remove and add to iterator
}
func (k Keeper) deleteCDP(owner CollateralType, CDP) {
	// get store
	// delete key
	// remove from iterator
}
// CDP could have a unique id, or a key could be generated from the owner's address and collateral type
func (k Keeper) getCDPKey(owner, collateralType) []byte {
	return []byte{owner.String() + collateralType.Denom}
}
// GetCollateralType
// setCollateralType
// getCollateralTypeKey

// ---------- Params ----------
// - globalDebtCeiling
// - authorizedCollateralTypes []denom ?

// ---------- Types ----------
type CDP struct {
	ID // is an id needed or can we create a key out of owner and collateral type?
	owner
	collateral
	debt
}

// Collateral types could be in params or stored normally in store
type CollateralType struct {
	denom
	debtCeiling
	totalDebt
}

type globalDebt sdk.Coin

// ---------- Handler, Msgs, EndBlocker ----------

// MsgCreateOrModifyCDP creates, adds/removes collateral/usdx from a cdp
type MsgCreateOrModifyCDP{
	// TODO
}
// MsgTransferCDP changes the ownership of a cdp
type MsgTransferCDP{
	// TODO
}

// No endblocker, cdp monitoring happens in liquidator module


// ---------- Weird Bank Stuff ----------
// This only exists because module accounts aren't a thing yet.
// Also because we need module accounts that allow for burning/minting.

// These functions make the CDP module act as a bank keeper, intercepting calls to move coins from the liquidator module account.

// Not sure if module accounts are good, but they make the auction module more general:
// - startAuction would just "mints" coins, relying on calling function to decrement them somewhere
// - closeAuction would have to call something specific for the receiver module to accept coins (like liquidationKeeper.AddStableCoins)

// With account modules all CDP functions can probably use just SendCoins to keep things safe (instead of AddCoins and SubtractCoins)

const LiquidationAccountAddress = []byte{"whatever"}

type LiquidatorModuleAccount struct {
	coins sdk.Coins // keeps track of seized collateral and surplus usdx
}

func (k Keeper) AddCoins(address, amount) {
	if address = LiquidationAccountAddress {
		switch amount.Denom
		case "xrs":
			return // do nothing - effectively burns sent XRS
		default:
			modifyLiquidatorAccount(amount) // adds collateral or usdx
	} else {
		return bank.AddCoins(address, amount)
	}
}
func (k Keeper) SubtractCoins{
	if address = LiquidationAccountAddress {
		switch amount.Denom
		case "xrs":
			return // do nothing - effectively mints XRS
		default:
			modifyLiquidatorAccount(amount) // adds collateral or usdx
	} else {
		return bank.SubtractCoins(address, amount)
	}
}
func (k Keeper) GetCoins(address) {
	// get store
	// get liquidator account
	// get coins, return // what should it return for XRS balance?
}

func (k Keeper) modifyLiquidatorAccount(amount) {
	// get store
	// get liquidator account
	// get coins
	// add/subtract coins
	// set account
}
