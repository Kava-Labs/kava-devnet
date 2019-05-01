package auction

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

type Keeper struct {
	bankKeeper bank.Keeper
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	// TODO codespace
}

// NewKeeper returns a new auction keeper.
func NewKeeper(cdc *codec.Codec, bankKeeper bank.Keeper, storeKey sdk.StoreKey) Keeper {
	return Keeper{
		bankKeeper: bankKeeper,
		storeKey:   storeKey,
		cdc:        cdc,
	}
}

// TODO these 3 start functions be combined or abstracted away?

// StartForwardAuction starts a normal auction. Known as flap in maker.
func (k Keeper) StartForwardAuction(ctx sdk.Context, seller sdk.AccAddress, lot sdk.Coin, initialBid sdk.Coin) sdk.Error {
	// create auction
	auction, initiatorOutput := NewForwardAuction(seller, lot, initialBid, endTime(ctx.BlockHeight())+maxAuctionDuration)
	// start the auction
	err := k.startAuction(ctx, &auction, initiatorOutput)
	if err != nil {
		return err
	}
	return nil
}

// StartReverseAuction starts an auction where sellers compete by offering decreasing prices. Known as flop in maker.
func (k Keeper) StartReverseAuction(ctx sdk.Context, buyer sdk.AccAddress, bid sdk.Coin, initialLot sdk.Coin) sdk.Error {
	// create auction
	auction, initiatorOutput := NewReverseAuction(buyer, bid, initialLot, endTime(ctx.BlockHeight())+maxAuctionDuration)
	// start the auction
	err := k.startAuction(ctx, &auction, initiatorOutput)
	if err != nil {
		return err
	}
	return nil
}

// StartForwardReverseAuction starts an auction where bidders bid up to a maxBid, then switch to bidding down on price. Known as flip in maker.
func (k Keeper) StartForwardReverseAuction(ctx sdk.Context, seller sdk.AccAddress, lot sdk.Coin, maxBid sdk.Coin, otherPerson sdk.AccAddress) sdk.Error {
	// create auction
	initialBid := sdk.NewInt64Coin(maxBid.Denom, 0) // set the bidding coin denomination from the specified max bid
	auction, initiatorOutput := NewForwardReverseAuction(seller, lot, initialBid, endTime(ctx.BlockHeight())+maxAuctionDuration, maxBid, otherPerson)
	// start the auction
	err := k.startAuction(ctx, &auction, initiatorOutput)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) startAuction(ctx sdk.Context, auction Auction, initiatorOutput bankOutput) sdk.Error {
	// get ID
	newAuctionID, err := k.getNextAuctionID(ctx)
	if err != nil {
		return err
	}
	// set ID
	auction.SetID(newAuctionID)

	// subtract coins from initiator
	_, _, err = k.bankKeeper.SubtractCoins(ctx, initiatorOutput.Address, sdk.Coins{initiatorOutput.Coin})
	if err != nil {
		return err
	}

	// store auction
	k.setAuction(ctx, auction)
	k.incrementNextAuctionID(ctx)
	return nil
}

// PlaceBid places a bid on any auction.
func (k Keeper) PlaceBid(ctx sdk.Context, auctionID auctionID, bidder sdk.AccAddress, bid sdk.Coin, lot sdk.Coin) sdk.Error {

	// get auction from store
	auction, found := k.getAuction(ctx, auctionID)
	if !found {
		return sdk.ErrInternal("auction doesn't exist")
	}

	// place bid
	coinOutputs, coinInputs, err := auction.PlaceBid(endTime(ctx.BlockHeight()), bidder, lot, bid) // update auction according to what type of auction it is // TODO should this return updated Auction to be more immutable?
	if err != nil {
		return err
	}
	// TODO this will fail if someone tries to update their bid without the full bid amount sitting in their account
	// sub outputs
	for _, output := range coinOutputs {
		_, _, err = k.bankKeeper.SubtractCoins(ctx, output.Address, sdk.Coins{output.Coin}) // TODO handle errors properly here. All coin transfers should be atomic. InputOutputCoins may work
		if err != nil {
			panic(err)
		}
	}
	// add inputs
	for _, input := range coinInputs {
		_, _, err = k.bankKeeper.AddCoins(ctx, input.Address, sdk.Coins{input.Coin}) // TODO errors
		if err != nil {
			panic(err)
		}
	}

	// store updated auction
	k.setAuction(ctx, auction)

	return nil
}

// CloseAuction closes an auction and distributes funds to the seller and highest bidder.
func (k Keeper) CloseAuction(ctx sdk.Context, auctionID auctionID) sdk.Error {

	// get the auction from the store
	auction, found := k.getAuction(ctx, auctionID)
	if !found {
		return sdk.ErrInternal("auction doesn't exist")
	}
	// check if auction has timed out
	if ctx.BlockHeight() > int64(auction.GetEndTime()) { // TODO > or ≥ ?
		return sdk.ErrInternal("auction has already ended")
	}
	// payout to the last bidder
	coinInput := auction.GetPayout()
	_, _, err := k.bankKeeper.AddCoins(ctx, coinInput.Address, sdk.Coins{coinInput.Coin})
	if err != nil {
		return err
	}

	// delete auction from store (and queue)
	k.deleteAuction(ctx, auctionID)

	return nil
}

