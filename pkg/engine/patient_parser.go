package engine

import (
	"strconv"
	"strings"
)

// ParsePatient parses the tokens after the `pt` command.
// Supported formats:
//   - age:       34, 30min, 6mo, 2w
//   - sex:       M / F
//   - age+sex:   34M / 5dF / 30minM
func ParsePatient(tokens []string, c *Case) {
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

		// Obstetrics: MOA (Months of Amenorrhoea)
		if strings.HasPrefix(lowerTok, "moa") && isAllDigits(tok[3:]) {
			c.Patient.MOA = tok[3:]
			continue
		}

		// Obstetrics: GPAL scores
		if (lowerTok[0] == 'g') && strings.Contains(lowerTok, "p") {
			c.Patient.GPAL = strings.ToUpper(tok)
			continue
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

		// Evaluate remaining token as a duration (e.g. 30min, 30m, 6mo, 2w, 3d)
		numStr, unitStr := ParseDurationToken(tok)
		if numStr != "" && IsValidDurationUnit(unitStr) {
			c.Patient.Age, _ = strconv.Atoi(numStr)
			c.Patient.AgeUnit = NormalizeDurationUnitShort(unitStr)
			continue
		}

		// We didn't resolve anything explicitly
		c.AddWarning("Unrecognized patient token: " + rawTok)
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
