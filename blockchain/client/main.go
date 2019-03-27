package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/sacOO7/gowebsocket"

	//"github.com/cosmos/cosmos-sdk/client/context"
	//"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptoKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"

	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/kava-labs/usdx/blockchain/app"
	"github.com/kava-labs/usdx/blockchain/x/peg"
)

func main() {
	// Setup Interrupt -------------------------------------------------------------
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Setup SDK stuff -------------------------------------------------------------
	// TODO create queue and worker (with sdk stuff)
	// make a channel with a capacity of 100.
	jobChan := make(chan Job, 100)

	cdc := app.MakeCodec()
	// TODO
	// create Keybase from dir - cosmos-sdk/client/keys#
	kb, err := keys.NewKeyBaseFromDir("~/.usdxcli")
	// set key name
	validatorKeyName := "val"
	// get key addr
	validatorAddressInfo, _ := kb.Get(validatorKeyName)
	validatorAddress := validatorAddressInfo.GetAddress() // sdk.AccAddress
	// get key password - from stdin
	passphrase, err := keys.GetPassphrase(validatorKeyName)
	if err != nil {
		panic(err)
	}
	encoder := sdk.GetConfig().GetTxEncoder()
	// 	create / get node
	node := rpcclient.NewHTTP("tcp://localhost:26657", "/websocket")
	chainID := "something"

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
		handleNewXrpTx(jobChan, validatorAddress, passphrase, message)
	}

	// Subscribe to new txs on the XRP multisig ------------------------------------
	socket.Connect()
	socket.SendText(`{"id": "Example watch Multisig Wallet","command": "subscribe","accounts": ["rs16hESfGChwAnK97oSdRJq4A18gcJbE7j"]}`)

	// Start the Job Worker --------------------------------------------------------
	go worker(jobChan, chainID, validatorAddress, kb, validatorKeyName, passphrase, encoder, node, cdc)

	// Handle shutdown -------------------------------------------------------------
	for {
		select {
		case <-interrupt:
			log.Println("interrupt")
			socket.Close()
			// TODO shutdown worker, print remaining in queue?
			return
		}
	}
}

type Job = []sdk.Msg

func handleNewXrpTx(jobChan chan Job, validatorAddress sdk.AccAddress, passphrase string, message string) {
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
	job := []sdk.Msg{msg}
	// enqueue a job
	jobChan <- job
	//signAndSubmit(passphrase, []sdk.Msg{msg}) // TODO replace with add to queue
}

func parseMessage(msg string) WebSocketXrpTransactionInfo {
	txInfo := WebSocketXrpTransactionInfo{}
	json.Unmarshal([]byte(msg), &txInfo)
	return txInfo
}

// Simple job queue thing taken from https://www.opsdash.com/blog/job-queues-in-go.html
func worker(jobChan <-chan Job, chainID string, validatorAddress sdk.AccAddress, kb cryptoKeys.Keybase, validatorKeyName string, passphrase string, encoder sdk.TxEncoder, node rpcclient.Client, cdc *codec.Codec) {
	for job := range jobChan {
		submitTx(job, chainID, validatorAddress, kb, validatorKeyName, passphrase, encoder, node, cdc)

	}
}

func submitTx(msgs []sdk.Msg, chainID string, validatorAddress sdk.AccAddress, kb cryptoKeys.Keybase, validatorKeyName string, passphrase string, encoder sdk.TxEncoder, node rpcclient.Client, cdc *codec.Codec) error {
	// TODO replace some of below by creating a txBuilder at the top level then using it here to build and sign the txs
	// Probably want to use some of the cliContext methods to fetch account and sequence numbers

	// BuildAndSign
	// 	Create StdSignMsg - bunch of details
	account, _ := getAccount(node, cdc, validatorAddress) // fetches info from node, needs to be in worker
	stdSignMsg := authtxb.StdSignMsg{
		ChainID:       chainID,
		AccountNumber: account.GetAccountNumber(),
		Sequence:      account.GetSequence(),
		Memo:          "",
		Msgs:          msgs,
		Fee:           auth.NewStdFee(10000, sdk.Coins{sdk.NewCoin("pxrp", sdk.NewInt(1))}), //TODO set really high for now, use simulator on txBuilder in future
	}
	// 	get signBytes and create auth.StdSignature (using helper function)
	sig, err := authtxb.MakeSignature(kb, validatorKeyName, passphrase, stdSignMsg)
	// 	create auth.StdTx
	stdTx := auth.NewStdTx(stdSignMsg.Msgs, stdSignMsg.Fee, []auth.StdSignature{sig}, stdSignMsg.Memo)
	// 	encode it using txEncoder
	tx, _ := encoder(stdTx)

	// 	submit transaction
	res, err := node.BroadcastTxCommit(tx)
	if err != nil {
		panic(err)
	}
	// 	Check it was delivered correctly
	if res.CheckTx.IsOK() || res.DeliverTx.IsOK() {
		panic("tx not included in block")
	}

	return nil
}

func getAccount(node rpcclient.Client, cdc *codec.Codec, address sdk.AccAddress) (auth.Account, error) {

	bz, err := cdc.MarshalJSON(auth.NewQueryAccountParams(address)) // This is not present in v0.32
	if err != nil {
		return nil, err
	}

	route := fmt.Sprintf("custom/%s/%s", "account?", auth.QueryAccount) //TODO

	result, err := node.ABCIQuery(route, bz)
	if err != nil {
		return nil, err
	}

	resp := result.Response
	if !resp.IsOK() {
		return nil, errors.New(resp.Log)
	}

	var account auth.Account
	if err := cdc.UnmarshalJSON(resp.Value, &account); err != nil {
		return nil, err
	}

	return account, nil
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

/* How to construct tx submitter.
 - storing account sequence nums doesn't seem helpful - can never mirror blockchain 100%
 - pull them in as needed
 - Don't need to submit tx and get it confirmed before submitting next one
 - submitting
	- submit and forget
	- submit, check received into mempool cache (if possible?)
	- submit, wait until `CheckTx` is ok (ie it's in the mempool)
	- submit, wait until `DeliverTx` is ok (ie it's in a valid block, and that block has updated the sdk app state)
 - keeping things simple - just use a queue

*/

// // Simple job queue thing taken from https://www.opsdash.com/blog/job-queues-in-go.html
// func worker(jobChan <-chan Job) {
//     for job := range jobChan {
//         process(job)
//     }
// }

// // make a channel with a capacity of 100.
// jobChan := make(chan Job, 100)

// // start the worker
// go worker(jobChan)

// // enqueue a job
// jobChan <- job

/* TODO stuff to get:
- chainID
- keybase
- cdc or txEncoder
- passphrase, keyName
- account Object thing

- maybe txEncoder, node
*/
