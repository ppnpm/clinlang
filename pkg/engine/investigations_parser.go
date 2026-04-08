package engine

import "strings"

// parseInvestigationToken breaks down complex clinical investigations.
func parseInvestigationToken(tok string) (key, value string) {
	// 0. Special ClinicalCase: Keys that naturally contain numbers (e.g. pco2, hba1c, b12)
	numKeys := []string{"pco2", "po2", "fio2", "spo2", "co2", "o2", "b12", "hba1c", "a1c"}
	lowerTok := strings.ToLower(tok)
	for _, k := range numKeys {
		if strings.HasPrefix(lowerTok, k) {
			val := tok[len(k):]
			val = strings.TrimPrefix(val, ":") // Clean pco2:45 down to 45
			if val != "" {
				return tok[:len(k)], val
			}
		}
	}

	// 1. Explicit Colon Override
	// If the user explicitly types a colon anywhere, we treat EVERYTHING 
	// before the FIRST colon as the key, and everything after as the value.
	// This safely handles edge cases like ca125:40 or cxr:wnl
	colonIdx := strings.Index(tok, ":")
	if colonIdx != -1 {
		parts := strings.SplitN(tok, ":", 2)
		return parts[0], parts[1]
	}

	// 2. Check for Serology / Positive/Negative flags ending in + or - (hiv+, crp+++, dengue-)
	if strings.HasSuffix(tok, "+") || strings.HasSuffix(tok, "-") {
		splitAt := -1
		for i := len(tok) - 1; i >= 0; i-- {
			if tok[i] != '+' && tok[i] != '-' {
				splitAt = i + 1
				break
			}
		}
		if splitAt != -1 && splitAt > 0 {
			return tok[:splitAt], tok[splitAt:]
		}
	}

	// 3. Fallback: numerical split for zero-punctuation fast dictation (e.g. hb9.2)
	for i, ch := range tok {
		if ch >= '0' && ch <= '9' {
			return tok[:i], tok[i:]
		}
	}

	// No numbers, colons, or +/-, treat as a boolean flag
	return tok, "true"
}

// ParseLab forces results directly into c.Labs
func ParseLab(tokens []string, c *ClinicalCase) {
	for _, raw := range tokens {
		tok := strings.TrimSpace(raw)
		if tok == "" {
			continue
		}
		key, value := parseInvestigationToken(tok)
		c.Labs[key] = value
	}
}

// ParseRad forces results directly into c.Imaging
func ParseRad(tokens []string, c *ClinicalCase) {
	for _, raw := range tokens {
		tok := strings.TrimSpace(raw)
		if tok == "" {
			continue
		}
		key, value := parseInvestigationToken(tok)
		c.Imaging[key] = value
	}
}

// radKeys is a dictionary of common radiology and imaging prefixes to auto-route ix commands.
var radKeys = map[string]bool{
	"cxr": true, "xray": true, "x-ray": true, "ct": true, "mri": true,
	"usg": true, "echo": true, "ecg": true, "ekg": true, "pet": true,
	"ultrasound": true,
}

// ParseIx acts as an umbrella, routing elements into either Labs or Imaging
func ParseIx(tokens []string, c *ClinicalCase) {
	for _, raw := range tokens {
		tok := strings.TrimSpace(raw)
		if tok == "" {
			continue
		}
		key, value := parseInvestigationToken(tok)

		// Auto routing based on prefix
		routeToRad := false
		lowerKey := strings.ToLower(key)
		for r := range radKeys {
			if strings.HasPrefix(lowerKey, r) {
				routeToRad = true
				break
			}
		}

		if routeToRad {
			c.Imaging[key] = value
		} else {
			c.Labs[key] = value
		}
	}
}
