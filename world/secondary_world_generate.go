package world

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// capPopulation floors rolled down to maxPopulation if it exceeds it —
// Book 3 p.29: "Subject to: Max Pop= MW Pop - 1".
func capPopulation(rolled, maxPopulation ehex.Value) ehex.Value {
	if rolled > maxPopulation {
		return maxPopulation
	}

	return rolled
}

// dmPopulation rolls Population via the standard 2D6-2 formula plus dm,
// then caps at max. The mainworld's own RollPopulation has a reroll-on-10
// exception (extending into a very-populous band) that this doesn't
// reproduce — a secondary-world population is far less consequential
// than the mainworld's, and that reroll rule reads as mainworld-specific
// flavor in the source, not a general Population-rolling rule.
func dmPopulation(r *dice.Roller, dm int, maxPopulation ehex.Value) ehex.Value {
	return capPopulation(ClampEhex(r.TwoD6()-2+dm, 0, int(ehex.Max)), maxPopulation)
}

// secondaryPopulation rolls Population for a secondary world: the
// standard RollPopulation (including its reroll-on-10 exception) when
// dm==0, since that matches a category formula with no DM at all
// ("roll normally"); dmPopulation (no reroll exception — see its own doc
// comment) otherwise.
func secondaryPopulation(r *dice.Roller, dm int, maxPopulation ehex.Value) ehex.Value {
	if dm == 0 {
		return capPopulation(RollPopulation(r), maxPopulation)
	}

	return dmPopulation(r, dm, maxPopulation)
}

// dmAtmosphere rolls Atmosphere via the standard RollAtmosphere formula
// plus an additional dm, clamped to 0..15 (F). dm=0 reproduces
// RollAtmosphere's own result exactly.
func dmAtmosphere(r *dice.Roller, size ehex.Value, dm int) ehex.Value {
	return ClampEhex(int(RollAtmosphere(r, size))+dm, 0, 15)
}

// dmHydrographics rolls Hydrographics via the standard RollHydrographics
// formula plus an additional dm, clamped to 0..10 (A). dm=0 reproduces
// RollHydrographics's own result exactly.
func dmHydrographics(r *dice.Roller, size, atm ehex.Value, dm int) ehex.Value {
	return ClampEhex(int(RollHydrographics(r, size, atm))+dm, 0, 10)
}

func rollBigWorldSize(r *dice.Roller) ehex.Value   { return ClampEhex(r.TwoD6()+7, 0, int(ehex.Max)) }
func rollWorldletSize(r *dice.Roller) ehex.Value   { return ClampEhex(r.D6()-3, 0, int(ehex.Max)) }
func rollStormWorldSize(r *dice.Roller) ehex.Value { return ClampEhex(r.TwoD6(), 0, int(ehex.Max)) }

// rollSecondaryUWPBody rolls the StSAHPGL-T skeleton every "normal"
// secondary-world category shares (Book 3 p.29): Starport through
// TechLevel, with Size from sizeRoll and Atmosphere/Hydrographics/
// Population each getting an additional dm on top of their standard roll
// (0 for "no override," matching a category's own formula exactly when
// it says nothing more than "roll normally"). Shared by every category
// except RadWorld/Inferno/Planetoids, whose formulas fix enough fields
// (Government/Law/TechLevel, or the whole UWP) that this skeleton
// doesn't fit them.
func rollSecondaryUWPBody(
	r *dice.Roller,
	sizeRoll func(*dice.Roller) ehex.Value,
	atmDM, hydDM, popDM int,
	maxPopulation ehex.Value,
) UWP {
	var u UWP

	u.Starport = RollStarport(r)
	u.Size = sizeRoll(r)
	u.Atmosphere = dmAtmosphere(r, u.Size, atmDM)
	u.Hydrographics = dmHydrographics(r, u.Size, u.Atmosphere, hydDM)
	u.Population = secondaryPopulation(r, popDM, maxPopulation)
	u.Government = RollGovernment(r, u.Population)
	u.Law = RollLaw(r, u.Government)
	u.TechLevel = RollTechLevel(r, u)

	return u
}

// generateHospitableWorld rolls a full UWP normally (Book 3 p.29:
// Hospitable= StSAHPGL-T).
func generateHospitableWorld(r *dice.Roller, maxPopulation ehex.Value) UWP {
	return rollSecondaryUWPBody(r, rollSize, 0, 0, 0, maxPopulation)
}

