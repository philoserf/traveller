package character

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// citizenTableC is Book 1 p.78's "C CITIZEN SKILLS" table (7 columns,
// rows keyed by 1D 1-6), transcribed directly from the page image.
// Column 1 (Personal) entries are characteristic boosts, normalized to
// this project's own Str/Dex/End/Int/Edu/Soc convention (matching
// scoutSkillTable's own column 1) rather than Book 1's own Citizen-box
// notation ("C2 +1"/"C3 +1"/"C5 +1"/"C6 +1**") — see this package's own
// chargen plan history for why. "Hostile Environ" and "JOT" are spelled
// out to their canonical names ("Hostile Environment", "Jack of all
// Trades") — the same class of abbreviation this project already caught
// as a real transcription bug in Scout's own table.
var citizenTableC = [7][6]string{
	{"Str", "Dex", "End", "Int", "Edu", "Soc"},
	{"Major", "Major", "Minor", "Minor", "One Trade", "One Trade"},
	{"Seafarer", "Navigator", "Hostile Environment", "Flyer", "Driver", "Vacc Suit"},
	{"Admin", "Broker", "Computer", "Animals", "Bureaucrat", "Trader"},
	{"Advocate", "Broker", "Trader", "Liaison", "Counsellor", "Teacher"},
	{"One Art", "One Science", "One Trade", "Driver", "Bureaucrat", "Computer"},
	{"One Art", "One Science", "Jack of all Trades", "Athlete", "Medic", "One Trade"},
}

// BeginCitizen resolves Book 1 p.78's "To Begin Auto" — a Citizen always
// qualifies, no roll needed, unlike every other career implemented so
// far (Scout's own BeginScout rolls against the best of Str/Dex/End,
// with a Retry against Edu). Kept as its own function, even though it
// always returns true, for symmetry with BeginScout and so this rule is
// visible and testable rather than silently assumed by a caller.
func BeginCitizen() bool {
	return true
}

// resolveCitizenLife resolves Book 1 p.78's Citizen Life roll for one
// term: a single 2D6 check against cc, no Mods (p.64: "Only one roll is
// made to determine Success or Failure. No Mods are used") and, unlike
// Scout's Risk & Reward, no characteristic reduction and no wound of any
// kind on failure — "the Citizen continues the term stuck in a dull,
// boring, unfulfilling life," not injury. What a success actually grants
// (a Job or Hobby skill, from Book 1's own separate 3-key "E CITIZEN
// SKILLS AND KNOWLEDGES" table, p.78) is deferred to a future slice —
// this package's own chargen plan history has the full rationale.
func resolveCitizenLife(r *dice.Roller, cc ehex.Value) bool {
	return rollAgainstTarget(r, int(cc), 0)
}

// ResolveCitizenTerm resolves one 4-year Citizen term (Book 1 p.78) for
// the controlling characteristic ccPos, drawn from Citizen Life's own
// C1 C2 C3 C4 set — one more than Scout's C1 C2 C3. Rotating across
// multiple terms is deferred to a future multi-term loop, the same split
// Scout itself took (BeginScout/ResolveScoutTerm shipped before
// ResolveScoutCareer's own rotation logic existed). Every term grants 4
// skills from Table C (p.78's own "Per Term 4 on Table C") unconditionally
// — unlike Scout's duty-dependent 4-or-8, this doesn't depend on
// CitizenLifeSucceeded at all. Job/Hobby skill grants (gated on that
// success) are deferred — see resolveCitizenLife's own doc comment.
// RiskResult is deliberately never touched, staying at its zero value
// (Unharmed): Citizen Life has no wound mechanic, so there's nothing to
// record there, matching ResolveScoutTerm's own precedent for Courier
// Duty (which also leaves RiskResult at Unharmed because no Risk roll
// happened).
func ResolveCitizenTerm(r *dice.Roller, upp UPP, ccPos Position) Term {
	cc := upp.Characteristics[ccPos]

	return Term{
		Length:                    4,
		ControllingCharacteristic: ccPos,
		CitizenLifeSucceeded:      resolveCitizenLife(r, cc),
		SkillsAwarded:             rollSkillsFromTable(r, citizenTableC, 4),
	}
}
