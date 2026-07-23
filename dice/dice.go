// Package dice implements Traveller5's core randomizers: the six-sided die,
// 2D6, and Flux (1D6-1D6). Used throughout generation — worlds, characters,
// ships — wherever the rules call for a roll.
package dice

import (
	"math/rand/v2"
	"time"
)

// Roller rolls dice from an injectable random source, so generation stays
// deterministic and testable given a seeded source. Never use a
// package-level global generator — always go through a Roller built from
// an explicit source.
type Roller struct {
	rng *rand.Rand
}

// New returns a Roller backed by src.
func New(src rand.Source) *Roller {
	return &Roller{rng: rand.New(src)}
}

// D6 rolls a single six-sided die: 1-6.
func (r *Roller) D6() int {
	return r.rng.IntN(6) + 1
}

// TwoD6 rolls two six-sided dice and sums them: 2-12.
func (r *Roller) TwoD6() int {
	return r.D6() + r.D6()
}

// Flux rolls T5's Flux: one D6 minus another D6, range -5..+5.
func (r *Roller) Flux() int {
	a, b := r.D6(), r.D6()

	return a - b
}

// Uniform rolls a uniform ("even distribution") random integer in [1,n],
// unlike D6/TwoD6/Flux, which are bell-curve die-sum rolls. T5 uses this
// where a value should feel arbitrary rather than weighted toward the
// middle — e.g. a world's PBG population digit.
func (r *Roller) Uniform(n int) int {
	return r.rng.IntN(n) + 1
}

// ResolveSeed returns *seed if seed is non-nil, or a time-derived seed if
// seed is nil (absent). A pointer, not a plain int64, is the point: it lets
// a caller represent "no seed was requested" distinctly from "the caller
// explicitly wants seed 0" — collapsing those (as an earlier version of
// this function did) means an API consumer can never actually get seed 0.
func ResolveSeed(seed *int64) int64 {
	if seed == nil {
		return time.Now().UnixNano()
	}

	return *seed
}

// RollerFromSeed builds a Roller from an already-resolved seed (see
// ResolveSeed).
func RollerFromSeed(seed int64) *Roller {
	//nolint:gosec // any int64 bit pattern, including negative, is a valid PRNG seed
	return New(rand.NewPCG(uint64(seed), uint64(seed)))
}
