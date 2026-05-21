package engine

import (
	"math"
	"strconv"
	"strings"
)

// ParsePatient parses the tokens after the `pt` command.
// Supported formats:
//   - age:       34, 30min, 6mo, 2w
//   - sex:       M / F
//   - age+sex:   34M / 5dF / 30minM
//
// reg is the per-parse CommandTokenRegistry. Pass nil when no plugin is active.
// Any token not handled by core logic is offered to the plugin via reg.Try("pt", ...)
// before a warning is emitted, so plugins can extend pt without modifying this file.
func ParsePatient(tokens []string, c *ClinicalCase, reg *CommandTokenRegistry) {
	for _, rawTok := range tokens {
		rawTok = strings.TrimSpace(rawTok)
		if rawTok == "" {
			continue
		}

		tok := rawTok
		lowerTok := strings.ToLower(tok)

		// Weight prefix: wt<num>
		if strings.HasPrefix(lowerTok, "wt") {
			numPart := tok[2:]
			// Handle decimals and common suffixes like 'kg'
			numPart = strings.TrimRight(numPart, "kgKG")
			val, err := strconv.ParseFloat(numPart, 64)
			if err == nil {
				c.Patient.Weight = val
				continue
			}
		}

		// Height prefix: ht<num>
		if strings.HasPrefix(lowerTok, "ht") {
			numPart := tok[2:]
			numPart = strings.TrimRight(numPart, "cmCM")
			val, err := strconv.ParseFloat(numPart, 64)
			if err == nil {
				c.Patient.Height = val
				continue
			}
		}

		if strings.HasPrefix(lowerTok, "bed") {
			numPart := tok[3:]
			numPart = strings.TrimRight(numPart, "bedBED")
			val, err := strconv.Atoi(numPart)
			if err == nil {
				c.Patient.Bed = val
				continue
			}
		}

		if strings.HasPrefix(lowerTok, "unit:") {
			unitName := tok[5:] // everything after "unit:" is the unit name as-is
			if unitName != "" {
				c.Patient.Unit = unitName
				continue
			}
		}

		// Try separating trailing 'M' or 'F' (case-sensitive) for biological sex
		sexChar := ""
		if len(tok) > 0 {
			lastChar := tok[len(tok)-1]
			if lastChar == 'M' || lastChar == 'F' {
				sexChar = string(lastChar)
				tok = tok[:len(tok)-1]
			}
		}

		if sexChar != "" {
			c.Patient.Sex = sexChar
		}

		if tok == "" {
			continue // token was just "M" or "F"
		}

		// Pure Digits -> default age unit
		if isAllDigits(tok) && tok != "" {
			c.Patient.Age, _ = strconv.Atoi(tok)
			c.Patient.AgeUnit = "Y"
			continue
		}

		// ── Plugin token extension ────────────────────────────────────────────
		// Run BEFORE ParseDurationToken so that tokens like ga32w are intercepted
		// by the plugin before the duration parser misreads them as age values.
		// Passes tok (sex-suffix already stripped) so handlers see clean values.
		if reg.Try("pt", tok, c) {
			continue
		}

		// Evaluate remaining token as a duration (e.g. 30min, 30m, 6mo, 2w, 3d)
		numStr, unitStr := ParseDurationToken(tok)
		if numStr != "" && IsValidDurationUnit(unitStr) {
			c.Patient.Age, _ = strconv.Atoi(numStr)
			c.Patient.AgeUnit = NormalizeDurationUnitShort(unitStr)
			continue
		}

		c.AddWarning("Unrecognized patient token: " + rawTok)
	}

	// Auto-calculate BMI and BSA once both weight and height are known.
	if c.Patient.Weight > 0 && c.Patient.Height > 0 {
		htM := c.Patient.Height / 100.0                                           // cm → m
		c.Patient.BMI = c.Patient.Weight / (htM * htM)                            // kg/m²
		c.Patient.BSA = math.Sqrt((c.Patient.Height * c.Patient.Weight) / 3600.0) // Mosteller formula
	}

	// Validate
	if c.Patient.Age == 0 {
		c.AddWarning("Patient age not specified")
	}
	if c.Patient.Sex == "" {
		c.AddWarning("Patient sex not specified")
	}
}

// isAllDigits returns true if every character in s is 0-9.
func isAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
