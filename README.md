# USDX

[![CircleCI](https://circleci.com/gh/Kava-Labs/usdx/tree/master.svg?style=shield)](https://circleci.com/gh/Kava-Labs/usdx/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/kava-labs/usdx)](https://goreportcard.com/report/github.com/kava-labs/usdx)
[![API Reference](https://godoc.org/github.com/Kava-Labs/usdx?status.svg
)](https://godoc.org/github.com/Kava-Labs/usdx)
[![license](https://img.shields.io/github/license/Kava-Labs/usdx.svg)](https://github.com/Kava-Labs/usdx/blob/master/LICENSE)

A protocol for creating a collateral-backed stablecoin using any digital asset. Built on the [cosmos-sdk](https://github.com/cosmos/cosmos-sdk).

## Quick Start

To read about the design of USDX, see [here](./spec/usdx.md).

### Installing
  To install, clone the repo and go to the new directory.

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
