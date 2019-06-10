# Kava

[![CircleCI](https://circleci.com/gh/Kava-Labs/kava-devnet/tree/master.svg?style=shield)](https://circleci.com/gh/Kava-Labs/kava-devnet/tree/master)
[![Go Report Card](https://goreportcard.com/badge/github.com/kava-labs/kava-devnet)](https://goreportcard.com/report/github.com/kava-labs/kava-devnet)
[![API Reference](https://godoc.org/github.com/Kava-Labs/kava-devnet?status.svg
)](https://godoc.org/github.com/Kava-Labs/kava-devnet)
[![license](https://img.shields.io/github/license/Kava-Labs/kava-devnet.svg)](https://github.com/Kava-Labs/kava-devnet/blob/master/LICENSE)

A protocol for creating a collateral-backed stablecoin using any digital asset. Built on the [cosmos-sdk](https://github.com/cosmos/cosmos-sdk).

## Quick Start

To read about the design of Kava, see [here](./spec/kava.md).

### Installing
  To install, clone the repo and go to the new directory.

    go install ./blockchain/cmd/kavad ./blockchain/cmd/kavacli

For local development:

    kavad init --chain-id kava-test testing
    kavacli keys add alice
    # enter a new password
    # re-enter password
    kavad add-genesis-account $(kavad keys show alice -a) 10kava,1oracle,100000000stake
    kavad gentx --name alice
    # enter password
    kavad collect-gentxs
    kavacli config trust-node true
    kavad start

Check account balance using `kavacli query account $(kavacli keys show alice -a)`
Check which assets are in the pricefeed `kavacli query pricefeed assets`


### Testnet

  To join the latest testnet. see [here](https://github.com/Kava-Labs/kava)

