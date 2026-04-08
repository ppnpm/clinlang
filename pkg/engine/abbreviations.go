package engine

import "strings"

// Abbreviations maps common clinical shorthand to full medical terms.
// These are applied ONLY at output time. The internal Case struct
// always stores the original abbreviated form (fast to type and store).
//
// Keys must be lowercase (input is normalized before lookup).

var Abbreviations = map[string]string{
	// Chronic diseases
	"dm1":   "Type 1 Diabetes Mellitus",
	"dm2":   "Type 2 Diabetes Mellitus",
	"htn":   "Hypertension",
	"ihd":   "Ischaemic Heart Disease",
	"cad":   "Coronary Artery Disease",
	"chf":   "Congestive Heart Failure",
	"af":    "Atrial Fibrillation",
	"copd":  "COPD",
	"tb":    "Tuberculosis",
	"ckd":   "Chronic Kidney Disease",
	"ckd1":  "CKD Stage 1",
	"ckd2":  "CKD Stage 2",
	"ckd3":  "CKD Stage 3",
	"ckd4":  "CKD Stage 4",
	"ckd5":  "CKD Stage 5 / ESRD",
	"esrd":  "End-Stage Renal Disease",
	"cld":   "Chronic Liver Disease",
	"nafld": "Non-Alcoholic Fatty Liver Disease",
	"gerd":  "GERD",
	"ibs":   "Irritable Bowel Syndrome",
	"ra":    "Rheumatoid Arthritis",
	"sle":   "Systemic Lupus Erythematosus",
	"ms":    "Multiple Sclerosis",
	"cva":   "Cerebrovascular Accident (Stroke)",
	"tia":   "Transient Ischaemic Attack",
	"dvt":   "Deep Vein Thrombosis",
	"pe":    "Pulmonary Embolism",
	"cap":   "Community Acquired Pneumonia",
	"hap":   "Hospital Acquired Pneumonia",
	"uti":   "Urinary Tract Infection",
	"urti":  "Upper Respiratory Tract Infection",
	"lrti":  "Lower Respiratory Tract Infection",
	"mi":    "Myocardial Infarction",
	"stemi": "ST-Elevation MI",
	"nstemi":"Non-ST-Elevation MI",
	"dka":   "Diabetic Ketoacidosis",
	"hhs":   "Hyperosmolar Hyperglycaemic State",

	// Symptoms / exam
	"sob":   "Shortness of Breath",
	"cp":    "Chest Pain",
	"abd":   "Abdominal",
	"lbp":   "Low Back Pain",
	"ha":    "Headache",
	"nv":    "Nausea & Vomiting",
	"loc":   "Loss of Consciousness",
	"pnd":   "Paroxysmal Nocturnal Dyspnoea",
	"doe":   "Dyspnoea on Exertion",
	"jvp":   "Raised JVP",
	"ams":   "Altered Mental Status",

	// Drug frequency
	"od":   "Once daily",
	"bd":   "Twice daily",
	"tds":  "Three times daily",
	"qds":  "Four times daily",
	"prn":  "As needed",
	"stat": "Immediately",
	"nocte":"At night",
	"ac":   "Before meals",
	"pc":   "After meals",

	// Routes
	"iv":   "Intravenous",
	"im":   "Intramuscular",
	"sc":   "Subcutaneous",
	"sl":   "Sublingual",
	"po":   "Oral",
	"top":  "Topical",
	"neb":  "Nebulised",
	"inh":  "Inhaled",

	// Lab abbreviations used in pmh/hpi context
	"hyperlipidemia": "Hyperlipidaemia",
	"hyperlipidaemia":"Hyperlipidaemia",
	"asthma":         "Bronchial Asthma",
}

