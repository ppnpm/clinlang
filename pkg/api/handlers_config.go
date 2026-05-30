package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"clinlang/pkg/auth"
	"clinlang/pkg/engine"
)

var configWhitelist = map[string]bool{
	"abbreviations.json":    true,
	"drugs.json":            true,
	"reference_ranges.json": true,
	"frequencies.json":      true,
	"routes.json":           true,
	"symptoms.json":         true,
	"rad_keys.json":         true,
	"durations.json":        true,
}

type configHandler struct {
	cfg Config
	ws  *workspaceState
}

func (h *configHandler) handle(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == "" {
		writeError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	filename := strings.TrimPrefix(r.URL.Path, "/api/v1/config/")
	filename = strings.TrimSuffix(filename, "/")

	if !configWhitelist[filename] {
		writeError(w, "invalid configuration filename", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r, user, filename)
	case http.MethodPut:
		h.handlePut(w, r, user, filename)
	default:
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *configHandler) handleGet(w http.ResponseWriter, r *http.Request, user, filename string) {
	configPath, err := configPathForUser(h.ws, h.cfg.Mode, user, filename)
	if err != nil {
		writeError(w, err.Error(), http.StatusConflict)
		return
	}

	// Check if the custom config file exists in the workspace root .config/ folder
	data, err := os.ReadFile(configPath)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
		return
	}

	// If it does not exist, get the embedded defaults
	defaultBytes := engine.DefaultConfigJSON(filename)
	if defaultBytes == nil {
		writeError(w, "default configuration not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(defaultBytes)
}

func (h *configHandler) handlePut(w http.ResponseWriter, r *http.Request, user, filename string) {
	configPath, err := configPathForUser(h.ws, h.cfg.Mode, user, filename)
	if err != nil {
		writeError(w, err.Error(), http.StatusConflict)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, "could not read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Validate input is valid JSON
	var parsed interface{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		writeError(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Pretty print it
	formatted, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		writeError(w, "failed to format JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create .config directory if it does not exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		writeError(w, "failed to create config directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the file
	if err := os.WriteFile(configPath, formatted, 0644); err != nil {
		writeError(w, "failed to write config file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "saved", "path": configPath})
}

func configPathForUser(ws *workspaceState, mode string, user string, filename string) (string, error) {
	root := ws.Get()
	if root == "" {
		return "", errWorkspaceUnconfigured
	}
	if mode == ModeHosted {
		clean := filepath.Clean("/" + user)
		root = filepath.Join(root, clean)
	}
	return filepath.Join(root, ".config", filename), nil
}
