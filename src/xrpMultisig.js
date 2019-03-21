const RippleAPI = require('ripple-lib').RippleAPI

async function run() {
  // Example of setting up a 2 of 3 multisig address using ripple-lib

  // Notes:
  // 1. The best docs on ripple-lib are here: https://github.com/ripple/ripple-lib/blob/develop/docs/index.md
  // 2. Ripple testnet explorer that shows signers of multi-sig: http://ripplerm.github.io/ripple-wallet/
  // 3. From a technical perspective, there's nothing stopping you from having the
  // multisignature address by a multisig of multisigs. So, if it's an issue, we
  // can tell people that the existing XRP ledger can support more than 8 validators for the
  // multisig.
  // 4. You can set a multisign to be the only option for an address.
  // 5. How the private key for the original multisign is derived is unclear.

  // The most straightforward is to start with a regular XRP keypair, then setup multisign, then remove the keypair, and only accept funds to the address if it's multi-sign only. Clients could enforce this, so as long as you're running the right software and the validators are commiting their XRP address on the USDX chain, you have BFT (I think?) guarantee that you're sending funds to a multi-sig controlled by the validators of the UDSX chain.
  // You could aslo do MPC ECDSA, and then I'm not sure you would even need to create the multisig with Ripple-Lib. However, I don't think you could prove that it's a multisig to clients then.

  const api = new RippleAPI({
    server: 'wss://s.altnet.rippletest.net:51233' // Public rippled server
  })
  await api.connect()

  const signerAddresses = [
    {
    address: "rJGnjrcCmRnaecWYtXg57WuoVHUxiNkjzj",
    secret: "sstZdZjm1NSSUcsmp1UtVxjjsJgyC"
    },
    {
      address: "rNSfagh9usmvrBVWUTeqyAAW9BCP1NdWyY",
      secret: "ssDEJJKNVv1GHEskCSMeTqJsXkwoN"
    },
    {
      address: "rJpZtBRwAwpTwABapNGs5ojzT8C24d9ewx",
      secret: "sa3VWXx1ZDMCUXitKGsMx2TXJUiab"
    }
  ]

  const multiSignAddress = {
    address: "rs16hESfGChwAnK97oSdRJq4A18gcJbE7j",
    secret: "sptZMEp27uRzAbdh7VarnWotkqC8v",
    sequence: async(a) => {
      const api = new RippleAPI({
        server: 'wss://s.altnet.rippletest.net:51233' // Public rippled server
      })
      await api.connect()
      const info = await api.getAccountInfo(a)
      return info.sequence
    }
  }

  const signerEntries = [
    {
        "SignerEntry": {
            "Account": "rJGnjrcCmRnaecWYtXg57WuoVHUxiNkjzj",
            "SignerWeight": 1
        }
    },
    {
        "SignerEntry": {
            "Account": "rNSfagh9usmvrBVWUTeqyAAW9BCP1NdWyY",
            "SignerWeight": 1
        }
    },
    {
        "SignerEntry": {
            "Account": "rJpZtBRwAwpTwABapNGs5ojzT8C24d9ewx",
            "SignerWeight": 1
        }
    }
  ]

  const setupXrpMultisig= async (api, xrpAddress, signers, quorum) => {
    const  txJson = {
      "Flags": 0,
      "TransactionType": "SignerListSet",
      "Account": xrpAddress.address,
      "Sequence": await xrpAddress.sequence(xrpAddress.address),
      "Fee": "12",
      "SignerQuorum": quorum,
      "SignerEntries": signers
    }

    const signedTx = await api.sign(JSON.stringify(txJson), xrpAddress.secret)
    console.log(signedTx)
    const receipt = await api.submit(signedTx.signedTransaction)
    console.log(receipt)
    return
  }

  // Set up two of three multisign for rs16hESfGChwAnK97oSdRJq4A18gcJbE7j
  await setupXrpMultisig(api, multiSignAddress, signerEntries, 2)
}

run().then(res => process.exit(0)).catch(e => { console.log(e); process.exit(1) })