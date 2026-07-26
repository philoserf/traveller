package character

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
	"github.com/philoserf/traveller/world"
)

// LandGrant is Book 1 p.88's own award of territory — "awards of
// territory given by the government in recognition of political power or
// important works", made by the Emperor to Nobles (as part of a Patent
// of Nobility) and to Explorer-Discoverers (in recognition of their
// discoveries).
//
// Retained at Mustering Out per p.68: "Any character who has received a
// Land Grant retains it at Mustering Out".
type LandGrant struct {
	// Source is what earned the grant — a noble title ("Baron") or
	// "Discovery".
	Source string
	// World is the UWP of the world the grant sits on. Book 1 gives the
	// Imperium a preference for particular worlds per rank (p.88's own
	// "Preferred World" column), which this codebase records but cannot
	// act on: it has no star map to search for a matching world.
	World string
	// TradeCodes are the granted world's own Trade Classifications, which
	// are what the grant's income is computed from.
	TradeCodes []string
	// MainworldHexes and OtherHexes are p.88's own "HEXES MW / other"
	// columns: "For each hex on a mainworld, a noble is also allocated
	// one hex on a non-mainworld in the same system."
	MainworldHexes int
	OtherHexes     int
	// AnnualIncome is the grant's own yearly profit in Credits, per
	// p.88's "LAND GRANT VALUE".
	AnnualIncome int
}

// Book 1 p.88's own "LAND GRANT VALUE": "An unimproved Land Grant
// generates income based on the Trade Classifications of the world and
// equal to Cr10,000 per TC annually (equal to Cr5,000 if there are no
// TCs)", stated per-hex in the same page's own "Each Hex generates a
// profit equal to Cr10,000 per Trade Classification per year (a Hex with
// no TC generates Cr5,000 annually)".
const (
	landGrantIncomePerTradeCode = 10_000
	landGrantIncomeNoTradeCodes = 5_000
)

// landGrantHexIncome is one hex's own annual profit.
func landGrantHexIncome(tradeCodes int) int {
	if tradeCodes == 0 {
		return landGrantIncomeNoTradeCodes
	}

	return tradeCodes * landGrantIncomePerTradeCode
}

// landGrantAnnualIncome totals a grant's own yearly profit. Mainworld
// hexes earn against the granted world's own Trade Classifications;
// companion hexes are on a non-mainworld in the same system, which
// p.88's own worked example treats as having none.
//
// That example is what pins this function: "recently knighted Sir
// Richard of Hefry (Trade Classifications Ni Va) has a Land Grant of one
// Terrain Hex on Hefry producing an income of Cr20,000 annually, and a
// companion Land Grant on a minor world (no Trade Classifications)
// elsewhere in the system producing Cr5,000 annually." A Knight's own
// row is MW 1 / other 1, and Hefry has two TCs: 1x(2xCr10,000) +
// 1xCr5,000 = Cr25,000, exactly the two figures the book prints.
func landGrantAnnualIncome(mainworldHexes, otherHexes, tradeCodes int) int {
	return mainworldHexes*landGrantHexIncome(tradeCodes) +
		otherHexes*landGrantIncomeNoTradeCodes
}

// nobleRank is one row of Book 1 p.88's own "NOBLE LAND GRANTS" table
// (reprinted on p.85 as "NOBLE RANK AND LAND GRANTS" without the voting
// columns) — the single table behind both Noble rank titles and Noble
// Land Grants, which is why they are implemented together.
type nobleRank struct {
	// Code is the table's own Soc column as printed. It is a string, not
	// an ehex.Value, because the table distinguishes lowercase from
	// uppercase at the same Soc: c Baronet and C Baron are both Soc 12,
	// e Viscount and E Count are both Soc 14, and f (Lesser) Duke and F
	// (Greater) Duke are both Soc 15.
	Code string
	// Soc is Code's own numeric value. Not unique across rows — see Code.
	Soc ehex.Value
	// Title is the rank's own name.
	Title string
	// Where is p.88's own "Where?" column: the scope within which the
	// grant's hexes are allocated.
	Where string
	// PreferredWorld is p.88's own "Preferred World" column. Recorded but
	// not acted on: "The Imperium prefers grants on worlds with high
	// potential for development", a preference this codebase has no star
	// map to satisfy.
	PreferredWorld string
	// MainworldHexes and OtherHexes are the "HEXES MW / other" columns.
	MainworldHexes int
	OtherHexes     int
	// Votes and ProxyValue are p.88's own Moot voting columns, recorded
	// for completeness. Proxies are post-generation politics, not a
	// CharGen mechanic, so nothing reads ProxyValue yet.
	Votes      int
	ProxyValue int
}

