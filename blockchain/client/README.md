## Notes on how to submit transactions in the cosmos sdk

This is roughly the flow from starting with a sdk.Msg (like a sendCoins msg) to signing and submitting it to the blockchain. In the sdk this logic is wrapped in layers of not very helpful helper functions, residing in:

 - cosmos-sdk/client/utils - helper functions
 - cosmos-sdk/client/context - CLIContext
 - cosmos-sdk/x/auth/client/txbuilder - TxBuilder

```go
    // Fetch account object from the chain (talks to the node over rpc)(requires a whole CliContext to be setup, but underneath just sends a request to the node)
    account, err := cliCtx.GetAccount(address)
    // Wrap the plain sdk.Msg's in a format to be signed
    stdSignMsg := authtxb.StdSignMsg{
        ChainID:       chainID,
        AccountNumber: account.GetAccountNumber(),
        Sequence:      account.GetSequence(),
        Memo:          "",
        Msgs:          msgs,
        Fee:           auth.NewStdFee(
                                    10000,
                                    sdk.Coins{sdk.NewCoin("pxrp", sdk.NewInt(1))}
                                ),
    }
    // get signBytes and create auth.StdSignature (using helper function) (needs a keybase - the encrypted priv key database)
    sig, err := authTxBuilderPkg.MakeSignature(keybase, keyName, passphrase, stdSignMsg)
    // wrap the signature and wrapped msgs into a standard transaction
    stdTx := auth.NewStdTx(
        stdSignMsg.Msgs,
        stdSignMsg.Fee,
        []auth.StdSignature{sig},
        stdSignMsg.Memo
    )
    // 	encode it to binary
    encoder := sdk.GetConfig().GetTxEncoder()
    tx, _ := encoder(stdTx)

    // 	submit transaction to the node (over rpc)
    res, err := node.BroadcastTxCommit(tx)
    if err != nil {
        panic(err)
    }
    // 	Check it made it into the mempool (CheckTx) and was commited into a block (DeliverTx)
    if res.CheckTx.IsOK() || res.DeliverTx.IsOK() {
        panic("tx not confirmed")
    }
```

The helper functions are attached to two structs `CLIContext` and `TxBuilder`. These are basically buckets of assorted of data and helper functions. The normal flow is to create an instance of each, populate them with flags from the cli (or default values). Then pass it to a `client/utils` function CompleteAndBroadcastTxCLI which calls the methods to create, sign, and submit the tx.

``` go
    type CLIContext struct {
        Codec         *codec.Codec
        AccDecoder    auth.AccountDecoder
        Client        rpcclient.Client
        Keybase       cryptokeys.Keybase
        Output        io.Writer
        OutputFormat  string
        Height        int64
        NodeURI       string
        From          string
        AccountStore  string
        TrustNode     bool
        UseLedger     bool
        Async         bool
        PrintResponse bool
        Verifier      tmlite.Verifier
        VerifierHome  string
        Simulate      bool
        GenerateOnly  bool
        FromAddress   sdk.AccAddress
        FromName      string
        Indent        bool
    }
```

``` go
    type TxBuilder struct {
        txEncoder          sdk.TxEncoder
        keybase            crkeys.Keybase
        accountNumber      uint64
        sequence           uint64
        gas                uint64
        gasAdjustment      float64
        simulateAndExecute bool
        chainID            string
        memo               string
        fees               sdk.Coins
        gasPrices          sdk.DecCoins
    }
```

```go
    func CompleteAndBroadcastTxCLI(txBldr authtxb.TxBuilder, cliCtx context.CLIContext, msgs []sdk.Msg) error
```