package liquidator

// --------- Keeper ----------
type Keeper struct {
	// CDP keeper
	// auction keeper
}

const moduleAccountAddress = cdp.LiquidationAccountAddress

func (k Keeper) StartCollateralAuction(seizedCdpId uint64) { // aka Cat.flip
	// get Seized CDP (fom ID)
	// calculate how much dai to try and raise in this auction	// daiToRaise := min(seizedCdp.Debt, params.maxDaiToBeRaised)
	// calculate how much collateral to sell (based on above)
	// subtract these values from the seizedCDP
	// start "forward reverse" auction type // auctionKeeper.StartForwardReverseAuction(moduleAccountAddress, lot, daiToRaise, seizedCDP.OriginalOwner)
	// store seizedCDP
	// result: dai in transferred to moduleAccount, collateral is transferred from module account to buyer, (and any excess collateral is transferred to original CDP owner)
}

func (k Keeper) StartDebtAuction() { // aka Vow.flop
	// check there is enough seized debt to be "sold"
	// start reverse auction, selling minted XRS for USDX // auctionKeeper.StartReverseAuction(moduleAccountAddress, params.daiToBeRaised, sdk.Coin{"MKR", 2 ^ 256 - 1}) // infinite MKR :(
	// result: minted MKR moved to highest bidder, dai moved to moduleAccount
}
func (k Keeper) StartSurplusAuction() { // aka Vow.flap
	// check there is enough surplus to be sold
	// start normal auction for usdx // aK.StartForwardAuction(moduleAccountAddress, daiToBeSold, sdk.Coin{"XRS", 0})
	// end result: dai removed from module account (eventually to buyer), XRS transferred to module account
}

func (k Keeper) ConfiscateUnderCollateralizedCDPs() { // aka Cat.bite (except over all cdps)
	// get all the undercollateralized cdps. This returns an iterator // cdps := k.cdpK.GetUnderCollateralizedCDPs()
	// loop through them:
	//    get the CDP // cdp := k.cdpKeeper.GetCDP(id)
	//    create a "seizedCDP". This is needed because the collateral is auctioned off per CDP. // k.setSeizedCDP(SeizedCDP{cdp.Debt, cdp.Collateral}) // aka create Cat.Flip object
	//    add cdp.collateral amount of coins to the moduleAccount (so they can be transferred to the auction later)
	//    increment the total seized debt by cdp.debt. // aka Awe
	//    cdpKeeper.ConfiscateCDP(id)         // just empties it and updates the global debt in cdp the module
}

func (k Keeper) settleDebt() { // When should this be called? At the start of every entry point here?
	// calculate max amount of debt/usdx that can be settled ("annihilated") // settlement := min(GetCoins(accountModule), GetTotalSeizedDebt())
	// decrement total seized debt by above amount // SubtractTotalSeizedDebt(settlement)
	// subtract usdx from moduleAccout // SubtractCoins(moduleAccountAddress, settlement)
}

// ---------- Params ----------
// Need access to the collateral types for fees I think. Can probably use the same param thing as the cdp module uses.

// Default auction sizes:
// maxDaiToBeRaised     // aka lump (flip)
// daiToBeSold          // aka bump (flap)
// daiToBeRaised        // aka sump (flop)

// --------- Types ----------
type SeizedCDP struct {
	ID            uint64 // is this needed or can we make a store key out of owner and collateral type?
	OriginalOwner sdk.AccAddress
	Collateral    sdk.Coin
	Debt          sdk.Coin // what should the type be here?
}

type totalSeizedDebt sdk.Coin // seized collateral and usdx are stored in the module account, but debt is stored here // aka Sin

// ---------- Handler, Msgs, EndBlocker ----------

func EndBlocker() {
	// just runs k.ConfiscateUnderCollateralizedCDPs()
}

// Message types for starting various auctions. Should they place an initial bid as well?
type MsgStartCollateralAuction struct {
	// TODO
}
type MsgStartDebtAuction struct {
	// TODO
}
type MsgStartSurplusAuction struct {
	// TODO
}

// ---------- Notes ----------
// - could maybe just use the original CDP object rather than a separate SeizedCDP object. Need to add an "IsSeized" flag
// - missing the debt queue thing from Vow
