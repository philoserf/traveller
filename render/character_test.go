package render_test

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/philoserf/traveller/character"
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
	"github.com/philoserf/traveller/render"
)

var scoutSheet = character.Character{
	Species:        "Human",
	GeneticProfile: "SDEIES",
	UPP:            character.UPP{Characteristics: [6]ehex.Value{8, 9, 7, 6, 5, 4}},
	Homeworld:      "A788899-C",
	Birthworld:     "A788899-C",
	Careers: []character.Career{
		{
			Name: "Scout",
			Terms: []character.Term{
				{
					ControllingCharacteristic: character.C1,
					RiskResult:                character.Unharmed,
					RewardResult:              "None",
					SkillsAwarded: []character.SkillLevel{
						{Name: "Vacc Suit", Level: 1, Kind: character.Skill},
					},
				},
				{
					ControllingCharacteristic: character.C2,
					RiskResult:                character.Wounded,
					RewardResult:              "Discovery",
					SkillsAwarded: []character.SkillLevel{
						{Name: "Zero-G", Level: 0, Kind: character.Skill},
					},
				},
			},
			MusteringOut: character.MusteringOut{
				Benefits: []string{"Ship Share"},
				Money:    []string{"Cr30,000"},
			},
		},
	},
	Skills: []character.SkillLevel{
		{Name: "Zero-G", Level: 1, Kind: character.Skill},
		{Name: "Vacc Suit", Level: 0, Kind: character.Skill},
	},
	WoundBadges: 1,
	Fame:        2,
	Cash:        30000,
}

func TestCharacterContainsAllFields(t *testing.T) {
	t.Parallel()

	out := render.Character(scoutSheet)

	want := []string{
		"Human",
		"SDEIES",
		scoutSheet.UPP.String(),
		"A788899-C",
		"**Wound Badges:** 1",
		"**Fame:** 2",
		"**Cash:** Cr30000",
		"### Scout",
		"Term 1 (Str): Unharmed",
		"Term 2 (Dex): Wounded, Reward: Discovery",
		"Vacc Suit-1",
		"Zero-G",
		"- Benefits: Ship Share",
		"- Money: Cr30,000",
	}

	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("render.Character missing %q in output:\n%s", w, out)
		}
	}
}

func TestCharacterTitleFallsBackToUPP(t *testing.T) {
	t.Parallel()

	out := render.Character(scoutSheet) // Name is unset
	if !strings.HasPrefix(out, "# "+scoutSheet.UPP.String()+"\n") {
		t.Errorf("render.Character with no Name should title with the UPP code, got:\n%s", out)
	}

	named := scoutSheet
	named.Name = "Eneri Dinsha"

	out = render.Character(named)
	if !strings.HasPrefix(out, "# Eneri Dinsha\n") {
		t.Errorf("render.Character with a Name set should use it as the title, got:\n%s", out)
	}
}

func TestCharacterShowsBirthworldOnlyWhenDifferent(t *testing.T) {
	t.Parallel()

	out := render.Character(scoutSheet) // Birthworld == Homeworld
	if strings.Contains(out, "Birthworld") {
		t.Errorf("render.Character should omit Birthworld when equal to Homeworld, got:\n%s", out)
	}

	moved := scoutSheet
	moved.Birthworld = "B000000-0"

	out = render.Character(moved)
	if !strings.Contains(out, "**Birthworld:** B000000-0") {
		t.Errorf("render.Character should show Birthworld when it differs from Homeworld, got:\n%s", out)
	}
}

func TestCharacterShowsRankOnlyWhenSet(t *testing.T) {
	t.Parallel()

	out := render.Character(scoutSheet) // Rank is unset
	if strings.Contains(out, "Rank") {
		t.Errorf("render.Character should omit Rank when unset, got:\n%s", out)
	}

	ranked := scoutSheet
	ranked.Rank = "Ensign"

	out = render.Character(ranked)
	if !strings.Contains(out, "**Rank:** Ensign") {
		t.Errorf("render.Character should show a set Rank, got:\n%s", out)
	}
}

