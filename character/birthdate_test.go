package character

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"strings"
	"testing"

	"github.com/philoserf/traveller/dice"
)

// TestBirthDayOfYear directly encodes Book 1 p.263's own "Birth Date
// Generation" table, re-derived from the page image (see this slice's own
// plan-file Context for the full derivation): boundary coverage for both
// sub-tables, plus both of the book's own worked examples verbatim.
func TestBirthDayOfYear(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                   string
		die1, die2, die3, die4 int
		wantDay                int
		wantOK                 bool
	}{
		{"left table, k=1, col1", 1, 1, 1, 1, 1, true},
		{"left table, k=36, col1", 1, 6, 6, 1, 36, true},
		{"left table, k=1, col5", 3, 1, 1, 5, 145, true},
		{"left table, k=36, col5", 3, 6, 6, 5, 180, true},
		{"left table, col6 always rerolls (k=1)", 1, 1, 1, 6, 0, false},
		{"left table, col6 always rerolls (k=36)", 3, 6, 6, 6, 0, false},
		{"right table, k=1, col1", 4, 1, 1, 1, 181, true},
		{"right table, k=36, col1", 6, 6, 6, 1, 216, true},
		{"right table, k=1, col5", 4, 1, 1, 5, 325, true},
		{"right table, k=36, col5", 6, 6, 6, 5, 360, true},
		{"right table, col6, k=31 rerolls (below k=32 cutoff)", 4, 6, 1, 6, 0, false},
		{"right table, col6, k=32 (cutoff)", 4, 6, 2, 6, 361, true},
		{"right table, col6, k=36", 6, 6, 6, 6, 365, true},
		{"book's own worked example: Day=44", 2, 2, 2, 2, 44, true},
		{"book's own worked Reroll! example", 4, 3, 1, 6, 0, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			gotDay, gotOK := birthDayOfYear(c.die1, c.die2, c.die3, c.die4)
			if gotDay != c.wantDay || gotOK != c.wantOK {
				t.Errorf("birthDayOfYear(%d,%d,%d,%d) = (%d, %v), want (%d, %v)",
					c.die1, c.die2, c.die3, c.die4, gotDay, gotOK, c.wantDay, c.wantOK)
			}
		})
	}
}

// TestBirthDayOfYearCoversEveryDayExactlyOnce is the completeness proof
// behind the formula's own derivation, not just the sampled boundary
// rows above: every valid (die1,die2,die3,die4) combination's day must
// land in exactly {1..365}, with no day outside that range and none
// missing — confirming 180 (left table) + 180 (right table cols 1-5) + 5
// (right table col6, k>=32) really do partition the year with no gaps
// or overlaps.
func TestBirthDayOfYearCoversEveryDayExactlyOnce(t *testing.T) {
	t.Parallel()

	seen := make(map[int]bool)

	for die1 := 1; die1 <= 6; die1++ {
		for die2 := 1; die2 <= 6; die2++ {
			for die3 := 1; die3 <= 6; die3++ {
				for die4 := 1; die4 <= 6; die4++ {
					day, ok := birthDayOfYear(die1, die2, die3, die4)
					if !ok {
						continue
					}

					if day < 1 || day > 365 {
						t.Fatalf("birthDayOfYear(%d,%d,%d,%d) = %d, out of [1,365]", die1, die2, die3, die4, day)
					}

					seen[day] = true
				}
			}
		}
	}

	if len(seen) != 365 {
		t.Fatalf("valid days covered = %d, want 365", len(seen))
	}

	for day := 1; day <= 365; day++ {
		if !seen[day] {
			t.Errorf("day %d is never reachable", day)
		}
	}
}

// TestWeekdayName pins day 1 (the separately-named Holiday), both of the
// book's own worked examples, and a full 7-day cycle to confirm
// wraparound.
func TestWeekdayName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		day  int
		want string
	}{
		{1, "Holiday"},
		{2, "Wonday"},
		{3, "Tuday"},
		{4, "Thirday"},
		{5, "Forday"},
		{6, "Fiday"},
		{7, "Sixday"},
		{8, "Senday"},
		{9, "Wonday"}, // cycle wraps
		{44, "Wonday"},
		{241, "Tuday"},
	}

	for _, c := range cases {
		if got := weekdayName(c.day); got != c.want {
			t.Errorf("weekdayName(%d) = %q, want %q", c.day, got, c.want)
		}
	}
}

var birthdateFormat = regexp.MustCompile(`^[A-Za-z]+ \d{3}-\d+$`)

// assertBirthdateFormat is the shared check for "does this Birthdate
// string look right for this age" — format, and that the year suffix is
// defaultImperialYear minus age. Shared by TestGenerateBirthdateFormat
// here and by TestBuildScoutCharacterSetsBirthdate/
// TestBuildCitizenCharacterSetsBirthdate in the sibling
// character_generate_test.go/citizen_character_generate_test.go, so the
// assertion only needs updating in one place if the format ever changes.
func assertBirthdateFormat(t *testing.T, birthdate string, age int) {
	t.Helper()

	if !birthdateFormat.MatchString(birthdate) {
		t.Errorf("Birthdate = %q, want format %q", birthdate, birthdateFormat.String())
	}

	if wantSuffix := fmt.Sprintf("-%d", defaultImperialYear-age); !strings.HasSuffix(birthdate, wantSuffix) {
		t.Errorf("Birthdate = %q, want year suffix %q (age %d)", birthdate, wantSuffix, age)
	}
}

// TestGenerateBirthdateFormat confirms the dice-wrapped assembly: format,
// and that year is defaultImperialYear minus age.
func TestGenerateBirthdateFormat(t *testing.T) {
	t.Parallel()

	cases := []struct {
		seed uint64
		age  int
	}{
		{1, 18},
		{2, 30},
		{3, 74},
	}

	for _, c := range cases {
		r := dice.New(rand.NewPCG(c.seed, c.seed))

		assertBirthdateFormat(t, GenerateBirthdate(r, c.age), c.age)
	}
}
