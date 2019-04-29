# XRP Sidechains

##

This specification builds on previous work of Tendermint based sidechains, including [Peggy](https://github.com/cosmos/peggy) and [BitcoinPeg](https://github.com/nomic-io/bitcoin-peg).

We present a design for a one-way XRP pegged sidechain based on the Tendermint BFT consensus protocol. This design enables decentralized groups of validators to manage a reserve of XRP within a sidechain environment, whereby the native functionality of XRP can be expanded to include custom application code and smart contracts.

## Technical Overview
The XRP pegged sidechain is comprised of:
* A dynamic reserve multisignature wallet on the XRP ledger [Reserve]
* A peg client that relays and verifies transactions between the two networks [Client]
* The sidechain [Sidechain]


### Reserve

The validators of the sidechain manage XRP reserves on the XRP ledger. They do this by committing to their XRP address on the sidechain and creating a multisignature wallet on the XRP ledger that requires 2/3 of validator voting power to execute a transaction. The following script shows the creation of such a reserve.

``` javascript
const RippleAPI = require('ripple-lib').RippleAPI

(async () => {
  const signers = [
    {"SingerEntry":{Account: <xrpAddress1>, SignerWeight: <votingPower1>}},{"SingerEntry":{Account: <xrpAddress2>, SignerWeight: <votingPower2>}},{"SingerEntry":{Account: <xrpAddressn>, SignerWeight: <votingPowern>}}]

  const totalWeight = signers.reduce((s, a) => s + a.SignerEntry.SignerWeight, 0)

  const reserve = {
    Account: <reserveAccount>,
    Secret: <reserveSecret>
  }


  const api = new RippleAPI({
    server: 'wss://s.altnet.rippletest.net:51233' // Public rippled server
  })
  await api.connect()

  const  txJson = {
      "Flags": 0,
      "TransactionType": "SignerListSet",
      "Account": reserve.Account,
      "Sequence": 1,
      "Fee": "12",
      "SignerQuorum": Math.ceil(((totalWeight*2)/(totalWeight*3) * totalWeight)),
      "SignerEntries": signers,
    }

  await api.sign(JSON.stringify(txJson), reserve.Secret)
  await api.submit(signedTx.signedTransaction)
})()
```

A current limitation of the multisignature XRP address is that is has at most 8 accounts. However, a multisignature address can have multisignature addresses as it's signers, so the XRP ledger can accommodate 64 signers by batching signers once, and more if necessary.

### Deposits

Users of the XRP ledger can deposit their XRP to the reserve and receive pegged XRP on the sidechain. To do this, they send an XRP transaction to the validator multisignature address, along with the address they will use on the sidechain.

This transaction MUST be a [Payment](https://github.com/ripple/ripple-lib/blob/develop/docs/index.md#payment) with the following specification:
1. The only destination address is the reserve account
2. There is exactly one memo with a data field that specifies a valid address on the sidechain.
    * The address MUST be 20 bytes, bech32 encoded, with prefix 'usdx'.

The peg client listens to the reserve multisig address and broadcasts valid deposit transactions, along with a proof that they were included in a closed XRP ledger. When the sidechain receives the proof of deposit, it mints tokens on the sidechains equal to the amount that was deposited.

### Withdrawals
Users of the sidechain can withdraw from the reserve by sending a withdrawal transaction to the sidechain. When a withdrawal is included in a block in the sidechain, validators individually sign and broadcast transactions that withdraw funds from the reserve on the XRP ledger to the user. Peg clients combine these signed transactions and submit them to the XRP ledger when 2/3 of validators have broadcast transactions.

### Validator Set Notarization

When the validator set changes and new XRP addresses and weights are committed to on the sidechain, the multisignature reserve wallet can be updated by updating the signer list and weights. This can be done in response to each change, or at a specified threshold. As long as the transition in the validator set is not greater than 1/3 of the validator voting power, clients can consider the reserve wallet safe.

A notarization transaction can be signed by each validator independently and broadcast after a change to the validator set is committed to the side chain. This transaction MUST be a [Settings](https://github.com/ripple/ripple-lib/blob/develop/docs/index.md#settings) with the following specification
1. The updated signer list is specified in SignerEntries
2. There is exactly one memo with a data field that specifies the sha256 hash of the string concatenation of all validator addresses.

Clients can listen for these signed transactions, combine them, and broadcast them to the XRP ledger when a sufficient quorum of validators have signed transactions.