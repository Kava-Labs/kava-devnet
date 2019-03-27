const RippleAPI = require('ripple-lib').RippleAPI
const sha256 = require('js-sha256')

;(async () => {
  const api = new RippleAPI({
    server: 'wss://s.altnet.rippletest.net:51233' // Public rippled server
  })
  await api.connect()

  const multiSignAddress = {
    Address: "rs16hESfGChwAnK97oSdRJq4A18gcJbE7j",
  }

  const currentSignerAddresses = [
    {
    address: "rJGnjrcCmRnaecWYtXg57WuoVHUxiNkjzj",
    secret: "sstZdZjm1NSSUcsmp1UtVxjjsJgyC",
    weight: 1
    },
    {
      address: "rNSfagh9usmvrBVWUTeqyAAW9BCP1NdWyY",
      secret: "ssDEJJKNVv1GHEskCSMeTqJsXkwoN",
      weight: 1
    },
    {
      address: "rJpZtBRwAwpTwABapNGs5ojzT8C24d9ewx",
      secret: "sa3VWXx1ZDMCUXitKGsMx2TXJUiab",
      weight: 1
    }
  ]

  const newSignerAddresses = [
    {
      address: "rPcNyFXrgDm9cqXMEu64KSTPiWpVBc9Yss",
      secret: "ssJQRd4UwQnkQTBRGMFEyhG3XKKZH",
      weight: 1
    }
  ]

  const allSigners = currentSignerAddresses.concat(newSignerAddresses)
  const allAddresses = allSigners.map((val) => val.address)
  const concatSignerAddresses = allAddresses.reduce((acc, address) => address + acc, '')
  // console.log(concatSignerAddresses)
  const hashedSignerAddresses = sha256(concatSignerAddresses)
  // console.log(hashedSignerAddresses)
  const totalWeight = allSigners
    .map((val) => val.weight)
    .reduce((acc, weight) => weight + acc, 0)
  // console.log(totalWeight)

  const newSigners = allSigners.map((val) => ({address: val.address, weight: val.weight}))
  // console.log(newSigners)
  const settingsJson = {
    signers: {
      threshold: Math.ceil(((totalWeight * 2)/(totalWeight* 3)* totalWeight)),
      weights: newSigners
    },
    memos: [
      {
        "data": hashedSignerAddresses
      }
    ]
  }
  // console.log(JSON.stringify(settingsJson))

  const settingsTx = await api.prepareSettings(
    multiSignAddress.Address,
    settingsJson,
    {signersCount: 2}
  )

  // console.log(settingsTx.txJSON)

  const sig1 = api.sign(
    settingsTx.txJSON,
    currentSignerAddresses[0].secret,
    {signAs: currentSignerAddresses[0].address}

  )

  const sig2 = api.sign(
    settingsTx.txJSON,
    currentSignerAddresses[1].secret,
    {signAs: currentSignerAddresses[1].address}
  )

  const combinedTx = api.combine([sig1.signedTransaction, sig2.signedTransaction])
  const receipt  = await api.submit(combinedTx.signedTransaction)
  console.log(receipt)



})()