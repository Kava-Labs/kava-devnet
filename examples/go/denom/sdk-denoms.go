package main

import (
	"fmt"
	"regexp"
)

func main() {
	var (
		// Denominations can be 3 ~ 16 characters long.
		reDnmString = `[a-z][a-z0-9]{2,15}`
		reAmt       = `[[:digit:]]+`
		reDecAmt    = `[[:digit:]]*\.[[:digit:]]+`
		reSpc       = `[[:space:]]*`
		// regex matching a denomination (lower case, 3-16 characters)
		reDnm = regexp.MustCompile(fmt.Sprintf(`^%s$`, reDnmString))
		// regex matching an integer amount followed by an optional space followed by a denom
		reCoin = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reAmt, reSpc, reDnmString))
		// regex matching a decimal amount followed by a space followed by a denom
		reDecCoin = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, reDecAmt, reSpc, reDnmString))
	)

	fmt.Printf("%v\n", reDnm.MatchString("bt"))
	fmt.Printf("%v\n", reDnm.MatchString("btcdeftgldncdebg"))
	fmt.Printf("%v\n", reDnm.MatchString("btcdeftgldncdebgd"))
	fmt.Printf("%v\n", reCoin.MatchString("100 btc"))
	fmt.Printf("%v\n", reCoin.MatchString("100btc"))
	fmt.Printf("%v\n", reCoin.MatchString("100btc."))
	fmt.Printf("%v\n", reDecCoin.MatchString("100.1 btc"))

}
