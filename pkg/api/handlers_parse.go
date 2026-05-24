package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"clinlang/pkg/engine"
)

// parseRequest is the JSON shape accepted by parsing endpoints.
// Markers is honored only by /soap and /markdown.
type parseRequest struct {
	Input   string `json:"input"`
	Markers bool   `json:"markers,omitempty"`
}

// handleHealth returns a simple liveness probe payload. Public (no auth).
func handleHealth(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":     "ok",
			"service":    "clinlang",
			"mode":       cfg.Mode,
			"disclaimer": Disclaimer,
		})
	}
}

func readParseRequest(r *http.Request) (parseRequest, error) {
	if r.Method != http.MethodPost {
		return parseRequest{}, fmt.Errorf("method %s not allowed", r.Method)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return parseRequest{}, fmt.Errorf("could not read request body")
	}
	defer r.Body.Close()

	var req parseRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return parseRequest{}, fmt.Errorf(`invalid JSON body — expected {"input":"...","markers":false}`)
	}
	if strings.TrimSpace(req.Input) == "" {
		return parseRequest{}, fmt.Errorf("'input' field is empty")
	}
	return req, nil
}

// handleParse returns the full parsed ClinicalCase as JSON.
func handleParse(w http.ResponseWriter, r *http.Request) {
	req, err := readParseRequest(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, engine.ParseString(req.Input))
}

// handleNote returns the canonical plain-text clinical note.
func handleNote(w http.ResponseWriter, r *http.Request) {
	req, err := readParseRequest(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	c := engine.ParseString(req.Input)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"note":          engine.FormatPlainNote(c),
		"warnings":      c.Warnings,
		"range_markers": c.RangeMarkers,
	})
}

// handleSOAP returns a SOAP-formatted note. The optional "markers"
// field in the request body toggles the "Notes (out of ref)" section;
// default off.
func handleSOAP(w http.ResponseWriter, r *http.Request) {
	req, err := readParseRequest(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	c := engine.ParseString(req.Input)
	out := engine.FormatSOAPWithOptions(c, engine.FormatOptions{ShowRangeMarkers: req.Markers})
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"soap":          out,
		"warnings":      c.Warnings,
		"range_markers": c.RangeMarkers,
	})
}

// handleMarkdown returns a markdown-formatted note. "markers" toggle
// follows the same rule as /soap.
func handleMarkdown(w http.ResponseWriter, r *http.Request) {
	req, err := readParseRequest(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	c := engine.ParseString(req.Input)
	out := engine.FormatMarkdownWithOptions(c, engine.FormatOptions{ShowRangeMarkers: req.Markers})
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"markdown":      out,
		"warnings":      c.Warnings,
		"range_markers": c.RangeMarkers,
	})
}

// handleLint returns parser warnings and out-of-range markers only.
func handleLint(w http.ResponseWriter, r *http.Request) {
	req, err := readParseRequest(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	c := engine.ParseString(req.Input)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"warnings":      c.Warnings,
		"range_markers": c.RangeMarkers,
	})
}

// handleValidateDeprecated proxies to handleLint and emits an HTTP
// Deprecation header. Will be removed in a future release; callers
// should migrate to /api/v1/lint.
func handleValidateDeprecated(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Deprecation", "true")
	w.Header().Set("Sunset", "future")
	w.Header().Set("Link", `</api/v1/lint>; rel="successor-version"`)
	handleLint(w, r)
}
