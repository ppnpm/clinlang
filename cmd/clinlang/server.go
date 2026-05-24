package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/skratchdot/open-golang/open"

	"clinlang/pkg/api"
)

// StartServer is the thin wiring that lifts an api.Config into a
// running HTTP server. All HTTP behavior lives in pkg/api.
//
// openBrowser=true asks the binary to launch the user's default
// browser at the bind URL once the server is up. Honored only in
// local mode (hosted-mode invocations are typically daemons, not
// interactive sessions).
func StartServer(modeOverride, portOverride string, openBrowser bool) {
	cfg, err := api.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config error:", err)
		os.Exit(2)
	}
	if modeOverride != "" {
		cfg.Mode = modeOverride
		// Reconcile bind address default if --mode flipped us.
		if portOverride == "" {
			if cfg.Mode == api.ModeLocal && cfg.BindAddr == "0.0.0.0:8080" {
				cfg.BindAddr = "127.0.0.1:8080"
			} else if cfg.Mode == api.ModeHosted && cfg.BindAddr == "127.0.0.1:8080" {
				cfg.BindAddr = "0.0.0.0:8080"
			}
		}
	}
	if portOverride != "" {
		cfg.BindAddr = api.OverridePort(cfg.BindAddr, portOverride)
	}

	srv, err := api.NewServer(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "server error:", err)
		os.Exit(2)
	}

	fmt.Printf("ClinLang API server running on http://%s\n", cfg.BindAddr)
	fmt.Println(disclaimer)
	fmt.Printf("Mode      : %s\n", cfg.Mode)
	fmt.Printf("Workspace : %s\n", cfg.WorkspaceRoot)
	fmt.Println("Routes    : POST /api/v1/{parse,note,soap,markdown,lint,autocomplete}")
	fmt.Println("            GET  /api/v1/{health,drugs,plugins,files,files/{path}}")
	fmt.Println("            PUT  /api/v1/files/{path}")
	fmt.Println("            DELETE /api/v1/files/{path}")
	fmt.Println("Press Ctrl+C to stop.")

	// Auto-launch the browser in local mode after a brief delay so
	// the server is actually listening before the URL is opened.
	if openBrowser && cfg.Mode == api.ModeLocal {
		go func() {
			time.Sleep(300 * time.Millisecond)
			url := "http://" + cfg.BindAddr
			// Bind addrs like 0.0.0.0:8080 won't open as a URL in a
			// browser sensibly; rewrite to localhost.
			if strings.HasPrefix(url, "http://0.0.0.0:") {
				url = "http://localhost:" + strings.TrimPrefix(url, "http://0.0.0.0:")
			}
			_ = open.Start(url)
		}()
	}

	if err := srv.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, "server error:", err)
		os.Exit(1)
	}
}
