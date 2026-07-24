package character

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/world"
)

// homeworldSkillByTradeCode is Book 1 p.56's "Homeworld and Birthworld
// Skills" table: the one specific skill a Trade Classification grants
// (p.58: "A character receives one specified skill for each Trade
// Classification or Remark from the homeworld" — e.g. an Ag world
// "automatically receives Animals-1"). Codes the table marks "(no
// skill)" (Ab, An, Ba, Di, Fo, Lk, Mr, Ph, Px, Pz, Re) are omitted
// here — a map miss means no skill — which also correctly covers the
// three world.TradeCode values Book 1's table has no row for at all
// (Sa, Pe, Cy) without any special-casing. Ri and In are omitted too:
// each names a short "(Choose One)" list instead of a single skill —
// see oneArtChoices/theTradeChoices and homeworldSkillForTradeCode.
//
// Also deliberately omitted, even though Book 1 gives them a skill:
// GenerateHomeworldSkills' homeworld comes from world.Generate, which
// only ever produces a standalone mainworld's UWP-derived trade codes
// (world.DeriveTradeCodes plus the Travel-Zone-derived code) — never
// the orbit-relative-to-HZ codes (Cold, Farming, Frozen, Hot, Tropic,
// Tundra, TwilightZone; those need system.GenerateSystem's own orbit
// placement, out of scope for this phase — world/orbit_tradecode.go's
// climateTradeCodes produces six of the seven (Hot, Cold, Tropic, Tundra,
// Farming, Frozen); TwilightZone is a separate check its caller
// DeriveOrbitTradeCodes makes directly, not part of climateTradeCodes
// itself), the sector-context ones (SubsectorCapital, SectorCapital;
// those need sector.assignCapitals' whole-sector Importance comparison),
// Capital (permanently referee-assigned in Book 3, no dice mechanic
// anywhere in this project), or Mining (Book 3 marks it "Not MW" —
// non-mainworld only, so a homeworld, which is always a mainworld here,
// can never legitimately have it regardless of generator completeness).
// Keeping their table rows here would be dead code a homeworld can never
// actually trigger.
//
// Skill names are Book 1's own canonical names, not the p.56 table's
// occasionally-abbreviated column text — e.g. "Hostile Env" there is
// "Hostile Environment" spelled out ("Hostile Environ (Hostile
// Environment) is skill in function...", printed p.85), the same class
// of abbreviation "JOT" turned out to be for "Jack of all Trades".
// "Driver" and "Fighter", despite reading like abbreviations too, are
// verified as genuine top-level T5 skill names in their own right —
// Book 1's "THE KNOWLEDGES-ONLY SKILLS" section (printed p.62) lists
// both directly alongside Animals/Engineer/Flyer/Gunner/Heavy Weapons/
// Pilot/Seafarer, each with its own sub-Knowledges (Driver: ACV,
// Automotive, Grav, ...; Fighter: Battle Dress, Beams, Blades, ...).
var homeworldSkillByTradeCode = map[world.TradeCode]string{
	world.Agricultural:    "Animals",
	world.AsteroidBelt:    "Zero-G",
	world.Dangerous:       "Fighter",
	world.Desert:          "Survival",
	world.Fluid:           "Hostile Environment",
	world.Garden:          "Trader",
	world.Hellworld:       "Hostile Environment",
	world.HighPopulation:  "Streetwise",
	world.IceCapped:       "Vacc Suit",
	world.LowPopulation:   "Flyer",
	world.NonAgricultural: "Survey",
	world.NonIndustrial:   "Driver",
	world.Ocean:           "Hi-G",
	world.PreAgricultural: "Trader",
	world.PreIndustrial:   "Jack of all Trades",
	world.Poor:            "Steward",
	world.PreRich:         "Craftsman",
	world.Vacuum:          "Vacc Suit",
	world.WaterWorld:      "Seafarer",
}

// oneArtChoices is Book 1 p.56's "One Art (Choose One)" list, granted by
// a Rich (Ri) homeworld.
var oneArtChoices = []string{"Actor", "Artist", "Author", "Chef", "Dancer", "Musician"}