// ExpandAbbreviations takes a space-separated string (e.g. from pmh or cc)
// and replaces known abbreviations with their full forms.
// Unknown words are passed through unchanged.
func ExpandAbbreviations(input string) string {
	if input == "" {
		return ""
	}
	words := splitWords(input)
	out := make([]string, 0, len(words))
	for _, w := range words {
		lower := toLower(w)
		if expanded, ok := Abbreviations[lower]; ok {
			out = append(out, expanded)
		} else if isDuration(w) {
			out = append(out, ExpandDuration(w))
		} else {
			out = append(out, w)
		}
	}
	return joinWords(out)
}

// ExpandAbbr expands a single token. Used for individual fields.
func ExpandAbbr(token string) string {
	if exp, ok := Abbreviations[toLower(token)]; ok {
		return exp
	}
	return token
}

// Helper: split on spaces, preserving tokens
func splitWords(s string) []string {
	out := []string{}
	cur := ""
	for _, ch := range s {
		if ch == ' ' || ch == ',' {
			if cur != "" {
				out = append(out, cur)
			}
			if ch == ',' {
				out = append(out, ",")
			}
			cur = ""
		} else {
			cur += string(ch)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func joinWords(words []string) string {
	result := ""
	for i, w := range words {
		if w == "," {
			result += ","
		} else {
			if i > 0 && words[i-1] != "," {
				result += " "
			} else if i > 0 {
				result += " "
			}
			result += w
		}
	}
	return result
}

func toLower(s string) string {
	result := ""
	for _, ch := range s {
		if ch >= 'A' && ch <= 'Z' {
			result += string(ch + 32)
		} else {
			result += string(ch)
		}
	}
	return result
}

// ParseDurationToken splits "30min" into "30" and "min".
func ParseDurationToken(s string) (string, string) {
	var numStr string
	var i int
	for i = 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			numStr += string(s[i])
		} else {
			break
		}
	}
	return numStr, s[i:]
}

func IsValidDurationUnit(unit string) bool {
	switch strings.ToLower(unit) {
	case "y", "yr", "yrs", "year", "years":
		return true
	case "mo", "mos", "m", "month", "months":
		return true
	case "w", "wk", "wks", "week", "weeks":
		return true
	case "d", "dy", "dys", "day", "days":
		return true
	case "h", "hr", "hrs", "hour", "hours":
		return true
	case "min", "mins", "minute", "minutes":
		return true
	case "s", "sec", "secs", "second", "seconds":
		return true
	}
	return false
}

func GetDurationUnitWord(unit string, plural bool) string {
	var word string
	switch strings.ToLower(unit) {
	case "y", "yr", "yrs", "year", "years":
		word = "year"
	case "mo", "mos", "m", "month", "months":
		word = "month"
	case "w", "wk", "wks", "week", "weeks":
		word = "week"
	case "d", "dy", "dys", "day", "days":
		word = "day"
	case "h", "hr", "hrs", "hour", "hours":
		word = "hour"
	case "min", "mins", "minute", "minutes":
		word = "minute"
	case "s", "sec", "secs", "second", "seconds":
		word = "second"
	default:
		return unit
	}
	if plural {
		return word + "s"
	}
	return word
}

func NormalizeDurationUnitShort(unit string) string {
	switch strings.ToLower(unit) {
	case "y", "yr", "yrs", "year", "years":
		return "y"
	case "mo", "mos", "m", "month", "months":
		return "mo"
	case "w", "wk", "wks", "week", "weeks":
		return "w"
	case "d", "dy", "dys", "day", "days":
		return "d"
	case "h", "hr", "hrs", "hour", "hours":
		return "h"
	case "min", "mins", "minute", "minutes":
		return "min"
	case "s", "sec", "secs", "second", "seconds":
		return "s"
	}
	return unit
}

// ExpandDuration converts short duration formats (e.g. 30min, 2w, 1mo, 4h) into full words.
func ExpandDuration(duration string) string {
	numStr, unitStr := ParseDurationToken(duration)
	if numStr == "" || unitStr == "" || !IsValidDurationUnit(unitStr) {
		return duration
	}
	
	plural := numStr != "1"
	word := GetDurationUnitWord(unitStr, plural)
	return numStr + " " + word
}
