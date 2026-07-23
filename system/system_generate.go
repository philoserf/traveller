package system

import (
	"slices"

	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
	"github.com/philoserf/traveller/world"
)

// primaryOrbitNumber is the sentinel Orbit.Number for the Primary star
// itself — "Primary = at center of System" (Book 3 p.28), not a numbered
// planetary orbit around anything. Distinct from every real orbit number
// (0-20 per orbitAUTable), which are all non-negative.
const primaryOrbitNumber = -1

// rollStar rolls a star's SpectralType, SpectralDecimal, LuminosityClass,
// and HabitableZoneOrbit. isPrimary=true rolls two fresh Flux values (one
// for type, one for size) — Book 3 p.28 Table 2: "Roll Flux for Primary."
// isPrimary=false derives both from primaryFlux (the Primary's own type
// roll) instead: "For all others, Primary Flux + (1D-1)" for type,
// "Primary Flux + (1D+2)" for size — this applies to Close/Near/Far stars
// and to any Companion, not just companions specifically ("for all
// others" reads as everyone but the Primary).
func rollStar(r *dice.Roller, primaryFlux int, isPrimary bool) Star {
	typeFlux := primaryFlux
	sizeFlux := r.Flux()

	if !isPrimary {
		typeFlux = primaryFlux + (r.D6() - 1)
		sizeFlux = primaryFlux + (r.D6() + 2)
	}

	t := rollSpectralType(r, typeFlux)

	var decimal int
	if t != SpectralDegenerate {
		decimal = r.Uniform(10) - 1
	}

	size := rollLuminosityClass(sizeFlux, t)

	star := Star{SpectralType: t, SpectralDecimal: decimal, LuminosityClass: size}

	// Believed unreachable given how rollLuminosityClass and
	// habitableZoneTable interlock (see habitableZoneOrbit's doc
	// comment) — HabitableZoneOrbit is left at its zero value rather
	// than guessing at a fallback for a combination this project has no
	// data for.
	if hz, ok := habitableZoneOrbit(t, size); ok {
		star.HabitableZoneOrbit = hz
	}

	return star
}

// attachCompanion rolls a Flux; if it meets starPresenceFlux, rolls and
// returns a Companion star sharing role, per Table 1's "Flux for
// Companions for each Star present." Shared by every star-rolling site
// (Primary, and each Close/Near/Far via rollAndPlaceStar) so this rule
// has exactly one implementation.
func attachCompanion(r *dice.Roller, primaryFlux int, role StellarRole) *Star {
	if r.Flux() < starPresenceFlux {
		return nil
	}

	companion := rollStar(r, primaryFlux, false)
	companion.Role = role

	return &companion
}

// rollAndPlaceStar rolls a Close/Near/Far star (and, per attachCompanion,
// its own optional Companion), returning it as an Orbit at orbitNumber.
func rollAndPlaceStar(r *dice.Roller, primaryFlux int, role StellarRole, orbitNumber int) Orbit {
	star := rollStar(r, primaryFlux, false)
	star.Role = role
	star.Companion = attachCompanion(r, primaryFlux, role)

	return Orbit{Number: orbitNumber, AU: orbitAU(orbitNumber), Star: &star}
}

