package engine

import (
	"strconv"
	"strings"
)

// ParseVitals parses tokens after the `vitals` command into a structured Vitals.
// Supported token formats:
//   - bp140/90    → Blood pressure (systolic/diastolic)
//   - hr110       → Heart rate (bpm)
//   - spo298      → SpO2 (percentage)
//   - temp99      → Temperature (Fahrenheit, numeric suffix)
//   - rr18        → Respiratory rate (breaths/min)
func ParseVitals(tokens []string, c *ClinicalCase) {
	for _, tok := range tokens {
		tok = strings.TrimSpace(strings.ToLower(tok))
		if tok == "" {
			continue
		}

		switch {
		case strings.HasPrefix(tok, "bp"):
			c.Vitals.BP = tok[2:] // Store as "140/90"

		case strings.HasPrefix(tok, "hr"):
			val, err := strconv.Atoi(tok[2:])
			if err != nil {
				c.AddWarning("Invalid HR value: " + tok)
			} else {
				c.Vitals.HR = val
			}

		case strings.HasPrefix(tok, "spo2"):
			val, err := strconv.Atoi(tok[4:])
			if err != nil {
				c.AddWarning("Invalid SpO2 value: " + tok)
			} else {
				c.Vitals.SpO2 = val
			}

		// Also handle spo (without the 2) for fast typing: spo298
		case strings.HasPrefix(tok, "spo"):
			val, err := strconv.Atoi(tok[3:])
			if err != nil {
				c.AddWarning("Invalid SpO2 value: " + tok)
			} else {
				c.Vitals.SpO2 = val
			}

		case strings.HasPrefix(tok, "temp"):
			c.Vitals.Temp = tok[4:]

		case strings.HasPrefix(tok, "rr"):
			val, err := strconv.Atoi(tok[2:])
			if err != nil {
				c.AddWarning("Invalid RR value: " + tok)
			} else {
				c.Vitals.RR = val
			}

		default:
			c.AddWarning("Unrecognized vitals token: " + tok)
		}
	}
}

// FormatVitals returns a human-readable vitals string.
func FormatVitals(v Vitals) string {
	parts := []string{}
	if v.BP != "" {
		parts = append(parts, "BP: "+v.BP)
	}
	if v.HR != 0 {
		parts = append(parts, "HR: "+strconv.Itoa(v.HR)+" bpm")
	}
	if v.SpO2 != 0 {
		parts = append(parts, "SpO2: "+strconv.Itoa(v.SpO2)+"%")
	}
	if v.Temp != "" {
		parts = append(parts, "Temp: "+v.Temp)
	}
	if v.RR != 0 {
		parts = append(parts, "RR: "+strconv.Itoa(v.RR)+" /min")
	}
	if len(parts) == 0 {
		return "N/A"
	}
	return strings.Join(parts, " | ")
}
