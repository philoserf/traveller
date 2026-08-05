package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/philoserf/traveller/character"
)

// Character renders c as a Markdown character sheet: Species, Genetic
// Profile, UPP, Age, Life Stage, Birthdate, Homeworld (and Birthworld
// only when it differs), Wound Badges, the full Skills list, CharGen
// step C's own Education (if attended), each Career's Terms and
// Mustering Out benefits, any Land Grants, and any Masterpieces — in
// that order, matching Book 1's own CharGen step order (Education is
// step C, before career resolution; Land Grants and Masterpieces are
// consequences of careers, so they sit after). Age, Life Stage, and
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
// Equipment is likewise shown only when non-empty — Craftsman is the
// first (and so far only) source that ever populates it, one entry per
// Masterpiece created; the Masterpieces section below shows the same
// creations in structured form (QREBS, Vintage-appreciated value).
// Position and RiskResult get local label functions
// (positionAbbrev, riskResultLabel) rather than String() methods on the
// character package's own types, matching this project's existing
// precedent (world.TradeCode/world.Base have no String() either — render
// uses world.TradeCodeStrings/world.BaseStrings instead); lifeStageLabel
// follows the same convention.
func Character(c character.Character) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", characterTitle(c))

	writeMetadata(&b, c)

	fmt.Fprint(&b, "## Skills\n\n")
	writeSkills(&b, c.Skills)

	writeEducation(&b, c.Education)

	// writeEducation already supplies its own leading blank line and its
	// own trailing one when it writes anything at all — an unconditional
	// second "\n" here would double up. Only add one when Education
	// stayed silent, matching writeSkills' own no-Education gap exactly.
	if c.Education.Attended() {
		fmt.Fprint(&b, "## Careers\n\n")
	} else {
		fmt.Fprint(&b, "\n## Careers\n\n")
	}

	for _, career := range c.Careers {
		writeCareer(&b, career)
	}

	writeLandGrants(&b, c)
	writeMasterpieces(&b, c)

	return strings.TrimRight(b.String(), "\n") + "\n"
}

// writeMetadata renders the character's own header fields as one
// compact block: consecutive lines joined by Markdown's two-space hard
// break, with a single blank line after the block rather than after
// every field. Each field previously became its own paragraph, which
// spent a screen of vertical space announcing mostly-constant values.
//
// A field appears only when it carries information the reader doesn't
// already have. Species and Genetic Profile are omitted at their Human
// defaults (nothing in this codebase generates anything else yet, so
// printing them says only "this generator ran"); UPP is omitted when
// characterTitle already used it as the heading, but kept when a Name
// took that slot; Wound Badges is omitted at zero, matching the
// existing treatment of Fame, Cash, Rank, Notes and Medals; Birthworld
// keeps its existing "only when it differs from Homeworld" rule.
// Commendations follows Medals' own shape exactly. Noble Title
// (Character.NobleTitle) follows the same "only when it differs" rule
// as Birthworld — against Rank, not Homeworld — since NobleTitle
// already equals Rank for a Noble-career character; it earns its own
// line only for a Knighthood-derived title with no career rank behind
// it. Proxy Income is omitted at zero the same way Fame/Cash are.
//
// Age, Life Stage, Birthdate and Homeworld always print: every
// generation path computes them, and none has a value that means
// "absent". Life Stage is derived from Age but kept — it names the
// rules bracket (Book 1 p.89) rather than restating the number.
func writeMetadata(b *strings.Builder, c character.Character) {
	lines := append(metadataIdentityLines(c), metadataStatusLines(c)...)

	// Two trailing spaces are Markdown's hard line break: the block reads
	// as one paragraph of labelled lines rather than as many paragraphs.
	fmt.Fprintf(b, "%s\n\n", strings.Join(lines, "  \n"))
}

// metadataIdentityLines is who/what the character is: Species, Genetic
// Profile, UPP (only when a Name displaced it from the heading), Age,
// Life Stage, Birthdate, Homeworld, Birthworld (only when it differs),
// and Rank.
func metadataIdentityLines(c character.Character) []string {
	var lines []string

	add := func(format string, args ...any) {
		lines = append(lines, fmt.Sprintf(format, args...))
	}

	if c.Species != "" && c.Species != character.HumanSpecies {
		add("**Species:** %s", c.Species)
	}

	if c.GeneticProfile != "" && c.GeneticProfile != character.HumanGeneticProfile {
		add("**Genetic Profile:** %s", c.GeneticProfile)
	}

	// Redundant with the heading unless a Name displaced it there.
	if c.Name != "" {
		add("**UPP:** %s", c.UPP)
	}

	add("**Age:** %d", c.Age)
	add("**Life Stage:** %s", lifeStageLabel(c.LifeStage))
	add("**Birthdate:** %s", c.Birthdate)
	add("**Homeworld:** %s", c.Homeworld)

	if c.Birthworld != "" && c.Birthworld != c.Homeworld {
		add("**Birthworld:** %s", c.Birthworld)
	}

	if c.Rank != "" {
		add("**Rank:** %s", c.Rank)
	}

	return lines
}

