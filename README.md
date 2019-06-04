# USDX

A protocol for collateralized loans using any digital asset and dollar-denominated debts (stablecoins) built on the [cosmos-sdk](https://github.com/cosmos/cosmos-sdk).

## Quick Start

To read about the design of the system, see [here](./spec/usdx.md).

### Installing

    go install ./blockchain/cmd/usdxd ./blockchain/cmd/usdxcli

For local development:

    usdxd init --chain-id usdx-test testing
    usdxcli keys add alice
    # enter a new password
    # re-enter password
    usdxd add-genesis-account $(usdxcli keys show alice -a) 10xrs,1oracle,100000000stake
    usdxd gentx --name alice
    # enter password
    usdxd collect-gentxs
    usdxcli config trust-node true
    usdxd start

Check account balance using `usdxcli query account $(usdxcli keys show alice -a)`
Check which assets are in the pricefeed `usdxcli query pricefeed assets`
