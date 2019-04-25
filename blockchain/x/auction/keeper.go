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

// StartAuction creates and starts an auction that ends at `endTime`.
func (k Keeper) StartAuction(ctx sdk.Context, seller sdk.AccAddress, amount sdk.Coins, endtime endTime) sdk.Error {
	// TODO validation

	// subtract coins from seller
	_, _, err := k.bankKeeper.SubtractCoins(ctx, seller, amount)
	if err != nil {
		return err
	}
	// create auction struct
	newAuctionID, _ := k.getNewAuctionID(ctx) // TODO if this fails then need to unsubtract coins above
	auction := Auction{
		ID:           newAuctionID,
		Seller:       seller,
		Amount:       amount,
		EndTime:      endtime,
		LatestBidder: seller,                // send the proceeds back to the seller if no one bids, and send the first bid
		LatestBid:    sdk.Coins{sdk.Coin{}}, // TODO check this doesn't cause problems if auction closed without any bids
	}
	// store auction (also adds it to the queue)
	k.setAuction(ctx, auction)

	return nil
}

// PlaceBid places a bid on an auction.
func (k Keeper) PlaceBid(ctx sdk.Context, auctionID auctionID, bidder sdk.AccAddress, bid sdk.Coins) sdk.Error {
	// TODO validation

	// get auction from store
	auction, found := k.getAuction(ctx, auctionID)
	if !found {
		return sdk.ErrInternal("auction doesn't exist") // TODO custom error types ?
	}

	// check if new bid ok and update the auction
	if bid.IsAllGT(auction.LatestBid) { // TODO implement min bid size ?
		// check is bidder has enough total funds
		// catch the edge case where a bidder is incrementing their own bid, but doesn't have much funds in their account
		availableFunds := k.bankKeeper.GetCoins(ctx, bidder)
		if auction.LatestBidder.Equals(bidder) {
			availableFunds = availableFunds.Plus(auction.LatestBid)
		}
		if availableFunds.IsAllGTE(bid) {
			// update bid
			// return previous bidder's funds (can be back to current bidder if someone is updating their bid)
			_, _, err := k.bankKeeper.AddCoins(ctx, auction.LatestBidder, auction.LatestBid)
			if err != nil {
				return err // failing here is ok
			}
			// add new bidder's coins to the auction
			_, _, err = k.bankKeeper.SubtractCoins(ctx, bidder, bid)
			if err != nil {
				return err // TODO This shouldn't fail but it's bad if it does. What is the best way to handle this?
			}
			// Add difference to the seller
			_, _, err = k.bankKeeper.AddCoins(ctx, auction.Seller, bid.Minus(auction.LatestBid))
			if err != nil {
				return err // TODO this should also not fail
			}
			auction.LatestBidder = bidder
			auction.LatestBid = bid
		}
	} else {
		return sdk.ErrInternal("bid size not greater than existing bid") // TODO custom error types?
	}

	// store updated auction
	k.setAuction(ctx, auction)

	return nil
}

// CloseAuction closes an auction and distributes funds to the seller and highest bidder.
func (k Keeper) CloseAuction(ctx sdk.Context, auctionID auctionID) sdk.Error {
	// TODO check if auction has timed out?

	// get the auction from the store
	auction, found := k.getAuction(ctx, auctionID)
	if !found {
		return sdk.ErrInternal("auction doesn't exist") // TODO custom error types ?
	}
	// send bidder's funds to seller
	_, _, err := k.bankKeeper.AddCoins(ctx, auction.Seller, auction.LatestBid)
	if err != nil {
		return err
	}
	// send seller's funds to bidder
	_, _, err = k.bankKeeper.AddCoins(ctx, auction.LatestBidder, auction.Amount)
	if err != nil {
		return err
	}
	// delete auction from store (and queue)
	k.deleteAuction(ctx, auctionID)

	return nil
}

// ---------- Store methods ----------
// Use these to add and remove auction from the store.

// getNewAuctionID gets the next available AuctionID and increments it
func (k Keeper) getNewAuctionID(ctx sdk.Context) (auctionID, sdk.Error) {
	// get next ID from store
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getNextAuctionIDKey())
	if bz == nil {
		panic("initial auctionID never set in genesis")
		//return 0, ErrInvalidGenesis(keeper.codespace, "InitialProposalID never set") // TODO is this needed? Why not just set it zero here?
	}
	var newAuctionID auctionID
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &newAuctionID)

	// increment the stored next ID
	bz = k.cdc.MustMarshalBinaryLengthPrefixed(newAuctionID + 1)
	store.Set(k.getNextAuctionIDKey(), bz)

	return newAuctionID, nil
}

// setAuction puts the auction into the database and adds it to the queue
// it overwrites any pre-existing auction with same ID
func (k Keeper) setAuction(ctx sdk.Context, auction Auction) {
	// remove the auction from the queue if it is already in there
	existingAuction, found := k.getAuction(ctx, auction.ID)
	if found {
		k.removeFromQueue(ctx, existingAuction.EndTime, existingAuction.ID)
	}

	// store auction
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(auction)
	store.Set(k.getAuctionKey(auction.ID), bz)

	// add to the queue
	k.insertIntoQueue(ctx, auction.EndTime, auction.ID)
}

// getAuction gets an auction from the store by auctionID
func (k Keeper) getAuction(ctx sdk.Context, auctionID auctionID) (Auction, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(k.getAuctionKey(auctionID))
	if bz == nil {
		return Auction{}, false // TODO what is the correct behavior when an auction is not found? gov module follows this pattern of returning a bool
	}
	var auction Auction
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &auction)
	return auction, true
}

// deleteAuction removes an auction from the store without any validation
func (k Keeper) deleteAuction(ctx sdk.Context, auctionID auctionID) {
	// remove from queue
	auction, found := k.getAuction(ctx, auctionID)
	if found {
		k.removeFromQueue(ctx, auction.EndTime, auctionID)
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
		sdk.PrefixEndBytes(getQueueElementKeyPrefix(endTime)), // end key (exclusive)
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
