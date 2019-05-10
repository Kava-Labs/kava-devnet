/*
Package Liquidator settles bad debt from undercollateralized CDPs by seizing them and raising funds through auctions.

Notes
 - Missing the debt queue thing from Vow
 - seized collateral and usdx are stored in the module account, but debt (aka Sin) is stored in keeper
 - The boundary between the liquidator and the cdp modules is messy.
	- The CDP type is used in liquidator
	- cdp knows about seizing
	- seizing of a CDP is split across each module
	- recording of debt is split across modules
	- liquidator needs get access to stable and gov denoms from the cdp module

TODO
 - Call settleDebt somewhere, currently it just racks up.
 - Fix bug in settleDebt where CDP module globalDebt is not reduced
 - Add an endblocker that seizes all undercollateralized CDPs (requires queue structure in CDP module)
 - Add params (need access to the collateral types for fees. Can this module access the cdp module params?)
 - rename SeizedDebt to TotalSeizedDebt ?
 - Add constants for the module and route names
 - user facing things like cli, rest, querier, tags
 - custom error types, codespace
*/
package liquidator
