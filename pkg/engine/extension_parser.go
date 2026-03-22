package engine

import "strings"

// ParseExtension handles any command that is NOT in the core command set.
// It stores all tokens as key-value pairs under the command namespace.
//
// Token format: key<value>   e.g. hb13.5, wbc12000, na138
//
// Parsing Strategy:
//   - Scan from right-to-left to find where the numeric/unit value starts.
//   - Everything before the number is the key; everything from the first
//     digit onwards is the value.
//
// Example:
//
//	cmd = "lab", tokens = ["hb13", "wbc12000", "plt150000"]
//	Result: Extra["lab"]["hb"]="13", Extra["lab"]["wbc"]="12000"
func ParseExtension(cmd string, tokens []string, c *Case) {
	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}

		key, value := splitKeyValue(tok)
		if key == "" {
			// Treat the whole token as a flag/marker (e.g., "fasting")
			c.SetExtra(cmd, tok, "true")
		} else {
			c.SetExtra(cmd, key, value)
		}
	}
}

// splitKeyValue splits a token like "hb13.5" into key="hb" and value="13.5".
// It finds the first digit and splits there.
// If no digit is found, it returns ("", token) so the whole token is used as key.
func splitKeyValue(tok string) (key, value string) {
	splitAt := -1
	for i, ch := range tok {
		if (ch >= '0' && ch <= '9') || ch == '.' {
			splitAt = i
			break
		}
	}
	if splitAt == -1 {
		return "", tok
	}
	return tok[:splitAt], tok[splitAt:]
}
