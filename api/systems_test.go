package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/philoserf/traveller/api"
)

func TestSystemsRandom(t *testing.T) {
	t.Parallel()

	rec := doRequest(t, api.NewMux(), "/systems/random")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got api.SystemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if got.Seed == 0 {
		t.Error("Seed = 0, want a resolved (non-zero) seed")
	}

	if len(got.StarGroups) == 0 {
		t.Fatal("StarGroups is empty, want at least the Primary")
	}

	if len(got.Mainworld.UWP) != 9 {
		t.Errorf("Mainworld.UWP = %q, want a 9-character StSAHPGL-T code", got.Mainworld.UWP)
	}

	if !isValidTravelZone(got.Mainworld.TravelZone) {
		t.Errorf("Mainworld.TravelZone = %q, want one of Green/Amber/Red", got.Mainworld.TravelZone)
	}

	nilOrbitCount := 0

	for _, g := range got.StarGroups {
		if g.Star.Orbit == nil {
			nilOrbitCount++
		}
	}

	if nilOrbitCount != 1 {
		t.Errorf("found %d stars with a nil Orbit, want exactly 1 (the Primary)", nilOrbitCount)
	}
}

func TestSystemsRandomSatelliteShape(t *testing.T) {
	t.Parallel()

	// Seed 5 is known (from #3's manual verification) to produce a
	// mainworld that is a Satellite of a Gas Giant.
	rec := doRequest(t, api.NewMux(), "/systems/random?seed=5")

	var got api.SystemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if !got.Mainworld.Satellite {
		t.Fatalf(
			"seed 5: Mainworld.Satellite = false, want true (regression: verify against cmd/sysgen -seed 5 if this changes)",
		)
	}

	if got.Mainworld.GasGiant == nil {
		t.Error("Mainworld.Satellite is true but GasGiant is nil, want it set")
	}

	if got.Mainworld.AU != 0 {
		t.Errorf("Mainworld.AU = %v for a Satellite orbit, want 0 (omitted)", got.Mainworld.AU)
	}

	// Seed 5's mainworld is a Close satellite (system.TestGenerateSystemPreservesMainworldSatelliteCloseFar
	// pins the same seed for the same reason on the domain-model side).
	if !got.Mainworld.Close {
		t.Error(
			"seed 5: Mainworld.Close = false, want true (regression: verify against cmd/sysgen -seed 5 if this changes)",
		)
	}
}

func TestSystemsRandomSeedReproducible(t *testing.T) {
	t.Parallel()

	mux := api.NewMux()

	rec1 := doRequest(t, mux, "/systems/random?seed=12345")
	rec2 := doRequest(t, mux, "/systems/random?seed=12345")

	var s1, s2 api.SystemResponse
	if err := json.Unmarshal(rec1.Body.Bytes(), &s1); err != nil {
		t.Fatalf("unmarshal response 1: %v", err)
	}

	if err := json.Unmarshal(rec2.Body.Bytes(), &s2); err != nil {
		t.Fatalf("unmarshal response 2: %v", err)
	}

	if s1.Mainworld.UWP != s2.Mainworld.UWP {
		t.Errorf("same seed produced different mainworld UWPs: %q vs %q", s1.Mainworld.UWP, s2.Mainworld.UWP)
	}

	if len(s1.StarGroups) != len(s2.StarGroups) {
		t.Errorf("same seed produced different star counts: %d vs %d", len(s1.StarGroups), len(s2.StarGroups))
	}
}

