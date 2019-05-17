package liquidator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/kava-labs/usdx/blockchain/x/cdp"
)

type SeizedCDP = cdp.CDP // TODO is this a reasonable thing to do?

type SeizedDebt struct {
	Total         sdk.Int // Total debt seized from CDPs. Known as Awe in maker.
	SentToAuction sdk.Int // Portion of seized debt that has had a (reverse) auction was started for it. Known as Ash in maker.
}

// Available gets the seized debt that has not been sent for auction. Known as Woe in maker.
func (sd SeizedDebt) Available() sdk.Int {
	return sd.Total.Sub(sd.SentToAuction)
}
