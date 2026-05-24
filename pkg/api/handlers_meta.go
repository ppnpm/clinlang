package api

import (
	"encoding/json"
	"io"
	"net/http"

	"clinlang/pkg/autocomplete"
	"clinlang/pkg/engine"
)

// autocompleteRequest is the JSON shape accepted by /autocomplete.
type autocompleteRequest struct {
	Command string `json:"command"`
	Query   string `json:"query"`
}

// handleAutocomplete returns suggestions for the given command + query.
// For "rx" it delegates to the engine drug list; other commands use the
// static suggestion dictionary in pkg/autocomplete.
func handleAutocomplete(w http.ResponseWriter, r *http.Request) {
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

	var req autocompleteRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, `invalid JSON body — expected {"command":"...","query":"..."}`, http.StatusBadRequest)
		return
	}
	if req.Command == "" {
		writeError(w, "'command' field is required", http.StatusBadRequest)
		return
	}

	suggestions := autocomplete.GetSuggestions(req.Command, req.Query)
	if suggestions == nil {
		suggestions = []autocomplete.Suggestion{}
	}
	writeJSON(w, http.StatusOK, suggestions)
}

// handleDrugs returns drug names matching the ?prefix= query.
func handleDrugs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	prefix := r.URL.Query().Get("prefix")
	matches := engine.SearchDrugs(prefix)
	if matches == nil {
		matches = []string{}
	}
	writeJSON(w, http.StatusOK, matches)
}

// handlePlugins returns metadata about all registered specialty plugins.
func handlePlugins(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	plugins := engine.ListPlugins()
	if plugins == nil {
		plugins = []engine.PluginInfo{}
	}
	writeJSON(w, http.StatusOK, plugins)
}
