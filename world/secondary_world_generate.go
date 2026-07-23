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
// then caps at max. The mainworld's own rollPopulation has a reroll-on-10
// exception (extending into a very-populous band) that this doesn't
// reproduce — a secondary-world population is far less consequential
// than the mainworld's, and that reroll rule reads as mainworld-specific
// flavor in the source, not a general Population-rolling rule.
func dmPopulation(r *dice.Roller, dm int, maxPopulation ehex.Value) ehex.Value {
	return capPopulation(clampEhex(r.TwoD6()-2+dm, 0, int(ehex.Max)), maxPopulation)
}

// dmAtmosphere rolls Atmosphere via the standard rollAtmosphere formula
// plus an additional dm, clamped to 0..15 (F).
func dmAtmosphere(r *dice.Roller, size ehex.Value, dm int) ehex.Value {
	return clampEhex(int(rollAtmosphere(r, size))+dm, 0, 15)
}

// dmHydrographics rolls Hydrographics via the standard rollHydrographics
// formula plus an additional dm, clamped to 0..10 (A).
func dmHydrographics(r *dice.Roller, size, atm ehex.Value, dm int) ehex.Value {
	return clampEhex(int(rollHydrographics(r, size, atm))+dm, 0, 10)
}

// generateHospitableWorld rolls a full UWP normally (Book 3 p.29:
// Hospitable= StSAHPGL-T).
func generateHospitableWorld(r *dice.Roller, maxPopulation ehex.Value) UWP {
	var u UWP

	u.Starport = rollStarport(r)
	u.Size = rollSize(r)
	u.Atmosphere = rollAtmosphere(r, u.Size)
	u.Hydrographics = rollHydrographics(r, u.Size, u.Atmosphere)
	u.Population = capPopulation(rollPopulation(r), maxPopulation)
	u.Government = rollGovernment(r, u.Population)
	u.Law = rollLaw(r, u.Government)
	u.TechLevel = rollTechLevel(r, u)

	return u
}

// generateBigWorld overrides Size = 2D+7 (Book 3 p.29: BigWorld=
// StSAHPGL-T Siz= 2D+7 — "any with Siz=B+ is BW"), rolling everything
// else from that Size normally.
func generateBigWorld(r *dice.Roller, maxPopulation ehex.Value) UWP {
	var u UWP

	u.Starport = rollStarport(r)
	u.Size = clampEhex(r.TwoD6()+7, 0, int(ehex.Max))
	u.Atmosphere = rollAtmosphere(r, u.Size)
	u.Hydrographics = rollHydrographics(r, u.Size, u.Atmosphere)
	u.Population = capPopulation(rollPopulation(r), maxPopulation)
	u.Government = rollGovernment(r, u.Population)
	u.Law = rollLaw(r, u.Government)
	u.TechLevel = rollTechLevel(r, u)

	return u
}

// generateRadWorld: Book 3 p.29's RadWorld= StSAH000-0 Siz=2D —
// Population/Government/Law/TechLevel fixed at 0 (UWP's own zero value).
func generateRadWorld(r *dice.Roller) UWP {
	var u UWP

	u.Starport = rollStarport(r)
	u.Size = clampEhex(r.TwoD6(), 0, int(ehex.Max))
	u.Atmosphere = rollAtmosphere(r, u.Size)
	u.Hydrographics = rollHydrographics(r, u.Size, u.Atmosphere)

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
		Size:     clampEhex(6+r.D6(), 0, int(ehex.Max)),
	}
}

// generateWorldlet: Book 3 p.29's Worldlet= StSAHPGL-T Siz=1D-3 (floored
// at 0), rest normal.
func generateWorldlet(r *dice.Roller, maxPopulation ehex.Value) UWP {
	var u UWP

	u.Starport = rollStarport(r)
	u.Size = clampEhex(r.D6()-3, 0, int(ehex.Max))
	u.Atmosphere = rollAtmosphere(r, u.Size)
	u.Hydrographics = rollHydrographics(r, u.Size, u.Atmosphere)
	u.Population = capPopulation(rollPopulation(r), maxPopulation)
	u.Government = rollGovernment(r, u.Population)
	u.Law = rollLaw(r, u.Government)
	u.TechLevel = rollTechLevel(r, u)

	return u
}

// generateIceworld: Book 3 p.29's Iceworld= StSAHPGL-T Pop=DM-6 — normal
// UWP, Population via dmPopulation instead of the standard roll.
func generateIceworld(r *dice.Roller, maxPopulation ehex.Value) UWP {
	var u UWP

	u.Starport = rollStarport(r)
	u.Size = rollSize(r)
	u.Atmosphere = rollAtmosphere(r, u.Size)
	u.Hydrographics = rollHydrographics(r, u.Size, u.Atmosphere)
	u.Population = dmPopulation(r, -6, maxPopulation)
	u.Government = rollGovernment(r, u.Population)
	u.Law = rollLaw(r, u.Government)
	u.TechLevel = rollTechLevel(r, u)

	return u
}

// generateInnerWorld: Book 3 p.29's Inner World= StSAHPGL-T Pop=DM-4
// Hyd=DM-4.
func generateInnerWorld(r *dice.Roller, maxPopulation ehex.Value) UWP {
	var u UWP

	u.Starport = rollStarport(r)
	u.Size = rollSize(r)
	u.Atmosphere = rollAtmosphere(r, u.Size)
	u.Hydrographics = dmHydrographics(r, u.Size, u.Atmosphere, -4)
	u.Population = dmPopulation(r, -4, maxPopulation)
	u.Government = rollGovernment(r, u.Population)
	u.Law = rollLaw(r, u.Government)
	u.TechLevel = rollTechLevel(r, u)

	return u
}

// generateStormWorld: Book 3 p.29's Stormworld= StSAHPGL-T Siz=2D
// Atm=DM+4 Hyd=DM-4 Pop=DM-6.
func generateStormWorld(r *dice.Roller, maxPopulation ehex.Value) UWP {
	var u UWP

	u.Starport = rollStarport(r)
	u.Size = clampEhex(r.TwoD6(), 0, int(ehex.Max))
	u.Atmosphere = dmAtmosphere(r, u.Size, 4)
	u.Hydrographics = dmHydrographics(r, u.Size, u.Atmosphere, -4)
	u.Population = dmPopulation(r, -6, maxPopulation)
	u.Government = rollGovernment(r, u.Population)
	u.Law = rollLaw(r, u.Government)
	u.TechLevel = rollTechLevel(r, u)

	return u
}

// generatePlanetoidWorld: Book 3 p.29's Planetoids= St000PGL-T —
// Size/Atmosphere/Hydrographics fixed at 0, matching AsteroidBelt's own
// trigger condition in tradeCodeTriggers, so DeriveTradeCodes naturally
// tags a placed belt world AsteroidBelt.
func generatePlanetoidWorld(r *dice.Roller, maxPopulation ehex.Value) UWP {
	var u UWP

	u.Starport = rollStarport(r)
	u.Population = capPopulation(rollPopulation(r), maxPopulation)
	u.Government = rollGovernment(r, u.Population)
	u.Law = rollLaw(r, u.Government)
	u.TechLevel = rollTechLevel(r, u)

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