// GenerateSystem builds a StarSystem around an already-generated
// mainworld (from Generate): rolls the Primary star (and, independently,
// whether a Close/Near/Far star and any Companions exist), computes each
// star's HZ orbit and precluded-orbit ceiling (for oversized stars —
// precludedOrbitHost), places the mainworld (as a Planet, as a Satellite
// of a freshly rolled Gas Giant, or — if the mainworld is itself an
// Asteroid Belt — via the Belt placement roll instead of HZ+Var), then
// rolls and places every other Gas Giant, Belt, and secondary world
// across all stars (placeGasGiants/placeBelts/placeOtherWorlds), each
// with its own satellites and Rings, and merges the newly-derivable
// orbit-dependent trade codes (DeriveOrbitTradeCodes) into every placed
// world's TradeCodes.
func GenerateSystem(r *dice.Roller, mainworld world.World) StarSystem {
	primaryFlux := r.Flux()
	primary := rollStar(r, primaryFlux, true)
	primary.Role = Primary
	primary.Companion = attachCompanion(r, primaryFlux, Primary)

	orbits := []Orbit{{Number: primaryOrbitNumber, Star: &primary}}

	if r.Flux() >= starPresenceFlux {
		orbits = append(orbits, rollAndPlaceStar(r, primaryFlux, Close, r.D6()-1))
	}

	if r.Flux() >= starPresenceFlux {
		orbits = append(orbits, rollAndPlaceStar(r, primaryFlux, Near, 5+r.D6()))
	}

	if r.Flux() >= starPresenceFlux {
		orbits = append(orbits, rollAndPlaceStar(r, primaryFlux, Far, 11+r.D6()))
	}

	orbits, mainworldOrbitIndex := placeMainworld(r, orbits, primary, mainworld)
	satelliteOfGasGiant := orbits[mainworldOrbitIndex].Satellite
	mw := *orbits[mainworldOrbitIndex].World

	// Book 3 p.29's "W Worlds" formula (Total Worlds = MW + GG + Belts +
	// 2D) needs Gas Giant / Belt counts as inputs, not outputs — they're
	// mw.PBG.GasGiants/.Belts, already rolled by Generate (rollPBG's own
	// doc comment: "describe the whole system, not just this world").
	// Phase 1 generated them but never consulted them for placement.
	gasGiantsToPlace := int(mw.PBG.GasGiants)

	initialSGGCount := 0

	if satelliteOfGasGiant {
		// The satellite-hosting Gas Giant already placed above counts
		// against this total, per P1's own sequence ("Place Mainworld"
		// — including its satellite GG — immediately followed by "Place
		// Gas Giants" for the rest). PBG.GasGiants can still be 0 even
		// though a satellite mainworld always gets one placed anyway (a
		// real edge case — see placeMainworld's doc comment) — clamped
		// here so gasGiantsToPlace never goes negative.
		gasGiantsToPlace = max(gasGiantsToPlace-1, 0)

		// The GG-vs-SGG "every second SGG converts to an IG" counter
		// (placeGasGiants) needs to know about this one too, or it
		// mis-numbers every SGG rolled after it.
		if gg := orbits[mainworldOrbitIndex-1].GasGiant; gg != nil && gg.Bracket == "SGG" {
			initialSGGCount = 1
		}
	}

	maxPopulation := ehex.Value(0)
	if mw.UWP.Population > 0 {
		maxPopulation = mw.UWP.Population - 1 // Book 3 p.29: "Subject to: Max Pop= MW Pop - 1"
	}

	hosts := availableHosts(orbits)
	placeGasGiants(r, &orbits, hosts, gasGiantsToPlace, initialSGGCount)
	placeBelts(r, &orbits, hosts, int(mw.PBG.Belts), maxPopulation)
	placeOtherWorlds(r, &orbits, hosts, r.TwoD6(), maxPopulation)

	// Every top-level body — the mainworld, its host Gas Giant if any,
	// and everything just placed above — gets its own satellite roll
	// (Book 3 p.29: "For Each World in the System" + Gas Giants) — except
	// Asteroid Belts, which the satellite-count table (Gas Giants/Inners/
	// Hospitables/Outers) has no row for at all. Snapshotting the current
	// top-level bodies first, rather than ranging over orbits live, is
	// what keeps newly-appended satellites from being mistaken for more
	// top-level bodies to recurse into — satellites don't get their own
	// satellites.
	topLevel := make([]Orbit, 0, len(orbits))

	for _, o := range orbits {
		if o.Satellite {
			continue
		}

		if o.GasGiant != nil || (o.World != nil && !slices.Contains(o.World.TradeCodes, world.AsteroidBelt)) {
			topLevel = append(topLevel, o)
		}
	}

	for _, parent := range topLevel {
		generateSatellitesForBody(r, &orbits, parent, parent.HostHZOrbit, maxPopulation)
	}

	return StarSystem{
		Orbits:         orbits,
		MainworldOrbit: mainworldOrbitIndex,
	}
}

