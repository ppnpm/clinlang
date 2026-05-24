package api

import (
	"net/http"

	"clinlang/pkg/auth"
)

// Disclaimer is shown in the /health endpoint and is kept in sync with
// DISCLAIMER.md at the repository root.
const Disclaimer = "ClinLang is a personal note-taking and templating tool — not a medical device. No diagnosis, dosing, or decision support."

// Server is the assembled HTTP API. The Mode and BindAddr are fixed
// at construction time, but WorkspaceRoot can change at runtime via
// PUT /api/v1/workspace (local mode only). The workspaceState pointer
// is shared with the files and workspace handlers so writes show up
// on the next request.
type Server struct {
	cfg     Config
	ws      *workspaceState
	handler http.Handler
}

// NewServer builds a Server from the given Config. It registers all
// /api/v1/ routes, the public /health route, and wraps the protected
// routes with CORS + auth middleware in the correct order.
func NewServer(cfg Config) (*Server, error) {
	ws := &workspaceState{path: cfg.WorkspaceRoot}

	mux := http.NewServeMux()

	// Public endpoint — no auth required so probes (load balancer,
	// uptime monitor) can reach it.
	mux.HandleFunc("/api/v1/health", handleHealth(cfg))

	// Protected subtree. Auth middleware injects the acting user;
	// every protected handler reads it via auth.UserFromContext.
	protected := http.NewServeMux()
	registerProtectedRoutes(protected, cfg, ws)

	mux.Handle("/api/v1/", auth.Middleware(cfg.Mode, protected))

	// Embedded SPA — only mounted if the build actually contains a
	// frontend bundle. When empty (API-only build), the SPA route is
	// left unbound and "/" returns 404, which is the expected
	// behavior for an API-only deployment.
	if hasIndexHTML() {
		mux.Handle("/", staticHandler())
	}

	// CORS wraps the whole tree (public + protected + static).
	return &Server{
		cfg:     cfg,
		ws:      ws,
		handler: corsMiddleware(cfg.Mode, mux),
	}, nil
}

// Handler returns the fully-wrapped http.Handler. Useful in tests via
// httptest.NewServer(srv.Handler()).
func (s *Server) Handler() http.Handler {
	return s.handler
}

// ListenAndServe binds to cfg.BindAddr and serves until error.
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.cfg.BindAddr, s.handler)
}

// Config returns the configuration the server was built with.
func (s *Server) Config() Config {
	return s.cfg
}
