package main

type WebSocketXrpTransactionInfo struct {
	Status              string `json:"status"`
	Type                string `json:"type"`
	EngineResult        string `json:"engine_result"`
	EngineResultCode    int    `json:"engine_result_code"`
	EngineResultMessage string `json:"engine_result_message"`
	LedgerHash          string `json:"ledger_hash"`
	LedgerIndex         int    `json:"ledger_index"`
	Meta                struct {
		AffectedNodes []struct {
			ModifiedNode struct {
				FinalFields struct {
					Flags         int    `json:"Flags"`
					IndexPrevious string `json:"IndexPrevious"`
					Owner         string `json:"Owner"`
					RootIndex     string `json:"RootIndex"`
				} `json:"FinalFields"`
				LedgerEntryType string `json:"LedgerEntryType"`
				LedgerIndex     string `json:"LedgerIndex"`
			} `json:"ModifiedNode,omitempty"`
			DeletedNode struct {
				FinalFields struct {
					Account           string `json:"Account"`
					BookDirectory     string `json:"BookDirectory"`
					BookNode          string `json:"BookNode"`
					Flags             int    `json:"Flags"`
					OwnerNode         string `json:"OwnerNode"`
					PreviousTxnID     string `json:"PreviousTxnID"`
					PreviousTxnLgrSeq int    `json:"PreviousTxnLgrSeq"`
					Sequence          int    `json:"Sequence"`
					TakerGets         string `json:"TakerGets"`
					TakerPays         struct {
						Currency string `json:"currency"`
						Issuer   string `json:"issuer"`
						Value    string `json:"value"`
					} `json:"TakerPays"`
				} `json:"FinalFields"`
				LedgerEntryType string `json:"LedgerEntryType"`
				LedgerIndex     string `json:"LedgerIndex"`
			} `json:"DeletedNode,omitempty"`
			ModifiedNode struct {
				FinalFields struct {
					Account    string `json:"Account"`
					Balance    string `json:"Balance"`
					Flags      int    `json:"Flags"`
					OwnerCount int    `json:"OwnerCount"`
					Sequence   int    `json:"Sequence"`
				} `json:"FinalFields"`
				LedgerEntryType string `json:"LedgerEntryType"`
				LedgerIndex     string `json:"LedgerIndex"`
				PreviousFields  struct {
					Balance    string `json:"Balance"`
					OwnerCount int    `json:"OwnerCount"`
					Sequence   int    `json:"Sequence"`
				} `json:"PreviousFields"`
				PreviousTxnID     string `json:"PreviousTxnID"`
				PreviousTxnLgrSeq int    `json:"PreviousTxnLgrSeq"`
			} `json:"ModifiedNode,omitempty"`
		} `json:"AffectedNodes"`
		TransactionIndex  int    `json:"TransactionIndex"`
		TransactionResult string `json:"TransactionResult"`
	} `json:"meta"`
	Transaction struct {
		Account         string `json:"Account"`
		Fee             string `json:"Fee"`
		Flags           int64  `json:"Flags"`
		OfferSequence   int    `json:"OfferSequence"`
		Sequence        int    `json:"Sequence"`
		SigningPubKey   string `json:"SigningPubKey"`
		TransactionType string `json:"TransactionType"`
		TxnSignature    string `json:"TxnSignature"`
		Date            int    `json:"date"`
		Hash            string `json:"hash"`
	} `json:"transaction"`
	Validated bool `json:"validated"`
}
