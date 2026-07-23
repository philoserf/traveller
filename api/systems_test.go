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

	if len(got.Stars) == 0 {
		t.Fatal("Stars is empty, want at least the Primary")
	}

	if len(got.Mainworld.UWP) != 9 {
		t.Errorf("Mainworld.UWP = %q, want a 9-character StSAHPGL-T code", got.Mainworld.UWP)
	}

	if !isValidTravelZone(got.Mainworld.TravelZone) {
		t.Errorf("Mainworld.TravelZone = %q, want one of Green/Amber/Red", got.Mainworld.TravelZone)
	}

	nilOrbitCount := 0

	for _, s := range got.Stars {
		if s.Orbit == nil {
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

	// Seed 5's mainworld is a Close satellite (world.TestGenerateSystemPreservesMainworldSatelliteCloseFar
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

	if len(s1.Stars) != len(s2.Stars) {
		t.Errorf("same seed produced different star counts: %d vs %d", len(s1.Stars), len(s2.Stars))
	}
}

func TestSystemsRandomOtherBodyHostRole(t *testing.T) {
	t.Parallel()

	rec := doRequest(t, api.NewMux(), "/systems/random?seed=12345")

	var got api.SystemResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	validRoles := map[string]bool{"Primary": true, "Close": true, "Near": true, "Far": true}

	for _, ob := range got.OtherBodies {
		if !validRoles[ob.HostRole] {
			t.Errorf("orbit %d: HostRole = %q, want one of Primary/Close/Near/Far", ob.Orbit, ob.HostRole)
		}
	}
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
