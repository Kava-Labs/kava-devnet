# WIP CDP Zone

## Install

    go install ./cmd/usdxd ./cmd/usdxcli

## Setup

To start a local development testnet

    usdxd init --chain-id usdx-test testing
    usdxcli keys add alice
    # enter a new password
    # re-enter password
    usdxd add-genesis-account $(usdxcli keys show alice -a) 10xrs,1oracle,100000000stake
    usdxd gentx --name alice
    # enter password
    usdxd collect-gentxs
    usdxcli config trust-node true

## Run (in separate windows)

    usdxd start

Check account balance using `usdxcli query account $(usdxcli keys show alice -a)`
Check which assets are in the pricefeed `usdxcli query pricefeed assets`
