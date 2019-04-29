The MakerDAO system uses 'price feeds', which are trusted oracles that determine the price of ETH/USD, MKR/USD, WETH/PETH

Good reference (and place to ask questions):
  https://chat.makerdao.com/channel/feeds

The client implementation that oracles are supposed to run is here:
https://github.com/makerdao/setzer

The contracts governing single collateral dai:
https://github.com/makerdao/sai

Good glossary for the code:
https://github.com/makerdao/dss/wiki/Glossary

Reference implementation with some more definitions of terms:
https://makerdao.com/purple

Location of some of the auxilary contracts (DS-thing, etc) used in sai:
https://github.com/dapphub

From maker chat:

> You can see exactly how the Oracle clients are polling their prices. You can run the software and verify for yourself.
> Now the issue here in my mind is we don't have a way to guarantee Oracles aren't running modified clients where they change how the price is queried. What I mentioned above that I'd like to do is have Oracles generating a zero knowledge proof of how they derived the price which is then verified on-chain.
> This is currently not really feasible because it's super expensive to do zk stuff on-chain, but there's definitely progress being made in this area especially with precompiles. So I'm optimistic this is something we can tackle in the future.

> @nik Thanks! I've actually heard that term before, but somehow didn't remember it when looking for the bot.

> So the price feeds for ETH are based on Bitstamp, Kraken, GDAX and Gemini. Median of their ETH/USD prices is used as the final price. Price needs to be fetched from at least 3 of them or it's declared unsuccessful.
> Assuming oracles run setzer with default sources and I read the code right.


Draft oracle module PR on the sdk (no longer in develop):
https://github.com/cosmos/cosmos-sdk/pull/1069/files

DAI/USD feed:
https://dai.stablecoin.science/


What is the difference in liquidation process between single collateral and multi-collateral dai?

In single collateral dai, when a liquidation happens:
1. The system immediately takes control of all collateral.
2. The collateral is auctioned off at a 3% discount until the debt is covered, as well as the stability and liquidation fees. Any remaining collateral is then returned to the CDP owner.

In multi collateral dai, when a liquidation happens:
1a. A "debt auction" begins, which sells MKR for DAI up to the amount of DAI that was drawn from the CDP. That DAI is burned. (Dilution)
1b. Simultaneously, a liquidation auction begins, which sells the collateral in the CDP for MKR. The collateral is sold to cover the debt, stability fees, and liquidation fee. The MKR is burned.

Assuming the CDP was not under-collateralized, the MKR burned should be greater than the MKR minted (greater by the stability fee + liquidation fee). Thus, the total supply of MKR will decrease during these auctions, with the benefits presumably distributed to holders of remaining MKR.


Why is DAI not convertible for collateral?

A straight-forward stability mechanism for DAI would be to allow DAI holders to redeem dai for $1 worth of collateral at the current price feed. While this would almost certainly increase the [stability of dai](https://dai.stablecoin.science/), it would make the price feeds susceptible to front-running. If one could predict the future price direction of ETH faster than the price feed, they could convert DAI for ETH at a reliable profit. Without a real-time price oracle, or a front-running resistant oracle, the system couldn't do redemption in a decentralized way.


<!-- 100 ETH

2500 DAI

Accumulated fees 250 DAI

Scenario 1

If fees not applied against draw()

I can create 1500 more DAI before getting wiped

Scenario 2

If fees applied against draw()

I can create 1250 DAI before getting wiped?

Total collateral locked

Scenario 1:

40 ETH
4000 Dai

Scenario 2
37.5 ETH
3750 Dai

In liquidation

Scenario 1
4000 Dai * .14 = 560 DAI

Scenario 2
3750 Dai * .14 = 525 DAi

After liquidation

Scenario 1
4000 Dai + 54.4 ETH = 94.4 ETH

Scenario 2
3750 Dai + 57.25 ETH = 94.75 ETH

 -->


 Inheritance model of DS-stuff

            DSThing
    ---------------------
    |         |         |
    |         |         |
  DSAuth    DSNote     DSMath



            DSToken
    ---------------------
    |                   |
    |                   |
  DSStop          DSTokenBase
                  -----------
                  |         |
                  |         |
                DSMath     ERC20


            DSGuard
            -------
              |
            DSAuth


            DSRoles
            -------
               |
             DSAuth


            DSValue
            -------
              |
            DSThing
    ---------------------
    |         |         |
    |         |         |
  DSAuth    DSNote     DSMath


  <!-- Important read on how difficult implementing this system on Solidty/ETH is:
  https://makerdao.com/CodeReview/Sai_Final_Report.pdf
  Hopefully cosmos-sdk makes this more managable! -->

Price feed mechanism:
A group of whitelisted oracles are posting prices to the USDX blockchain for each given collateral type:
* COL1:USD
* COL2:USD
* ...

The oracles submit transactions with the following format:
```
MsgPostPrice{
			From: from,
			AssetCode: assetCode,
			Price: price,
			Expiry: expiry
}
```

These transactions are submitted to the `pricefeed` module and commited to a block. The block number that the price was commited at is appended to the transaction.

The `pricefeed` module gathers are MsgPostPrice transactions from the previous n blocks, and selects the most recent MsgPostPrice transaction from each validator (measured by the transaction with the highest blocknumber). If a majority of price oracles have submitted MsgPostPrice transactions, the `pricefeed` module takes the median of these prices and calls `UpdatePrice` on the keeper. This updates the price in storage of the `pricefeed` module and becomes the latest price for that collateral type in the USDX CDP system.