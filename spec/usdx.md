# USDX

USDX is a collateralized debt position system built on Cosmos.

## Background

Measured by encumbered collateral, debt issuance for the purpose of leveraged exposure has been the most succesful secondary financial usecase of cryptocurrencies. MakerDao, and the associated DeFi ecosystem on Ethereum, represent arguably the best example of product-market fit for any blockchain product or service that is not a base-layer protocol. While the number of users of these products is generally small, the potential for synthetic asset issuance (USD pegged stablecoins, or debt denominated in a stable basket of goods) that spans jurisdictions and runs largely autonomously is large.

Cosmos (Gaia) is a new blockchain protocol that uses Tendermint BFT for consensus and is designed with a hub-and-spoke model of cross-blockchain interoperabilty that emphasizes composability and self-sovereignty of application specific blockchains. We believe one of the primary financial usecases for the Cosmos ecosystem will be the the issuance of decentralized pegged assets like pegged Bitcoin (https://github.com/nomic-io/bitcoin-peg/blob/master/bitcoinPeg.md), as well as other pegged crpyto-native and traditional financial assets.

We are building a blockchain on the cosmos-sdk for the purpose of issuing Collateralized Debt Positions (CDPs) for assets in the cosmos ecosystem. The design of the CDP zone is inspired by Multi-Collateral Dai (https://github.com/makerdao/dss) and will allow users to lock their assets as collateral and draw a dollar-denominated debt off of their collateral. We believe this zone is a useful addition to the Cosmos ecosystem, providing a native way for users to gain leveraged exposure to a basket of assets in the cosmos ecosystem, as well as to create a collateral-backed stablecoin that is native to Cosmos.

## Design

There are 4 modules that make up the system:

* Pricefeed
* Auction
* CDP
* Liquidator

### Pricefeed
The pricefeed module implements a simple price oracle where a group of white-listed oracles post prices for various assets in the system. The median price of all valid oracle prices is taken as the current price in the system. Adding and removing of assets and oracles is controlled by governance proposals.

#### Messages and Types

``` go
// Asset struct that represents an asset in the pricefeed
type Asset struct {
	AssetCode   string `json:"asset_code"`
	Description string `json:"description"`
}

// Oracle struct that documents which address an oracle is using
type Oracle struct {
	OracleAddress string `json:"oracle_address"`
}

// CurrentPrice struct that contains the metadata of a current price for a particular asset in the pricefeed module.
type CurrentPrice struct {
	AssetCode string  `json:"asset_code"`
	Price     sdk.Dec `json:"price"`
	Expiry    sdk.Int `json:"expiry"`
}

// PostedPrice struct represented a price for an asset posted by a specific oracle
type PostedPrice struct {
	AssetCode     string  `json:"asset_code"`
	OracleAddress string  `json:"oracle_address"`
	Price         sdk.Dec `json:"price"`
	Expiry        sdk.Int `json:"expiry"`
}

// MsgPostPrice struct representing a posted price message.
// Used by oracles to input prices to the pricefeed
type MsgPostPrice struct {
	From      sdk.AccAddress // client that sent in this address
	AssetCode string         // asset code used by exchanges/api
	Price     sdk.Dec        // price in decimal (max precision 18)
	Expiry    sdk.Int        // block height
}
```

### Auction

The Auction module implements three distinct auction types that control the supply of bad debt and surplus in the CDP system.

**Forward Auction** A standard auction where a seller takes increasing bids for an item. Each bid increments the price, as well as the duration of the auction. This auction type is used when there is a surplus of collected fees in the system. The surplus is converted to stablecoins and sold for governance tokens.

**Reverse Auction** An auction where a buyer solicits decreasing bids for a particular item or lot of items. This auction type is used when governance tokens are sold (minted) in exchange for stablecoins, to cover shortfalls after failed collateral auctions.

**Forward Reverse Auction** An auction where a buyer solicits increasing bids for a lot of goods, up to some ceiling. After the ceiling is reached, each bid lowers the amount of goods being sold for the ceiling  price. This type of auction is used when collateral is siezed from a risky CDP and sold for stablecoins to cover the debt.

#### Messages and Types

``` go
// Auction is an interface to several types of auction.
type Auction interface {
	GetID() ID
	SetID(ID)
	PlaceBid(currentBlockHeight endTime, bidder sdk.AccAddress, lot sdk.Coin, bid sdk.Coin) ([]bankOutput, []bankInput, sdk.Error)
	GetEndTime() endTime // auctions close at the end of the block with blockheight EndTime (ie bids placed in that block are valid)
	GetPayout() bankInput
	String() string
}

// BaseAuction type shared by all Auctions
type BaseAuction struct {
	ID         ID
	Initiator  sdk.AccAddress // Person who starts the auction. Giving away Lot (aka seller in a forward auction)
	Lot        sdk.Coin       // Amount of coins up being given by initiator (FA - amount for sale by seller, RA - cost of good by buyer (bid))
	Bidder     sdk.AccAddress // Person who bids in the auction. Receiver of Lot. (aka buyer in forward auction, seller in RA)
	Bid        sdk.Coin       // Amount of coins being given by the bidder (FA - bid, RA - amount being sold)
	EndTime    endTime        // Block height at which the auction closes. It closes at the end of this block
	MaxEndTime endTime        // Maximum closing time. Auctions can close before this but never after.
}

type ID uint64
type endTime int64
// ForwardAuction type for forward auctions
type ForwardAuction struct {
	BaseAuction
}
// ReverseAuction type for reverse auctions
type ReverseAuction struct {
	BaseAuction
}
// ForwardReverseAuction type for forward reverse auction
type ForwardReverseAuction struct {
	BaseAuction
	MaxBid      sdk.Coin
	OtherPerson sdk.AccAddress
}

// MsgPlaceBid is the message type used to place a bid on any type of auction.
type MsgPlaceBid struct {
	AuctionID ID
	Bidder    sdk.AccAddress // This can be a buyer (who increments bid), or a seller (who decrements lot) TODO rename to be clearer?
	Bid       sdk.Coin
	Lot       sdk.Coin
}
```

### CDP

The CDP module is a factory for creating CDPs and storing the global state of the debt system. It allows users to create, modify, and close CDPs for any collateral type in the pricefeed module. It also sets the global parameters of the system, which can be altered by governance proposals. These parameters include the global debt limit (the total amount of stablecoins that can be in circulation), the debt limit for each collateral type, and the collateralization ratio for each collateral type.

#### Messages and Types

``` go
// CDP is the state of a single Collateralized Debt Position.
type CDP struct {
	//ID             []byte                                    // removing IDs for now to make things simpler
	Owner            sdk.AccAddress `json:"owner"`             // Account that authorizes changes to the CDP
	CollateralDenom  string         `json:"collateral_denom"`  // Type of collateral stored in this CDP
	CollateralAmount sdk.Int        `json:"collateral_amount"` // Amount of collateral stored in this CDP
	Debt             sdk.Int        `json:"debt"`              // Amount of stable coin drawn from this CDP
}

// CollateralState stores global information tied to a particular collateral type.
type CollateralState struct {
	Denom     string  // Type of collateral
	TotalDebt sdk.Int // total debt collateralized by a this coin type
	//AccumulatedFees sdk.Int // Ignoring fees for now
}

// MsgCreateOrModifyCDP creates, adds/removes collateral/stable coin from a cdp
type MsgCreateOrModifyCDP struct {
	Sender           sdk.AccAddress
	CollateralDenom  string
	CollateralChange sdk.Int
	DebtChange       sdk.Int
}
```
### Liquidator

The liquidator module tracks the status of CDPs based on prices in the pricefeed module and is responsible for siezing collateral from CDPs whose collateralization ratio is below the threshold set for that collateral type and sending it to the auction module.

#### Messages and Types

``` go
type SeizedDebt struct {
	Total         sdk.Int // Total debt seized from CDPs.
	SentToAuction sdk.Int // Portion of seized debt that has had a (reverse) auction was started for it.
	// SentToAuction should always be < Total
}

type MsgSeizeAndStartCollateralAuction struct {
	Sender          sdk.AccAddress // needed to pay the tx fees
	CdpOwner        sdk.AccAddress
	CollateralDenom string
}
```

### Kava - Governance and Staking Token

The system is secured by Kava, a staking and governance token. Staked Kava tokens receive inflationary block rewards and are eligible to create and vote on governance proposals. Fees for transactions and for closing and liquidating CDPs are also collected in Kava tokens. Fees associated with CDPs are auctioned as part of the `auction` module.

### Liquidation and Recolateralization
In the event of a CDP falling below the required collateral ratio, that CDP will be siezed by the liquidator module. When a `lot` of collateral has been siezed due to liquidations, that collateral is auctioned by the auction module for stable tokens using a forward reverse auction. In normal times, this auction is expected to raise sufficient stable tokens to wipe out the debt originally held by the CDP owners, along with a small liquidation penalty.

In the event collateral auctions fail to raise the requisite amount of stable tokens, Kava tokens are auctioned by the auction module for stable tokens using a reverse auction until the global collateral ratio is reached. In this way, the Kava token serves as a lender of last resort in times of under-colleteralization.