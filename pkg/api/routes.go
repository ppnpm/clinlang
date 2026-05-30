package api

import "net/http"

// registerProtectedRoutes attaches all auth-required handlers to mux.
// The routes here are mounted under /api/v1/ by the parent server.
func registerProtectedRoutes(mux *http.ServeMux, cfg Config, ws *workspaceState) {
	// Parsing / formatting endpoints (handlers_parse.go)
	ph := &parseHandler{cfg: cfg, ws: ws}
	mux.HandleFunc("/api/v1/parse", ph.handleParse)
	mux.HandleFunc("/api/v1/note", ph.handleNote)
	mux.HandleFunc("/api/v1/soap", ph.handleSOAP)
	mux.HandleFunc("/api/v1/markdown", ph.handleMarkdown)
	mux.HandleFunc("/api/v1/lint", ph.handleLint)
	mux.HandleFunc("/api/v1/validate", ph.handleValidateDeprecated)

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

	// Config endpoints (handlers_config.go)
	ch := &configHandler{cfg: cfg, ws: ws}
	mux.HandleFunc("/api/v1/config/", ch.handle)
}
