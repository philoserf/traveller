package character

import "github.com/philoserf/traveller/dice"

// citizenMusterOutMoney is Book 1 p.78's Citizen "D MUSTER OUT" table,
// Money column (1D-indexed, row 1 at index 0). "Low/Middle/High Passage"
// spelled out from the box's own "Low/Mid/High Psg", matching the exact
// expansion already established for Scout's identical abbreviations
// (career_muster_out.go's scoutMusterOutMoney).
var citizenMusterOutMoney = [11]string{
	"Low Passage", "Low Passage", "Middle Passage", "High Passage",
	"Cr15,000", "StarPass", "Cr25,000", "Cr30,000", "Cr35,000", "Cr40,000", "Cr50,000",
}

// citizenMusterOutBenefits is Book 1 p.78's Citizen "D MUSTER OUT" table,
// Benefits column. "TAS Fellow Membership" spelled out from "TAS Fellow",
// matching Scout's own identical expansion (career_muster_out.go's
// scoutMusterOutBenefits). "Cx +1"/"Str +1" keep the box's own
// characteristic-boost shorthand, matching every skill table's existing
// convention.
var citizenMusterOutBenefits = [11]string{
	"Str +1", "Str +1", "Wafer Jack", "C5 +1",
	"Str +1", "C2 +1", "C3 +1", "Int +1", "Soc +1", "TAS Fellow Membership", "Ship Share",
}

// ResolveCitizenMusterOut resolves Book 1 p.78's Citizen Mustering Out
// table (step E, p.57 — a distinct step following Career Resolution; see
// ResolveScoutMusterOut's own doc comment for why Mustering Out stays a
// separate function from the career loop, not fused in). Not called from
// ResolveCitizenCareer; a caller invokes this explicitly once a Career is
// final.
//
// One roll per term served (p.68's "One Per Term"), with no doubling or
// zeroing: unlike Scout, Citizen Life has no wound or death mechanic at
// all (established across every prior Citizen slice), so neither p.65's
// Disability Muster Out ("Double Benefits") nor p.69's "Dying During
// Character Generation" (zero rolls, void the attempt) can ever apply
// here — the roll count is simply len(career.Terms), with no
// branching logic that would justify its own helper function.
//
// Each roll independently picks the Money or Benefits column via a
// uniform random pick (p.68, the same open-player-choice resolution
// ResolveScoutMusterOut already uses), then rolls rollCitizenMusterOutRow
// with dm = len(career.Terms) (p.78's "DM +Terms +Terms" — uniform for
// both columns, unlike Scout's Benefits-column +Fame/2, so there's no
// Fame-dependent gap to document here).
func ResolveCitizenMusterOut(r *dice.Roller, career Career) MusteringOut {
	var out MusteringOut

	dm := len(career.Terms)

	for range len(career.Terms) {
		appendMusterOutRoll(r, &out, dm, citizenMusterOutMoney[:], citizenMusterOutBenefits[:])
	}

	// p.70: "Any character who has been a Citizen or Functionary for at
	// least one Term is eligible for a Citizen's Pension."
	if len(career.Terms) > 0 {
		applyPension(&out, citizenPensionRate)
	}

	return out
}
