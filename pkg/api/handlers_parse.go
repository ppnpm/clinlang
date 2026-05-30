package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"clinlang/pkg/auth"
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

type parseHandler struct {
	cfg Config
	ws  *workspaceState
}

func (h *parseHandler) getParserConfigAndRanges(user string) (*engine.ParserConfig, engine.ReferenceRanges) {
	cfg := engine.DefaultConfig
	ranges := engine.DefaultReferenceRanges()

	root := h.ws.Get()
	if root == "" {
		return &cfg, ranges
	}

	if h.cfg.Mode == ModeHosted {
		clean := filepath.Clean("/" + user)
		root = filepath.Join(root, clean)
	}
	configDir := filepath.Join(root, ".config")

	loadJSON := func(filename string, dest interface{}) bool {
		p := filepath.Join(configDir, filename)
		data, err := os.ReadFile(p)
		if err != nil {
			return false
		}
		if err := json.Unmarshal(data, dest); err == nil {
			return true
		}
		return false
	}

	var customAbbr map[string]string
	if loadJSON("abbreviations.json", &customAbbr) {
		cfg.Abbreviations = customAbbr
	}
	var customFreq engine.FrequencyConfig
	if loadJSON("frequencies.json", &customFreq) {
		cfg.Frequencies = customFreq
	}
	var customRoutes engine.RouteConfig
	if loadJSON("routes.json", &customRoutes) {
		cfg.Routes = customRoutes
	}
	var customSymptoms map[string]string
	if loadJSON("symptoms.json", &customSymptoms) {
		cfg.Symptoms = customSymptoms
	}
	var customRadKeys []string
	if loadJSON("rad_keys.json", &customRadKeys) {
		cfg.RadKeys = customRadKeys
	}
	var customDur map[string]engine.DurationUnit
	if loadJSON("durations.json", &customDur) {
		cfg.Durations = customDur
	}
	var customDrugs []string
	if loadJSON("drugs.json", &customDrugs) {
		cfg.Drugs = customDrugs
	}
	var customRanges engine.ReferenceRanges
	if loadJSON("reference_ranges.json", &customRanges) {
		ranges = customRanges
	}

	return &cfg, ranges
}

// handleParse returns the full parsed ClinicalCase as JSON.
func (h *parseHandler) handleParse(w http.ResponseWriter, r *http.Request) {
	req, err := readParseRequest(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	user := auth.UserFromContext(r.Context())
	parserCfg, ranges := h.getParserConfigAndRanges(user)
	writeJSON(w, http.StatusOK, engine.ParseStringWithOptions(req.Input, parserCfg, ranges))
}

// handleNote returns the provisional plain-text clinical note.
func (h *parseHandler) handleNote(w http.ResponseWriter, r *http.Request) {
	req, err := readParseRequest(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	user := auth.UserFromContext(r.Context())
	parserCfg, ranges := h.getParserConfigAndRanges(user)
	c := engine.ParseStringWithOptions(req.Input, parserCfg, ranges)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"note":          engine.FormatPlainNote(c),
		"warnings":      c.Warnings,
		"range_markers": c.RangeMarkers,
	})
}

// handleSOAP returns a SOAP-formatted note. The optional "markers"
// field in the request body toggles the "Notes (out of ref)" section;
// default off.
func (h *parseHandler) handleSOAP(w http.ResponseWriter, r *http.Request) {
	req, err := readParseRequest(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	user := auth.UserFromContext(r.Context())
	parserCfg, ranges := h.getParserConfigAndRanges(user)
	c := engine.ParseStringWithOptions(req.Input, parserCfg, ranges)
	out := engine.FormatSOAPWithOptions(c, engine.FormatOptions{ShowRangeMarkers: req.Markers})
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"soap":          out,
		"warnings":      c.Warnings,
		"range_markers": c.RangeMarkers,
		"images":        c.Images,
	})
}

// handleMarkdown returns a markdown-formatted note. "markers" toggle
// follows the same rule as /soap.
func (h *parseHandler) handleMarkdown(w http.ResponseWriter, r *http.Request) {
	req, err := readParseRequest(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	user := auth.UserFromContext(r.Context())
	parserCfg, ranges := h.getParserConfigAndRanges(user)
	c := engine.ParseStringWithOptions(req.Input, parserCfg, ranges)
	out := engine.FormatMarkdownWithOptions(c, engine.FormatOptions{ShowRangeMarkers: req.Markers})
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"markdown":      out,
		"warnings":      c.Warnings,
		"range_markers": c.RangeMarkers,
		"images":        c.Images,
	})
}

// handleLint returns parser warnings and out-of-range markers only.
func (h *parseHandler) handleLint(w http.ResponseWriter, r *http.Request) {
	req, err := readParseRequest(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	user := auth.UserFromContext(r.Context())
	parserCfg, ranges := h.getParserConfigAndRanges(user)
	c := engine.ParseStringWithOptions(req.Input, parserCfg, ranges)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"warnings":      c.Warnings,
		"range_markers": c.RangeMarkers,
	})
}

// handleValidateDeprecated proxies to handleLint and emits an HTTP
// Deprecation header. Will be removed in a future release; callers
// should migrate to /api/v1/lint.
func (h *parseHandler) handleValidateDeprecated(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Deprecation", "true")
	w.Header().Set("Sunset", "future")
	w.Header().Set("Link", `</api/v1/lint>; rel="successor-version"`)
	h.handleLint(w, r)
}
