// Command chargen generates a Traveller5 character and renders it as
// markdown. Not yet implemented: character generation logic (careers,
// skills, dice resolution) doesn't exist yet in the character package.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "chargen: not yet implemented")
	os.Exit(1)
}
