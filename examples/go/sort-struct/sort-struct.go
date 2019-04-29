package main

import (
	"fmt"
	"sort"
)

type PostedPrice struct {
	Oracle    string
	AssetCode string
	Price     int
	Expiry    int
}

func main() {

	price1 := 10
	price2 := 34
	expiry := 100

	var prices = []PostedPrice{
		PostedPrice{
			Oracle:    "alice",
			AssetCode: "xrp",
			Price:     price2,
			Expiry:    expiry,
		},
		PostedPrice{
			Oracle:    "bob",
			AssetCode: "xrp",
			Price:     price1,
			Expiry:    expiry,
		},
	}
	fmt.Println(prices)
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Price < prices[j].Price
	})

	fmt.Println(prices)
}
