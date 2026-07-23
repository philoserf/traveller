package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/philoserf/traveller/api"
	"github.com/philoserf/traveller/ehex"
)

func doRequest(t *testing.T, mux *http.ServeMux, target string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	return rec
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	rec := doRequest(t, api.NewMux(), "/healthz")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if body.Status != "ok" {
		t.Errorf("status field = %q, want %q", body.Status, "ok")
	}
}

func TestWorldsRandom(t *testing.T) {
	t.Parallel()

	rec := doRequest(t, api.NewMux(), "/worlds/random")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got api.WorldResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if got.Seed == 0 {
		t.Error("Seed = 0, want a resolved (non-zero) seed")
	}

	if len(got.UWP) != 9 { // StSAHPGL-T: 7 digits, a dash, tech level
		t.Fatalf("UWP = %q, want a 9-character StSAHPGL-T code", got.UWP)
	}

	for i, c := range got.UWP {
		if i == 7 {
			if c != '-' {
				t.Errorf("UWP[%d] = %q, want '-'", i, c)
			}

			continue
		}

		if _, err := ehex.Parse(string(c)); err != nil && i != 0 {
			t.Errorf("UWP digit %q at position %d is not a valid ehex digit: %v", c, i, err)
		}
	}
}

func TestWorldsRandomSeedReproducible(t *testing.T) {
	t.Parallel()

	mux := api.NewMux()

	rec1 := doRequest(t, mux, "/worlds/random?seed=12345")
	rec2 := doRequest(t, mux, "/worlds/random?seed=12345")

	var w1, w2 api.WorldResponse
	if err := json.Unmarshal(rec1.Body.Bytes(), &w1); err != nil {
		t.Fatalf("unmarshal response 1: %v", err)
	}

	if err := json.Unmarshal(rec2.Body.Bytes(), &w2); err != nil {
		t.Fatalf("unmarshal response 2: %v", err)
	}

	if w1.UWP != w2.UWP {
		t.Errorf("same seed produced different UWPs: %q vs %q", w1.UWP, w2.UWP)
	}

	if len(w1.TradeCodes) != len(w2.TradeCodes) {
		t.Errorf("same seed produced different trade code counts: %v vs %v", w1.TradeCodes, w2.TradeCodes)
	}
}

func TestWorldsRandomBadSeed(t *testing.T) {
	t.Parallel()

	rec := doRequest(t, api.NewMux(), "/worlds/random?seed=notanumber")

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
