package engine

import "strings"

// Abbreviations maps common clinical shorthand to full medical terms.
// These are applied ONLY at output time. The internal Case struct
// always stores the original abbreviated form (fast to type and store).
//
// Keys must be lowercase (input is normalized before lookup).

var Abbreviations map[string]string

// ExpandAbbreviations takes a space-separated string (e.g. from pmh or cc)
// and replaces known abbreviations with their full forms.
// Unknown words are passed through unchanged.
func ExpandAbbreviations(input string) string {
	return ExpandAbbreviationsWithOptions(input, nil, nil)
}

// ExpandAbbreviationsWithOptions implements ExpandAbbreviations with optional custom config.
func ExpandAbbreviationsWithOptions(input string, customAbbr map[string]string, customDur map[string]DurationUnit) string {
	if input == "" {
		return ""
	}
	abbrMap := Abbreviations
	if customAbbr != nil {
		abbrMap = customAbbr
	}
	words := splitWords(input)
	out := make([]string, 0, len(words))
	for _, w := range words {
		lower := toLower(w)
		if expanded, ok := abbrMap[lower]; ok {
			out = append(out, expanded)
		} else if isDurationWithOptions(w, customDur) {
			out = append(out, ExpandDurationWithOptions(w, customDur))
		} else {
			out = append(out, w)
		}
	}
	return joinWords(out)
}

// ExpandAbbr expands a single token. Used for individual fields.
func ExpandAbbr(token string) string {
	return ExpandAbbrWithOptions(token, nil)
}

// ExpandAbbrWithOptions expands a single token with custom abbreviations map.
func ExpandAbbrWithOptions(token string, customAbbr map[string]string) string {
	abbrMap := Abbreviations
	if customAbbr != nil {
		abbrMap = customAbbr
	}
	if exp, ok := abbrMap[toLower(token)]; ok {
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
	return IsValidDurationUnitWithOptions(unit, nil)
}

func IsValidDurationUnitWithOptions(unit string, customDur map[string]DurationUnit) bool {
	durMap := DefaultConfig.Durations
	if customDur != nil {
		durMap = customDur
	}
	lower := strings.ToLower(unit)
	for _, du := range durMap {
		for _, alias := range du.Aliases {
			if strings.ToLower(alias) == lower {
				return true
			}
		}
	}
	return false
}

func GetDurationUnitWord(unit string, plural bool) string {
	return GetDurationUnitWordWithOptions(unit, plural, nil)
}

func GetDurationUnitWordWithOptions(unit string, plural bool, customDur map[string]DurationUnit) string {
	durMap := DefaultConfig.Durations
	if customDur != nil {
		durMap = customDur
	}
	lower := strings.ToLower(unit)
	for _, du := range durMap {
		for _, alias := range du.Aliases {
			if strings.ToLower(alias) == lower {
				word := du.Word
				if plural {
					return word + "s"
				}
				return word
			}
		}
	}
	return unit
}

func NormalizeDurationUnitShort(unit string) string {
	return NormalizeDurationUnitShortWithOptions(unit, nil)
}

func NormalizeDurationUnitShortWithOptions(unit string, customDur map[string]DurationUnit) string {
	durMap := DefaultConfig.Durations
	if customDur != nil {
		durMap = customDur
	}
	lower := strings.ToLower(unit)
	for _, du := range durMap {
		for _, alias := range du.Aliases {
			if strings.ToLower(alias) == lower {
				return du.Short
			}
		}
	}
	return unit
}

// ExpandDuration converts short duration formats (e.g. 30min, 2w, 1mo, 4h) into full words.
func ExpandDuration(duration string) string {
	return ExpandDurationWithOptions(duration, nil)
}

// ExpandDurationWithOptions converts short duration formats with custom configuration.
func ExpandDurationWithOptions(duration string, customDur map[string]DurationUnit) string {
	numStr, unitStr := ParseDurationToken(duration)
	if numStr == "" || unitStr == "" || !IsValidDurationUnitWithOptions(unitStr, customDur) {
		return duration
	}
	
	plural := numStr != "1"
	word := GetDurationUnitWordWithOptions(unitStr, plural, customDur)
	return numStr + " " + word
}

func isDuration(s string) bool {
	return isDurationWithOptions(s, nil)
}

func isDurationWithOptions(s string, customDur map[string]DurationUnit) bool {
	num, unit := ParseDurationToken(s)
	return num != "" && IsValidDurationUnitWithOptions(unit, customDur)
}
