package main

import (
	"fmt"
)

type PostedPrice struct {
	Oracle    string
	AssetCode string
	Price     int
	Expiry    int
}

func main() {

	price1 := 33
	price2 := 34
	expiry := 100

	var prices = []PostedPrice{
		PostedPrice{
			Oracle:    "alice",
			AssetCode: "xrp",
			Price:     price1,
			Expiry:    expiry,
		},
		PostedPrice{
			Oracle:    "bob",
			AssetCode: "xrp",
			Price:     price2,
			Expiry:    expiry,
		},
	}
	var index int

	for i := range prices {
		if prices[i].Oracle == "bob" {
			index = i
			break
		}
	}

	fmt.Println(index)
}