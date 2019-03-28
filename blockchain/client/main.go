package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/sacOO7/gowebsocket"
	rpcclient "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptokeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"

	"github.com/kava-labs/usdx/blockchain/app"
	"github.com/kava-labs/usdx/blockchain/x/peg"
)

func main() {
	// Setup Interrupt -------------------------------------------------------------
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Create the job queue -------------------------------------------------------
	// make a channel with a capacity of 100.
	jobQueue := make(chan Job, 100)

	// Setup SDK stuff -------------------------------------------------------------
	app.SetAddressPrefixes()
	validatorKeyName := "validatorName" // not actually a validator, just a local account
	kb, err := keys.NewKeyBaseFromDir(os.ExpandEnv("$HOME/.usdxcli"))
	if err != nil {
		panic(err)
	}
	passphrase, err := keys.ReadPassphraseFromStdin(validatorKeyName)
	if err != nil {
		panic(err)
	}
	// create custom configured CliContext and txBuilder
	cdc := app.MakeCodec()
	cliCtx := newCLIContext(cdc, kb, validatorKeyName)
	txBldr := newTxBuilder(cdc, kb)

	// Setup Websocket -------------------------------------------------------------
	socket := gowebsocket.New("wss://s.altnet.rippletest.net:51233/")

	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Fatal("Received connect error - ", err)
	}
	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server")
	}
	socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received ping")
	}
	socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received pong - " + data)
	}
	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
	}
	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		log.Println("Received message - " + message)
		err := handleNewXrpTx(jobQueue, cliCtx.GetFromAddress(), message)
		if err != nil {
			log.Println(err)
		}
	}

	// Subscribe to new txs on the XRP multisig ------------------------------------
	socket.Connect()
	socket.SendText(`{"id": "Example watch Multisig Wallet","command": "subscribe","accounts": ["rs16hESfGChwAnK97oSdRJq4A18gcJbE7j"]}`)

	// Start the Job Worker --------------------------------------------------------
	jobRunner := getJobRunner(txBldr, cliCtx, passphrase)
	go worker(jobQueue, jobRunner)

	// Handle shutdown -------------------------------------------------------------
	for {
		select {
		case <-interrupt:
			log.Println("interrupt")
			socket.Close()
			log.Printf("number of remaining jobs: %d", len(jobQueue))
			return
		}
	}
}

type Job = []sdk.Msg

func handleNewXrpTx(jobQueue chan Job, validatorAddress sdk.AccAddress, message string) error {
	// Parse message into struct, extract txHash
	parsedMessage := parseMessage(message)
	txHash := parsedMessage.Transaction.Hash

	// Check message is good
	if len(txHash) != 0 {

		// Create Msg
		msg := peg.NewMsgXrpTx(txHash, validatorAddress)
		err := msg.ValidateBasic()
		if err != nil {
			return err
		}

		// Add it to the queue to be submitted to the chain
		job := []sdk.Msg{msg}
		jobQueue <- job
	}
	return nil
}

func parseMessage(msg string) WebSocketXrpTransactionInfo {
	txInfo := WebSocketXrpTransactionInfo{}
	json.Unmarshal([]byte(msg), &txInfo)
	return txInfo
}

// Simple job queue thing taken from https://www.opsdash.com/blog/job-queues-in-go.html
func worker(jobQueue <-chan Job, runJob func(Job)) {
	for job := range jobQueue {
		runJob(job)
	}
}

func getJobRunner(txBldr authtxb.TxBuilder, cliCtx context.CLIContext, passphrase string) func(Job) {
	return (func(job Job) {
		log.Printf("Submiting new tx...")
		time.Sleep(10 * time.Second) // hackityhackhack
		err := CompleteAndBroadcastTxCLI(txBldr, cliCtx, passphrase, job)
		if err != nil {
			log.Printf("...failed %s", err)
		} else {
			log.Println("...done")
		}
	})
}

func newCLIContext(cdc *codec.Codec, kb cryptokeys.Keybase, keyName string) context.CLIContext {
	nodeURI := "tcp://localhost:26657"
	rpc := rpcclient.NewHTTP(nodeURI, "/websocket")
	info, err := kb.Get(keyName)
	if err != nil {
		panic(err)
	}

	// TODO add any of the other args here?
	cliCtx := context.CLIContext{
		Codec:        cdc,
		Client:       rpc,
		Output:       os.Stdout,
		NodeURI:      nodeURI,
		AccountStore: auth.StoreKey,
		From:         keyName,
		//OutputFormat:  viper.GetString(cli.OutputFlag),
		//Height:        viper.GetInt64(client.FlagHeight),
		TrustNode:     true,
		UseLedger:     false,
		Async:         false,
		PrintResponse: false,
		//Verifier:      tmlite.Verifier{}, // nil value, don't want to verifying anything
		Simulate:     false,
		GenerateOnly: false,
		FromAddress:  info.GetAddress(),
		FromName:     keyName,
		Indent:       false,
	}
	cliCtx = cliCtx.WithAccountDecoder(cdc)
	return cliCtx
}

func newTxBuilder(cdc *codec.Codec, kb cryptokeys.Keybase) authtxb.TxBuilder {
	// These field values all come from default cli flag values

	fees, _ := sdk.ParseCoins("")
	gasPrices, _ := sdk.ParseDecCoins("")
	// can't use struct literal because fields aren't exported, have to use new function.
	txBldr := authtxb.NewTxBuilder(
		utils.GetTxEncoder(cdc), //txEncoder
		0, //accountNumber
		0, //sequence
		client.DefaultGasLimit,      //gas
		client.DefaultGasAdjustment, //gasAdjustment
		false,       //simulateAndExecute //maybe
		"usdx-test", //chainID
		"",          //memo
		fees,        //fees
		gasPrices,   //gasPrices
	)
	txBldr = txBldr.WithKeybase(kb) // NewTxBuilder doesn't allow this to be set from an arg, so has to be set afterwards
	return txBldr
}

// TODO handle error cases
// Function from cosmos-sdk/client/utils but modified to accept passphrase as arg rather than stopping to pull it from stdin
func CompleteAndBroadcastTxCLI(txBldr authtxb.TxBuilder, cliCtx context.CLIContext, passphrase string, msgs []sdk.Msg) error {
	txBldr, err := utils.PrepareTxBuilder(txBldr, cliCtx)
	if err != nil {
		return err
	}

	fromName := cliCtx.GetFromName()

	if txBldr.SimulateAndExecute() || cliCtx.Simulate {
		txBldr, err = utils.EnrichWithGas(txBldr, cliCtx, msgs)
		if err != nil {
			return err
		}

		gasEst := utils.GasEstimateResponse{GasEstimate: txBldr.Gas()}
		fmt.Fprintf(os.Stderr, "%s\n", gasEst.String())
	}

	if cliCtx.Simulate {
		return nil
	}

	// build and sign the transaction
	txBytes, err := txBldr.BuildAndSign(fromName, passphrase, msgs)
	if err != nil {
		return err
	}

	// broadcast to a Tendermint node
	res, err := cliCtx.BroadcastTx(txBytes)
	log.Printf("Tx submission response: %s", res)
	return err
}
