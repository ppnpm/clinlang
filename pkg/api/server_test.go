package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"clinlang/pkg/api"
	"clinlang/pkg/auth"
)

func newLocalServer(t *testing.T) *httptest.Server {
	t.Helper()
	// Sandbox persistence so PUT /workspace doesn't write to the
	// developer's real ~/.config/clinlang/config.json.
	t.Setenv("CLINLANG_CONFIG_DIR", t.TempDir())
	root := t.TempDir()
	cfg := api.Config{
		Mode:          api.ModeLocal,
		BindAddr:      "127.0.0.1:0",
		WorkspaceRoot: root,
	}
	srv, err := api.NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func newHostedServer(t *testing.T) *httptest.Server {
	t.Helper()
	t.Setenv("CLINLANG_CONFIG_DIR", t.TempDir())
	root := t.TempDir()
	cfg := api.Config{
		Mode:          api.ModeHosted,
		BindAddr:      "127.0.0.1:0",
		WorkspaceRoot: root,
	}
	srv, err := api.NewServer(cfg)
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func mustPostJSON(t *testing.T, url string, body interface{}, headers map[string]string) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	return resp
}

func TestHealth_IsPublic(t *testing.T) {
	ts := newHostedServer(t)
	resp, err := http.Get(ts.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("hosted /health should be public (no auth), got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "disclaimer") {
		t.Errorf("/health body should contain disclaimer, got %s", string(body))
	}
}

func TestParse_LocalMode(t *testing.T) {
	ts := newLocalServer(t)
	resp := mustPostJSON(t, ts.URL+"/api/v1/parse", map[string]string{"input": "pt 40M"}, nil)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("status: want 200, got %d", resp.StatusCode)
	}
}

func TestParse_HostedMode_RequiresAuth(t *testing.T) {
	ts := newHostedServer(t)
	// No X-Forwarded-User header → 401.
	resp := mustPostJSON(t, ts.URL+"/api/v1/parse", map[string]string{"input": "pt 40M"}, nil)
	defer resp.Body.Close()
	if resp.StatusCode != 401 {
		t.Errorf("hosted no auth: status: want 401, got %d", resp.StatusCode)
	}
}

func TestParse_HostedMode_WithAuth(t *testing.T) {
	ts := newHostedServer(t)
	resp := mustPostJSON(t, ts.URL+"/api/v1/parse",
		map[string]string{"input": "pt 40M"},
		map[string]string{auth.HostedUserHeader: "dr.smith"})
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("hosted with auth: want 200, got %d body=%s", resp.StatusCode, body)
	}
}

func TestSOAP_MarkersOptIn(t *testing.T) {
	ts := newLocalServer(t)
	// Default (markers off)
	resp := mustPostJSON(t, ts.URL+"/api/v1/soap",
		map[string]interface{}{"input": "pt 40M\nvitals hr150"}, nil)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(body), "Notes (out of ref)") {
		t.Errorf("SOAP default should NOT include marker section; got %s", string(body))
	}

	// With markers
	resp2 := mustPostJSON(t, ts.URL+"/api/v1/soap",
		map[string]interface{}{"input": "pt 40M\nvitals hr150", "markers": true}, nil)
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)
	if !strings.Contains(string(body2), "Notes (out of ref)") {
		t.Errorf("SOAP with markers should include marker section; got %s", string(body2))
	}
}

func TestValidate_DeprecationHeader(t *testing.T) {
	ts := newLocalServer(t)
	resp := mustPostJSON(t, ts.URL+"/api/v1/validate",
		map[string]string{"input": "pt 40M"}, nil)
	defer resp.Body.Close()
	if resp.Header.Get("Deprecation") == "" {
		t.Errorf("/validate should emit Deprecation header")
	}
}

func TestCORS_LocalAllowsLocalhostOrigin(t *testing.T) {
	ts := newLocalServer(t)
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("local mode should echo localhost Origin, got %q", got)
	}
}

func TestCORS_HostedNoHeaders(t *testing.T) {
	ts := newHostedServer(t)
	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/health", nil)
	req.Header.Set("Origin", "http://evil.example.com")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("hosted mode should not emit CORS headers, got %q", got)
	}
}
