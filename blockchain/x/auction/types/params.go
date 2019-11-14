package types

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Defaults for auction params
const (
	// DefaultMaxAuctionDuration max length of auction, roughly 2 days in blocks
	DefaultMaxAuctionDuration EndTime = 2 * 24 * 3600 / 5
	// DefaultBidDuration how long an auction gets extended when someone bids, roughly 3 hours in blocks
	DefaultMaxBidDuration EndTime = 3 * 3600 / 5
	// DefaultStartingAuctionID what the id of the first auction will be
	DefaultStartingAuctionID ID = ID(0)
)

// Parameter keys
var (
	// ParamStoreKeyAuctionParams Param store key for auction params
	KeyAuctionBidDuration = []byte("MaxBidDuration")
	KeyAuctionDuration = []byte("MaxAuctionDuration")
	KeyAuctionStartingID = []byte("StartingAuctionID")
)

var _ subspace.ParamSet = &AuctionParams{}

// AuctionParams governance parameters for auction module
type AuctionParams struct {
	MaxAuctionDuration EndTime `json:"max_auction_duration" yaml:"max_auction_duration"` // max length of auction, in blocks
	MaxBidDuration        EndTime `json:"max_bid_duration" yaml:"max_bid_duration"`
	StartingAuctionID  ID      `json:"starting_auction_id" yaml:"starting_auction_id"`
}

// NewAuctionParams creates a new AuctionParams object
func NewAuctionParams(maxAuctionDuration EndTime, bidDuration EndTime, startingID ID) AuctionParams {
	return AuctionParams{
		MaxAuctionDuration: maxAuctionDuration,
		MaxBidDuration:        bidDuration,
		StartingAuctionID:  startingID,
	}
}

// DefaultAuctionParams default parameters for auctions
func DefaultAuctionParams() AuctionParams {
	return NewAuctionParams(
		DefaultMaxAuctionDuration,
		DefaultMaxBidDuration,
		DefaultStartingAuctionID,
	)
}

// ParamKeyTable Key declaration for parameters
func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&AuctionParams{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of auth module's parameters.
// nolint
func (ap *AuctionParams) ParamSetPairs() subspace.ParamSetPairs {
	return subspace.ParamSetPairs{
		{KeyAuctionBidDuration, &ap.MaxBidDuration},
		{KeyAuctionDuration, &ap.MaxAuctionDuration},
		{KeyAuctionStartingID, &ap.StartingAuctionID},
	}
}

// Equal returns a boolean determining if two AuctionParams types are identical.
func (ap AuctionParams) Equal(ap2 AuctionParams) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&ap)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&ap2)
	return bytes.Equal(bz1, bz2)
}

// String implements stringer interface
func (ap AuctionParams) String() string {
	return fmt.Sprintf(`Auction Params:
	Max Auction Duration: %s
	Max Bid Duration: %s
	Starting Auction ID: %v`, ap.MaxAuctionDuration, ap.MaxBidDuration, ap.StartingAuctionID)
}

// Validate checks that the parameters have valid values.
func (ap AuctionParams) Validate() error {
	if ap.MaxAuctionDuration <= EndTime(0) {
		return fmt.Errorf("max auction duration should be positive, is %s", ap.MaxAuctionDuration)
	}
	if ap.MaxBidDuration <= EndTime(0) {
		return fmt.Errorf("bid duration should be positive, is %s", ap.MaxBidDuration)
	}
	if ap.StartingAuctionID <= ID(0) {
		return fmt.Errorf("starting auction ID should be positive, is %v", ap.StartingAuctionID)
	}
	return nil
}
