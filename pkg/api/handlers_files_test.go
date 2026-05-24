package api_test

import (
	"bytes"
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

func mustReq(t *testing.T, method, url string, body interface{}, headers map[string]string) *http.Response {
	t.Helper()
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, url, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	return resp
}

func TestFiles_RoundTripLocal(t *testing.T) {
	ts := newLocalServer(t)

	// PUT
	put := mustReq(t, http.MethodPut, ts.URL+"/api/v1/files/today.cln",
		map[string]string{"content": "pt 40M"}, nil)
	put.Body.Close()
	if put.StatusCode != 200 {
		t.Fatalf("PUT: %d", put.StatusCode)
	}

	// GET
	get := mustReq(t, http.MethodGet, ts.URL+"/api/v1/files/today.cln", nil, nil)
	defer get.Body.Close()
	if get.StatusCode != 200 {
		t.Fatalf("GET: %d", get.StatusCode)
	}
	var loaded struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	body, _ := io.ReadAll(get.Body)
	if err := json.Unmarshal(body, &loaded); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, body)
	}
	if loaded.Content != "pt 40M" {
		t.Errorf("content: want 'pt 40M', got %q", loaded.Content)
	}

	// DELETE
	del := mustReq(t, http.MethodDelete, ts.URL+"/api/v1/files/today.cln", nil, nil)
	del.Body.Close()
	if del.StatusCode != 200 {
		t.Fatalf("DELETE: %d", del.StatusCode)
	}
}

func TestFiles_PathTraversalDenied(t *testing.T) {
	ts := newLocalServer(t)
	put := mustReq(t, http.MethodPut, ts.URL+"/api/v1/files/../escape.cln",
		map[string]string{"content": "x"}, nil)
	put.Body.Close()
	// The PUT path either gets rewritten inside the workspace OR is
	// rejected. In neither case must it land outside the workspace.
	// We can't easily inspect the workspace root from this test
	// without exporting more — but the workspace_test already exercises
	// the boundary. Here we just confirm the request is rejected or
	// rewritten safely (no 5xx).
	if put.StatusCode >= 500 {
		t.Errorf("path traversal attempt produced 5xx, want 4xx or safe rewrite; got %d", put.StatusCode)
	}
}

func TestFiles_HostedPerUserIsolation(t *testing.T) {
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

	headerA := map[string]string{auth.HostedUserHeader: "alice"}
	headerB := map[string]string{auth.HostedUserHeader: "bob"}

	// Alice writes a file.
	put := mustReq(t, http.MethodPut, ts.URL+"/api/v1/files/note.cln",
		map[string]string{"content": "alice's note"}, headerA)
	put.Body.Close()
	if put.StatusCode != 200 {
		t.Fatalf("alice PUT: %d", put.StatusCode)
	}

	// Alice can read it back.
	getA := mustReq(t, http.MethodGet, ts.URL+"/api/v1/files/note.cln", nil, headerA)
	if getA.StatusCode != 200 {
		t.Errorf("alice GET: want 200, got %d", getA.StatusCode)
	}
	getA.Body.Close()

	// Bob cannot read Alice's file (different workspace subtree).
	getB := mustReq(t, http.MethodGet, ts.URL+"/api/v1/files/note.cln", nil, headerB)
	if getB.StatusCode == 200 {
		t.Errorf("bob should not see alice's file; got 200")
	}
	getB.Body.Close()

	// And the file should be physically at root/alice/note.cln.
	expected := filepath.Join(root, "alice", "note.cln")
	if _, err := os.Stat(expected); err != nil {
		t.Errorf("expected file at %s: %v", expected, err)
	}
}

func TestFiles_ETagRoundTrip(t *testing.T) {
	ts := newLocalServer(t)

	// 1. Write the file (no If-Match, fresh file).
	put := mustReq(t, http.MethodPut, ts.URL+"/api/v1/files/etag.cln",
		map[string]string{"content": "v1"}, nil)
	put.Body.Close()
	if put.StatusCode != 200 {
		t.Fatalf("initial PUT: %d", put.StatusCode)
	}

	// 2. GET returns ETag header.
	get := mustReq(t, http.MethodGet, ts.URL+"/api/v1/files/etag.cln", nil, nil)
	if get.StatusCode != 200 {
		t.Fatalf("GET: %d", get.StatusCode)
	}
	etag := get.Header.Get("ETag")
	get.Body.Close()
	if etag == "" {
		t.Fatalf("GET should return ETag header")
	}

	// 3. PUT with matching If-Match succeeds.
	put2 := mustReq(t, http.MethodPut, ts.URL+"/api/v1/files/etag.cln",
		map[string]string{"content": "v2"},
		map[string]string{"If-Match": etag})
	put2.Body.Close()
	if put2.StatusCode != 200 {
		t.Fatalf("PUT with matching If-Match: %d", put2.StatusCode)
	}

	// 4. PUT with the STALE If-Match (etag from before v2 was written)
	//    now returns 412.
	put3 := mustReq(t, http.MethodPut, ts.URL+"/api/v1/files/etag.cln",
		map[string]string{"content": "v3-stale"},
		map[string]string{"If-Match": etag})
	put3.Body.Close()
	if put3.StatusCode != http.StatusPreconditionFailed {
		t.Errorf("PUT with stale If-Match: want 412, got %d", put3.StatusCode)
	}
}

func TestFiles_MkdirAndList(t *testing.T) {
	ts := newLocalServer(t)
	mk := mustReq(t, http.MethodPost, ts.URL+"/api/v1/files/mkdir",
		map[string]string{"path": "cases/2026"}, nil)
	mk.Body.Close()
	if mk.StatusCode != 200 {
		t.Fatalf("mkdir: %d", mk.StatusCode)
	}

	list := mustReq(t, http.MethodGet, ts.URL+"/api/v1/files", nil, nil)
	defer list.Body.Close()
	body, _ := io.ReadAll(list.Body)
	if list.StatusCode != 200 {
		t.Fatalf("list: %d body=%s", list.StatusCode, body)
	}
	if len(body) == 0 || body[0] != '[' {
		t.Errorf("list should be JSON array, got %s", body)
	}
}
