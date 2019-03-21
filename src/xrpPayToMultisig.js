const RippleAPI = require('ripple-lib').RippleAPI;

(async () => {
  const api = new RippleAPI({
    server: 'wss://s.altnet.rippletest.net:51233' // Public rippled server
  })
  await api.connect()

  const xrpUser = {
    Address: "rsGPNkSLt36BDLMgPAYKifFvCphQJZ2qJw",
    Secret: "shr6CrnZqh7CgrS8izxd2rWsSgHCn"
  }

  const multiSignAddress = {
    Address: "rs16hESfGChwAnK97oSdRJq4A18gcJbE7j",
  }

  const txJson = {
    source: {
      address: xrpUser.Address,
      amount: {value: "1000000000", currency: 'drops'}
    },
    destination: {
      address: multiSignAddress.Address,
      minAmount: {
        value: '' + "1000000000",
        currency: 'drops'
      }
    },
    memos: [
      {
        "data": "usdxaddrrsGPNkSLt36BDLMgPAYKifFvCphQJZ2qJw"
      }
    ]
  }

  const payment = await api.preparePayment(xrpUser.Address, txJson)
  const signedTx = await api.sign(payment.txJSON, xrpUser.Secret)
  const receipt  = await api.submit(signedTx.signedTransaction)
  console.log(receipt)
  // memo returns 7573647861646472727347504E6B534C74333642444C4D675041594B69664676437068514A5A32714A77
  // to recover:
  // Buffer.from("7573647861646472727347504E6B534C74333642444C4D675041594B69664676437068514A5A32714A77", "hex").toString()

})()