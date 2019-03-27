package main

import (
	"encoding/hex"
	"fmt"
	"log"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {
	// Example showing going bech32 address, to hex, and confirming they are equal
	// note that we have to set the prefix first.
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("usdx", "usdx"+"pub")
	config.Seal()
	accAddress1, err := sdk.AccAddressFromBech32("usdx1suktcgynw2dgqnpjcfdp6ay6j32f7j5emhvemd")
	if err != nil {
		log.Fatal(err)
	}
	accAddress1RawBytes := accAddress1.Bytes()
	accAddress1HexEncoded := hex.EncodeToString(accAddress1RawBytes)
	accAddress1FromHex, _ := sdk.AccAddressFromHex(accAddress1HexEncoded)
	accAddressesAreEqual := accAddress1.Equals(accAddress1FromHex)
	fmt.Printf("%t\n", accAddressesAreEqual)
	fmt.Printf("%s\n", accAddress1.String())
	fmt.Printf("%s\n", accAddress1FromHex.String())

}