// TestCharacterAlwaysShowsAgeLifeStageAndBirthdate confirms Age, Life
// Stage, and Birthdate all render unconditionally, unlike
// Rank/Fame/Cash/Notes — every real generation path now computes them
// (finalizeAging and GenerateBirthdate, character/aging.go and
// character/birthdate.go), so there's no zero-value ambiguity left to
// guard against. Name remains the one field character generation still
// never sets, but render.Character has no dedicated "Name:" label to
// guard — it only affects characterTitle's own fallback, already covered
// by TestCharacterTitleFallsBackToUPP.
func TestCharacterAlwaysShowsAgeLifeStageAndBirthdate(t *testing.T) {
	t.Parallel()

	aged := scoutSheet
	aged.Age = 42
	aged.LifeStage = 6
	aged.Birthdate = "Wonday 044-1075"

	out := render.Character(aged)

	for _, w := range []string{"**Age:** 42", "**Life Stage:** Mid-Life", "**Birthdate:** Wonday 044-1075"} {
		if !strings.Contains(out, w) {
			t.Errorf("render.Character missing %q in output:\n%s", w, out)
		}
	}
}

// TestCharacterShowsNotesOnlyWhenSet mirrors
// TestCharacterShowsRankOnlyWhenSet: Notes is empty unless Aging actually
// produced an illness/death event.
func TestCharacterShowsNotesOnlyWhenSet(t *testing.T) {
	t.Parallel()

	out := render.Character(scoutSheet) // Notes is unset
	if strings.Contains(out, "Notes") {
		t.Errorf("render.Character should omit Notes when unset, got:\n%s", out)
	}

	noted := scoutSheet
	noted.Notes = "Age 70: extremely major illness"

	out = render.Character(noted)
	if !strings.Contains(out, "**Notes:** Age 70: extremely major illness") {
		t.Errorf("render.Character should show set Notes, got:\n%s", out)
	}
}

func TestCharacterOmitsEmptySkills(t *testing.T) {
	t.Parallel()

	bare := character.Character{Skills: nil}

	out := render.Character(bare)
	if !strings.Contains(out, "## Skills\n\nNone.") {
		t.Errorf("render.Character with no Skills should show \"None.\", got:\n%s", out)
	}
}

func TestCharacterShowsNeverQualifiedCareer(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{{Name: "Scout", Terms: nil}},
	}

	out := render.Character(c)
	if !strings.Contains(out, "Never qualified for this career.") {
		t.Errorf("render.Character with an empty career should say so, got:\n%s", out)
	}

	if !strings.Contains(out, "- Automatics: None") {
		t.Errorf(
			"render.Character's Mustering Out section should still render (all None) for a never-qualified career, got:\n%s",
			out,
		)
	}
}

func TestCharacterShowsMusteringOutEntitlements(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{
			{
				Name: "Scout",
				MusteringOut: character.MusteringOut{
					Pension:       5000,
					RetirementPay: 2000,
				},
			},
		},
	}

	out := render.Character(c)

	if !strings.Contains(out, "- Pension: Cr5000/year") {
		t.Errorf("render.Character should show a set Pension, got:\n%s", out)
	}

	if !strings.Contains(out, "- Retirement Pay: Cr2000/year") {
		t.Errorf("render.Character should show a set Retirement Pay, got:\n%s", out)
	}

	if strings.Contains(render.Character(scoutSheet), "Pension") {
		t.Errorf("render.Character should omit Pension when 0, got:\n%s", render.Character(scoutSheet))
	}
}

