package character

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
)

// CitizenCareerName is Citizen's own Career.Name value — exported and
// shared (ResolveCitizenCareer's own Career{Name: CitizenCareerName}
// literal below, render.Character's career-outcome dispatch) as a single
// source of truth, rather than two independent "Citizen" string literals
// that could silently drift apart on a future rename.
const CitizenCareerName = "Citizen"

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
// boring, unfulfilling life," not injury.
func resolveCitizenLife(r *dice.Roller, cc ehex.Value) (bool, int) {
	roll := r.TwoD6()

	return succeedsAgainst(roll, int(cc), 0), roll
}

// citizenTableE is Book 1 p.78's "E CITIZEN SKILLS AND KNOWLEDGES" table:
// a 3-key lookup (A: 1-3 — rerolled if a D6 shows 4-6; B: 1D6 rows 1-6;
// C: 1D6 columns 1-6, "top row C"), transcribed directly from the page
// image. [3][6][6]string indexed [A-1][B-1][C-1]. "JOT"->"Jack of all
// Trades", "Fwd Obs"->"Forward Observer", "Bay Wpns"->"Bay Weapons",
// "Heavy Wpns"->"Heavy Weapons", "BattleDress"->"Battle Dress" (an OCR
// spacing artifact), "Hostile Env"->"Hostile Environment", and "Power
// System"->"Power Systems" (the plural is the canonical form used
// elsewhere in the text; this table's own singular is the outlier) are
// all spelled out to their canonical names, each individually confirmed
// against the reference text's own usage elsewhere. "ACV", "WMD", and
// "LTA" are kept as printed — confirmed via direct grep to be the book's
// own standing short-form names used throughout (e.g. "ACV. Air Cushion
// Vehicle. Hovercraft."), the same class as "Vacc Suit" (itself short
// for "Vacuum Suit" but the book's own standard name everywhere), not
// column-width shorthand for a longer phrase that appears in actual
// play.
var citizenTableE = [3][6][6]string{
	{
		{"ACV", "Comms", "High-G", "Steward", "Ordnance", "Naval Arch"},
		{"Jack of all Trades", "Rider", "Sensors", "Forward Observer", "Survival", "Streetwise"},
		{"LTA", "Spines", "Flapper", "Seafarer", "No Skill", "Astrogator"},
		{"WMD", "Leader", "Tracked", "Engineer", "Computer", "Navigation"},
		{"Chef", "Survey", "Animals", "Fluidics", "Bay Weapons", "Explosives"},
		{"Mole", "Dancer", "Tactics", "Launcher", "Magnetics", "Jump Drive"},
	},
	{
		{"Grav", "Artist", "Turrets", "Teamster", "Photonics", "Counsellor"},
		{"Boat", "Legged", "Teacher", "Designer", "Vacc Suit", "Submersible"},
		{"Ship", "Sapper", "Unarmed", "Engineer", "Artillery", "Aeronautics"},
		{"Wing", "Driver", "Exotics", "Language", "Craftsman", "Aquanautics"},
		{"Recon", "Gunner", "Stealth", "Musician", "Gravitics", "Battle Dress"},
		{"Actor", "Blades", "Trainer", "Strategy", "Forensics", "Electronics"},
	},
	{
		{"Flyer", "Zero-G", "Animals", "Maneuver", "Biologics", "Hostile Environment"},
		{"Pilot", "Author", "Liaison", "Polymers", "Ortillery", "Power Systems"},
		{"Rotor", "Broker", "Athlete", "Advocate", "Automotive", "Life Support"},
		{"Admin", "Trader", "Fighter", "Computer", "Bureaucrat", "Slug Thrower"},
		{"Beams", "Sprays", "Wheeled", "Diplomat", "Heavy Weapons", "Fleet Tactics"},
		{"Medic", "Gambler", "Screens", "Mechanic", "Programmer", "Spacecraft"},
	},
}

// rollCitizenTableEName rolls Book 1 p.78's own described procedure —
// "Roll A (reroll if >3), then roll B, and finally top row C" — as a
// literal reroll loop, not a Uniform(3) pick, matching the book's own
// wording. Returns the name and true, or false for the one unresolvable
// entry, "No Skill" (A=1, B=3, C=5) — the same "lost benefit" treatment
// already established for Table C's Major/Minor/One Science entries.
func rollCitizenTableEName(r *dice.Roller) (string, bool) {
	var a int

	for {
		a = r.D6()
		if a <= 3 {
			break
		}
	}

	name := citizenTableE[a-1][r.D6()-1][r.D6()-1]
	if name == "No Skill" {
		return "", false
	}

	return name, true
}

// citizenLifeGrantIsJob reports whether priorSuccesses (Citizen Life
// successes before the current one) means the current success grants
// Job (true, odd-numbered successes: 1st, 3rd, 5th...) or Hobby (false,
// even-numbered: 2nd, 4th...) — Book 1 p.78's own "Citizen Life Roll"
// alternation (First=Job, Second=Hobby, Third=Job, ...).
func citizenLifeGrantIsJob(priorSuccesses int) bool {
	return priorSuccesses%2 == 0
}

