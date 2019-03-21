## USDX Validator Economics

The USDX blockchain is a decentralized protocol for creating an XRP peg zone and a stable asset called USDX backed by pegged XRP (pXRP). Similar to Ether in Single Collateral Dai, users of the USDX blockchain lock pXRP into collateralized debt positions (CDPs) and are able to draw the USDX stable coin against the CDPs. The USDX blockchain is built using the Cosmos-sdk and will implement the USDX stable coin system as a series of modules built on the SDK, rather than smart contracts built on the EVM. The native token of the USDX blockchain is XRS, a governance and validation token that is staked by validators, who vote on blocks and participate in governance.

## Validator Overview
For an introduction to staking see the (Cosmos Validator FAQ)[https://github.com/cosmos/cosmos/blob/master/VALIDATORS_FAQ.md]
Validators of the USDX blockchain must meet two basic requirements:

1. Run the correct version of the USDX software. This means that the validator must be highly available and the validators private keys must not be compromised. Validators who run incorrect software or are compromised are subject to slashing.
2. Participate in governance of the USDX protocol. This includes voting on all governance proposals. Governance proposals will control the parameters of the collateral backed stable coin and decisions to upgrade the USDX software and protocol.

### Staking and Delegation
Validators can stake XRS directly or have XRS delegated to them. There is so minimum number of XRS required to validate, and the top 100 validators by weight (self-delegated plus delegated XRS) will be eligible to validate the USDX blockchain.

## XRS Token Economics

### Transaction Fees
Transaction fees are required for all USDX transactions, which includes transferring pXRP to other users, transferring pXRP to the USDX collateral module, drawing USDX from a CDP, closing a CDP, and sending USDX between users. Transaction fees for each block are split by validators according to weighted stake.

### Deflation
When USDX users send pXRP collateral to the USDX collateralized debt position (CDP) module, they can draw USDX up to the liquidation percentage of their collateral. Each CDP is subject to a USDX denominated stability fee that must be paid in XRS when the CDP is closed. The initial rate of the stability fee will be 1% APR. When CDPs are closed, the XRS that is used to pay the stability fee is burned, decreasing the total outstanding supply of XRS.

#### Example

If CDP users close CDPs worth 10 million USDX during the first year of operation, then 100,000 USDX worth of XRS will be burned during the first year of operation. Assuming a 10 million USDX market cap of XRS and an initial total supply of 100,000,000 XRS, 1,000,000 XRS would be burned during the first year of operation of the USDX blockchain.


## Investment and Vesting Period
Validators participating in the XRS token sale can buy XRS tokens with the following vesting and commitments:

XRS total supply: 100,000,000
XRS available to validators: 10,000,000 for direct investment plus vesting tokens
XRS token price: 0.10 USD/XRS
XRS pre-sale market cap: $10 Million


Validators can receive up to 10 additional XRS tokens per $1 USD invested. This investment is subject to a 2 year vesting schedule and is tied to validation of the USDX blockchain. At the end of each thirty (30) day period, validators will receive a distribution of XRS tokens proportional to the number of pre-commits they have signed during the thirty day period. For example, if a validator investor has a total vest of 1,216.667 XRS tokens, they can receive a maximum of 50,000 XRS tokens at the end of each 30 day period. If they sign 90% of pre-commits during a 30 day period, they would receive 45,000 XRS tokens.
