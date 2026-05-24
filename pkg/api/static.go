package api

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// webDist embeds the production build of the Vite frontend
// (web/dist/, copied by `make web` into pkg/api/web-dist/ at build
// time). When the directory is empty (e.g. during early development
// before the frontend has been built), the embedded FS is still valid
// but contains no files — staticHandler degrades gracefully and
// returns 404 for all non-API paths.
//
//go:embed all:web-dist
var webDistRaw embed.FS

// webDistFS is webDistRaw rooted at web-dist/, so URL "/" maps to
// web-dist/index.html instead of web-dist/web-dist/index.html.
func webDistFS() fs.FS {
	sub, err := fs.Sub(webDistRaw, "web-dist")
	if err != nil {
		return webDistRaw
	}
	return sub
}

// hasIndexHTML reports whether the embedded frontend bundle contains
// an index.html. If not, we treat the binary as API-only and skip the
// SPA fallback entirely.
func hasIndexHTML() bool {
	_, err := fs.Stat(webDistFS(), "index.html")
	return err == nil
}

// staticHandler serves the embedded SPA bundle with a SPA-style
// fallback: any request that does not start with /api/ and doesn't
// match a real asset returns index.html, letting client-side routing
// handle the URL.
func staticHandler() http.Handler {
	files := webDistFS()
	fileServer := http.FileServer(http.FS(files))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API requests should never land here; the router mounts the
		// static handler last. Defensive double-check.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// If a real asset exists at this path, serve it.
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if f, err := files.Open(path); err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Otherwise fall back to index.html so the React router can
		// handle the route. Read it from the embedded FS directly.
		index, err := fs.ReadFile(files, "index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// SPA shell must not be cached by intermediaries — service
		// worker handles fast refresh in the browser.
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(index)
	})
}
