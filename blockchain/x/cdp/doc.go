/*
Package CDP manages the storage of Collateralized Debt Positions. It handles their creation, modification, and stores the global state of all CDPs.

Notes
- sdk.Int is used for all the number types to maintain compatibility with internal type of sdk.Coin - saves type conversion when doing maths.
  Also it allows for changes to a CDP to be expressed as a +ve or -ve number.
- Only allowing one CDP per account-collateralDenom pair for now to keep things simple.

TODO
- what happens if a collateral type is removed from the list of allowed ones?
- Should the values used to generate a key for a stored struct be in the struct?

*/
package cdp
