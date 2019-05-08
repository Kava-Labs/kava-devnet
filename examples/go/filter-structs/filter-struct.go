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
	test := func(s PostedPrice, n string) bool { return s.Oracle == n }
	alicesPrice := GetLatestPriceForOracle(prices, test, "alice")
	fmt.Println(alicesPrice)
	charlesPrice := GetLatestPriceForOracle(prices, test, "charles")
	fmt.Println(charlesPrice)
}

func GetLatestPriceForOracle(vs []PostedPrice, f func(PostedPrice, string) bool, n string) []PostedPrice {
	vsf := make([]PostedPrice, 0)
	for _, v := range vs {
		if f(v, n) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}
