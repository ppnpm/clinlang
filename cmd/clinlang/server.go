package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"clinlang/pkg/engine"
)

// =============================================================================
// HTTP API SERVER
// =============================================================================

// apiRequest is the standard request body for all endpoints.
type apiRequest struct {
	Input string `json:"input"`
}

// StartServer starts the HTTP API server on the given port (e.g., "8080").
func StartServer(port string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/parse",    corsMiddleware(handleParse))
	mux.HandleFunc("/note",     corsMiddleware(handleNote))
	mux.HandleFunc("/soap",     corsMiddleware(handleSOAP))
	mux.HandleFunc("/validate", corsMiddleware(handleValidate))
	mux.HandleFunc("/health",   handleHealth)

	addr := ":" + port
	fmt.Printf("ClinLang API server running on http://localhost%s\n", addr)
	fmt.Println("Endpoints: POST /parse | /note | /soap | /validate")
	fmt.Println("Press Ctrl+C to stop.")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Println("Server error:", err)
	}
}

// =============================================================================
// HANDLERS
// =============================================================================

// handleParse returns the full parsed Case as JSON.
func handleParse(w http.ResponseWriter, r *http.Request) {
	c, err := readAndParse(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, c)
}

// handleNote returns a plain-text clinical note wrapped in JSON.
func handleNote(w http.ResponseWriter, r *http.Request) {
	c, err := readAndParse(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	note := buildNoteString(c)
	writeJSON(w, map[string]interface{}{
		"note":           note,
		"warnings":       c.Warnings,
		"abnormal_flags": c.AbnormalFlags,
	})
}

// handleSOAP returns a SOAP-formatted note wrapped in JSON.
func handleSOAP(w http.ResponseWriter, r *http.Request) {
	c, err := readAndParse(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]interface{}{
		"soap":           engine.FormatSOAP(c),
		"warnings":       c.Warnings,
		"abnormal_flags": c.AbnormalFlags,
	})
}

// handleValidate returns warnings and abnormal flags without the full note.
func handleValidate(w http.ResponseWriter, r *http.Request) {
	c, err := readAndParse(r)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	valid := len(c.Warnings) == 0 && len(c.AbnormalFlags) == 0
	writeJSON(w, map[string]interface{}{
		"valid":          valid,
		"warnings":       c.Warnings,
		"abnormal_flags": c.AbnormalFlags,
	})
}

// handleHealth is a simple liveness probe for load balancers / apps.
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{"status": "ok", "service": "clinlang"})
}

// =============================================================================
// HELPERS
// =============================================================================

// readAndParse decodes the request body and calls ParseString.
func readAndParse(r *http.Request) (engine.ClinicalCase, error) {
	if r.Method == http.MethodOptions {
		return engine.ClinicalCase{}, fmt.Errorf("preflight")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return engine.ClinicalCase{}, fmt.Errorf("could not read request body")
	}
	defer r.Body.Close()

	var req apiRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return engine.ClinicalCase{}, fmt.Errorf("invalid JSON body — expected {\"input\": \"...\"}")
	}
	if strings.TrimSpace(req.Input) == "" {
		return engine.ClinicalCase{}, fmt.Errorf("'input' field is empty")
	}
	return engine.ParseString(req.Input), nil
}

// buildNoteString captures PrintClinicalNote output as a string.
func buildNoteString(c engine.ClinicalCase) string {
	sb := strings.Builder{}
	sep := strings.Repeat("─", 50)

	writeLine := func(s string) { sb.WriteString(s + "\n") }
	writeField := func(label, value string) {
		if value != "" {
			sb.WriteString(fmt.Sprintf("%-22s: %s\n", label, value))
		}
	}

	writeLine(sep)
	writeLine("         CLINICAL NOTE")
	writeLine(sep)
	writeLine(fmt.Sprintf("ID     : %s", c.Patient.Id))

	age := fmt.Sprintf("%d", c.Patient.Age)
	if c.Patient.Age == 0 {
		age = "?"
	}
	patLine := fmt.Sprintf("Patient: %s/%s", age, c.Patient.Sex)
	if c.Patient.Weight > 0 {
		patLine += fmt.Sprintf("  Wt: %gkg", c.Patient.Weight)
	}
	if c.Patient.Height > 0 {
		patLine += fmt.Sprintf("  Ht: %gcm", c.Patient.Height)
	}
	writeLine(patLine)
	writeLine(sep)

	writeField("Chief Complaint", c.CC)
	writeField("HPI", c.HPI)
	writeField("Past Medical History", c.PMH)
	writeLine(fmt.Sprintf("Vitals : %s", engine.FormatVitals(c.Vitals)))

	if len(c.AbnormalFlags) > 0 {
		writeLine("")
		for _, f := range c.AbnormalFlags {
			icon := "⚠"
			if f.Severity == engine.SeverityCritical {
				icon = "🔴 CRITICAL"
			}
			writeLine(fmt.Sprintf("  %s  %s: %s — %s", icon, f.Field, f.Value, f.Message))
		}
	}

	if len(c.Symptoms) > 0 {
		writeLine("Symptoms:")
		for _, s := range c.Symptoms {
			label := engine.IntensityLabel(s.Intensity)
			dur := ""
			if s.Duration != "" {
				dur = " × " + s.Duration
			}
			writeLine(fmt.Sprintf("  ▸ %-20s [%s%s]", s.Name, label, dur))
		}
	}

	if len(c.Extra) > 0 {
		writeLine(sep)
		writeLine("Additional Data:")
		for cmd, kv := range c.Extra {
			writeLine(fmt.Sprintf("  [%s]", strings.ToUpper(cmd)))
			for k, v := range kv {
				if v == "true" {
					writeLine(fmt.Sprintf("    ▸ %s", k))
				} else {
					writeLine(fmt.Sprintf("    ▸ %-12s: %s", k, v))
				}
			}
		}
	}

	if len(c.Prescriptions) > 0 {
		writeLine(sep)
		writeLine(engine.FormatPrescriptions(c.Prescriptions))
	}

	writeLine(sep)
	writeField("Diagnosis", c.DX)
	writeLine(sep)

	return sb.String()
}

// writeJSON serializes v to JSON and writes to the ResponseWriter.
func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, `{"error":"internal encoding error"}`, http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

// writeError returns a JSON error response.
func writeError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	b, _ := json.Marshal(map[string]string{"error": msg})
	w.Write(b)
}

// corsMiddleware adds CORS headers so any web frontend can call the API.
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}