func TestCharacterRendersNobleTermOutcome(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{
			{
				Name: character.NobleCareerName,
				Terms: []character.Term{
					{
						ControllingCharacteristic: character.C2,
						NobleAction:               "Intrigue",
						NobleSucceeded:            true,
						Elevated:                  true,
					},
					{ControllingCharacteristic: character.C3, NobleAction: "Return", NobleSucceeded: false},
				},
			},
		},
	}

	out := render.Character(c)

	for _, want := range []string{
		"Term 1 (Dex): Intrigue: Success, Elevated",
		"Term 2 (End): Return: Failure",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("render.Character missing %q in output:\n%s", want, out)
		}
	}
}

// TestCharacterRendersRogueTermOutcome mirrors
// TestCharacterRendersNobleTermOutcome's own shape, covering every
// combination rogueTermLabel must distinguish: a scaled Payoff, a Ship
// Share, no Reward, a plain Imprisoned term (Reward also failed), and —
// the case a code-review pass caught missing — an Imprisoned term where
// Reward independently succeeded on both a Payoff and a Ship Share
// Scheme. Reward is rolled unconditionally regardless of Risk's own
// outcome (ResolveRogueTerm), so all four Imprisoned x Reward
// combinations are real, reachable states, not just the "both failed"
// case.
func TestCharacterRendersRogueTermOutcome(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{
			{
				Name: character.RogueCareerName,
				Terms: []character.Term{
					{
						ControllingCharacteristic: character.C6,
						Scheme:                    "Entertainer",
						RewardSucceeded:           true,
						SchemePayoff:              600000,
					},
					{
						ControllingCharacteristic: character.C6,
						Scheme:                    "Merchant",
						RewardSucceeded:           true,
						SchemeShipShare:           true,
					},
					{
						ControllingCharacteristic: character.C6,
						Scheme:                    "Marine",
						RewardSucceeded:           false,
					},
					{
						ControllingCharacteristic: character.C6,
						Scheme:                    "Soldier",
						Imprisoned:                true,
						PrisonYears:               2,
					},
					{
						ControllingCharacteristic: character.C6,
						Scheme:                    "Soldier",
						Imprisoned:                true,
						PrisonYears:               3,
						RewardSucceeded:           true,
						SchemePayoff:              75000,
					},
					{
						ControllingCharacteristic: character.C6,
						Scheme:                    "Merchant",
						Imprisoned:                true,
						PrisonYears:               1,
						RewardSucceeded:           true,
						SchemeShipShare:           true,
					},
				},
			},
		},
	}

	out := render.Character(c)

	for _, want := range []string{
		"Term 1 (Soc): Scheme: Entertainer, Success (Payoff Cr600000)",
		"Term 2 (Soc): Scheme: Merchant, Success (Ship Share)",
		"Term 3 (Soc): Scheme: Marine, Success (No Reward)",
		"Term 4 (Soc): Scheme: Soldier, Imprisoned 2 years",
		"Term 5 (Soc): Scheme: Soldier, Imprisoned 3 years, Reward: Payoff Cr75000",
		"Term 6 (Soc): Scheme: Merchant, Imprisoned 1 years, Reward: Ship Share",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("render.Character missing %q in output:\n%s", want, out)
		}
	}
}

// TestCharacterRendersScholarTermOutcome mirrors
// TestCharacterRendersRogueTermOutcome's own shape, covering
// scholarTermLabel's own branches: Wounded/Disabled skip the Publication
// clause entirely (Reward is only rolled on Unharmed, per
// ResolveScholarTerm), an Unharmed term shows Publication Success/
// Rejected/Award-Winning, and Tenure Granted appends independently of
// the Publication outcome.
func TestCharacterRendersScholarTermOutcome(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{
			{
				Name: character.ScholarCareerName,
				Terms: []character.Term{
					{ControllingCharacteristic: character.C1, RiskResult: character.Wounded},
					{ControllingCharacteristic: character.C1, RiskResult: character.Disabled},
					{
						ControllingCharacteristic: character.C1,
						RiskResult:                character.Unharmed,
						PublicationSucceeded:      false,
					},
					{
						ControllingCharacteristic: character.C1,
						RiskResult:                character.Unharmed,
						PublicationSucceeded:      true,
					},
					{
						ControllingCharacteristic: character.C1,
						RiskResult:                character.Unharmed,
						PublicationSucceeded:      true,
						AwardWinning:              true,
					},
					{
						ControllingCharacteristic: character.C1,
						RiskResult:                character.Unharmed,
						PublicationSucceeded:      true,
						TenureGranted:             true,
					},
				},
			},
		},
	}

	out := render.Character(c)

	for _, want := range []string{
		"Term 1 (Str): Research: Wounded",
		"Term 2 (Str): Research: Disabled",
		"Term 3 (Str): Research: Unharmed, Publication: Rejected",
		"Term 4 (Str): Research: Unharmed, Publication: Success",
		"Term 5 (Str): Research: Unharmed, Publication: Award-Winning",
		"Term 6 (Str): Research: Unharmed, Publication: Success, Tenure Granted",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("render.Character missing %q in output:\n%s", want, out)
		}
	}
}