// ---------- Store methods ----------
// Use these to add and remove auction from the store.

// getNextAuctionID gets the next available global AuctionID
func (k Keeper) getNextAuctionID(ctx sdk.Context) (auctionID, sdk.Error) { // TODO don't need error return here
	// get next ID from store
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getNextAuctionIDKey())
	if bz == nil {
		// if not found, set the id at 0
		bz = k.cdc.MustMarshalBinaryLengthPrefixed(auctionID(0))
		store.Set(k.getNextAuctionIDKey(), bz)
		// TODO Why does the gov module set the id in genesis? :
		//return 0, ErrInvalidGenesis(keeper.codespace, "InitialProposalID never set")
	}
	var auctionID auctionID
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &auctionID)
	return auctionID, nil
}

// incrementNextAuctionID increments the global ID in the store by 1
func (k Keeper) incrementNextAuctionID(ctx sdk.Context) sdk.Error {
	// get next ID from store
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getNextAuctionIDKey())
	if bz == nil {
		panic("initial auctionID never set in genesis")
		//return 0, ErrInvalidGenesis(keeper.codespace, "InitialProposalID never set") // TODO is this needed? Why not just set it zero here?
	}
	var auctionID auctionID
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &auctionID)

	// increment the stored next ID
	bz = k.cdc.MustMarshalBinaryLengthPrefixed(auctionID + 1)
	store.Set(k.getNextAuctionIDKey(), bz)

	return nil
}

// setAuction puts the auction into the database and adds it to the queue
// it overwrites any pre-existing auction with same ID
func (k Keeper) setAuction(ctx sdk.Context, auction Auction) {
	// remove the auction from the queue if it is already in there
	existingAuction, found := k.getAuction(ctx, auction.GetID())
	if found {
		k.removeFromQueue(ctx, existingAuction.GetEndTime(), existingAuction.GetID())
	}

	// store auction
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(auction)
	store.Set(k.getAuctionKey(auction.GetID()), bz)

	// add to the queue
	k.insertIntoQueue(ctx, auction.GetEndTime(), auction.GetID())
}

// getAuction gets an auction from the store by auctionID
func (k Keeper) getAuction(ctx sdk.Context, auctionID auctionID) (Auction, bool) {
	var auction Auction

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getAuctionKey(auctionID))
	if bz == nil {
		return auction, false // TODO what is the correct behavior when an auction is not found? gov module follows this pattern of returning a bool
	}

	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &auction)
	return auction, true
}

// deleteAuction removes an auction from the store without any validation
func (k Keeper) deleteAuction(ctx sdk.Context, auctionID auctionID) {
	// remove from queue
	auction, found := k.getAuction(ctx, auctionID)
	if found {
		k.removeFromQueue(ctx, auction.GetEndTime(), auctionID)
	}

	// delete auction
	store := ctx.KVStore(k.storeKey)
	store.Delete(k.getAuctionKey(auctionID))
}

// ---------- Queue and key methods ----------
// These are lower level function used by the store methods above.

func (k Keeper) getNextAuctionIDKey() []byte {
	return []byte("nextAuctionID")
}
func (k Keeper) getAuctionKey(auctionID auctionID) []byte {
	return []byte(fmt.Sprintf("auctions:%d", auctionID))
}

// Inserts a AuctionID into the queue at endTime
func (k Keeper) insertIntoQueue(ctx sdk.Context, endTime endTime, auctionID auctionID) {
	// get the store
	store := ctx.KVStore(k.storeKey)
	// marshal thing to be inserted
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(auctionID)
	// store it
	store.Set(
		getQueueElementKey(endTime, auctionID),
		bz,
	)
}

// removes an auctionID from the queue
func (k Keeper) removeFromQueue(ctx sdk.Context, endTime endTime, auctionID auctionID) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(getQueueElementKey(endTime, auctionID))
}

// Returns an iterator for all the auctions in the queue that expire by endTime
func (k Keeper) getQueueIterator(ctx sdk.Context, endTime endTime) sdk.Iterator { // TODO rename to "getAuctionsByExpiry" ?
	// get store
	store := ctx.KVStore(k.storeKey)
	// get an interator
	return store.Iterator(
		queueKeyPrefix, // start key
		sdk.PrefixEndBytes(getQueueElementKeyPrefix(endTime+1)), // end key (exclusive) // +1 to make it inclusive
	)
}

var queueKeyPrefix = []byte("queue")
var keyDelimiter = []byte(":")

// Returns half a key for an auctionID in the queue, it missed the id off the end
func getQueueElementKeyPrefix(endTime endTime) []byte {
	return bytes.Join([][]byte{
		queueKeyPrefix,
		sdk.Uint64ToBigEndian(uint64(endTime)), // TODO check this gives correct ordering
	}, keyDelimiter)
}

// Returns the key for an auctionID in the queue
func getQueueElementKey(endTime endTime, auctionID auctionID) []byte {
	return bytes.Join([][]byte{
		queueKeyPrefix,
		sdk.Uint64ToBigEndian(uint64(endTime)), // TODO check this gives correct ordering
		sdk.Uint64ToBigEndian(uint64(auctionID)),
	}, keyDelimiter)
}
