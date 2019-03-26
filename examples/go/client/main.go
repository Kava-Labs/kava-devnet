package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/sacOO7/gowebsocket"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"

	// remove 'app' and it should work
	app "github.com/kava-labs/usdx/blockchain"   // If we put code in this path, then this should import as 'app' because that's the pkg declaration in the package name of 'blockchain'
	"github.com/kava-labs/usdx/blockchain/x/peg" // TODO these need fixed
)

func main() {
	// Setup Interrupt -------------------------------------------------------------
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Setup SDK stuff -------------------------------------------------------------
	cdc := app.MakeCodec()
	cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)
	txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
	if err := cliCtx.EnsureAccountExists(); err != nil {
		panic(err)
	}
	fromName := cliCtx.GetFromName()
	passphrase, err := keys.GetPassphrase(fromName)
	if err != nil {
		panic(err)
	}
	cliCtx.PrintResponse = true // might want to disable this

	// Setup Websocket -------------------------------------------------------------
	socket := gowebsocket.New("wss://s.altnet.rippletest.net:51233/")

	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Fatal("Received connect error - ", err)
	}
	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")
	}
	socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received ping - " + data)
	}
	socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received pong - " + data)
	}
	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
	}
	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		log.Println("Received message - " + message)
		handleNewXrpTx(txBldr, cliCtx, passphrase, message)
	}

	// Subscribe to new txs on the XRP multisig ------------------------------------
	socket.Connect()
	socket.SendText(`{"id": "Example watch Multisig Wallet","command": "subscribe","accounts": ["rs16hESfGChwAnK97oSdRJq4A18gcJbE7j"]}`)

	// Handle shutdown -------------------------------------------------------------
	for {
		select {
		case <-interrupt:
			log.Println("interrupt")
			socket.Close()
			return
		}
	}
}

func handleNewXrpTx(txBldr authtxb.TxBuilder, cliCtx context.CLIContext, passphrase string, message string) {
	// Parse message into struct
	parsedMessage := parseMessage(message)
	txHash := parsedMessage.Transaction.Hash

	// Check message is good
	// TODO

	// Create, sign and submit tx
	from := cliCtx.GetFromAddress()
	msg := peg.NewMsgXrpTx(txHash, from)
	err := msg.ValidateBasic()
	if err != nil {
		panic(err)
	}
	signAndSubmit(txBldr, cliCtx, passphrase, msg)
}

func parseMessage(msg string) WebSocketXrpTransactionInfo {
	txInfo := WebSocketXrpTransactionInfo{}
	json.Unmarshal([]bytes(msg), &txInfo)
	return txInfo
}

func signAndSubmit(txBldr authtxb.TxBuilder, cliCtx context.CLIContext, passphrase string, msg sdk.Msg) {
	// set sequence and account numbers
	txBldr, err := utils.PrepareTxBuilder(txBldr, cliCtx)
	if err != nil {
		panic(err)
	}

	if txBldr.SimulateAndExecute() || cliCtx.Simulate {
		txBldr, err = utils.EnrichWithGas(txBldr, cliCtx, msgs)
		if err != nil {
			panic(err)
		}

		gasEst := utils.GasEstimateResponse{GasEstimate: txBldr.Gas()} // Don't need this
		fmt.Fprintf(os.Stderr, "%s\n", gasEst.String())
	}

	// if cliCtx.Simulate { // Don't think this needs to be here
	// 	return nil
	// }

	// build and sign the transaction
	fromName := cliCtx.GetFromName()
	txBytes, err := txBldr.BuildAndSign(fromName, passphrase, msgs)
	if err != nil {
		panic(err)
	}

	// broadcast to a Tendermint node
	res, err := cliCtx.BroadcastTx(txBytes)
	cliCtx.PrintOutput(res)
}

// NOTES

// CLI
// cliContext - stuff about command line flags, txBuilder - maintain sequence numbers n stuff.
// txBytes = txBldr.BuildAndSign(fromName, passphrase, msgs)
// cliCtx.BroadcastTx(txBytes)
// REST
// rest handler functions in all the modules just return a json tx. They don't submit anything.
// looks like the lcd doesn't ever submit txs
// rest server itself creates a cliContext
// creates new txBldr using sequence num from http request
// uses txBuilder

/*
TODO
 - better error handling
 - find alternative to passing password around everywhere (look at what validators use for their private keys)
 - tidy up tx submission functions to get more general
 - handle concurrency
*/