// nobleRanks is Book 1 p.88's own table, in printed order.
//
// The column boundaries were recovered from the extraction's own
// character offsets rather than read off visually: the page interleaves
// the adjacent "NOBLE PROXIES" body text into every row, and the
// Gentleman row is missing its "Where?" cell, which makes the remaining
// cells look shifted. By offset, "Where?" sits at column 41, "Preferred
// World" ends at 75, the two hex columns at 76-89, Votes at 94 and Total
// Value at 102 — placing Gentleman's own two "any"s in Preferred World
// and MW, and its lone 1 in the "other" column.
//
// Gentleman is therefore transcribed as MW "any" — the one cell in the
// table that is not a hex count. Read here as "no mainworld hex is
// allocated": a Gentleman (Soc A) sits below the Soc B+ the Noble career
// itself requires to Begin, and the row grants a single companion hex
// with no world restriction; recorded as MainworldHexes 0 rather than
// guessed upward, so the one genuinely ambiguous cell in the table
// resolves to the smaller claim.
var nobleRanks = []nobleRank{
	{"A", 10, "Gentleman", "", "any", 0, 1, 0, 0},
	{"B", 11, "Knight", "homeworld", "any", 1, 1, 0, 0},
	{"c", 12, "Baronet", "one system", "Pre-Ag or Pre-Ri", 2, 2, 1, 100_000},
	{"C", 12, "Baron", "one system", "Ag or Ri", 4, 4, 1, 100_000},
	{"D", 13, "Marquis", "one subsector", "Pre-Ind", 8, 8, 2, 200_000},
	{"e", 14, "Viscount", "one subsector", "Pre-Hi", 16, 16, 3, 300_000},
	{"E", 14, "Count", "one sector", "In or Hi", 32, 32, 3, 300_000},
	{"f", 15, "Duke", "one sector", "Importance 4+, but not a Capital", 64, 64, 4, 400_000},
	{"F", 15, "Duke", "one sector", "Subsector or Sector Capital", 128, 128, 4, 400_000},
	{"G", 16, "Archduke", "one domain", "any", 256, 256, 5, 500_000},
	{"H", 17, "Imperial Family", "in the empire", "any", 256, 256, 6, 600_000},
	{"H", 17, "Emperor", "in the empire", "any", 256, 256, 0, 0},
}

// nobleRankIndexForSoc is Book 1 p.65's own "Nobles begin with rank
// equal to their Social Standing" — the row a character enters the
// career at.
//
// Where two rows share a Soc, this returns the *first* — Baronet over
// Baron, Viscount over Count, Lesser Duke over Greater. That is the
// lower of the two, which is where a character beginning at that Soc
// starts and elevates up from; nobleRankAfterElevation walks to the
// second row as its own step.
//
// Returns false below Soc A, which is not a noble rank at all. Soc above
// the table's own top row clamps to the last non-Emperor row: GenerateUPP
// cannot produce it, but Mustering Out's own Soc awards and a Knighthood
// can carry a character past Soc 17, and Emperor is not a rank anyone is
// elevated into.
func nobleRankIndexForSoc(soc ehex.Value) (int, bool) {
	best, found := 0, false

	for i, rank := range nobleRanks {
		if rank.Soc > soc || rank.Title == "Emperor" {
			continue
		}

		// The first row at a given Soc wins, so a later row of equal Soc
		// does not overwrite it.
		if found && nobleRanks[best].Soc == rank.Soc {
			continue
		}

		best, found = i, true
	}

	return best, found
}

// nobleRankAfterElevation is p.65's own "Elevated to the next higher
// Noble rank and its associated increase in Social Standing (if any)".
//
// The "(if any)" is load-bearing and is why this walks the table by row
// rather than incrementing Soc: c Baronet and C Baron are both Soc 12,
// e Viscount and E Count both Soc 14, f and F Duke both Soc 15. Elevating
// from Baronet to Baron is a real rank step that raises no
// characteristic, and treating elevation as "Soc +1" would skip those
// three rows entirely while inflating Soc past what the table allows.
//
// Returns the index of the new row and whether Soc actually rose —
// which is what p.85's own "Each increase in Soc during CharGen awards a
// Land Grant" keys off. Elevation past the top of the ladder returns the
// current index unchanged.
//
// The index is returned rather than the rank itself because two adjacent
// rows share the title "Duke": a caller comparing titles to decide
// whether it moved would silently skip the Lesser-to-Greater Duke step,
// which raises no Soc and changes no name but is a real elevation.
func nobleRankAfterElevation(current int) (int, bool) {
	next := current + 1
	for next < len(nobleRanks) && nobleRanks[next].Title == "Emperor" {
		next++
	}

	if next >= len(nobleRanks) {
		return current, false
	}

	return next, nobleRanks[next].Soc > nobleRanks[current].Soc
}

