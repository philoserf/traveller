package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/philoserf/traveller/api"
)

func TestSectorsRandom(t *testing.T) {
	t.Parallel()

	rec := doRequest(t, api.NewMux(), "/sectors/random")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got api.SectorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if got.Name != "Unnamed" {
		t.Errorf("Name = %q, want %q (default)", got.Name, "Unnamed")
	}

	if len(got.Hexes) != 1280 {
		t.Fatalf("len(Hexes) = %d, want 1280", len(got.Hexes))
	}

	populated := 0

	for _, h := range got.Hexes {
		if h.System != nil {
			populated++
		}
	}

	if populated == 0 {
		t.Error("no populated hexes in a 1280-hex Standard-density sector, want at least one")
	}
}

func TestSectorsRandomNameAndDensity(t *testing.T) {
	t.Parallel()

	rec := doRequest(t, api.NewMux(), "/sectors/random?name=Deneb&density=Core&seed=1")

	var got api.SectorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if got.Name != "Deneb" {
		t.Errorf("Name = %q, want %q", got.Name, "Deneb")
	}

	populated := 0

	for _, h := range got.Hexes {
		if h.System != nil {
			populated++
		}
	}

	// Core is 2D6<=11 (97.2% true rate — see world.densityTable's own doc
	// comment on the book's inconsistent 91% figure), so all but a
	// handful of 1280 hexes should be populated.
	if populated < 1000 {
		t.Errorf("Core density populated %d/1280 hexes, want a large majority", populated)
	}
}

func TestSectorsRandomBadDensity(t *testing.T) {
	t.Parallel()

	rec := doRequest(t, api.NewMux(), "/sectors/random?density=Nonexistent")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestSectorsRandomBadSubsector(t *testing.T) {
	t.Parallel()

	for _, subsector := range []string{"Z", "AB", "1"} {
		rec := doRequest(t, api.NewMux(), "/sectors/random?subsector="+subsector)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("subsector=%q: status = %d, want %d", subsector, rec.Code, http.StatusBadRequest)
		}
	}
}

func TestSectorsRandomSubsectorFilters(t *testing.T) {
	t.Parallel()

	rec := doRequest(t, api.NewMux(), "/sectors/random?subsector=A&seed=1")

	var got api.SectorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if len(got.Hexes) != 80 {
		t.Fatalf("len(Hexes) = %d, want 80 (one Subsector)", len(got.Hexes))
	}

	if got.Hexes[0].Location != "0101" || got.Hexes[len(got.Hexes)-1].Location != "0810" {
		t.Errorf("Hexes[0]/[last].Location = %q/%q, want \"0101\"/\"0810\"",
			got.Hexes[0].Location, got.Hexes[len(got.Hexes)-1].Location)
	}
}

func TestSectorsRandomSeedReproducible(t *testing.T) {
	t.Parallel()

	mux := api.NewMux()

	rec1 := doRequest(t, mux, "/sectors/random?seed=99")
	rec2 := doRequest(t, mux, "/sectors/random?seed=99")

	if rec1.Body.String() != rec2.Body.String() {
		t.Error("same seed produced different sector responses")
	}
}