// generateBigWorld overrides Size = 2D+7 (Book 3 p.29: BigWorld=
// StSAHPGL-T Siz= 2D+7 — "any with Siz=B+ is BW"), rolling everything
// else from that Size normally.
func generateBigWorld(r *dice.Roller, maxPopulation ehex.Value) UWP {
	return rollSecondaryUWPBody(r, rollBigWorldSize, 0, 0, 0, maxPopulation)
}

// generateRadWorld: Book 3 p.29's RadWorld= StSAH000-0 Siz=2D —
// Population/Government/Law/TechLevel fixed at 0 (UWP's own zero value).
func generateRadWorld(r *dice.Roller) UWP {
	var u UWP

	u.Starport = RollStarport(r)
	u.Size = ClampEhex(r.TwoD6(), 0, int(ehex.Max))
	u.Atmosphere = RollAtmosphere(r, u.Size)
	u.Hydrographics = RollHydrographics(r, u.Size, u.Atmosphere)

	return u
}

// generateInferno: Book 3 p.29 transcribes Inferno= "YSB0000-0, Siz=6+1D".
// Size is unambiguous; the exact intended Atmosphere/Hydrographics values
// beyond "fixed at 0" (matching the "0000" run) couldn't be resolved with
// full confidence from this project's source transcription — possibly an
// OCR ambiguity around the Starport letter, transcribed "Y" rather than
// this project's usual StarportNone representation. Read here as
// Starport=StarportNone (T5's "no starport" case) and every other field
// fixed at 0 — a defensible, simple reading, not a claim of certainty.
func generateInferno(r *dice.Roller) UWP {
	return UWP{
		Starport: StarportNone,
		Size:     ClampEhex(6+r.D6(), 0, int(ehex.Max)),
	}
}

// generateWorldlet: Book 3 p.29's Worldlet= StSAHPGL-T Siz=1D-3 (floored
// at 0), rest normal.
func generateWorldlet(r *dice.Roller, maxPopulation ehex.Value) UWP {
	return rollSecondaryUWPBody(r, rollWorldletSize, 0, 0, 0, maxPopulation)
}

// generateIceworld: Book 3 p.29's Iceworld= StSAHPGL-T Pop=DM-6 — normal
// UWP, Population via secondaryPopulation instead of the standard roll.
func generateIceworld(r *dice.Roller, maxPopulation ehex.Value) UWP {
	return rollSecondaryUWPBody(r, rollSize, 0, 0, -6, maxPopulation)
}

// generateInnerWorld: Book 3 p.29's Inner World= StSAHPGL-T Pop=DM-4
// Hyd=DM-4.
func generateInnerWorld(r *dice.Roller, maxPopulation ehex.Value) UWP {
	return rollSecondaryUWPBody(r, rollSize, 0, -4, -4, maxPopulation)
}

// generateStormWorld: Book 3 p.29's Stormworld= StSAHPGL-T Siz=2D
// Atm=DM+4 Hyd=DM-4 Pop=DM-6.
func generateStormWorld(r *dice.Roller, maxPopulation ehex.Value) UWP {
	return rollSecondaryUWPBody(r, rollStormWorldSize, 4, -4, -6, maxPopulation)
}

// generatePlanetoidWorld: Book 3 p.29's Planetoids= St000PGL-T —
// Size/Atmosphere/Hydrographics fixed at 0, matching AsteroidBelt's own
// trigger condition in tradeCodeTriggers, so DeriveTradeCodes naturally
// tags a placed belt world AsteroidBelt.
func generatePlanetoidWorld(r *dice.Roller, maxPopulation ehex.Value) UWP {
	var u UWP

	u.Starport = RollStarport(r)
	u.Population = capPopulation(RollPopulation(r), maxPopulation)
	u.Government = RollGovernment(r, u.Population)
	u.Law = RollLaw(r, u.Government)
	u.TechLevel = RollTechLevel(r, u)

	return u
}

// generateSecondaryWorldUWP dispatches to the matching generate*World
// function for category.
func generateSecondaryWorldUWP(r *dice.Roller, category secondaryWorldCategory, maxPopulation ehex.Value) UWP {
	switch category {
	case categoryInferno:
		return generateInferno(r)
	case categoryInnerWorld:
		return generateInnerWorld(r, maxPopulation)
	case categoryBigWorld:
		return generateBigWorld(r, maxPopulation)
	case categoryStormWorld:
		return generateStormWorld(r, maxPopulation)
	case categoryRadWorld:
		return generateRadWorld(r)
	case categoryWorldlet:
		return generateWorldlet(r, maxPopulation)
	case categoryIceworld:
		return generateIceworld(r, maxPopulation)
	default: // categoryHospitable
		return generateHospitableWorld(r, maxPopulation)
	}
}
