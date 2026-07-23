package dice

import "flag"

// SeedFlag registers a "-seed" int64 flag on the command line, parses
// os.Args[1:], and returns the resolved seed via ResolveSeed: a
// time-derived value if -seed was never explicitly passed, or the exact
// value passed (including 0) if it was. Extracted after cmd/worldgen and
// cmd/sysgen's main.go each duplicated this flag.Visit dance verbatim —
// every traveller cmd that rolls dice from a CLI seed should call this
// instead of reimplementing it.
func SeedFlag() int64 {
	seed := flag.Int64("seed", 0, "PRNG seed (omit the flag entirely to derive one from current time)")

	flag.Parse()

	var seedPtr *int64

	flag.Visit(func(f *flag.Flag) {
		if f.Name == "seed" {
			seedPtr = seed
		}
	})

	return ResolveSeed(seedPtr)
}