// citizenGrantLevel returns the level a Citizen Life success grants,
// keyed strictly to its ordinal position among ALL Citizen Life
// successes so far (Book 1 p.78's own "Citizen Life Roll" table:
// First=Job-4, Second=Hobby-2, every later success=Level 1) — not to
// whether a name has already been determined. This matters when an
// earlier same-type roll landed on "No Skill" (see citizenLifeSkillGrant's
// own doc comment): a later determining roll still grants only the level
// its own ordinal calls for, per RAW's table, even though it's the first
// time a name is actually being recorded.
func citizenGrantLevel(priorSuccesses int) int {
	switch priorSuccesses {
	case 0:
		return 4
	case 1:
		return 2
	default:
		return 1
	}
}

// citizenLifeGrantOne resolves a single Job-or-Hobby grant: reuse
// persistedName at level if already determined, else roll Table E for a
// new name at level, returned as the newly-persisted name — "Once
// determined, Job and Hobby cannot be changed" (p.78). On "No Skill,"
// grants nothing and leaves persistedName unchanged (still undetermined)
// for a later success of the same type to try again — RAW doesn't address
// a first-determining roll landing on "No Skill"; this is this project's
// own interpretive call, not a silent assumption. Shared by both Job and
// Hobby via citizenLifeSkillGrant, since the two are otherwise identical
// logic over different persisted fields.
func citizenLifeGrantOne(r *dice.Roller, persistedName string, level int) (SkillLevel, bool, string) {
	if persistedName != "" {
		return SkillLevel{Name: persistedName, Level: level, Kind: Skill}, true, persistedName
	}

	name, ok := rollCitizenTableEName(r)
	if !ok {
		return SkillLevel{}, false, persistedName
	}

	return SkillLevel{Name: name, Level: level, Kind: Skill}, true, name
}

// citizenLifeSkillGrant resolves what one Citizen Life success grants,
// given priorSuccesses and the already-persisted job/hobby names (empty
// if not yet determined) — see citizenGrantLevel and citizenLifeGrantOne
// for the level and reuse-or-roll logic respectively.
//
// A later same-type grant is recorded as its own separate SkillLevel
// entry rather than merged into the earlier one to produce an
// accumulating higher level — the same simplification already documented
// for Character.Skills in general (character_generate.go).
func citizenLifeSkillGrant(
	r *dice.Roller,
	priorSuccesses int,
	jobSkill, hobbySkill string,
) (SkillLevel, bool, string, string) {
	level := citizenGrantLevel(priorSuccesses)

	if citizenLifeGrantIsJob(priorSuccesses) {
		grant, ok, newJob := citizenLifeGrantOne(r, jobSkill, level)

		return grant, ok, newJob, hobbySkill
	}

	grant, ok, newHobby := citizenLifeGrantOne(r, hobbySkill, level)

	return grant, ok, jobSkill, newHobby
}

// ResolveCitizenTerm resolves one 4-year Citizen term (Book 1 p.78) for
// the controlling characteristic ccPos, drawn from Citizen Life's own
// C1 C2 C3 C4 set — one more than Scout's C1 C2 C3. Rotating across
// multiple terms is the multi-term loop's own job (career_loop.go's
// citizen loop). Every term grants 4 skills from Table C (p.78's own
// "Per Term 4 on Table C") unconditionally — unlike Scout's duty-dependent
// 4-or-8, this doesn't depend on CitizenLifeSucceeded at all. RiskResult
// is deliberately never touched, staying at its zero value (Unharmed):
// Citizen Life has no wound mechanic, so there's nothing to record there,
// matching ResolveScoutTerm's own precedent for Courier Duty (which also
// leaves RiskResult at Unharmed because no Risk roll happened).
//
// Additionally takes and returns the career's persisted Job/Hobby state
// — priorSuccesses (Citizen Life successes before this term, for
// citizenLifeGrantIsJob's own alternation) and jobSkill/hobbySkill (empty
// if not yet determined) — since a Citizen Life success this term may
// grant a Job or Hobby skill on top of Table C's own unconditional 4
// (p.78's "SKILL ELIGIBILITY: Per Term ... Job per Citizen Life ...
// Hobby per Citizen Life" — three separate entitlement lines, not
// alternatives).
func ResolveCitizenTerm(
	r *dice.Roller,
	upp UPP,
	ccPos Position,
	priorSuccesses int,
	jobSkill, hobbySkill string,
) (Term, string, string) {
	cc := upp.Characteristics[ccPos]
	succeeded, roll := resolveCitizenLife(r, cc)

	term := Term{
		Length:                    4,
		ControllingCharacteristic: ccPos,
		CitizenLifeSucceeded:      succeeded,
		CitizenLifeRoll:           roll,
		CitizenLifeTarget:         int(cc),
		SkillsAwarded:             rollSkillsFromTable(r, citizenTableC, 4),
	}

	if !succeeded {
		return term, jobSkill, hobbySkill
	}

	grant, ok, newJob, newHobby := citizenLifeSkillGrant(r, priorSuccesses, jobSkill, hobbySkill)
	if ok {
		term.SkillsAwarded = append(term.SkillsAwarded, grant)
		term.CitizenLifeGrant = grant.Name
		term.CitizenLifeGrantIsJob = citizenLifeGrantIsJob(priorSuccesses)
	}

	return term, newJob, newHobby
}
