package render

import (
	"fmt"
	"strings"

	"github.com/philoserf/traveller/character"
)

// Character renders c as a Markdown character sheet: Species, Genetic
// Profile, UPP, Age, Life Stage, Birthdate, Homeworld (and Birthworld
// only when it differs), Wound Badges, the full Skills list, and each
// Career's Terms and Mustering Out benefits. Age, Life Stage, and
// Birthdate are always shown, like WoundBadges: every real generation
// path computes them now (finalizeAging and GenerateBirthdate,
// character/aging.go and character/birthdate.go), with a minimum Age of
// 18, so there's no ambiguous zero-value case to guard against the way
// Fame/Cash/Rank/Notes still have. Name, Notes, Rank, Fame, and Cash are
// shown only when set — nothing in the character package generates Name
// yet (see character_generate.go's own doc comment), Notes is empty
// unless Aging actually produced an event, and Fame/Cash are genuinely 0
// whenever ApplyMusteringOut (character/muster_out_apply.go) found no
// "Fame +N"/"CrN,NNN" entry in the raw Benefits/Money lines to
// accumulate — printing "0" there would be indistinguishable from a real
// zero result, the same reasoning already applied to WoundBadges.
// Position and RiskResult get local label functions
// (positionAbbrev, riskResultLabel) rather than String() methods on the
// character package's own types, matching this project's existing
// precedent (world.TradeCode/world.Base have no String() either — render
// uses world.TradeCodeStrings/world.BaseStrings instead); lifeStageLabel
// follows the same convention.
func Character(c character.Character) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", characterTitle(c))
	fmt.Fprintf(&b, "**Species:** %s\n\n", c.Species)
	fmt.Fprintf(&b, "**Genetic Profile:** %s\n\n", c.GeneticProfile)
	fmt.Fprintf(&b, "**UPP:** %s\n\n", c.UPP)
	fmt.Fprintf(&b, "**Age:** %d\n\n", c.Age)
	fmt.Fprintf(&b, "**Life Stage:** %s\n\n", lifeStageLabel(c.LifeStage))
	fmt.Fprintf(&b, "**Birthdate:** %s\n\n", c.Birthdate)
	fmt.Fprintf(&b, "**Homeworld:** %s\n\n", c.Homeworld)

	if c.Birthworld != "" && c.Birthworld != c.Homeworld {
		fmt.Fprintf(&b, "**Birthworld:** %s\n\n", c.Birthworld)
	}

	if c.Rank != "" {
		fmt.Fprintf(&b, "**Rank:** %s\n\n", c.Rank)
	}

	fmt.Fprintf(&b, "**Wound Badges:** %d\n\n", c.WoundBadges)

	if len(c.Medals) > 0 {
		fmt.Fprintf(&b, "**Medals:** %s\n\n", strings.Join(c.Medals, ", "))
	}

	if c.Fame != 0 {
		fmt.Fprintf(&b, "**Fame:** %d\n\n", c.Fame)
	}

	if c.Cash != 0 {
		fmt.Fprintf(&b, "**Cash:** Cr%d\n\n", c.Cash)
	}

	if c.Notes != "" {
		fmt.Fprintf(&b, "**Notes:** %s\n\n", c.Notes)
	}

	fmt.Fprint(&b, "## Skills\n\n")
	writeSkills(&b, c.Skills)

	fmt.Fprint(&b, "\n## Careers\n\n")

	for _, career := range c.Careers {
		writeCareer(&b, career)
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

// characterTitle falls back to the UPP code when Name is empty — nothing
// in the character package generates Name yet, so this is the common
// case, matching World's own Name-or-UWP fallback (render/world.go).
func characterTitle(c character.Character) string {
	if c.Name != "" {
		return c.Name
	}

	return c.UPP.String()
}

func writeSkills(b *strings.Builder, skills []character.SkillLevel) {
	if len(skills) == 0 {
		fmt.Fprint(b, "None.\n")

		return
	}

	for _, s := range skills {
		fmt.Fprintf(b, "- %s\n", skillNotation(s))
	}
}

// skillNotation matches Traveller's own notation (SkillLevel's doc
// comment): "Level 0 ('default skill') is implicit and commonly
// omitted" — e.g. "Pilot-4", but bare "Zero-G" at Level 0. A Personal
// entry (scoutSkillTable's own column-0 grants, e.g. Name:"Str",
// Level:1) is a characteristic boost, not a trained skill — rendering it
// as "Str-1" would misleadingly read as a proficiency literally named
// "Str". Rendered instead with the book's own "+1" boost notation
// (matching scoutSkillTable's and scoutMusterOutBenefits' own "Str +1"
// literal table entries), which also visually distinguishes it from a
// same-named real skill.
func skillNotation(s character.SkillLevel) string {
	if s.Kind == character.Personal {
		return fmt.Sprintf("%s +%d", s.Name, s.Level)
	}

	if s.Level == 0 {
		return s.Name
	}

	return fmt.Sprintf("%s-%d", s.Name, s.Level)
}

func writeCareer(b *strings.Builder, c character.Career) {
	fmt.Fprintf(b, "### %s\n\n", c.Name)

	if len(c.Terms) == 0 {
		fmt.Fprint(b, "Never qualified for this career.\n\n")
	} else {
		for i, t := range c.Terms {
			fmt.Fprintf(b, "- %s\n", termOutcomeLine(c, i, t))

			for _, sk := range t.SkillsAwarded {
				fmt.Fprintf(b, "  - %s\n", skillNotation(sk))
			}
		}

		fmt.Fprint(b, "\n")

		if c.JobSkill != "" {
			fmt.Fprintf(b, "**Job:** %s\n\n", c.JobSkill)
		}

		if c.HobbySkill != "" {
			fmt.Fprintf(b, "**Hobby:** %s\n\n", c.HobbySkill)
		}

		if c.Specialty != "" {
			fmt.Fprintf(b, "**Specialty:** %s\n\n", c.Specialty)
		}
	}

	writeMusteringOut(b, c.MusteringOut)
}

// termOutcomeLine renders one term's own outcome — Citizen's
// CitizenLifeSucceeded (no wound/Reward concept), Noble's own
// NobleAction/NobleSucceeded/Elevated (Return & Intrigue, no wound/
// Reward concept either), or every other career's RiskResult/
// RewardResult (Scout's own shape today). Dispatches on c.Name against
// the career's own exported CareerName constant rather than a bare
// string literal or a Career-level bool flag — Name is already the
// career's own unique identifier, and the shared exported constants
// (also used by each career's own Career{Name: ...} literal) mean the
// two call sites for a given career can't silently drift apart the way
// two independent string literals could.
func termOutcomeLine(c character.Career, i int, t character.Term) string {
	// Entertainer has no Controlling Characteristic at all (Risk & Reward
	// targets Talent, not a UPP position) — t.ControllingCharacteristic
	// is never set for its own terms, so positionAbbrev's own zero-value
	// fallback ("Str") would misleadingly suggest one. Its own prefix
	// omits the "(Xxx)" suffix entirely rather than showing a fake CC.
	if c.Name == character.EntertainerCareerName {
		return fmt.Sprintf("Term %d", i+1) + ": " + entertainerTermLabel(t)
	}

	prefix := fmt.Sprintf("Term %d (%s)", i+1, positionAbbrev(t.ControllingCharacteristic))

	switch c.Name {
	case character.CitizenCareerName:
		return prefix + ": " + citizenLifeLabel(t.CitizenLifeSucceeded)
	case character.NobleCareerName:
		return prefix + ": " + nobleTermLabel(t)
	case character.RogueCareerName:
		return prefix + ": " + rogueTermLabel(t)
	case character.ScholarCareerName:
		return prefix + ": " + scholarTermLabel(t)
	case character.FunctionaryCareerName:
		return prefix + ": " + functionaryTermLabel(t)
	}

	line := prefix + ": " + riskResultLabel(t.RiskResult)
	if t.RewardResult != "" && t.RewardResult != "None" {
		line += ", Reward: " + t.RewardResult
	}

	return line
}

// nobleTermLabel renders "Return: Success", "Intrigue: Failure", or
// "Intrigue: Success, Elevated" — Elevated is only ever true alongside a
// successful Intrigue (ResolveNobleTerm's own invariant), so it's
// appended unconditionally rather than re-checked here.
func nobleTermLabel(t character.Term) string {
	label := t.NobleAction + ": "

	if t.NobleSucceeded {
		label += "Success"
	} else {
		label += "Failure"
	}

	if t.Elevated {
		label += ", Elevated"
	}

	return label
}

// rogueTermLabel renders "Scheme: Craftsman, Success (Payoff Cr45,000)"/
// "Scheme: Noble, Imprisoned 2 years" — Rogue's own outcome has no
// Wounded/Disabled/Dead concept to reuse riskResultLabel for.
//
// Reward is rolled unconditionally, independent of Risk's own outcome
// (ResolveRogueTerm), so an Imprisoned term can still carry a real
// Payoff or Ship Share — buildRogueCharacter always adds
// t.SchemePayoff into Character.Cash regardless of Imprisoned. The
// Imprisoned branch must therefore still report the Reward outcome
// rather than returning early, or an Imprisoned term's own earned
// Payoff/Ship Share would silently vanish from the rendered sheet.
func rogueTermLabel(t character.Term) string {
	label := "Scheme: " + t.Scheme + ", "

	if t.Imprisoned {
		label += fmt.Sprintf("Imprisoned %d years", t.PrisonYears)

		switch {
		case t.SchemeShipShare:
			label += ", Reward: Ship Share"
		case t.RewardSucceeded:
			label += fmt.Sprintf(", Reward: Payoff Cr%d", t.SchemePayoff)
		}

		return label
	}

	if t.SchemeShipShare {
		return label + "Success (Ship Share)"
	}

	if t.RewardSucceeded {
		return label + fmt.Sprintf("Success (Payoff Cr%d)", t.SchemePayoff)
	}

	return label + "Success (No Reward)"
}

// scholarTermLabel renders "Research: Unharmed, Publication: Success"/
// "Research: Wounded"/"Research: Unharmed, Publication: Award-Winning,
// Tenure Granted" — reuses riskResultLabel for the Risk half; Scholar's
// own PublicationSucceeded/AwardWinning/TenureGranted are typed bools
// like Rogue's own fields, not a fit for the generic bare-string
// RewardResult shape every other risk career uses.
func scholarTermLabel(t character.Term) string {
	label := "Research: " + riskResultLabel(t.RiskResult)

	if t.RiskResult == character.Unharmed {
		switch {
		case t.AwardWinning:
			label += ", Publication: Award-Winning"
		case t.PublicationSucceeded:
			label += ", Publication: Success"
		default:
			label += ", Publication: Rejected"
		}
	}

	if t.TenureGranted {
		label += ", Tenure Granted"
	}

	return label
}

// entertainerTermLabel renders "Fame 8, Talent 6, Unharmed, Reward:
// Success" — Fame/Talent's own per-term evolution is Entertainer-
// specific and worth surfacing alongside the generic Risk/Reward outcome,
// unlike the plain RiskResult+RewardResult shape the generic fallback
// already renders for Scout/Marine-shape careers.
//
// Does not reuse riskResultLabel's own "Dead" text for a Dead
// RiskResult: character/entertainer_generate.go's own doc comment is
// explicit that Entertainer's Dead means Talent completely spent, not
// physical death — printing the literal word "Dead" here would read
// identically to a Scout's/Marine's own real death on the same sheet. A
// code-review pass caught an earlier version doing exactly that.
func entertainerTermLabel(t character.Term) string {
	riskLabel := riskResultLabel(t.RiskResult)
	if t.RiskResult == character.Dead {
		riskLabel = "Talent Exhausted"
	}

	label := fmt.Sprintf("Fame %d, Talent %d, %s", t.FameAfterTerm, t.TalentAfterTerm, riskLabel)

	if t.RewardResult != "" && t.RewardResult != "None" {
		label += ", Reward: " + t.RewardResult
	}

	return label
}

// functionaryTermLabel renders "Unharmed, Reward: Promoted to F3
// Manager" / "Office Politics Failed" — Office Politics reuses
// RiskResult Disabled for "career ends, cannot Continue"
// (character/functionary_generate.go's own ResolveFunctionaryTerm doc
// comment), not a physical disability, so the literal word "Disabled"
// is never shown here — the same class of override
// entertainerTermLabel's own "Talent Exhausted" already establishes for
// a different reused RiskResult value.
func functionaryTermLabel(t character.Term) string {
	riskLabel := riskResultLabel(t.RiskResult)
	if t.RiskResult == character.Disabled {
		riskLabel = "Office Politics Failed"
	}

	label := riskLabel
	if t.RewardResult != "" && t.RewardResult != "None" {
		label += ", Reward: " + t.RewardResult
	}

	return label
}

func citizenLifeLabel(succeeded bool) string {
	if succeeded {
		return "Citizen Life: Success"
	}

	return "Citizen Life: Failure"
}

// writeMusteringOut always renders all four benefit lists, even for a
// never-qualified career (where every one is empty) — uniform structure
// over special-casing, and "Automatics: None" etc. is accurate there too
// (no Mustering Out rolls occurred, per career_muster_out.go's own
// scoutMusterOutRollCount).
//
// Uses joinPhrasesOrNone, not world.JoinOrNone: these entries are
// multi-word phrases ("Forbidden Knowledge", "Middle Passage"), unlike
// world.JoinOrNone's own short fixed-token use cases (trade codes,
// bases) — a plain space join would run two or more such phrases
// together with no way to tell where one ends and the next begins.
func writeMusteringOut(b *strings.Builder, m character.MusteringOut) {
	fmt.Fprint(b, "**Mustering Out**\n\n")
	fmt.Fprintf(b, "- Automatics: %s\n", joinPhrasesOrNone(m.Automatics))
	fmt.Fprintf(b, "- Benefits: %s\n", joinPhrasesOrNone(m.Benefits))
	fmt.Fprintf(b, "- Money: %s\n", joinPhrasesOrNone(m.Money))
	fmt.Fprintf(b, "- Entitlements: %s\n", joinPhrasesOrNone(m.Entitlements))

	if m.Pension != 0 {
		fmt.Fprintf(b, "- Pension: Cr%d/year\n", m.Pension)
	}

	if m.RetirementPay != 0 {
		fmt.Fprintf(b, "- Retirement Pay: Cr%d/year\n", m.RetirementPay)
	}

	fmt.Fprint(b, "\n")
}

// positionAbbrev returns p's Human characteristic abbreviation, matching
// the exact literal strings career_generate.go's own scoutSkillTable[0]
// already uses for Personal skill grants (rollScoutSkill) — the
// established convention for how a Position prints, not a new one.
func positionAbbrev(p character.Position) string {
	switch p {
	case character.C1:
		return "Str"
	case character.C2:
		return "Dex"
	case character.C3:
		return "End"
	case character.C4:
		return "Int"
	case character.C5:
		return "Edu"
	case character.C6:
		return "Soc"
	default:
		return "?"
	}
}

// lifeStageLabel names c.LifeStage per Book 1 p.89's own "THE STAGES OF
// LIFE" table. "?" is unreachable in practice — character.LifeStageForAge
// only ever returns 0-9 — but kept for the same reason positionAbbrev and
// riskResultLabel keep their own default cases: no silent zero value for
// an out-of-range input.
func lifeStageLabel(stage int) string {
	switch stage {
	case 0:
		return "Infant"
	case 1:
		return "Child"
	case 2:
		return "Adolescent"
	case 3:
		return "Young Adult"
	case 4:
		return "Adult"
	case 5:
		return "Peak"
	case 6:
		return "Mid-Life"
	case 7:
		return "Senior"
	case 8:
		return "Elder"
	case 9:
		return "Retirement"
	default:
		return "?"
	}
}

// joinPhrasesOrNone joins multi-word phrases with a comma separator (not
// world.JoinOrNone's plain space join, which is only unambiguous for
// short fixed tokens like trade codes) — "None" for an empty list,
// matching world.JoinOrNone's own empty-list convention.
func joinPhrasesOrNone(items []string) string {
	if len(items) == 0 {
		return "None"
	}

	return strings.Join(items, ", ")
}

func riskResultLabel(r character.RiskResult) string {
	switch r {
	case character.Unharmed:
		return "Unharmed"
	case character.Wounded:
		return "Wounded"
	case character.Disabled:
		return "Disabled"
	case character.Dead:
		return "Dead"
	default:
		return "?"
	}
}
