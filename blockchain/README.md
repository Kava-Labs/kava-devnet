# WIP CDP Zone

## Install

    go install ./cmd/usdxd ./cmd/usdxcli

## Setup

    usdxd init --chain-id usdx-test
    usdxcli keys add alice
    # enter a new password
    # re-enter password
    usdxd add-genesis-account $(usdxcli keys show alice -a) 10xrs,1oracle
    usdxcli config trust-node true

## Run (in separate windows)

    usdxd start
    
Check account balance using `usdxcli query account $(usdxcli keys show alice -a)`
Check which assets are in the pricefeed `usdxcli query pricefeed assets`