// TestCharacterOmitsZeroFameAndCash guards against showing "Fame: 0"/
// "Cash: Cr0" for a Character whose Fame/Cash were never actually
// computed — see Character's own doc comment on why that would
// contradict the Benefits/Money lines shown elsewhere on the sheet.
func TestCharacterOmitsZeroFameAndCash(t *testing.T) {
	t.Parallel()

	bare := character.Character{}

	out := render.Character(bare)
	if strings.Contains(out, "Fame") || strings.Contains(out, "Cash") {
		t.Errorf("render.Character should omit Fame/Cash when 0, got:\n%s", out)
	}

	if got := render.Character(
		scoutSheet,
	); !strings.Contains(got, "**Fame:** 2") ||
		!strings.Contains(got, "**Cash:** Cr30000") {
		t.Errorf("render.Character should show a nonzero Fame/Cash, got:\n%s", got)
	}
}

// TestCharacterShowsMedalsOnlyWhenSet mirrors TestCharacterOmitsZeroFameAndCash's
// own convention for Character.Medals (Marine's own p.91 medal grants).
func TestCharacterShowsMedalsOnlyWhenSet(t *testing.T) {
	t.Parallel()

	bare := character.Character{}
	if out := render.Character(bare); strings.Contains(out, "Medals") {
		t.Errorf("render.Character should omit Medals when empty, got:\n%s", out)
	}

	withMedals := character.Character{Medals: []string{"XS", "MCUF"}}
	if got := render.Character(withMedals); !strings.Contains(got, "**Medals:** XS, MCUF") {
		t.Errorf("render.Character should show a joined Medals list, got:\n%s", got)
	}
}

// TestCharacterRendersPersonalSkillAsBoost is the regression test for
// the Personal-kind rendering fix: a characteristic boost (e.g. Scout's
// own Str+1 Personal skill grant) must not look like a same-named
// trained skill.
func TestCharacterRendersPersonalSkillAsBoost(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Skills: []character.SkillLevel{{Name: "Str", Level: 1, Kind: character.Personal}},
	}

	out := render.Character(c)

	if !strings.Contains(out, "- Str +1") {
		t.Errorf("render.Character should render a Personal-kind grant as %q, got:\n%s", "Str +1", out)
	}

	if strings.Contains(out, "Str-1") {
		t.Errorf("render.Character should not render a Personal-kind grant as skill notation \"Str-1\", got:\n%s", out)
	}
}

// TestCharacterJoinsMultipleMusteringOutEntriesWithCommas is the
// regression test for the join-separator fix: multi-word Mustering Out
// entries must stay distinguishable when a career earns two or more.
func TestCharacterJoinsMultipleMusteringOutEntriesWithCommas(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{
			{
				Name: "Scout",
				MusteringOut: character.MusteringOut{
					Benefits: []string{"Forbidden Knowledge", "Ship Share"},
				},
			},
		},
	}

	out := render.Character(c)

	if !strings.Contains(out, "- Benefits: Forbidden Knowledge, Ship Share") {
		t.Errorf("render.Character should comma-join multiple Benefits entries, got:\n%s", out)
	}
}

