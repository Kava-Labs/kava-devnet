package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"

	"github.com/sacOO7/gowebsocket"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/client/keys"

	rpcclient "github.com/tendermint/tendermint/rpc/client"

	app "github.com/kava-labs/usdx/blockchain"
	"github.com/kava-labs/usdx/blockchain/x/peg"
)

func main() {
	// Setup Interrupt -------------------------------------------------------------
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Setup SDK stuff -------------------------------------------------------------
	cdc := app.MakeCodec()
	// TODO
	// create Keybase from dir - cosmos-sdk/client/keys#
	kb, err := keys.NewKeyBaseFromDir("~/.usdxcli")
	// set key name
	validatorKeyName := "val"
	// get key addr
	validatorAddress := kb.Get(validatorKeyName).GetAddress() // sdk.AccAddress
	// get key password - from stdin
	passphrase, err := keys.GetPassphrase(validatorKeyName)
	if err != nil {
		panic(err)
	}

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
		handleNewXrpTx(validatorAddress, passphrase, message)
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

func handleNewXrpTx(validatorAddress sdk.AccAddress, passphrase string, message string) {
	// Parse message into struct
	parsedMessage := parseMessage(message)
	txHash := parsedMessage.Transaction.Hash

	// Check message is good
	// TODO

	// Create, sign and submit tx
	msg := peg.NewMsgXrpTx(txHash, validatorAddress)
	err := msg.ValidateBasic()
	if err != nil {
		panic(err)
	}
	signAndSubmit(passphrase, []sdk.Msg{msg})
}

func parseMessage(msg string) WebSocketXrpTransactionInfo {
	txInfo := WebSocketXrpTransactionInfo{}
	json.Unmarshal([]byte(msg), &txInfo)
	return txInfo
}

func signAndSubmit(passphrase string, msgs []sdk.Msg) {
	// TODO replace some of below by creating a txBuilder at the top level then using it here to build and sign the txs
	// Probably want to use some of the cliContext methods to fetch account and sequence numbers

	// BuildAndSign
	// 	Create StdSignMsg - bunch of details
	stdSignMsg := authtxb.StdSignMsg{
		ChainID:       bldr.chainID,
		AccountNumber: bldr.accountNumber,
		Sequence:      bldr.sequence,
		Memo:          bldr.memo,
		Msgs:          msgs,
		Fee:           auth.NewStdFee(bldr.gas, fees),
	}
	// 	get signBytes and create auth.StdSignature (using helper function)
	sig, err := authtxb.MakeSignature(kb, validatorKeyName, passphrase, stdSignMsg)
	// 	create auth.StdTx
	stdTx := auth.NewStdTx(msg.Msgs, msg.Fee, []auth.StdSignature{sig}, msg.Memo)
	// 	encode it using txEncoder
	txEncoder := utils.GetTxEncoder(cdc)
	tx := txEncoder(stdTx)

	// BroadcastTx
	// 	create / get node
	node := rpcclient.NewHTTP("tcp://localhost:26657", "/websocket")
	// 	submit transaction
	res, err := node.BroadcastTxCommit(tx)
	if err != nil {
		panic(err)
	}
	// 	Check it was delivered correctly
	if res.CheckTx.IsOK() || res.DeliverTx.IsOK() {
		panic("tx not included in block")
	}

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
