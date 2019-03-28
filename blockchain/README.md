# WIP XRP Peg Zone

## Install

    go install ./cmd/usdxd ./cmd/usdxcli ./client

## Setup

    usdxd init --chain-id usdx-test
    usdxcli keys add validatorName
    # enter a new password
    # re-enter password
    usdxd add-genesis-account <validatorName's address> 10xrs
    usdxcli config trust-node true

## Run (in separate windows)

    usdxd start
    client

Send a xrp tx to the multisig using `node xrpPayToMultisig.js` in `examples/js`.
You will see the client process the transaction and sumbit it to the blockchain.

Check account balance using `usdxcli query account usdx1cwd9fxrxvz5yq5qdrtscvmc4h0l7mqu9hldkfa`
Querying won't work until there is at least one tx sent to that address.