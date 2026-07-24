package system

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
