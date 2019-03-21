The MakerDAO system uses 'price feeds', which are trusted oracles that determine the price of ETH/USD, MKR/USD, WETH/PETH

Good reference (and place to ask questions):
  https://chat.makerdao.com/channel/feeds

The client implementation that oracles are supposed to run is here:
https://github.com/makerdao/setzer


From maker chat:

> You can see exactly how the Oracle cients are polling their prices. You can run the software and verify for yourself.
> Now the issue here in my mind is we don't have a way to guarantee Oracles aren't running modified clients where they change how the price is queried. What I mentioned above that I'd like to do is have Oracles generating a zero knowledge proof of how they derived the price which is then verified on-chain.
> This is currently not really feasible because it's super expensive to do zk stuff on-chain, but there's definitely progress being made in this area especially with precompiles. So I'm optimistic this is something we can tackle in the future.

> @nik Thanks! I've actually heard that term before, but somehow didn't remember it when looking for the bot.

> So the price feeds for ETH are based on Bitstamp, Kraken, GDAX and Gemini. Median of their ETH/USD prices is used as the final price. Price needs to be fetched from at least 3 of them or it's declared unsuccessful.
> Assuming oracles run setzer with default sources and I read the code right.


Draft oracle module PR on the sdk (no longer in develop):
https://github.com/cosmos/cosmos-sdk/pull/1069/files

DAI/USD feed:
https://dai.stablecoin.science/