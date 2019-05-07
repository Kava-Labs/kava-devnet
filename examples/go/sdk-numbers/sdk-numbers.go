package main

import (
	"fmt"
	"log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {
	// Example showing going from a decimal string to a decimal
	// Note, max precision is 18!
	dec, err := sdk.NewDecFromStr("2.245050505050505505")
	if err != nil {
		log.Fatal(err)
	}

	_, ok := sdk.NewIntFromString("3")
	if !ok {
		log.Fatal("Bad Int ")
	}

	fmt.Println(dec.String())
	fmt.Println(sdk.NewInt(19))
	fmt.Println(sdk.MustNewDecFromStr("0.33"))

	p1 := sdk.MustNewDecFromStr("0.35")
	p2 := sdk.MustNewDecFromStr("0.36")
	sum := p1.Add(p2)
	divsor := sdk.MustNewDecFromStr("2")
	fmt.Println(sum)
	result := sum.Quo(divsor)
	fmt.Println(result)

	e1 := sdk.NewInt(101)
	e2 := sdk.NewInt(2)
	ex := e1.Add(e2).Quo(sdk.NewInt(2))
	fmt.Println(ex)

}

// [1, 2, 3, 4] -> 1, 2

// [1, 2, 3, 4, 5, 6] -> 2, 3