// placeMainworld places mainworld into orbits (which already holds the
// system's stars): as a Planet, as a Satellite of a freshly rolled Gas
// Giant, as a BigWorld (if Table 2C says Satellite but the system's own
// rolled PBG.GasGiants is 0 — Book 3 p.24's "If Satellite and No Giants,
// place a BigWorld in MW Orbit"), or — if mainworld is an Asteroid Belt —
// via the Belt placement roll instead of HZ+Var. Merges the
// newly-derivable orbit-dependent trade codes (DeriveOrbitTradeCodes)
// into the mainworld's own copy. Returns the updated orbits and the
// index of the mainworld's own Orbit entry within it — when that entry's
// Satellite is true, the immediately preceding orbits entry is its host
// Gas Giant (see the two-Orbit-append below).
func placeMainworld(r *dice.Roller, orbits []Orbit, primary Star, mainworld world.World) ([]Orbit, int) {
	hzOrbit := primary.HabitableZoneOrbit
	mw := mainworld

	var (
		orbitNumber int
		kind        = mainworldPlanet
	)

	if slices.Contains(mw.TradeCodes, world.AsteroidBelt) {
		// "If the Mainworld is an Asteroid Belt, it is placed using the
		// Belt Column of the Basic Placement Chart without regard to
		// Habitable Zone" — skips Table 2B's HZ+Var roll entirely.
		orbitNumber = hzOrbit + rollBeltOffset(r)
	} else {
		dm := 0

		switch primary.SpectralType { //nolint:exhaustive // only M/O/B carry a DM (Table 2B); everything else is +0
		case SpectralM:
			dm = 2
		case SpectralO, SpectralB:
			dm = -2
		}

		orbitNumber = hzOrbit + mainworldHZVar(r.Flux()+dm)
		kind = rollMainworldPlacementKind(r.Flux())
	}

	// "If Satellite and No Giants, place a BigWorld in MW Orbit" (Book 3
	// p.24) — the system's own rolled Gas Giant count can be 0 even when
	// Table 2C's own roll says this mainworld orbits one; when both are
	// true, regenerate the mainworld as a BigWorld and place it as an
	// ordinary planet instead of manufacturing a Gas Giant that would
	// contradict PBG. Only the fields GenerateWithSize itself computes
	// are overwritten — Name/Sector/Hex/Nobility/Allegiance/Worlds/Notes/
	// Ring (left zero by Generate, but not necessarily by an arbitrary
	// caller of GenerateSystem) are preserved rather than wholesale-
	// replaced. An Asteroid Belt mainworld can never reach here — kind
	// only gets set to a satellite kind in the non-belt branch above.
	if kind != mainworldPlanet && mw.PBG.GasGiants == 0 {
		bigWorld := world.GenerateWithSize(r, rollBigWorldSize)
		mw.UWP = bigWorld.UWP
		mw.TradeCodes = bigWorld.TradeCodes
		mw.TravelZone = bigWorld.TravelZone
		mw.Bases = bigWorld.Bases
		mw.PBG = bigWorld.PBG
		mw.Importance = bigWorld.Importance
		mw.Economic = bigWorld.Economic
		mw.Cultural = bigWorld.Cultural
		kind = mainworldPlanet
	}

	// HZVar/the Belt roll can both go negative enough to land below orbit
	// 0 (e.g. an M-type primary with hzOrbit=0 and a negative HZVar) —
	// floored here since orbit 0 is the innermost real orbit, and a
	// negative number would otherwise collide with primaryOrbitNumber's
	// own sentinel value and fall outside orbitAUTable's range.
	orbitNumber = max(orbitNumber, 0)

	// The computed number can independently coincide with a Close/Near/Far
	// star's own orbit, or fall inside the Primary's own photosphere
	// (both are computed separately, with nothing ruling out a match) —
	// nudge via the same collision/preclusion handling placeInOrbit gives
	// every other placement. If nothing in range is free (practically
	// impossible), keep the original number rather than leave the
	// mainworld unplaced.
	primaryHost := starHost{
		role: Primary, hzOrbit: hzOrbit, minOrbit: precludedOrbitHost(primary), maxOrbit: primaryMaxOrbit,
	}
	if n, ok := placeInOrbit(orbits, primaryHost, orbitNumber); ok {
		orbitNumber = n
	}

	mw.TradeCodes = append(mw.TradeCodes, world.DeriveOrbitTradeCodes(mw.UWP, orbitNumber, hzOrbit, true)...)

	if kind != mainworldPlanet {
		gg := rollGasGiant(r)
		orbits = append(
			orbits,
			Orbit{
				Number:      orbitNumber,
				AU:          orbitAU(orbitNumber),
				HostHZOrbit: hzOrbit,
				HostRole:    Primary,
				GasGiant:    &gg,
			},
		)
		orbits = append(orbits, Orbit{
			Number: orbitNumber, Satellite: true, Close: kind == mainworldCloseSatellite, World: &mw,
		})
	} else {
		orbits = append(
			orbits,
			Orbit{Number: orbitNumber, AU: orbitAU(orbitNumber), HostHZOrbit: hzOrbit, HostRole: Primary, World: &mw},
		)
	}

	return orbits, len(orbits) - 1
}

