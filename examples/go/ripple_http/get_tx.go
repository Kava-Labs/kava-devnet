package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func main() {
	// struct
	type XrpTx struct {
		Result      string `json:"result"`
		Transaction struct {
			Hash        string    `json:"hash"`
			LedgerIndex int       `json:"ledger_index"`
			Date        time.Time `json:"date"`
			Tx          struct {
				TransactionType    string `json:"TransactionType"`
				Flags              int64  `json:"Flags"`
				Sequence           int    `json:"Sequence"`
				LastLedgerSequence int    `json:"LastLedgerSequence"`
				Amount             string `json:"Amount"`
				Fee                string `json:"Fee"`
				SigningPubKey      string `json:"SigningPubKey"`
				TxnSignature       string `json:"TxnSignature"`
				Account            string `json:"Account"`
				Destination        string `json:"Destination"`
				Memos              []struct {
					Memo struct {
						MemoData string `json:"MemoData"`
					} `json:"Memo"`
				} `json:"Memos"`
			} `json:"tx"`
			Meta struct {
				TransactionIndex int `json:"TransactionIndex"`
				AffectedNodes    []struct {
					ModifiedNode struct {
						LedgerEntryType   string `json:"LedgerEntryType"`
						PreviousTxnLgrSeq int    `json:"PreviousTxnLgrSeq"`
						PreviousTxnID     string `json:"PreviousTxnID"`
						LedgerIndex       string `json:"LedgerIndex"`
						PreviousFields    struct {
							Sequence int    `json:"Sequence"`
							Balance  string `json:"Balance"`
						} `json:"PreviousFields"`
						FinalFields struct {
							Flags      int    `json:"Flags"`
							Sequence   int    `json:"Sequence"`
							OwnerCount int    `json:"OwnerCount"`
							Balance    string `json:"Balance"`
							Account    string `json:"Account"`
						} `json:"FinalFields"`
					} `json:"ModifiedNode"`
				} `json:"AffectedNodes"`
				TransactionResult string `json:"TransactionResult"`
				DeliveredAmount   string `json:"delivered_amount"`
			} `json:"meta"`
		} `json:"transaction"`
	}
	resp, err := http.Get("https://testnet.data.api.ripple.com/v2/transactions/F98899")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	jsonResult := XrpTx{}
	err = json.Unmarshal(body, &jsonResult)
	if err != nil {
		log.Fatal(string(body))
	}
	if jsonResult.Result == "error" {
		msg := fmt.Sprintf("Request returned error: %s", body)
		log.Fatal(msg)
	}
	fmt.Println(jsonResult.Transaction.Hash, jsonResult.Transaction.Tx.Memos[0].Memo.MemoData)
}
