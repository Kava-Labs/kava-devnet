package auction

import (
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

func (k Keeper) createAuction(ctx sdk.Context, seller sdk.AccAddress, amount sdk.Coins, endtime sdk.Int) sdk.Error {
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
		LatestBidder: seller,                // send the proceeds back to the seller if no one bids
		LatestBid:    sdk.Coins{sdk.Coin{}}, // TODO check this doesn't cause problems if auction closed without any bids
	}
	// store auction (also adds it to the queue)
	k.storeAuction(ctx, auction)

	return nil
}

func (k Keeper) placeBid(ctx sdk.Context, auctionID auctionID, bidder sdk.AccAddress, bid sdk.Coins) sdk.Error {
	// TODO validation

	// get auction from store
	auction, found := k.getAuction(ctx, auctionID)
	if !found {
		return sdk.ErrInternal("auction doesn't exist") // TODO custom error types ?
	}

	// check if new bid ok and update the auction
	if bid.IsAllGT(auction.LatestBid) { // TODO implement min bid size ?
		// check is bidder has enough total funds
		// catch the edge case where a bidder is incrementing their own bid
		biddersTotalFunds := k.bankKeeper.GetCoins(ctx, bidder)
		if auction.LatestBidder.Equals(bidder) {
			biddersTotalFunds = biddersTotalFunds.Plus(auction.LatestBid)
		}
		if biddersTotalFunds.IsAllGTE(bid) {
			// update bid
			// return previous bidder's funds (can be back to current bidder if someone is updating their bid)
			_, _, err := k.bankKeeper.AddCoins(ctx, auction.LatestBidder, auction.LatestBid)
			if err != nil {
				return err // failing here is ok
			}
			// subtract bidder's coins
			_, _, err = k.bankKeeper.SubtractCoins(ctx, bidder, bid)
			if err != nil {
				return err // TODO This shouldn't fail but is bad if it does. What is the best way to handle this?
			}
			auction.LatestBidder = bidder
			auction.LatestBid = bid
		}
	} else {
		return sdk.ErrInternal("bid size not greater than existing bid") // TODO custom error types?
	}

	// store updated auction
	k.storeAuction(ctx, auction) // TODO this might cause problems because storeAuction adds it to the queue as well

	return nil
}

func (k Keeper) closeAuction(ctx sdk.Context, auctionID auctionID) sdk.Error {
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
	// delete auction from store
	k.deleteAuction(ctx, auctionID)

	return nil
}

func (k Keeper) getNextAuctionIDKey() []byte {
	return []byte("nextAuctionID")
}
func (k Keeper) getAuctionKey(auctionID auctionID) []byte {
	return []byte(fmt.Sprintf("auctions:%d", auctionID))
}

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

// storeAuction puts the auction into the database and adds it to the queue
func (k Keeper) storeAuction(ctx sdk.Context, auction Auction) {
	// store auction
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(auction)
	store.Set(k.getAuctionKey(auction.ID), bz)

	// TODO add to the "queue"
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
	store := ctx.KVStore(k.storeKey)
	store.Delete(k.getAuctionKey(auctionID))

	// TODO remove from queue as well?
}

/////////// QUEUE STUFF
// get expired auctions iterator
// insert
// remove

/*
// Returns an iterator for all the proposals in the Active Queue that expire by endTime
func (keeper Keeper) AuctionQueueIterator(ctx sdk.Context, endTime time.Time) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return store.Iterator(PrefixActiveProposalQueue, sdk.PrefixEndBytes(PrefixActiveProposalQueueTime(endTime)))
}

// Inserts a AuctionID into the auction queue at endTime
func (keeper Keeper) InsertAuctionQueue(ctx sdk.Context, endTime time.Time, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(proposalID)
	store.Set(KeyActiveProposalQueueProposal(endTime, proposalID), bz)
}

// removes an auctionID from the Auction Queue
func (keeper Keeper) RemoveFromAuctionQueue(ctx sdk.Context, endTime time.Time, proposalID uint64) {
	store := ctx.KVStore(keeper.storeKey)
	store.Delete(KeyActiveProposalQueueProposal(endTime, proposalID))
}
*/