// primaryMaxOrbit is the Primary's own orbit-number ceiling: "The Primary
// Star may have orbits out to Orbit-19" (Book 3 p.21) reads as inclusive
// of 20 once cross-checked against orbitAUTable's own documented 0-20
// range — treated here as 20, not 19, so the table's full range is
// actually reachable.
const primaryMaxOrbit = 20

// starHost is one candidate star a non-mainworld body can be placed
// around: its own StellarRole (for Orbit.HostRole), HabitableZoneOrbit,
// the lowest orbit number it can host (minOrbit — see precludedOrbitHost),
// and the highest (maxOrbit). Close/Near/Far stars "may fill orbits
// around them to their own Orbit minus 3" (Book 3 p.21) — maxOrbit can
// come out negative for a Close/Near/Far star in a low orbit itself ("A
// Close Star in Orbit 2 can have no Planet Orbits"), which placeInOrbit's
// range check handles by simply never finding a free slot.
type starHost struct {
	role     StellarRole
	hzOrbit  int
	minOrbit int
	maxOrbit int
}

// precludedOrbitHost returns minOrbit for star: one past whatever orbit
// its own photosphere engulfs (Book 3 p.21's "Precluded Orbits" —
// precludedOrbitCeiling), or 0 if star's luminosity class never precludes
// any orbit at all.
func precludedOrbitHost(star Star) int {
	if ceiling, ok := precludedOrbitCeiling(star.SpectralType, star.SpectralDecimal, star.LuminosityClass); ok {
		return ceiling + 1
	}

	return 0
}

// availableHosts returns every star in orbits as a placement candidate,
// in the order they appear (Primary first, then Close/Near/Far — the
// order GenerateSystem appends them in).
func availableHosts(orbits []Orbit) []starHost {
	var hosts []starHost

	for i := range orbits {
		if orbits[i].Star == nil {
			continue
		}

		maxOrbit := primaryMaxOrbit
		if orbits[i].Number != primaryOrbitNumber {
			maxOrbit = orbits[i].Number - 3
		}

		hosts = append(hosts, starHost{
			role:     orbits[i].Star.Role,
			hzOrbit:  orbits[i].Star.HabitableZoneOrbit,
			minOrbit: precludedOrbitHost(*orbits[i].Star),
			maxOrbit: maxOrbit,
		})
	}

	return hosts
}

// orbitOccupied reports whether any non-Satellite entry in orbits already
// uses number — Satellite entries are exempt since they're expected to
// share a Number with their parent (see Orbit's doc comment).
func orbitOccupied(orbits []Orbit, number int) bool {
	for _, o := range orbits {
		if !o.Satellite && o.Number == number {
			return true
		}
	}

	return false
}

// placeInOrbit finds the free orbit within [host.minOrbit, host.maxOrbit]
// closest to candidate, per P2's own note ("If an orbit is duplicated or
// precluded, adjust to an adjacent or the closest possible orbit") —
// scanning the whole range rather than only forward from candidate,
// since a free orbit behind candidate is still "the closest possible" if
// nothing closer exists ahead of it. Both senses of "precluded" apply:
// already occupied (orbitOccupied), or physically inside host's own
// photosphere (host.minOrbit, from precludedOrbitHost). ok is false if
// no free slot exists anywhere in range — the caller skips this body
// rather than force an invalid placement. host.maxOrbit staying small (at
// most 20) keeps this full-range scan cheap.
func placeInOrbit(orbits []Orbit, host starHost, candidate int) (int, bool) {
	best, found, bestDist := 0, false, 0

	for n := host.minOrbit; n <= host.maxOrbit; n++ {
		if orbitOccupied(orbits, n) {
			continue
		}

		dist := candidate - n
		if dist < 0 {
			dist = -dist
		}

		if !found || dist < bestDist {
			best, found, bestDist = n, true, dist
		}
	}

	return best, found
}

