package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"clinlang/pkg/autocomplete"
)

func TestAutocomplete_SXCommand(t *testing.T) {
	ts := newLocalServer(t)
	resp := mustPostJSON(t, ts.URL+"/api/v1/autocomplete",
		map[string]string{"command": "sx", "query": "ches"}, nil)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	var out []autocomplete.Suggestion
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, body)
	}
	if len(out) == 0 {
		t.Errorf("expected at least one suggestion for 'ches', got 0")
	}
}

func TestAutocomplete_RXDelegatesToDrugs(t *testing.T) {
	ts := newLocalServer(t)
	resp := mustPostJSON(t, ts.URL+"/api/v1/autocomplete",
		map[string]string{"command": "rx", "query": "para"}, nil)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status: %d", resp.StatusCode)
	}
}

func TestAutocomplete_UnknownCommandEmpty(t *testing.T) {
	ts := newLocalServer(t)
	resp := mustPostJSON(t, ts.URL+"/api/v1/autocomplete",
		map[string]string{"command": "zzz", "query": "x"}, nil)
	defer resp.Body.Close()
	var out []autocomplete.Suggestion
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, body)
	}
	if len(out) != 0 {
		t.Errorf("unknown command should return empty list, got %d items", len(out))
	}
}

func TestDrugs_PrefixQuery(t *testing.T) {
	ts := newLocalServer(t)
	resp, err := http.Get(ts.URL + "/api/v1/drugs?prefix=para")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status: %d", resp.StatusCode)
	}
}

func TestPlugins_ListShape(t *testing.T) {
	ts := newLocalServer(t)
	resp, err := http.Get(ts.URL + "/api/v1/plugins")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 || body[0] != '[' {
		t.Errorf("plugins response should be a JSON array, got %s", body)
	}
}
