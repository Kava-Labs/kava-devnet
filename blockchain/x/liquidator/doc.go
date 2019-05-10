/*
Package Liquidator settles bad debt from undercollateralized CDPs by seizing bad CDPs and raising funds through auctions.

Notes
 - Could maybe just use the original CDP object rather than a separate SeizedCDP object. Need to add an "IsSeized" flag
 - Missing the debt queue thing from Vow
 - seized collateral and usdx are stored in the module account, but debt (aka Sin) is stored in keeper

TODO
 - Should there be an endblocker?
func EndBlocker() {
	// just runs k.ConfiscateUnderCollateralizedCDPs()
}

 - Should there be params?
// Need access to the collateral types for fees I think. Can probably use the same param thing as the cdp module uses.
// Default auction sizes:
// maxDaiToBeRaised     // aka lump (flip)
// daiToBeSold          // aka bump (flap)
// daiToBeRaised        // aka sump (flop)

 - rename SeizedDebt to TotalSeizedDebt ?
 - The boundary between the liquidator and the cdp modules is messy. The CDP type is in the liquidator, cdp knows about seizing
 - Call settleDebt somewhere, currently it just racks up.
*/
package liquidator
