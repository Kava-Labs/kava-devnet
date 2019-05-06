package cdp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CDP is the state of a single Collateralized Debt Position.
type CDP struct {
	//ID               []byte       // removing IDs for now to make things simpler
	Owner            sdk.AccAddress // Account that authorizes changes to the CDP
	CollateralDenom  string         // Type of collateral stored in this CDP
	CollateralAmount sdk.Int        // Amount of collateral stored in this CDP
	Debt             sdk.Int        // Amount of stable coin drawn from this CDP
}

// CollateralState stores global information tied to a particular collateral type.
type CollateralState struct {
	Denom     string  // Type of collateral
	TotalDebt sdk.Int // total debt collateralized by a this coin type
	//AccumulatedFees sdk.Int // Ignoring fees for now
}
