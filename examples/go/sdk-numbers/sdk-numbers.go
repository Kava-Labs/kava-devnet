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

	i, ok := sdk.NewIntFromString("3")
	if !ok {
		log.Fatal("Bad Int ")
	}
	fmt.Printf(dec.String())
	fmt.Println()
	fmt.Printf(i.String())

}
