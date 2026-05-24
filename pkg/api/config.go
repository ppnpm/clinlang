// Package api is the HTTP surface of ClinLang. It is intentionally
// independent of cmd/clinlang so that other binaries (the future
// embedded-frontend launcher in cmd/clinlang-web, the test harness)
// can construct a server from a Config without going through the CLI.
//
// Deployment posture is determined entirely by the Config:
//
//   - Mode "local"  — single user, bound to 127.0.0.1, no auth header
//                     required, CORS allows http://localhost:* origins.
//   - Mode "hosted" — multi-user, bound to 0.0.0.0, X-Forwarded-User
//                     header required, CORS rejects cross-origin
//                     requests (the reverse proxy owns CORS).
//
// One binary, two postures.
package api

import (
	"fmt"
	"net"
	"os"
)

// Mode constants for Config.Mode.
const (
	ModeLocal  = "local"
	ModeHosted = "hosted"
)

// Config holds the deployment-mode dependent settings of the API server.
type Config struct {
	// Mode is either ModeLocal or ModeHosted.
	Mode string

	// BindAddr is the host:port to bind on. Defaults: 127.0.0.1:8080
	// (local), 0.0.0.0:8080 (hosted).
	BindAddr string

	// WorkspaceRoot is the filesystem root for note storage.
	//
	//   - Local mode:  one directory holding all notes for the single
	//                  user. Defaults to $HOME/ClinCases.
	//   - Hosted mode: a PARENT directory under which each user gets
	//                  a subfolder (<root>/<user>/...). Required —
	//                  no default; LoadConfig errors if unset.
	WorkspaceRoot string
}

// LoadConfig resolves a Config from environment variables, filling in
// safe defaults for local mode. Returns an error if the resolved
// configuration is invalid (unknown mode, hosted with no workspace,
// home dir lookup failed).
//
//	CLINLANG_MODE       = "local" (default) | "hosted"
//	CLINLANG_BIND       = "host:port"
//	CLINLANG_WORKSPACE  = filesystem path
func LoadConfig() (Config, error) {
	cfg := Config{
		Mode:          os.Getenv("CLINLANG_MODE"),
		BindAddr:      os.Getenv("CLINLANG_BIND"),
		WorkspaceRoot: os.Getenv("CLINLANG_WORKSPACE"),
	}

	if cfg.Mode == "" {
		cfg.Mode = ModeLocal
	}
	if cfg.Mode != ModeLocal && cfg.Mode != ModeHosted {
		return cfg, fmt.Errorf("CLINLANG_MODE must be %q or %q, got %q",
			ModeLocal, ModeHosted, cfg.Mode)
	}

	if cfg.BindAddr == "" {
		if cfg.Mode == ModeLocal {
			cfg.BindAddr = "127.0.0.1:8080"
		} else {
			cfg.BindAddr = "0.0.0.0:8080"
		}
	}

	// Workspace resolution in local mode:
	//   1. CLINLANG_WORKSPACE env var wins (already loaded above).
	//   2. Saved config file from a previous launch.
	//   3. Otherwise left empty — the frontend will prompt on first
	//      launch via the welcome dialog and PUT /workspace will
	//      persist the choice.
	//
	// Hosted mode still requires CLINLANG_WORKSPACE explicitly; the
	// operator owns the layout, no interactive prompt available.
	if cfg.WorkspaceRoot == "" {
		if cfg.Mode == ModeHosted {
			return cfg, fmt.Errorf("CLINLANG_WORKSPACE is required in hosted mode")
		}
		saved := loadPersistedConfig()
		if saved.WorkspaceRoot != "" {
			cfg.WorkspaceRoot = saved.WorkspaceRoot
		}
		// If still empty, server starts unconfigured. Frontend
		// detects this via GET /workspace and shows the welcome
		// dialog. File handlers refuse operations until configured.
	}

	return cfg, nil
}

// OverridePort replaces the port portion of a host:port string. Used
// to honor the --port CLI flag without forcing the user to specify a
// full bind address.
func OverridePort(bindAddr, port string) string {
	if port == "" {
		return bindAddr
	}
	host, _, err := net.SplitHostPort(bindAddr)
	if err != nil {
		// bindAddr was malformed; fall back to localhost on the new port.
		return net.JoinHostPort("127.0.0.1", port)
	}
	return net.JoinHostPort(host, port)
}
