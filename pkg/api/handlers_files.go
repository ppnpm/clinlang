package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"clinlang/pkg/auth"
	"clinlang/pkg/workspace"
)

// etagFor returns a short content-hash ETag, quoted per RFC 7232.
// We use the first 16 hex chars of SHA-256 — plenty of uniqueness for
// concurrency detection, short enough to log without clutter.
func etagFor(content string) string {
	sum := sha256.Sum256([]byte(content))
	return `"` + hex.EncodeToString(sum[:8]) + `"`
}

// errWorkspaceUnconfigured is returned when a file operation is
// attempted before a workspace root has been set. The HTTP layer
// translates this to 409 Conflict.
var errWorkspaceUnconfigured = errors.New("workspace not configured")

// filesHandler reads the LIVE workspace root from workspaceState on
// every request so that a successful PUT /api/v1/workspace takes
// effect immediately for subsequent file operations.
type filesHandler struct {
	cfg Config
	ws  *workspaceState
}

// fsForUser returns a LocalFileSystem rooted at the per-user workspace.
// In local mode every user (always "local") shares the configured root.
// In hosted mode each user gets a subdirectory under the root.
//
// Returns errWorkspaceUnconfigured when the workspace root has not yet
// been set (only possible in local mode, on first launch). The HTTP
// handler maps that to 409 Conflict so the frontend knows to show the
// welcome dialog.
func (h *filesHandler) fsForUser(user string) (*workspace.LocalFileSystem, error) {
	root := h.ws.Get()
	if root == "" {
		return nil, errWorkspaceUnconfigured
	}
	if h.cfg.Mode == ModeHosted {
		// Defence in depth: never let a username path component
		// escape the workspace root.
		clean := filepath.Clean("/" + user)
		root = filepath.Join(root, clean)
	}
	return workspace.NewLocalFileSystem(root)
}

// handleRoot serves GET /api/v1/files (the tree root) and the
// POST verbs that don't have a path in the URL: /files/rename and
// /files/mkdir.
//
// http.ServeMux dispatches /api/v1/files (no slash) here; /api/v1/files/*
// goes to handleSubpath.
func (h *filesHandler) handleRoot(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	fs, err := h.fsForUser(user)
	if err != nil {
		if errors.Is(err, errWorkspaceUnconfigured) {
			writeError(w, err.Error(), http.StatusConflict)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		entries, err := fs.ListFiles(".")
		if err != nil {
			writeError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if entries == nil {
			entries = []workspace.FileEntry{}
		}
		writeJSON(w, http.StatusOK, entries)
	default:
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// renameRequest is the body of POST /api/v1/files/rename.
type renameRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// mkdirRequest is the body of POST /api/v1/files/mkdir.
type mkdirRequest struct {
	Path string `json:"path"`
}

// saveRequest is the body of PUT /api/v1/files/{path}.
type saveRequest struct {
	Content string `json:"content"`
}

// handleSubpath serves /api/v1/files/{path} for GET/PUT/DELETE, plus
// the two reserved action paths /files/rename and /files/mkdir.
func (h *filesHandler) handleSubpath(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	fs, err := h.fsForUser(user)
	if err != nil {
		if errors.Is(err, errWorkspaceUnconfigured) {
			writeError(w, err.Error(), http.StatusConflict)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rel := strings.TrimPrefix(r.URL.Path, "/api/v1/files/")
	rel = strings.TrimSuffix(rel, "/")

	// Reserved action paths.
	switch rel {
	case "rename":
		h.handleRename(w, r, fs)
		return
	case "mkdir":
		h.handleMkdir(w, r, fs)
		return
	}

	if rel == "" {
		writeError(w, "missing file path", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		content, err := fs.LoadFile(rel)
		if err != nil {
			writeError(w, err.Error(), http.StatusNotFound)
			return
		}
		etag := etagFor(content)
		w.Header().Set("ETag", etag)
		writeJSON(w, http.StatusOK, map[string]string{
			"path":    rel,
			"content": content,
			"etag":    etag,
		})

	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, "could not read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		var req saveRequest
		if err := json.Unmarshal(body, &req); err != nil {
			writeError(w, `invalid JSON body — expected {"content":"..."}`, http.StatusBadRequest)
			return
		}

		// Concurrency check: if the client sent an If-Match header,
		// reject the write when it doesn't match the file's current
		// content hash. Lets the frontend offer "overwrite or reload"
		// when someone else (or the user themselves in another tab,
		// or an external editor) changed the file since the client
		// last read it. Missing header = no check (creating new files,
		// CLI tooling, etc.).
		ifMatch := r.Header.Get("If-Match")
		if ifMatch != "" {
			if current, err := fs.LoadFile(rel); err == nil {
				if etagFor(current) != ifMatch {
					w.Header().Set("ETag", etagFor(current))
					writeError(w, "file changed on disk since last read", http.StatusPreconditionFailed)
					return
				}
			}
			// LoadFile error → file doesn't exist yet → treat as new
			// write; If-Match is irrelevant.
		}

		if err := fs.SaveFile(rel, req.Content); err != nil {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		newEtag := etagFor(req.Content)
		w.Header().Set("ETag", newEtag)
		writeJSON(w, http.StatusOK, map[string]string{
			"path": rel,
			"etag": newEtag,
		})

	case http.MethodDelete:
		if err := fs.DeleteFile(rel); err != nil {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"path": rel})

	default:
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *filesHandler) handleRename(w http.ResponseWriter, r *http.Request, fs *workspace.LocalFileSystem) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, "could not read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var req renameRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, `invalid JSON body — expected {"from":"...","to":"..."}`, http.StatusBadRequest)
		return
	}
	if req.From == "" || req.To == "" {
		writeError(w, "both 'from' and 'to' are required", http.StatusBadRequest)
		return
	}
	if err := fs.RenameFile(req.From, req.To); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"from": req.From, "to": req.To})
}

func (h *filesHandler) handleMkdir(w http.ResponseWriter, r *http.Request, fs *workspace.LocalFileSystem) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, "could not read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var req mkdirRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, `invalid JSON body — expected {"path":"..."}`, http.StatusBadRequest)
		return
	}
	if req.Path == "" {
		writeError(w, "'path' field is required", http.StatusBadRequest)
		return
	}
	if err := fs.CreateDirectory(req.Path); err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"path": req.Path})
}
