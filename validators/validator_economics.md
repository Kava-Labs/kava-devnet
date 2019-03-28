## USDX Validator Economics

USDX is a XRP-Backed stablecoin. The USDX stablecoin system is implemented as a Cosmos Zone designed to peg in XRP, issue an XRP-collateralized token (called USDX) and dynamically maintain its price stability to USD with decentralized mechanisms. The native token of the USDX blockchain is XRS, a governance and validation taken that is staked by validators, who vote on blocks and participate in governance.

## Validator Overview
For an introduction to staking see the (Cosmos Validator FAQ)[https://github.com/cosmos/cosmos/blob/master/VALIDATORS_FAQ.md]
Validators of the USDX blockchain must meet two basic requirements:

1. Run the correct version of the USDX software in a secure, high availability validator setup. This means that the validator must be highly available and the validators private keys must not be compromised. Validators who run incorrect software or are compromised are subject to slashing. This normally involves dedicated hardware, HSMs, secure colocation, and DDoS protection through sentry nodes.
2. Participate in governance of the USDX protocol. This includes voting on all governance proposals. Governance proposals will control the parameters of the collateral backed stable coin and decisions to upgrade the USDX software and protocol.

### Staking and Delegation
Validators can stake XRS directly or have XRS delegated to them. There is no minimum number of XRS required to validate, and the top 100 validators by weight (self-delegated plus delegated XRS) will be eligible to validate the USDX blockchain.

## XRS Token Economics

### Transaction Fees
Transaction fees are required for all USDX transactions, which includes transferring pXRP to other users, transferring pXRP to the USDX collateral module, drawing USDX from a CDP, closing a CDP, and sending USDX between users. Transaction fees for each block are split by validators according to weighted stake.

### Deflation
When USDX users send pXRP collateral to the USDX collateralized debt position (CDP) module, they can draw USDX up to the liquidation percentage of their collateral. Each CDP is subject to a USDX denominated stability fee that must be paid in XRS when the CDP is closed. The initial rate of the stability fee will be 1% APR. When CDPs are closed, the XRS that is used to pay the stability fee is burned, decreasing the total outstanding supply of XRS.

#### Example

If CDP users close CDPs worth 10 million USDX during the first year of operation, then 100,000 USDX worth of XRS will be burned during the first year of operation. Assuming a 10 million USDX market cap of XRS and an initial total supply of 100,000,000 XRS, then 1,000,000 XRS would be burned during the first year of operation of the USDX blockchain.


## XRS Token Price and Validator Incentives

### XRS Price and Distribution

XRS total supply: 100,000,000
XRS available to validators: 10,000,000
XRS token pre-sale price: 0.10 USD/XRS
XRS pre-sale market cap: $10 Million

### USDX Validation Incentives

Validators purchasing XRS tokens are additionally allocated one (1) XRS token (“Validator Allocation”) for every pre-sale XRS token they purchase. Validator Allocation is subject to a 720 day vesting schedule, under which 1/24th of the total allocation will vest at the end of each thirty (30) day period (“Vesting Period”) after the vesting commencement date. Allocation is additionally subject to blockchain validation. For every Vesting Period, validators will receive the percentage of Validator Allocation in that Vesting Period directly proportional to the number of pre-commits they have signed during that Vesting Period. The vesting commencement date occurs at USDX mainnet launch.

#### Example
1. A validator purchases 240,000 XRS tokens.
2. They will be allocated a maximum additional 240,000 XRS tokens as Validator Allocation in the following way:
    * At the launch of USDX mainnet, the 240,000 XRS tokens are reserved.
    * At the end of the first Vesting Period (30 days) 10,000 XRS tokens vest (240,000XRS/24). If the validator signed 90% of pre-commits during this first Vesting Period they will be distributed 9,000 XRS and are no longer eligible for the remaining 1,000 XRS vested in that period.
    * On the second Vesting Period another 10,000 XRS tokens vest. If the validator signed 95% of pre-commits during during this second Vesting Period they will be distributed 9,500 XRS and are no longer eligible for the remaining 500 XRS in that period.
    * This process repeats for the remaining Vesting Periods in the 720 day vesting schedule.
