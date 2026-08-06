// Command shipgen generates a Traveller5 starship design and renders it as
// markdown. Not yet implemented: ship design sequence logic (hull/drive
// sizing, tonnage budgets) doesn't exist yet in the starship package (#6).
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "shipgen: not yet implemented")
	os.Exit(1)
}
