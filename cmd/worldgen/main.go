// Command worldgen generates a Traveller5 world or star system and renders
// it as markdown. Not yet implemented: world/system generation logic (UWP
// rolls, trade codes, orbit placement) doesn't exist yet in the world
// package.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "worldgen: not yet implemented")
	os.Exit(1)
}