// TestCharacterRendersAllPositionsAndRiskResults is a full-coverage
// black-box pin of positionAbbrev/riskResultLabel's mapping, exercised
// through the exported render.Character API — this package's existing
// tests (world_test.go, system_test.go) are all external (package
// render_test) black-box tests with no direct access to unexported
// helpers, so this covers all 6 Positions and all 4 RiskResults the same
// way, matching this project's established "no partial pins" convention.
func TestCharacterRendersAllPositionsAndRiskResults(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{
			{
				Name: "Scout",
				Terms: []character.Term{
					{ControllingCharacteristic: character.C1, RiskResult: character.Unharmed},
					{ControllingCharacteristic: character.C2, RiskResult: character.Wounded},
					{ControllingCharacteristic: character.C3, RiskResult: character.Disabled},
					{ControllingCharacteristic: character.C4, RiskResult: character.Dead},
					{ControllingCharacteristic: character.C5, RiskResult: character.Unharmed},
					{ControllingCharacteristic: character.C6, RiskResult: character.Wounded},
				},
			},
		},
	}

	out := render.Character(c)

	want := []string{
		"Term 1 (Str): Unharmed",
		"Term 2 (Dex): Wounded",
		"Term 3 (End): Disabled",
		"Term 4 (Int): Dead",
		"Term 5 (Edu): Unharmed",
		"Term 6 (Soc): Wounded",
	}

	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("render.Character missing %q in output:\n%s", w, out)
		}
	}
}

// TestCharacterHandlesGeneratedOutput is an integration smoke test:
// render.Character must not panic on real character.GenerateScoutCharacter
// output, across both the ok and !ok paths — the fixture-based tests
// above already cover exact formatting.
func TestCharacterHandlesGeneratedOutput(t *testing.T) {
	t.Parallel()

	for seed := range uint64(20) {
		r := dice.New(rand.NewPCG(seed+1, seed+1))

		c, _ := character.GenerateScoutCharacter(r)

		if out := render.Character(c); out == "" {
			t.Errorf("seed %d: render.Character produced empty output", seed)
		}
	}
}

var citizenSheet = character.Character{
	Species:        "Human",
	GeneticProfile: "SDEIES",
	UPP:            character.UPP{Characteristics: [6]ehex.Value{8, 9, 7, 6, 5, 4}},
	Homeworld:      "A788899-C",
	Careers: []character.Career{
		{
			Name:       character.CitizenCareerName,
			JobSkill:   "Pilot",
			HobbySkill: "Broker",
			Terms: []character.Term{
				{
					ControllingCharacteristic: character.C1,
					CitizenLifeSucceeded:      true,
					SkillsAwarded: []character.SkillLevel{
						{Name: "Pilot", Level: 4, Kind: character.Skill},
					},
				},
				{
					ControllingCharacteristic: character.C2,
					CitizenLifeSucceeded:      false,
				},
			},
			MusteringOut: character.MusteringOut{
				Benefits: []string{"Str +1"},
			},
		},
	},
}

// TestCharacterRendersCitizenLifeOutcome confirms Citizen's own term
// lines show "Citizen Life: Success"/"Citizen Life: Failure", never a
// RiskResult word ("Unharmed"/"Wounded"/etc.) or a Reward suffix, which
// have no meaning for Citizen Life.
func TestCharacterRendersCitizenLifeOutcome(t *testing.T) {
	t.Parallel()

	out := render.Character(citizenSheet)

	want := []string{
		"Term 1 (Str): Citizen Life: Success",
		"Term 2 (Dex): Citizen Life: Failure",
	}

	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("render.Character missing %q in output:\n%s", w, out)
		}
	}

	for _, unwanted := range []string{"Unharmed", "Wounded", "Disabled", "Dead", "Reward:"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("render.Character should not render %q for a Citizen career, got:\n%s", unwanted, out)
		}
	}
}

