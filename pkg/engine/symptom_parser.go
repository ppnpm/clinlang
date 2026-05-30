package engine

import (
	"slices"
	"strings"
)

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

		sym := parseSymptomToken(tok, c)
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
func parseSymptomToken(tok string, c *ClinicalCase) Symptom {
	intensityMap := getIntensityMap(c)
	suffixes := make([]string, 0, len(intensityMap))
	for k := range intensityMap {
		suffixes = append(suffixes, k)
	}
	slices.SortFunc(suffixes, func(a, b string) int {
		return len(b) - len(a)
	})

	for _, suffix := range suffixes {
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
		duration := parseDurationWithCase(rest, c)

		return Symptom{
			Name:      name,
			Intensity: intensityMap[suffix],
			Duration:  duration,
		}
	}

	// No intensity marker found — check if the token ends in a duration (e.g. "cough3d")
	for i := 0; i < len(tok); i++ {
		ch := tok[i]
		if ch >= '0' && ch <= '9' {
			if i == 0 || tok[i-1] < '0' || tok[i-1] > '9' {
				// Potential duration start
				rest := tok[i:]
				if isDurationWithCase(rest, c) {
					return Symptom{
						Name:     tok[:i],
						Duration: rest,
					}
				}
			}
		}
	}

	// Treat the whole token as a symptom name
	return Symptom{Name: tok}
}

func getIntensityMap(c *ClinicalCase) map[string]string {
	if c != nil && c.Config != nil && len(c.Config.Symptoms) > 0 {
		return c.Config.Symptoms
	}
	return intensityMap
}

func isDurationWithCase(s string, c *ClinicalCase) bool {
	if c != nil && c.Config != nil {
		return isDurationWithOptions(s, c.Config.Durations)
	}
	return isDuration(s)
}

func parseDurationWithCase(s string, c *ClinicalCase) string {
	if isDurationWithCase(s, c) {
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
