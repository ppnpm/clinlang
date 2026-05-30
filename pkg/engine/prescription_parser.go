package engine

import (
	"strings"
)

// Prescription represents a single medication order.
type Prescription struct {
	Drug      string `json:"drug"`
	Dose      string `json:"dose"`       // e.g. "25mg", "500mg"
	Frequency string `json:"frequency"`  // e.g. "BD", "OD"
	Route     string `json:"route"`      // e.g. "IV", "PO" (optional)
	Duration  string `json:"duration"`   // e.g. "7d", "5d" (optional)
}

// frequencyAliases normalizes shorthand freq tokens to uppercase labels.
var frequencyAliases = map[string]string{
	"od":    "OD",
	"bd":    "BD",
	"tds":   "TDS",
	"qds":   "QDS",
	"qid":   "QDS",
	"tid":   "TDS",
	"bid":   "BD",
	"stat":  "STAT",
	"prn":   "PRN",
	"nocte": "Nocte",
	"q1h":   "Q1H",
	"qh":    "Q1H",
	"q4h":   "Q4H",
	"q6h":   "Q6H",
	"q8h":   "Q8H",
	"q12h":  "Q12H",
	"q24h":  "Q24H",
	"qam":   "QAM",
	"om":    "QAM",
	"qpm":   "QPM",
	"qod":   "QOD",
	"eod":   "QOD",
	"qw":    "QW",
	"qwk":   "QW",
	"biw":   "BIW",
	"qm":    "QM",
}

// routeAliases normalizes route tokens.
var routeAliases = map[string]string{
	"iv":  "IV",
	"im":  "IM",
	"sc":  "SC",
	"sl":  "SL",
	"po":  "PO",
	"top": "Topical",
	"neb": "Nebulised",
	"inh": "Inhaled",
}

// ParsePrescription parses tokens after the `rx` command.
//
// Expected token order (all after drug name are flexible):
//
//	rx <drug> <dose> <frequency> [route] [x<duration>]
//
// Examples:
//
//	rx metoprolol 25mg bd
//	rx amoxicillin 500mg tds po x7d
//	rx lasix 40mg od iv
//	rx paracetamol 1g qds prn
func ParsePrescription(tokens []string, c *ClinicalCase) {
	if len(tokens) == 0 {
		c.AddWarning("rx: missing drug name")
		return
	}

	p := Prescription{}

	// Match against known drugs to properly consume multi-word drugs (e.g. "vitamin b12")
	fullText := strings.ToLower(strings.Join(tokens, " "))
	longestMatch := ""
	consumedTokens := 1

	drugs := DrugsList
	if c != nil && c.Config != nil && len(c.Config.Drugs) > 0 {
		drugs = c.Config.Drugs
	}

	for _, drug := range drugs {
		drugLower := strings.ToLower(drug)
		if strings.HasPrefix(fullText, drugLower) {
			if len(fullText) == len(drugLower) || fullText[len(drugLower)] == ' ' {
				if len(drugLower) > len(longestMatch) {
					longestMatch = drug
					consumedTokens = len(strings.Split(drugLower, " "))
				}
			}
		}
	}

	if longestMatch != "" {
		p.Drug = longestMatch
	} else {
		// Fallback for unknown drugs: accumulate tokens until a dose, frequency, route, or duration is hit
		p.Drug = tokens[0]
		for i := 1; i < len(tokens); i++ {
			tok := strings.ToLower(strings.TrimSpace(tokens[i]))
			if tok == "" {
				consumedTokens++
				continue
			}

			// Check if this token is a modifier
			isMod := false
			freqAliases := getFrequencyAliases(c)
			rtAliases := getRouteAliases(c)
			if containsDigit(tok) {
				isMod = true
			} else if _, ok := freqAliases[tok]; ok {
				isMod = true
			} else if _, ok := rtAliases[tok]; ok {
				isMod = true
			} else if isDurationWithCase(tok, c) || (strings.HasPrefix(tok, "x") && len(tok) > 1 && isDurationWithCase(tok[1:], c)) {
				isMod = true
			}

			if isMod {
				break
			}
			p.Drug += " " + tokens[i]
			consumedTokens++
		}
	}

	for _, tok := range tokens[consumedTokens:] {
		tok = strings.ToLower(strings.TrimSpace(tok))
		if tok == "" {
			continue
		}

		// Duration: starts with 'x' → x7d, x5d
		if strings.HasPrefix(tok, "x") && len(tok) > 1 {
			rest := tok[1:]
			if isDurationWithCase(rest, c) {
				p.Duration = rest
				continue
			}
		}

		// Plain duration without 'x' prefix: 7d, 5d, 2w
		if isDurationWithCase(tok, c) {
			p.Duration = tok
			continue
		}

		// Frequency alias
		freqAliases := getFrequencyAliases(c)
		if freq, ok := freqAliases[tok]; ok {
			p.Frequency = freq
			continue
		}

		// Route alias
		rtAliases := getRouteAliases(c)
		if route, ok := rtAliases[tok]; ok {
			p.Route = route
			continue
		}

		// Dose: contains a digit  → 25mg, 500mg, 1g, 40mg, 0.5mg
		if containsDigit(tok) && p.Dose == "" {
			p.Dose = tok
			continue
		}

		// Unknown token — warn but don't crash
		c.AddWarning("rx " + p.Drug + ": unrecognized token '" + tok + "'")
	}

	if p.Frequency == "" {
		c.AddWarning("rx " + p.Drug + ": frequency not specified")
	}

	c.Prescriptions = append(c.Prescriptions, p)
}