func TestCharacterShowsJobAndHobbyOnlyWhenSet(t *testing.T) {
	t.Parallel()

	out := render.Character(citizenSheet)
	if !strings.Contains(out, "**Job:** Pilot") || !strings.Contains(out, "**Hobby:** Broker") {
		t.Errorf("render.Character should show a set Job/Hobby, got:\n%s", out)
	}

	bare := character.Character{
		Careers: []character.Career{
			{Name: character.CitizenCareerName, Terms: []character.Term{{ControllingCharacteristic: character.C1}}},
		},
	}

	out = render.Character(bare)
	if strings.Contains(out, "Job") || strings.Contains(out, "Hobby") {
		t.Errorf("render.Character should omit Job/Hobby when unset, got:\n%s", out)
	}
}

// TestCharacterHandlesCitizenGeneratedOutput is the Citizen analog of
// TestCharacterHandlesGeneratedOutput.
func TestCharacterHandlesCitizenGeneratedOutput(t *testing.T) {
	t.Parallel()

	for seed := range uint64(20) {
		r := dice.New(rand.NewPCG(seed+1, seed+1))

		c := character.GenerateCitizenCharacter(r)

		if out := render.Character(c); out == "" {
			t.Errorf("seed %d: render.Character produced empty output", seed)
		}
	}
}

// TestCharacterRendersEntertainerTermOutcome confirms Entertainer's own
// term prefix omits the "(Xxx)" characteristic suffix every other
// career's own termOutcomeLine shows (Entertainer has no Controlling
// Characteristic at all — t.ControllingCharacteristic is never set for
// its own terms), and that entertainerTermLabel shows Fame/Talent
// alongside the generic Risk/Reward outcome, including the no-Reward
// case (Dead skips the Reward roll entirely).
func TestCharacterRendersEntertainerTermOutcome(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{
			{
				Name: character.EntertainerCareerName,
				Terms: []character.Term{
					{
						RiskResult:      character.Unharmed,
						RewardResult:    "Success",
						FameAfterTerm:   8,
						TalentAfterTerm: 6,
					},
					{
						RiskResult:      character.Wounded,
						RewardResult:    "None",
						FameAfterTerm:   4,
						TalentAfterTerm: 4,
					},
					{
						RiskResult:      character.Dead,
						FameAfterTerm:   2,
						TalentAfterTerm: 0,
					},
				},
			},
		},
	}

	out := render.Character(c)

	for _, want := range []string{
		"Term 1: Fame 8, Talent 6, Unharmed, Reward: Success",
		"Term 2: Fame 4, Talent 4, Wounded",
		"Term 3: Fame 2, Talent 0, Talent Exhausted",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("render.Character missing %q in output:\n%s", want, out)
		}
	}

	if strings.Contains(out, "Term 1 (") || strings.Contains(out, "Term 2 (") || strings.Contains(out, "Term 3 (") {
		t.Errorf("render.Character should omit the \"(Xxx)\" characteristic suffix for Entertainer, got:\n%s", out)
	}
}

// TestCharacterRendersEntertainerSpecialty is the regression test for a
// code-review-caught bug: writeCareer rendered Citizen's own JobSkill/
// HobbySkill but never Entertainer's analogous Specialty field, silently
// dropping which of Artist/Actor/Author/Dancer/Musician/Chef a generated
// character has from the rendered sheet entirely.
func TestCharacterRendersEntertainerSpecialty(t *testing.T) {
	t.Parallel()

	c := character.Character{
		Careers: []character.Career{
			{
				Name:      character.EntertainerCareerName,
				Specialty: "Chef",
				Terms:     []character.Term{{RiskResult: character.Unharmed}},
			},
		},
	}

	out := render.Character(c)
	if !strings.Contains(out, "**Specialty:** Chef") {
		t.Errorf("render.Character missing the Specialty line, got:\n%s", out)
	}
}