// metadataStatusLines is what the character has accumulated: Wound
// Badges, Medals, Commendations, Noble Title (only when it differs from
// Rank — see writeMetadata's own doc comment), Proxy Income, Fame,
// Cash, Notes, and Equipment. All are omitted at their zero value.
func metadataStatusLines(c character.Character) []string {
	var lines []string

	add := func(format string, args ...any) {
		lines = append(lines, fmt.Sprintf(format, args...))
	}

	if c.WoundBadges != 0 {
		add("**Wound Badges:** %d", c.WoundBadges)
	}

	if len(c.Medals) > 0 {
		add("**Medals:** %s", strings.Join(c.Medals, ", "))
	}

	if len(c.Commendations) > 0 {
		add("**Commendations:** %s", strings.Join(c.Commendations, ", "))
	}

	// Only when it earns its place beside Rank: NobleTitle already equals
	// Rank for a Noble-career character (TestGeneratedNobleRankAndTitleNeverDisagree),
	// so this line exists for a Knighthood-derived title with no career
	// rank behind it at all — the same "only when it differs" shape
	// Birthworld already uses against Homeworld.
	if title := c.NobleTitle(); title != "" && title != c.Rank {
		add("**Noble Title:** %s", title)
	}

	if income := c.ProxyIncome(); income != 0 {
		add("**Proxy Income:** %s/year", formatCr(income))
	}

	if c.Fame != 0 {
		add("**Fame:** %d", c.Fame)
	}

	if c.Cash != 0 {
		add("**Cash:** %s", formatCr(c.Cash))
	}

	if c.Notes != "" {
		add("**Notes:** %s", c.Notes)
	}

	if len(c.Equipment) > 0 {
		add("**Equipment:** %s", strings.Join(c.Equipment, ", "))
	}

	return lines
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
	// Later Education (p.59) can occur inside any of the careers that
	// offer it, so it's checked before any career-specific dispatch
	// below — a term spent at school has no Controlling Characteristic,
	// Risk, or Reward to report, only the institution attended.
	if t.LaterEducation {
		return fmt.Sprintf("Term %d: Later Education — %s", i+1, t.LaterEducationSchool)
	}

	// Entertainer has no Controlling Characteristic at all (Risk & Reward
	// targets Talent, not a UPP position) — t.ControllingCharacteristic
	// is never set for its own terms, so positionAbbrev's own zero-value
	// fallback ("Str") would misleadingly suggest one. Its own prefix
	// omits the "(Xxx)" suffix entirely rather than showing a fake CC.
	if c.Name == character.EntertainerCareerName {
		return fmt.Sprintf("Term %d", i+1) + ": " + entertainerTermLabel(t)
	}

	// Citizen names its own Controlling Characteristic inside the check
	// it renders, so the generic "(CC)" parenthetical would print it
	// twice — and unexplained the first time.
	if c.Name == character.CitizenCareerName {
		return fmt.Sprintf("Term %d — %s", i+1, citizenLifeLabel(t))
	}

	prefix := fmt.Sprintf("Term %d (%s)", i+1, positionAbbrev(t.ControllingCharacteristic))

	switch c.Name {
	case character.NobleCareerName:
		return prefix + ": " + nobleTermLabel(t)
	case character.RogueCareerName:
		return prefix + ": " + rogueTermLabel(t)
	case character.ScholarCareerName:
		return prefix + ": " + scholarTermLabel(t)
	case character.FunctionaryCareerName:
		return prefix + ": " + functionaryTermLabel(t)
	}

	// Unharmed is the ordinary outcome of a Risk roll, so naming it on
	// every routine term buries the Wounded/Disabled/Dead ones that
	// actually matter. A term with nothing to report renders as the bare
	// "Term N (CC)"; a Reward still prints on its own.
	var parts []string

	if t.RiskResult != character.Unharmed {
		parts = append(parts, riskResultLabel(t.RiskResult))
	}

	if t.RewardResult != "" && t.RewardResult != "None" {
		parts = append(parts, "Reward: "+t.RewardResult)
	}

	if len(parts) == 0 {
		return prefix
	}

	return prefix + ": " + strings.Join(parts, ", ")
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

	// Two different things, both worth showing: a term served in prison
	// for last term's failure, and this term's own Scheme failing and
	// earning the next sentence. Book 1 p.84 imposes prison "at the
	// start of the next Term", so a Rogue can be doing both at once.
	if t.ServedYears > 0 {
		label = fmt.Sprintf("In prison %d years, ", t.ServedYears) + label
	}

	if t.Imprisoned {
		label += fmt.Sprintf("failed (sentenced %d years)", t.PrisonYears)

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

// citizenLifeLabel renders Citizen Life (Book 1 p.78) as the check it
// actually is — 2D6 against the term's own Controlling Characteristic,
// success on equal-or-under, no Mods — rather than as an unexplained
// "(Str): Success" tuple. Showing the roll beside the target it was
// compared against lets a reader confirm the outcome instead of taking
// it on trust, and names which characteristic was being tested and why.
//
// A successful term's extra Job or Hobby skill is called out by name.
// It also appears in the term's own skill list (it is a real skill), but
// nothing there distinguishes it from the four unconditional Table C
// skills beside it, and it is the entire mechanical consequence of
// succeeding.
//
// Failure carries no wound and no characteristic reduction — Book 1's
// own "the Citizen continues the term stuck in a dull, boring,
// unfulfilling life" — so the line deliberately stops at "failure"
// rather than implying a cost the rules don't impose.
func citizenLifeLabel(t character.Term) string {
	comparison, outcome := "≤", "success"
	if !t.CitizenLifeSucceeded {
		comparison, outcome = ">", "failure"
	}

	label := fmt.Sprintf("Citizen Life %d %s %s %d: %s",
		t.CitizenLifeRoll, comparison, positionAbbrev(t.ControllingCharacteristic),
		t.CitizenLifeTarget, outcome)

	if t.CitizenLifeGrant != "" {
		label += ", " + citizenLifeGrantLabel(t) + ": " + t.CitizenLifeGrant
	}

	return label
}

// citizenLifeGrantLabel names which of the two alternating tracks a
// successful term's grant came from — Book 1 p.78 alternates Job and
// Hobby across successful terms, so the label has to follow the same
// count rather than be stored per term.
func citizenLifeGrantLabel(t character.Term) string {
	if t.CitizenLifeGrantIsJob {
		return "Job"
	}

	return "Hobby"
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
//
// Money prints every raw roll, cash included, even though
// Character.Cash (rendered once, near the top of the sheet) already
// sums the same cash entries across every career. This is a per-career
// historical record, not a staging area the character-wide total
// supersedes: a career whose Mustering Out rolls happened to land on
// cash every time still produced real results that term, and hiding
// them behind "None" would misreport what actually happened, not just
// omit a redundant number. (An earlier version of this function
// filtered cash entries out here — reverted once that was shown to
// blank out the Money line for most veteran characters, since a high
// DM from Terms served pushes Money rolls toward a table's cash rows;
// see this repo's own PR discussion for #46.)
func writeMusteringOut(b *strings.Builder, m character.MusteringOut) {
	var lines []string

	add := func(label string, items []string) {
		if len(items) > 0 {
			lines = append(lines, fmt.Sprintf("- %s: %s", label, strings.Join(items, ", ")))
		}
	}

	add("Automatics", m.Automatics)
	add("Benefits", m.Benefits)
	add("Money", m.Money)
	add("Entitlements", m.Entitlements)

	if m.Pension != 0 {
		lines = append(lines, fmt.Sprintf("- Pension: Cr%d/year", m.Pension))
	}

	if m.RetirementPay != 0 {
		lines = append(lines, fmt.Sprintf("- Retirement Pay: Cr%d/year", m.RetirementPay))
	}

	// Nothing was granted — a never-qualified career, or one whose
	// character died before reaching Mustering Out. Four "None" lines
	// under a heading say exactly as much as no heading at all.
	if len(lines) == 0 {
		return
	}

	fmt.Fprint(b, "**Mustering Out**\n\n")
	fmt.Fprintf(b, "%s\n\n", strings.Join(lines, "\n"))
}

// formatCr renders amount as Book 1's own "CrN,NNN" thousands-grouped
// notation — the same comma-grouped form every muster-out table literal
// already uses (career_muster_out.go's own "Cr30,000" etc.), now also
// applied to Character.Cash's own accumulated total, which previously
// printed as an ungrouped "Cr720000".
func formatCr(amount int) string {
	sign := ""

	if amount < 0 {
		sign = "-"
		amount = -amount
	}

	digits := strconv.Itoa(amount)

	var grouped strings.Builder

	for i, d := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			grouped.WriteByte(',')
		}

		grouped.WriteRune(d)
	}

	return fmt.Sprintf("Cr%s%s", sign, grouped.String())
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

// writeEducation renders CharGen step C's own result (#113): School,
// Major/Minor, Degree, Honors, and — for a University graduate who
// qualified for one — the follow-on Graduate program chain (Masters,
// Medical School, or Law School, with Professors chained after a
// graduated Masters). Nothing is written at all for a character who
// attended no institution — Education.Attended() is false whenever
// chooseInstitution never found a qualifying school, which in practice
// is never for a generated character (education.go's own doc comment),
// but a hand-built Character has no such guarantee.
//
// CommissionCareers is deliberately not shown here: it's eligibility
// bookkeeping (which careers a Commission could still apply to), not an
// outcome, and its actual consequence — a career entered already
// Commissioned at Officer1 — is already visible below in that career's
// own first Term (Commissioned: true, an Officer1 Rank). Printing the
// raw list too would restate the same fact a second, more confusing way.
func writeEducation(b *strings.Builder, edu character.Education) {
	if !edu.Attended() {
		return
	}

	fmt.Fprint(b, "\n## Education\n\n")

	var lines []string

	add := func(format string, args ...any) {
		lines = append(lines, fmt.Sprintf(format, args...))
	}

	add("**School:** %s", edu.School)

	if edu.Major != "" {
		add("**Major:** %s", edu.Major)
	}

	if edu.Minor != "" {
		add("**Minor:** %s", edu.Minor)
	}

	if edu.Degree != "" {
		add("**Degree:** %s", edu.Degree)
	}

	if edu.Honors {
		add("**Honors:** yes")
	}

	fmt.Fprintf(b, "%s\n\n", strings.Join(lines, "  \n"))

	for grad := edu.Graduate; grad != nil; grad = grad.Next {
		writeGraduateProgram(b, *grad)
	}
}

// writeGraduateProgram renders one link of the Graduate chain — status
// distinguishes an attempt that never graduated (School attended, no
// Degree) from a completed one, the same "attended, even if it went
// nowhere" record Education.School itself already keeps for the
// primary institution.
func writeGraduateProgram(b *strings.Builder, grad character.GraduateProgram) {
	status := "attended"
	if grad.Graduated {
		status = "graduated as " + grad.Degree
	}

	fmt.Fprintf(b, "**Graduate Program:** %s (%s)\n\n", grad.School, status)
}

// writeLandGrants renders every Land Grant this character holds — p.88:
// "Land Grants Are Cumulative. Each title confers its own Land Grant: a
// Knight raised to Baronet receives it in addition to his Knighthood."
// A Total Annual Income line (the existing LandGrantIncome) appears
// only when there's more than one grant to sum; a single grant's own
// line already states its income.
func writeLandGrants(b *strings.Builder, c character.Character) {
	if len(c.LandGrants) == 0 {
		return
	}

	fmt.Fprint(b, "\n## Land Grants\n\n")

	for _, g := range c.LandGrants {
		fmt.Fprintf(b, "- %s on %s (MW %d, other %d hexes) — %s/year\n",
			g.Source, g.World, g.MainworldHexes, g.OtherHexes, formatCr(g.AnnualIncome))
	}

	if len(c.LandGrants) > 1 {
		fmt.Fprintf(b, "\n**Total Annual Income:** %s/year\n", formatCr(c.LandGrantIncome()))
	}

	fmt.Fprint(b, "\n")
}

// writeMasterpieces renders every Masterpiece this character created
// (#95's own QREBS allocation and Vintage appreciation): Master Points,
// the QREBS split, Perfect status, the age it was created, and the
// character-wide current Vintage-appreciated value (MasterpieceValue,
// which already sums across every Masterpiece). The same creations
// already appear as flat strings in Equipment; this is their
// structured record.
func writeMasterpieces(b *strings.Builder, c character.Character) {
	if len(c.Masterpieces) == 0 {
		return
	}

	fmt.Fprint(b, "\n## Masterpieces\n\n")

	for _, m := range c.Masterpieces {
		perfect := ""
		if m.Perfect {
			perfect = ", Perfect"
		}

		fmt.Fprintf(b, "- %d Master Points (Q%d R%d E%d B%d S%d)%s, created at Age %d\n",
			m.MasterPoints, m.QREBS.Quality, m.QREBS.Reliability, m.QREBS.EaseOfUse,
			m.QREBS.Burden, m.QREBS.Safety, perfect, m.CreatedAtAge)
	}

	fmt.Fprintf(b, "\n**Total Value:** %s\n\n", formatCr(c.MasterpieceValue()))
}
