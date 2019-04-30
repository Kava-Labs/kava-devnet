/*
Package auction is a module for creating generic auctions and allowing users to place bids until a timeout is reached.

TODO
 - expand placeBid tests - maybe refactor
 - add endblocker test
 - add keeper tests
 	- start_Auction
	- placeBid
	- closeAuction
	- auctionID
	- set/get/delete auction (also test queue?)
 - add minimum bid increment
 - decided wether to put auction params like default timeouts into the auctions themselves
 - user facing things like cli, rest, querier, tags
 - custom error types, codespace
*/
package auction