// theTradeChoices is Book 1 p.56's "The Trades (Choose One)" list,
// granted by an Industrial (In) homeworld.
var theTradeChoices = []string{
	"Biologics", "Craftsman", "Electronics", "Fluidics", "Gravitics",
	"Magnetics", "Mechanic", "Photonics", "Polymers", "Programmer",
}

// rollChoice uniformly picks one of choices. The book leaves Ri/In's own
// specific Art/Trade to the player, with no dice mechanic given — this
// is this generator's own resolution for running non-interactively, not
// a book-specified table. dice.Roller.Uniform(n) returns 1..n, hence the
// -1 to index into choices.
func rollChoice(r *dice.Roller, choices []string) string {
	return choices[r.Uniform(len(choices))-1]
}

// homeworldSkillForTradeCode returns the one skill tc grants (Book 1
// p.56), and whether it grants one at all — false for a "(no skill)"
// code. Ri/In resolve their own "(Choose One)" list via rollChoice.
func homeworldSkillForTradeCode(r *dice.Roller, tc world.TradeCode) (SkillLevel, bool) {
	switch tc { //nolint:exhaustive // only Ri/In need special handling; everything else falls through to the map
	case world.Rich:
		return SkillLevel{Name: rollChoice(r, oneArtChoices), Level: 1, Kind: Skill}, true
	case world.Industrial:
		return SkillLevel{Name: rollChoice(r, theTradeChoices), Level: 1, Kind: Skill}, true
	}

	name, ok := homeworldSkillByTradeCode[tc]
	if !ok {
		return SkillLevel{}, false
	}

	return SkillLevel{Name: name, Level: 1, Kind: Skill}, true
}

// rollDeepSpaceBonus rolls Book 1 p.58's "Born In Deep Space" check
// ("A very few characters are born offworld (roll 2 on 2D)... naturally
// learns the skills Zero-G and Vacc Suit") and returns the two bonus
// skills if it fires, or nil otherwise. Independent of, and stacking
// with, whatever the homeworld's own Trade Codes already granted — a
// deep-space-born character with an Asteroid Belt (As, already Zero-G)
// or Vacuum (Va, already Vacc Suit) homeworld ends up with two entries
// of the same skill, not one deduplicated.
func rollDeepSpaceBonus(r *dice.Roller) []SkillLevel {
	if r.TwoD6() != 2 {
		return nil
	}

	return []SkillLevel{
		{Name: "Zero-G", Level: 1, Kind: Skill},
		{Name: "Vacc Suit", Level: 1, Kind: Skill},
	}
}

// GenerateHomeworldSkills rolls a random homeworld and its background
// skills. Book 1's own random-homeworld option (the "Select a Homeworld
// (Spinward Marches)" chart, p.56) is 36 real, campaign-specific worlds
// with pre-baked UWPs — explicitly flagged by the book itself as
// campaign flavor referees should replace with their own equivalent,
// not a setting-agnostic generation procedure. world.Generate is this
// project's own in-spirit substitute: a real, rulebook-verified random
// world with genuine Trade Classifications to derive skills from,
// without baking in Spinward Marches-specific content. homeworld is the
// generated world's UWP string (world.Generate never sets a Name).
//
// Character.Birthworld and Character.Homeworld are the same world here
// — p.58 allows them to differ only via an optional player-choice
// reroll ("if the player is dissatisfied... may decide the character
// changed worlds as a child") with no dice mechanic given, so it's
// skipped, matching this project's existing precedent of not
// synthesizing dice for pure player-preference choices (e.g.
// Nobility/Allegiance in world generation).
func GenerateHomeworldSkills(r *dice.Roller) (string, []SkillLevel) {
	hw := world.Generate(r)

	var skills []SkillLevel

	for _, tc := range hw.TradeCodes {
		if skill, ok := homeworldSkillForTradeCode(r, tc); ok {
			skills = append(skills, skill)
		}
	}

	skills = append(skills, rollDeepSpaceBonus(r)...)

	return hw.UWP.String(), skills
}
