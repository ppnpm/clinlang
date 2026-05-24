package api

import "net/http"

// registerProtectedRoutes attaches all auth-required handlers to mux.
// The routes here are mounted under /api/v1/ by the parent server.
func registerProtectedRoutes(mux *http.ServeMux, cfg Config, ws *workspaceState) {
	// Parsing / formatting endpoints (handlers_parse.go)
	mux.HandleFunc("/api/v1/parse", handleParse)
	mux.HandleFunc("/api/v1/note", handleNote)
	mux.HandleFunc("/api/v1/soap", handleSOAP)
	mux.HandleFunc("/api/v1/markdown", handleMarkdown)
	mux.HandleFunc("/api/v1/lint", handleLint)
	mux.HandleFunc("/api/v1/validate", handleValidateDeprecated)

	// Metadata / completion endpoints (handlers_meta.go)
	mux.HandleFunc("/api/v1/autocomplete", handleAutocomplete)
	mux.HandleFunc("/api/v1/drugs", handleDrugs)
	mux.HandleFunc("/api/v1/plugins", handlePlugins)

	// Workspace endpoints (handlers_workspace.go)
	wh := &workspaceHandler{cfg: cfg, ws: ws}
	mux.HandleFunc("/api/v1/workspace", wh.handle)
	mux.HandleFunc("/api/v1/workspace/browse", wh.handleBrowse)

	// File workspace endpoints (handlers_files.go)
	fh := &filesHandler{cfg: cfg, ws: ws}
	mux.Handle("/api/v1/files", http.HandlerFunc(fh.handleRoot))
	mux.Handle("/api/v1/files/", http.HandlerFunc(fh.handleSubpath))
}
