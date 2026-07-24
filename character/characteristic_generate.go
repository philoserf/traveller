package character

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
	"github.com/philoserf/traveller/world"
)

// rollCharacteristic rolls a single characteristic: 2D6, recorded
// directly in eHex (Book 1 p.57: "Roll 2D for each characteristic...
// and record the result in eHex as a UPP"). No DM for Human characters.
// world.ClampEhex is the codebase's own shared int->ehex.Value
// conversion (already reused by package system for the same reason) —
// TwoD6() is provably in [2,12], well within [0, ehex.Max], but the
// clamp keeps this call site consistent with every other roll-to-ehex
// conversion in the codebase rather than a second, one-off suppression.
func rollCharacteristic(r *dice.Roller) ehex.Value {
	return world.ClampEhex(r.TwoD6(), 0, int(ehex.Max))
}

// GenerateUPP rolls a Human character's Universal Personality Profile:
// the six personal characteristics (Position's own doc comment gives
// the C1-C6 order — Strength, Dexterity, Endurance, Intelligence,
// Education, Social Standing — matching Book 1 p.57 exactly), each an
// independent 2D6 roll. Sanity and Psionics are left at their zero
// value — Book 1 p.57 states both are "created only as needed," not
// part of standard character generation.
func GenerateUPP(r *dice.Roller) UPP {
	var u UPP
	for i := range u.Characteristics {
		u.Characteristics[i] = rollCharacteristic(r)
	}

	return u
}