// TestSystemsRandomBodiesNestUnderTheirStarAndSatellitesNestUnderBodies
// pins seed 1 (known, via go run ./cmd/sysgen -seed 1, to place
// satellites on Orbit 0, 5, and 12 of its single Primary group) —
// confirming bodies decode nested under their hosting star's group, and
// that a body with satellites carries them nested rather than flattened
// into a sibling list (the shape toSystemResponse replaced).
func TestSystemsRandomBodiesNestUnderTheirStarAndSatellitesNestUnderBodies(t *testing.T) {
	t.Parallel()

	rec := doRequest(t, api.NewMux(), "/systems/random?seed=1")

	var got api.SystemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(got.StarGroups) != 1 {
		t.Fatalf(
			"seed 1: StarGroups has %d entries, want 1 (regression: verify against cmd/sysgen -seed 1 if this changes)",
			len(got.StarGroups),
		)
	}

	bodiesWithSatellites := 0

	for _, body := range got.StarGroups[0].Bodies {
		if len(body.Satellites) > 0 {
			bodiesWithSatellites++
		}
	}

	if bodiesWithSatellites != 3 {
		t.Errorf(
			"seed 1: %d bodies carry nested Satellites, want 3 (orbit 0's world, orbit 5's Gas Giant, orbit 12's world)",
			bodiesWithSatellites,
		)
	}
}

// TestSystemsRandomMultiStarBodiesGroupCorrectly pins seed 3 (known, via
// go run ./cmd/sysgen -seed 3, to produce a Primary/Near/Far system with
// bodies hosted by all three, and a Ring on the Primary's orbit-2 Gas
// Giant) — confirming toSystemResponse's role-keyed grouping is exercised
// across multiple stars at the API layer (not just the single-star case
// TestSystemsRandomBodiesNestUnderTheirStarAndSatellitesNestUnderBodies
// covers), and that Ring survives the JSON round-trip.
func TestSystemsRandomMultiStarBodiesGroupCorrectly(t *testing.T) {
	t.Parallel()

	rec := doRequest(t, api.NewMux(), "/systems/random?seed=3")

	var got api.SystemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(got.StarGroups) != 3 {
		t.Fatalf(
			"seed 3: StarGroups has %d entries, want 3 (regression: verify against cmd/sysgen -seed 3 if this changes)",
			len(got.StarGroups),
		)
	}

	bodiesByRole := map[string]int{}
	ringFound := false

	for _, g := range got.StarGroups {
		bodiesByRole[g.Star.Role] = len(g.Bodies)

		for _, b := range g.Bodies {
			if b.Ring {
				ringFound = true
			}
		}
	}

	for _, role := range []string{"Primary", "Near", "Far"} {
		if bodiesByRole[role] == 0 {
			t.Errorf("seed 3: %s group has no bodies, want at least one", role)
		}
	}

	if !ringFound {
		t.Error("seed 3: no body in any group has Ring set, want the Primary's orbit-2 Gas Giant to carry one")
	}
}

// TestSystemsRandomExactlyOneBodyIsMainworld confirms the mainworld's own
// Orbit — no longer excluded from StarGroups (world.SystemBodies) — is
// marked IsMainworld exactly once across the whole response, whether it's
// a freestanding body (seed 1) or itself a satellite of a Gas Giant
// (seed 5, per TestSystemsRandomSatelliteShape).
func TestSystemsRandomExactlyOneBodyIsMainworld(t *testing.T) {
	t.Parallel()

	for _, seed := range []string{"1", "5"} {
		t.Run("seed="+seed, func(t *testing.T) {
			t.Parallel()

			rec := doRequest(t, api.NewMux(), "/systems/random?seed="+seed)

			var got api.SystemResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal response: %v", err)
			}

			if count := countMainworldEntries(got); count != 1 {
				t.Errorf("seed %s: found %d IsMainworld entries across StarGroups, want exactly 1", seed, count)
			}
		})
	}
}

// countMainworldEntries counts how many bodies or satellites across every
// StarGroup carry IsMainworld — the mainworld's own Orbit should appear,
// and be marked, exactly once.
func countMainworldEntries(got api.SystemResponse) int {
	count := 0

	for _, g := range got.StarGroups {
		for _, b := range g.Bodies {
			if b.IsMainworld {
				count++
			}

			for _, sat := range b.Satellites {
				if sat.IsMainworld {
					count++
				}
			}
		}
	}

	return count
}

func TestSystemsRandomBadSeed(t *testing.T) {
	t.Parallel()

	rec := doRequest(t, api.NewMux(), "/systems/random?seed=notanumber")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if body.Error == "" {
		t.Error("error field is empty, want a message")
	}
}
