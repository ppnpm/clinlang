package engine

import "strings"

// intensityMap maps suffix strings to clinical descriptions.
var intensityMap = map[string]string{
	"+++": "very severe",
	"++":  "severe",
	"+":   "mild",
	"---": "resolved",
	"--":  "resolving",
	"-":   "improving",
}

// ParseSymptoms parses tokens after the `sx` command.
// Supported token formats:
//   - fever        → symptom only
//   - fever+       → mild
//   - cough++      → severe
//   - sob+++       → very severe
//   - pain--3d     → resolving, 3d duration
//   - fever++2h    → severe, 2h duration
//
// reg is the per-parse CommandTokenRegistry. Pass nil when no plugin is active.
// Tokens that produce no name (rare) are offered to plugin extensions before warning.
func ParseSymptoms(tokens []string, c *ClinicalCase, reg *CommandTokenRegistry) {
	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}

		sym := parseSymptomToken(tok)
		if sym.Name == "" {
			// ── Plugin token extension ────────────────────────────────────────
			if reg.Try("sx", tok, c) {
				continue
			}
			c.AddWarning("Unrecognized symptom token: " + tok)
			continue
		}
		c.Symptoms = append(c.Symptoms, sym)
	}
}

// parseSymptomToken breaks a single token into a Symptom struct.
func parseSymptomToken(tok string) Symptom {
	// Order matters: check longer suffixes first.
	intensitySuffixes := []string{"+++", "++", "+", "---", "--", "-"}

	for _, suffix := range intensitySuffixes {
		idx := strings.Index(tok, suffix)
		if idx == -1 {
			continue
		}

		// Everything before the intensity marker is the symptom name.
		name := tok[:idx]
		if name == "" {
			continue
		}

		// Everything after the intensity marker may be a duration.
		rest := tok[idx+len(suffix):]
		duration := parseDuration(rest)

		return Symptom{
			Name:      name,
			Intensity: intensityMap[suffix],
			Duration:  duration,
		}
	}

	// No intensity marker found — treat the whole token as a symptom name.
	// The token might still end in a duration (e.g. "fever3d"), but that's
	// unusual — we'll keep it as-is for simplicity.
	return Symptom{Name: tok}
}

// parseDuration extracts a duration string like "3d", "2h", "1w", "30min" from the
// remainder after the intensity marker. Returns empty string if not a duration.
func parseDuration(s string) string {
	num, unit := ParseDurationToken(s)
	if num != "" && IsValidDurationUnit(unit) {
		return s
	}
	return ""
}

// IntensityLabel returns the human-facing label for a given symbol string.
func IntensityLabel(intensity string) string {
	if intensity == "" {
		return "present"
	}
	return intensity
}
