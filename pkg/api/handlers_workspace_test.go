package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"clinlang/pkg/api"
	"clinlang/pkg/auth"
)

func TestWorkspace_GetLocal(t *testing.T) {
	ts := newLocalServer(t)
	resp, err := http.Get(ts.URL + "/api/v1/workspace")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var out struct {
		Path string `json:"path"`
		Mode string `json:"mode"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Mode != "local" {
		t.Errorf("mode: want 'local', got %q", out.Mode)
	}
	if out.Path == "" {
		t.Errorf("path should be populated")
	}
}

func TestWorkspace_PutLocal_SwitchesRoot(t *testing.T) {
	ts := newLocalServer(t)

	// Pick a brand-new temp dir as the new workspace.
	newRoot := filepath.Join(t.TempDir(), "switched")

	body, _ := json.Marshal(map[string]string{"path": newRoot})
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/workspace",
		newReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("PUT status: %d body=%s", resp.StatusCode, b)
	}

	// Now writing a file should land in the new root.
	put := mustReq(t, http.MethodPut, ts.URL+"/api/v1/files/in-new.cln",
		map[string]string{"content": "after-switch"}, nil)
	put.Body.Close()
	if put.StatusCode != 200 {
		t.Fatalf("file PUT after switch: %d", put.StatusCode)
	}
	if _, err := os.Stat(filepath.Join(newRoot, "in-new.cln")); err != nil {
		t.Errorf("expected file at new root, got: %v", err)
	}
}

func TestWorkspace_PutHostedForbidden(t *testing.T) {
	t.Setenv("CLINLANG_CONFIG_DIR", t.TempDir())
	root := t.TempDir()
	cfg := api.Config{
		Mode:          api.ModeHosted,
		BindAddr:      "127.0.0.1:0",
		WorkspaceRoot: root,
	}
	srv, err := api.NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	body, _ := json.Marshal(map[string]string{"path": "/tmp/whatever"})
	req, _ := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/workspace",
		newReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(auth.HostedUserHeader, "dr.smith")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("hosted PUT: want 403, got %d", resp.StatusCode)
	}
}

// newReader is a tiny adapter to keep the body-creation calls terse.
func newReader(b []byte) *byteReader { return &byteReader{b: b} }

type byteReader struct {
	b []byte
	i int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}