// nobleRankForSoc returns the rank a Soc value confers.
func nobleRankForSoc(soc ehex.Value) (nobleRank, bool) {
	i, ok := nobleRankIndexForSoc(soc)
	if !ok {
		return nobleRank{}, false
	}

	return nobleRanks[i], true
}

// NobleTitleForSoc is Book 1 p.88's own Soc-to-title mapping. Returns
// the empty string below Soc A, which is not a noble rank at all.
func NobleTitleForSoc(soc ehex.Value) string {
	rank, ok := nobleRankForSoc(soc)
	if !ok {
		return ""
	}

	return rank.Title
}

// newLandGrant builds a grant on w, sized by the given hex counts.
func newLandGrant(source string, w world.World, mainworldHexes, otherHexes int) LandGrant {
	codes := world.TradeCodeStrings(w.TradeCodes)

	return LandGrant{
		Source:         source,
		World:          w.UWP.String(),
		TradeCodes:     codes,
		MainworldHexes: mainworldHexes,
		OtherHexes:     otherHexes,
		AnnualIncome:   landGrantAnnualIncome(mainworldHexes, otherHexes, len(codes)),
	}
}

// discoveryLandGrantHexes is the size of a Scout's own Discovery grant.
//
// Book 1 p.88 says only that a discoverer's grant "operates (and
// produces income) much like a Noble grant" and never prints a hex count
// for one, so this takes the smallest allocation the Noble table
// actually grants a mainworld hex for — Knight's own MW 1 / other 1.
// Sized down rather than up, on the same reasoning as the Gentleman row
// above: where Book 1 declines to say, claim less.
const (
	discoveryLandGrantMainworldHexes = 1
	discoveryLandGrantOtherHexes     = 1
)

// newDiscoveryLandGrant is p.88's own "DISCOVERER LAND GRANTS": "The
// Imperium makes Land Grants to the scout discoverers of new worlds."
// The world is freshly generated because that is what a Discovery is —
// p.79's own Reward Success is "the Scout discovers a valuable new world
// or a valuable feature on a known world".
func newDiscoveryLandGrant(w world.World) LandGrant {
	return newLandGrant("Discovery", w, discoveryLandGrantMainworldHexes, discoveryLandGrantOtherHexes)
}

// newNobleLandGrant is p.85's own "Land Grants. Each increase in Soc
// during CharGen awards a Land Grant", sized by the row the new Soc
// lands on. Returns false below Soc A, where no rank — and so no
// fief — exists.
func newNobleLandGrant(soc ehex.Value, w world.World) (LandGrant, bool) {
	rank, ok := nobleRankForSoc(soc)
	if !ok {
		return LandGrant{}, false
	}

	return newLandGrant(rank.Title, w, rank.MainworldHexes, rank.OtherHexes), true
}

// totalLandGrantIncome sums the annual profit of every grant a character
// holds. Book 1 p.88: "Land Grants Are Cumulative. Each title confers
// its own Land Grant: a Knight raised to Baronet receives it in addition
// to his Knighthood".
func totalLandGrantIncome(grants []LandGrant) int {
	total := 0

	for _, grant := range grants {
		total += grant.AnnualIncome
	}

	return total
}

// scoutDiscoveryLandGrants is Book 1 p.79's own Reward Success: the
// Scout "discovers a valuable new world or a valuable feature on a known
// world (a Discovery), receives a Land Grant, and Fame +1" — one grant
// per Discovery, on a freshly generated world, since a Discovery is by
// definition a world not previously on the charts.
//
// The dice are drawn here rather than during the career loop so a term
// that discovered nothing costs no rolls, keeping the stream aligned
// with careers that have no Discovery mechanic at all.
func scoutDiscoveryLandGrants(r *dice.Roller, career Career) []LandGrant {
	var grants []LandGrant

	for _, t := range career.Terms {
		if t.RewardResult == "Discovery" {
			grants = append(grants, newDiscoveryLandGrant(world.Generate(r)))
		}
	}

	return grants
}
