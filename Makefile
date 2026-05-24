# ClinLang build orchestration.
#
# Usage:
#   make web         — install deps + build the Vite frontend, copy
#                      into pkg/api/web-dist/ for go:embed
#   make types       — regenerate web/src/lib/types.ts via tygo
#   make build       — web + Go binary for the host platform
#   make build-all   — Go binaries for linux/darwin/windows
#   make dev         — start the Go server + Vite dev server (split panes)
#   make clean       — remove web/dist, web-dist embed, dist/

GO            := go
NPM           := npm
WEB_DIR       := web
WEB_OUT       := $(WEB_DIR)/dist
EMBED_DIR     := pkg/api/web-dist
DIST_DIR      := dist
BIN_NAME      := clinlang
LDFLAGS       :=

.PHONY: web types check-types build build-all dev clean

web:
	cd $(WEB_DIR) && $(NPM) install
	cd $(WEB_DIR) && $(NPM) run build
	rm -rf $(EMBED_DIR)
	mkdir -p $(EMBED_DIR)
	cp -R $(WEB_OUT)/. $(EMBED_DIR)/

types:
	$(GO) run github.com/gzuidhof/tygo@latest generate

# check-types regenerates the TS bindings and fails if anything in the
# committed files differs from what tygo would write today. Use this in
# CI to catch drift between Go structs and the frontend type surface.
check-types: types
	@if ! git diff --quiet -- $(WEB_DIR)/src/lib/types-engine.ts $(WEB_DIR)/src/lib/types-workspace.ts $(WEB_DIR)/src/lib/types-autocomplete.ts; then \
		echo "ERROR: generated TS types are stale. Run 'make types' and commit."; \
		git --no-pager diff -- $(WEB_DIR)/src/lib/types-engine.ts $(WEB_DIR)/src/lib/types-workspace.ts $(WEB_DIR)/src/lib/types-autocomplete.ts; \
		exit 1; \
	fi
	@echo "Generated TS types are in sync with Go source."

build: web
	$(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BIN_NAME) ./cmd/clinlang

build-all: web
	mkdir -p $(DIST_DIR)
	GOOS=linux   GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BIN_NAME)-linux-amd64       ./cmd/clinlang
	GOOS=linux   GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BIN_NAME)-linux-arm64       ./cmd/clinlang
	GOOS=darwin  GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BIN_NAME)-darwin-amd64      ./cmd/clinlang
	GOOS=darwin  GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BIN_NAME)-darwin-arm64      ./cmd/clinlang
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(BIN_NAME)-windows-amd64.exe ./cmd/clinlang

dev:
	@echo "Run in two terminals:"
	@echo "  Terminal 1:  go run ./cmd/clinlang server"
	@echo "  Terminal 2:  cd web && npm run dev"
	@echo "Then open http://localhost:5173"

clean:
	rm -rf $(WEB_OUT) $(EMBED_DIR) $(DIST_DIR)
	mkdir -p $(EMBED_DIR)
	touch $(EMBED_DIR)/.gitkeep
