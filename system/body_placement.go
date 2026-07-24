package system

import (
	"github.com/philoserf/traveller/dice"
	"github.com/philoserf/traveller/ehex"
	"github.com/philoserf/traveller/world"
)

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
