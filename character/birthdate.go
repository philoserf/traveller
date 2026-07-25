package character

import (
	"fmt"

	"github.com/philoserf/traveller/dice"
)

// defaultImperialYear is the "current adventuring year" used to compute a
// Birthdate from Age. Not a RAW-mandated value — Book 1 p.262 leaves the
// campaign's current date to the Referee — but the exact year used in
// p.263's own fully-worked Birthdate example ("Sigg... Age=30... current
// adventuring year is 1105"), and also listed there as "The Golden Age"
// era boundary.
const defaultImperialYear = 1105

// weekdayNames are p.263's own calendar names for days 2-365 (7-day
// cycle); day 1 of every year is the separately-named "Holiday".
var weekdayNames = [7]string{"Wonday", "Tuday", "Thirday", "Forday", "Fiday", "Sixday", "Senday"}

// birthDayOfYear is rollBirthDayOfYear's own dice-free decision: given
// four D6 results, returns the day of the year (1-365) and whether the
// roll was valid, per Book 1 p.263's "Birth Date Generation" table
// (verified directly against the page image — collapses to this formula;
// see this slice's own plan-file Context for the full derivation and its
// confirmation against the book's two worked examples, including the
// "Reroll!" case).
func birthDayOfYear(die1, die2, die3, die4 int) (int, bool) {
	k := (die2-1)*6 + die3

	if die1 <= 3 {
		if die4 == 6 {
			return 0, false
		}

		return k + 36*(die4-1), true
	}

	if die4 <= 5 {
		return 180 + k + 36*(die4-1), true
	}

	if k >= 32 {
		return k + 329, true
	}

	return 0, false
}

// rollBirthDayOfYear rerolls all four dice until birthDayOfYear reports a
// valid day, per Book 1 p.263's own "Reroll!" instruction.
func rollBirthDayOfYear(r *dice.Roller) int {
	for {
		if day, ok := birthDayOfYear(r.D6(), r.D6(), r.D6(), r.D6()); ok {
			return day
		}
	}
}

// weekdayName names day (1-365) per p.263's own calendar: day 1 is
// "Holiday"; days 2-365 cycle through weekdayNames.
func weekdayName(day int) string {
	if day == 1 {
		return "Holiday"
	}

	return weekdayNames[(day-2)%7]
}

// GenerateBirthdate produces a Birthdate string ("Wonday 044-1075") for a
// character of the given age, per Book 1 p.263. year is
// defaultImperialYear minus age — see this function's own doc comment on
// why that base year isn't RAW-mandated.
func GenerateBirthdate(r *dice.Roller, age int) string {
	day := rollBirthDayOfYear(r)

	return fmt.Sprintf("%s %03d-%d", weekdayName(day), day, defaultImperialYear-age)
}