func getFrequencyAliases(c *ClinicalCase) map[string]string {
	if c != nil && c.Config != nil && len(c.Config.Frequencies.Aliases) > 0 {
		return c.Config.Frequencies.Aliases
	}
	return frequencyAliases
}

func getRouteAliases(c *ClinicalCase) map[string]string {
	if c != nil && c.Config != nil && len(c.Config.Routes.Aliases) > 0 {
		return c.Config.Routes.Aliases
	}
	return routeAliases
}

func getRouteExpansion(route string, config *ParserConfig) string {
	if config != nil && len(config.Routes.Expansions) > 0 {
		if val, ok := config.Routes.Expansions[route]; ok {
			return val
		}
	}
	if val, ok := DefaultConfig.Routes.Expansions[route]; ok {
		return val
	}
	if val, ok := expandedRoutes[route]; ok {
		return val
	}
	return route
}

func getFrequencyExpansion(freq string, config *ParserConfig) string {
	if config != nil && len(config.Frequencies.Expansions) > 0 {
		if val, ok := config.Frequencies.Expansions[freq]; ok {
			return val
		}
	}
	if val, ok := DefaultConfig.Frequencies.Expansions[freq]; ok {
		return val
	}
	if val, ok := expandedFrequencies[freq]; ok {
		return val
	}
	return freq
}

// containsDigit returns true if any character in s is 0-9.
func containsDigit(s string) bool {
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			return true
		}
	}
	return false
}

// expandedFrequencies provides user-friendly text for clinical notes.
var expandedFrequencies = map[string]string{
	"OD":    "Once daily",
	"BD":    "Twice daily",
	"TDS":   "Three times daily",
	"QDS":   "Four times daily",
	"STAT":  "Immediately",
	"PRN":   "As needed",
	"Nocte": "At night",
	"Q1H":   "Every hour",
	"Q4H":   "Every 4 hours",
	"Q6H":   "Every 6 hours",
	"Q8H":   "Every 8 hours",
	"Q12H":  "Every 12 hours",
	"Q24H":  "Every 24 hours",
	"QAM":   "Every morning",
	"QPM":   "Every night",
	"QOD":   "Every other day",
	"QW":    "Once weekly",
	"BIW":   "Twice weekly",
	"QM":    "Once monthly",
}

// expandedRoutes provides user-friendly text for clinical notes.
var expandedRoutes = map[string]string{
	"IV":        "Intravenously",
	"IM":        "Intramuscularly",
	"SC":        "Subcutaneously",
	"SL":        "Sublingually",
	"PO":        "Orally",
	"TOP":   "Topically",
	"NEB": "Nebulised",
	"INH":   "Inhaled",
}

// FormatPrescriptions formats the medication list for clinical note output.
func FormatPrescriptions(c ClinicalCase) string {
	prescriptions := c.Prescriptions
	if len(prescriptions) == 0 {
		return ""
	}
	sb := strings.Builder{}
	sb.WriteString("Medications:\n")
	for _, p := range prescriptions {
		line := "  ▸ " + titleCase(p.Drug)
		if p.Dose != "" {
			line += " " + p.Dose
		}
		
		routeStr := getRouteExpansion(p.Route, c.Config)
		if routeStr != "" {
			line += " " + routeStr
		}

		freqStr := getFrequencyExpansion(p.Frequency, c.Config)
		if freqStr != "" {
			line += ", " + freqStr
		}

		if p.Duration != "" {
			var customDur map[string]DurationUnit
			if c.Config != nil {
				customDur = c.Config.Durations
			}
			line += " for " + ExpandDurationWithOptions(p.Duration, customDur)
		}
		sb.WriteString(line + "\n")
	}
	return sb.String()
}

// titleCase capitalizes the first letter of a string.
func titleCase(s string) string {
	if s == "" {
		return ""
	}
	first := strings.ToUpper(string(s[0]))
	return first + s[1:]
}
