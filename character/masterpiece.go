package character

import "github.com/philoserf/traveller/dice"

// QREBS is Book 1's own five-dimension description of an object:
// "Objects are (or can be) described in the QREBS system to indicate
// their level of Quality, Reliability, Ease of Use, Burden, and Safety."
//
// Note what the B and S stand for. #35 described them as "Boldness of
// design" and "Style"; the book's own two expansions agree with each
// other and with neither — "Quality, Reliability, Ease Of Use, Bulk (or
// Burden) and Safety".
type QREBS struct {
	Quality     int
	Reliability int
	EaseOfUse   int
	Burden      int
	Safety      int
}

// Values returns the five dimensions in QREBS order.
func (q QREBS) Values() []int {
	return []int{q.Quality, q.Reliability, q.EaseOfUse, q.Burden, q.Safety}
}

// Book 1's own QREBS allocation scale: "Allocate the Master Points to
// QREBS (for the ranges -5 to +5, -5 = 1 point; +5 = 11 points)."
//
// So a dimension set to v costs v + qrebsPointOffset points, every
// dimension costs at least one, and the five together cost at most 55.
//
// That maximum is not a coincidence worth passing over: 5 x 11 = 55 is
// exactly Book 1's own Perfect Masterpiece threshold ("A Perfect
// Masterpiece has 55 or more Master Points"). A Perfect Masterpiece is
// precisely one with enough points to set every QREBS dimension to its
// maximum — which is what makes the next sentence necessary at all: "If
// all QREBS values are set at the Maximum, excess Master Points can be
// allocated equally in excess of +5".
const (
	qrebsMin         = -5
	qrebsMax         = 5
	qrebsPointOffset = 6 // v + 6 is v's cost: -5 costs 1, +5 costs 11
	qrebsDimensions  = 5
)

// qrebsMinTotal and qrebsMaxTotal are the cost of setting every
// dimension to its minimum and maximum.
const (
	qrebsMinTotal = qrebsDimensions * (qrebsMin + qrebsPointOffset) // 5
	qrebsMaxTotal = qrebsDimensions * (qrebsMax + qrebsPointOffset) // 55
)

// allocateQREBS distributes masterPoints across the five dimensions.
//
// Book 1 leaves the split to the player — it says only "Allocate the
// Master Points to QREBS" — so this resolves it the way this codebase
// resolves every other open player choice, through the dice roller. The
// allocation is random rather than optimizing because, unlike a Begin
// target or a Controlling Characteristic, there is no better or worse
// here: QREBS dimensions are descriptive, and a heavy, unsafe, superbly
// made object is a legitimate Masterpiece.
//
// Every dimension starts at the minimum, since the scale has no zero
// cost, and the remaining points are handed out one at a time to
// dimensions not yet at the maximum. Points beyond qrebsMaxTotal go to
// the "allocated equally in excess of +5" case: they are spread evenly
// across all five, which is what "equally" means and why the remainder
// is what the excess loop finally distributes.
func allocateQREBS(r *dice.Roller, masterPoints int) QREBS {
	values := make([]int, qrebsDimensions)
	for i := range values {
		values[i] = qrebsMin
	}

	remaining := masterPoints - qrebsMinTotal

	// Fill toward the maximum, one point at a time, choosing among the
	// dimensions that can still take one.
	for remaining > 0 {
		available := make([]int, 0, qrebsDimensions)

		for i, v := range values {
			if v < qrebsMax {
				available = append(available, i)
			}
		}

		if len(available) == 0 {
			break
		}

		values[available[r.Uniform(len(available))-1]]++
		remaining--
	}

	// "If all QREBS values are set at the Maximum, excess Master Points
	// can be allocated equally in excess of +5."
	if remaining > 0 {
		each := remaining / qrebsDimensions
		for i := range values {
			values[i] += each
		}
	}

	return QREBS{
		Quality:     values[0],
		Reliability: values[1],
		EaseOfUse:   values[2],
		Burden:      values[3],
		Safety:      values[4],
	}
}

// Masterpiece is one object a Craftsman created, structured rather than
// recorded as a narrative string — the "structured item record" #35
// identified as the missing prerequisite for QREBS and Vintage.
type Masterpiece struct {
	// MasterPoints is the score the Masterpiece was created against.
	MasterPoints int
	// Perfect is Book 1's own "A Perfect Masterpiece has 55 or more
	// Master Points".
	Perfect bool
	// QREBS is the allocation of those points across the five dimensions.
	QREBS QREBS
	// CreatedAtAge is the character's age when the Masterpiece was
	// finished — the time reference Vintage appreciation needs, and the
	// other half of what #35 said this codebase lacked.
	CreatedAtAge int
	// BaseValue is the sale price at creation, before any Vintage
	// appreciation.
	BaseValue int
}

// newMasterpiece builds the record for a successful creation.
func newMasterpiece(r *dice.Roller, masterPoints, ageAtCreation int) Masterpiece {
	return Masterpiece{
		MasterPoints: masterPoints,
		Perfect:      masterPoints >= craftsmanPerfectMasterPoints,
		QREBS:        allocateQREBS(r, masterPoints),
		CreatedAtAge: ageAtCreation,
		BaseValue:    craftsmanMasterpieceValue(masterPoints),
	}
}

// vintageAppreciationPercentPerYear is Book 1's own "A Masterpiece
// increases in value about 1% per year (simple interest)". Simple, not
// compound, is stated outright and is what VintageValue implements.
const vintageAppreciationPercentPerYear = 1

// VintageValue is the Masterpiece's own worth after it has aged
// ageNow-CreatedAtAge years:
//
//	p.75: "Vintage Masterpieces. A Masterpiece increases in value about
//	1% per year (simple interest), but are subject to Flux (in percent)
//	when sold."
//
// The Flux half is deliberately not applied here. It is explicitly a
// sale-time roll ("when sold"), not a property of the object, so
// applying it at generation would bake one particular sale into the
// character record and make the same Masterpiece worth different
// amounts on different reads. VintageValueAtSale is the function that
// rolls it.
//
// Ages before creation return the base value rather than depreciating:
// a Masterpiece cannot be worth less than new.
func (m Masterpiece) VintageValue(ageNow int) int {
	years := ageNow - m.CreatedAtAge
	if years <= 0 {
		return m.BaseValue
	}

	return m.BaseValue + m.BaseValue*vintageAppreciationPercentPerYear*years/100
}

// VintageValueAtSale applies p.75's own sale-time Flux to the aged
// value: "subject to Flux (in percent) when sold". Flux is a signed
// roll, so a sale can come in under the vintage value as easily as over
// it.
//
// Clamped at zero: Flux ranges to -5, which cannot take a Masterpiece
// below worthless however briefly it has been held.
func (m Masterpiece) VintageValueAtSale(r *dice.Roller, ageNow int) int {
	value := m.VintageValue(ageNow)

	return max(0, value+value*r.Flux()/100)
}

// masterpiecesFromTerms collects the structured record of every
// Masterpiece a career created, in the order they were made. Terms whose
// creation attempt failed carry no record — Book 1 p.75: "If The
// Creation Fails, the Piece (not Masterpiece) is flawed and worthless".
func masterpiecesFromTerms(terms []Term) []Masterpiece {
	var out []Masterpiece

	for _, t := range terms {
		if t.Masterpiece != nil {
			out = append(out, *t.Masterpiece)
		}
	}

	return out
}
