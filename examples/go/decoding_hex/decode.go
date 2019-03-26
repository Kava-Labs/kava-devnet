package main

import (
	"encoding/hex"
	"fmt"
	"log"
)

func main() {
	const s = "7573647861646472727347504E6B534C74333642444C4D675041594B69664676437068514A5A32714A77"
	decoded, err := hex.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", decoded)

}