// placeGasGiants places count Gas Giants, rotating through hosts (Book 3
// p.21: "place the first of the worlds concerned in orbit around the
// Primary, the second... around the Close..."). Every second SGG rolled
// converts to an IG (Ice Giant), per the GG table's own note.
// initialSGGCount carries forward the SGG/LGG split of any Gas Giant
// already placed elsewhere in the system (the satellite-mainworld host,
// placed in placeMainworld before this function ever runs) — without it,
// this function's own counter would start over at 0 and mis-number every
// SGG it rolls relative to the system as a whole.
func placeGasGiants(r *dice.Roller, orbits *[]Orbit, hosts []starHost, count, initialSGGCount int) {
	if len(hosts) == 0 {
		return
	}

	sggCount := initialSGGCount

	for i := range count {
		host := hosts[i%len(hosts)]
		gg := rollGasGiant(r)

		var offset int

		switch gg.Bracket {
		case "SGG":
			sggCount++

			if sggCount%2 == 0 {
				gg.Bracket = "IG"
				offset = rollIGOffset(r)
			} else {
				offset = rollSGGOffset(r)
			}
		default: // "LGG"
			offset = rollLGGOffset(r)
		}

		if n, ok := placeInOrbit(*orbits, host, host.hzOrbit+offset); ok {
			*orbits = append(
				*orbits,
				Orbit{Number: n, AU: orbitAU(n), HostHZOrbit: host.hzOrbit, HostRole: host.role, GasGiant: &gg},
			)
		}
	}
}

// placeBelts places count Planetoid Belts, rotating through hosts. Each
// gets its own World (Book 3 p.29: "Planetoids= St000PGL-T").
func placeBelts(r *dice.Roller, orbits *[]Orbit, hosts []starHost, count int, maxPopulation ehex.Value) {
	if len(hosts) == 0 {
		return
	}

	for i := range count {
		host := hosts[i%len(hosts)]

		n, ok := placeInOrbit(*orbits, host, host.hzOrbit+rollBeltOffset(r))
		if !ok {
			continue
		}

		u := generatePlanetoidWorld(r, maxPopulation)
		w := worldWithTradeCodes(u, n, host.hzOrbit)
		*orbits = append(
			*orbits,
			Orbit{Number: n, AU: orbitAU(n), HostHZOrbit: host.hzOrbit, HostRole: host.role, World: &w},
		)
	}
}

// placeOtherWorlds places count secondary worlds, rotating through hosts.
// All but the last use P2's World1 column for their orbit; the last uses
// World2 (farther out) — Book 3 p.29's P1 step: "Place Other Worlds...
// using P2 World1 Column. Last World, place using P2 World2 Column."
// Each world's category (Inferno/InnerWorld/BigWorld/StormWorld/RadWorld/
// Hospitable/Worldlet/Iceworld) comes from its resolved orbit's position
// relative to its host's own HZ orbit, not the mainworld's.
func placeOtherWorlds(r *dice.Roller, orbits *[]Orbit, hosts []starHost, count int, maxPopulation ehex.Value) {
	if len(hosts) == 0 {
		return
	}

	for i := range count {
		host := hosts[i%len(hosts)]

		candidate := rollWorld1Orbit(r)
		if i == count-1 {
			candidate = rollWorld2Orbit(r)
		}

		n, ok := placeInOrbit(*orbits, host, candidate)
		if !ok {
			continue
		}

		category := rollSecondaryWorldCategory(r, n-host.hzOrbit)
		u := generateSecondaryWorldUWP(r, category, maxPopulation)
		w := worldWithTradeCodes(u, n, host.hzOrbit)
		*orbits = append(
			*orbits,
			Orbit{Number: n, AU: orbitAU(n), HostHZOrbit: host.hzOrbit, HostRole: host.role, World: &w},
		)
	}
}

// worldWithTradeCodes builds a secondary World from u, deriving both its
// UWP-only and orbit-dependent trade codes (DeriveTradeCodes,
// DeriveOrbitTradeCodes — the same functions the mainworld uses,
// isMainworld=false here). Ix/Ex/Cx extensions, Bases, and PBG are
// mainworld-only concerns (Book 3 p.27: Ix/Ex apply to the entire
// system, computed once already) — left zero-valued, not omitted by
// oversight.
func worldWithTradeCodes(u world.UWP, orbit, hzOrbit int) world.World {
	tradeCodes := world.DeriveTradeCodes(u)
	tradeCodes = append(tradeCodes, world.DeriveOrbitTradeCodes(u, orbit, hzOrbit, false)...)

	return world.World{UWP: u, TradeCodes: tradeCodes}
}
