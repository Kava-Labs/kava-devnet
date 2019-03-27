const RippleAPI = require('ripple-lib').RippleAPI;

(async () => {
  const api = new RippleAPI({
    server: 'wss://s.altnet.rippletest.net:51233' // Public rippled server
  })
  await api.connect()

  const withdrawingUser = {
    Address: "rsGPNkSLt36BDLMgPAYKifFvCphQJZ2qJw"
  }

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
    Address: "rs16hESfGChwAnK97oSdRJq4A18gcJbE7j",
  }

  const multisigSequence = (await api.getAccountInfo(multiSignAddress.Address)).sequence
  const txJson = {
    source: {
      address: multiSignAddress.Address,
      amount: {value: "1000000000", currency: 'drops'}
    },
    destination: {
      address: withdrawingUser.Address,
      minAmount: {
        value: '' + "1000000000",
        currency: 'drops'
      }
    }
  }

  const payment  = await api.preparePayment(
    multiSignAddress.Address,
    txJson,
    {
      sequence: multisigSequence,
      signersCount: 2
    })
  const sig1 = api.sign(
    payment.txJSON,
    signerAddresses[0].secret,
    {signAs: signerAddresses[0].address}
    )
  const sig2 = api.sign(
    payment.txJSON,
    signerAddresses[1].secret,
    {signAs: signerAddresses[1].address}
    )

  const combinedTx = api.combine([sig1.signedTransaction, sig2.signedTransaction])
  const receipt  = await api.submit(combinedTx.signedTransaction)
  console.log(receipt)

})()