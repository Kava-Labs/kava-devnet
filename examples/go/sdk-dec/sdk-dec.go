package main

import (
	"fmt"
	"sort"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SortDecs provides the interface needed to sort sdk.Dec slices
type SortDecs []sdk.Dec

func (a SortDecs) Len() int { return len(a) }
func (a SortDecs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortDecs) Less(i, j int) bool { return a[i].LT(a[j])}

func main() {
	// Example showing going from a decimal string to a decimal
	// Note, max precision is 18!
	dec, _ := sdk.NewDecFromStr("2.245050505050505505")
	fmt.Printf("%f\n", dec)

	// Example sorting sdk.Dec slices

	dec1, _ := sdk.NewDecFromStr("5.13")
	dec2, _ := sdk.NewDecFromStr("1.13")
	dec3, _ := sdk.NewDecFromStr("3.13")


	decs := []sdk.Dec{
		dec1, dec2, dec3,
	}
	fmt.Println(decs)
	sort.Sort(SortDecs(decs))
	fmt.Println(decs)

	if len(decs) != 0 {
		fmt.Println("This shit is real")
	}

}
