package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// workspaceHandler serves /api/v1/workspace.
//
//	GET  → returns the current workspace root + deployment mode.
//	PUT  → changes the workspace root (LOCAL MODE ONLY).
//
// In hosted mode the operator owns the workspace layout, so a runtime
// PUT is rejected with 403 Forbidden. Local-mode users are expected
// to point ClinLang at any folder they like (their own laptop, an
// iCloud / OneDrive / Dropbox synced folder, a USB drive, etc.).
type workspaceHandler struct {
	cfg Config
	ws  *workspaceState
}

type workspaceResponse struct {
	Path       string `json:"path"`
	Mode       string `json:"mode"`
	Configured bool   `json:"configured"`
	Suggested  string `json:"suggested,omitempty"`
}

type workspaceRequest struct {
	Path string `json:"path"`
}

// browseResponse is the body of POST /api/v1/workspace/browse.
// Cancelled = true means the user closed the OS dialog without
// choosing. Frontends should treat that as "no change", not an error.
type browseResponse struct {
	Path      string `json:"path"`
	Cancelled bool   `json:"cancelled"`
}

func (h *workspaceHandler) handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		current := h.ws.Get()
		resp := workspaceResponse{
			Path:       current,
			Mode:       h.cfg.Mode,
			Configured: current != "",
		}
		// Suggestion is only useful in local mode and only when no
		// workspace is configured yet.
		if h.cfg.Mode == ModeLocal && !resp.Configured {
			resp.Suggested = suggestedWorkspace()
		}
		writeJSON(w, http.StatusOK, resp)

	case http.MethodPut:
		if h.cfg.Mode != ModeLocal {
			writeError(w,
				"workspace can only be changed at runtime in local mode; hosted mode is fixed by CLINLANG_WORKSPACE",
				http.StatusForbidden)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, "could not read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		var req workspaceRequest
		if err := json.Unmarshal(body, &req); err != nil {
			writeError(w, `invalid JSON body — expected {"path":"..."}`, http.StatusBadRequest)
			return
		}
		newPath, err := validateAndPrepareWorkspace(req.Path)
		if err != nil {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.ws.Set(newPath)
		// Persist the choice so the next launch skips the welcome
		// dialog. Best-effort: a failure to write the config file
		// doesn't block the request — the user can simply re-pick.
		_ = savePersistedConfig(persistedConfig{WorkspaceRoot: newPath})
		writeJSON(w, http.StatusOK, workspaceResponse{
			Path:       newPath,
			Mode:       h.cfg.Mode,
			Configured: true,
		})

	default:
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleBrowse opens the OS-native folder picker and returns the chosen
// path. Local-mode only. Browsers cannot return absolute paths through
// any web API (File System Access API gives sandboxed handles, not
// paths), so the Go binary itself spawns the native dialog.
func (h *workspaceHandler) handleBrowse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.cfg.Mode != ModeLocal {
		writeError(w,
			"native folder picker is not available in hosted mode",
			http.StatusForbidden)
		return
	}
	path, err := openFolderPicker()
	if err != nil {
		if err == errPickerUnavailable {
			writeError(w, err.Error(), http.StatusNotImplemented)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, browseResponse{
		Path:      path,
		Cancelled: path == "",
	})
}

// validateAndPrepareWorkspace cleans the user-supplied path and ensures
// the directory exists (creating it if necessary). Returns the absolute
// canonical path used to set the workspace.
func validateAndPrepareWorkspace(raw string) (string, error) {
	if raw == "" {
		return "", errEmptyPath
	}
	abs, err := filepath.Abs(raw)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(abs); err == nil {
		if !info.IsDir() {
			return "", errNotADirectory
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(abs, 0755); err != nil {
			return "", err
		}
	} else {
		return "", err
	}
	return abs, nil
}

var (
	errEmptyPath     = &workspaceError{"path is empty"}
	errNotADirectory = &workspaceError{"path exists but is not a directory"}
)

type workspaceError struct{ msg string }

func (e *workspaceError) Error() string { return e.msg }
