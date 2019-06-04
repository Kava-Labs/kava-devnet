# USDX

USDX is a collateralized debt position system built on Cosmos.

## Background

Measured by encumbered collateral, debt issuance for the purpose of leveraged exposure has been the most succesful secondary financial usecase of cryptocurrencies. MakerDao, and the associated DeFi ecosystem on Ethereum, represent arguably the best example of product-market fit for any blockchain product or service that is not a base-layer protocol. While the number of users of these products is generally small, the potential for synthetic asset issuance (USD pegged stablecoins, or debt denominated in a stable basket of goods) that spans jurisdictions and runs largely autonomously is large.

Cosmos is a new blockchain protocol that uses Tendermint BFT for consensus and is designed with a hub-and-spoke model of cross-blockchain interoperabilty that emphasizes composability and self-sovereignty of application specific blockchains. We believe one of the primary financial usecases for the Cosmos ecosystem will be the the issuance of decentralized pegged assets like pegged Bitcoin (https://github.com/nomic-io/bitcoin-peg/blob/master/bitcoinPeg.md), as well as other pegged crpyto-native and traditional financial assets.

We are building a blockchain on the cosmos-sdk for the purpose of issuing Collateralized Debt Positions (CDPs) for assets in the cosmos ecosystem. The design of the CDP zone is inspired by Multi-Collateral Dai (https://github.com/makerdao/dss) and will allow users to lock their assets as collateral and draw a dollar-denominated debt off of their collateral. We believe this zone is a useful addition to the Cosmos ecosystem, providing a native way for users to gain leveraged exposure to a basket of assets in the cosmos ecosystem.


## Design

There are 4 modules that make up the system:

* Pricefeed
* Auction
* CDP
* Liquidator

The pricefeed module implements a simple price oracle where a group of white-listed oracles post prices for various assets in the system. The median price of all valid oracle prices is taken as the current price in the system. Adding and removing of assets is controlled by governance proposals.

The Auction module implements three distinct auction types that control the supply of bad debt and surplus in the CDP system.

**Forward Auction** A standard auction where a seller takes increasing bids for an item. Each bid increments the price, as well as the duration of the auction. This auction type is used when there is a surplus of collected fees in the system. The surplus is converted to stablecoins and sold for governance tokens.

**Reverse Auction** An auction where a buyer solicits decreasing bids for a particular item or lot of items. This auction type is used when governance tokens are sold (minted) in exchange for stablecoins, to cover shortfalls after failed collateral auctions.

**Forward Reverse Auction** An auction where a buyer solicits increasing bids for a lot of goods, up to some ceiling. After the ceiling is reached, each bid lowers the amount of goods being sold for the ceiling  price. This type of auction is used when collateral is siezed from a risky CDP and sold for stablecoins to cover the debt.

The CDP module is a factory for creating CDPs and storing the global state of the debt system. It allows users to create, modify, and close CDPs for any collateral type in the pricefeed module. It also sets the global parameters of the system, which can be altered by governance proposals. These parameters include the global debt limit (the total amount of stablecoins that can be in circulation), the debt limit for each collateral type, and the collateralization ratio for each collateral type.

The liquidator module tracks the status of CDPs based on prices in the pricefeed module and is responsible for siezing collateral from risky CDPs and sending it to the auction module.

The system is secured by Kava, a staking and governance token. Staked Kava tokens receive inflationary block rewards and are eligible to create and vote on governance proposals.

### Liquidation and Recolateralization
In the event of a CDP falling below the required collateral ratio, that CDP will be siezed by the liquidator module. When a `lot` of collateral has been siezed due to liquidations, that collateral is auctioned by the auction module for stable tokens using a forward reverse auction. In normal times, this auction is expected to raise sufficient stable tokens to wipe out the debt originally held by the CDP owners, along with a small liquidation penalty.

In the event collateral auctions fail to raise the requisite amount of stable tokens, Kava tokens are auctioned by the auction module for stable tokens using a reverse auction until the global collateral ratio is reached.