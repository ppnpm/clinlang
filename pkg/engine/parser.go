package engine

import (
	"maps"
	"fmt"
	"os"
	"strings"
)

// =============================================================================
// CORE COMMAND REGISTRY
// =============================================================================

// ParserFunc is the signature for all command handler functions.
type ParserFunc func(tokens []string, c *Case)

// coreCommands maps command keywords to their dedicated parsers.
// Any command NOT in this map is passed to the extension system.
var coreCommands = map[string]ParserFunc{
	"pt": func(tokens []string, c *Case) {
		ParsePatient(tokens, c)
	},
	"cc": func(tokens []string, c *Case) {
		c.CC = strings.Join(tokens, " ")
	},
	"hpi": func(tokens []string, c *Case) {
		c.HPI = strings.Join(tokens, " ")
	},
	"pmh": func(tokens []string, c *Case) {
		c.PMH = strings.Join(tokens, " ")
	},
	"dx": func(tokens []string, c *Case) {
		c.DX = strings.Join(tokens, " ")
	},
	"ddx": func(tokens []string, c *Case) {
		c.DDX = strings.Join(tokens, " ")
	},
	"sx": func(tokens []string, c *Case) {
		ParseSymptoms(tokens, c)
	},
	"vitals": func(tokens []string, c *Case) {
		ParseVitals(tokens, c)
	},
	"rx": func(tokens []string, c *Case) {
		ParsePrescription(tokens, c)
	},
	"id": func(tokens []string, c *Case) {
		if len(tokens) > 0 {
			c.Patient.Id = tokens[0]
		}
	},
}

// GetCoreCommandNames returns all standard keywords.
func GetCoreCommandNames() []string {
	var names []string
	for k := range coreCommands {
		names = append(names, k)
	}
	return names
}

// =============================================================================
// LIBRARY ENTRY POINT
// =============================================================================

// ParseString is the primary library function. It takes raw .cln text
// (not a file path) and returns a fully parsed Case.
//
// This is the function your app/server should call.
// It never reads from disk — the caller passes the string content.
func ParseString(input string) Case {
	c := NewCase()
	c.Profile = "general"

	// Local copy of commands to allow plugin injection
	activeCommands := make(map[string]ParserFunc)
	maps.Copy(activeCommands, coreCommands)

	lines := strings.Split(input, "\n")

	for lineNum, line := range lines {
		// Normalize: strip Windows \r, trim whitespace
		line = strings.TrimSpace(strings.ReplaceAll(line, "\r", ""))

		// Skip blanks, comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// Skip the old .cln terminator dot (if someone uses it)
		if line == "." {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		// Normalize command to lowercase
		cmd := strings.ToLower(parts[0])
		tokens := parts[1:]

		// Handle @profile plugin loading
		if cmd == "@profile" && len(tokens) > 0 {
			profileName := strings.ToLower(tokens[0])
			c.Profile = profileName
			plugin := GetPlugin(profileName)
			if plugin != nil {
				c.SpecialtyData = plugin.InitData()
				maps.Copy(activeCommands, plugin.GetCommands())
			} else {
				c.AddWarning(fmt.Sprintf("Line %d: Unknown profile '%s'", lineNum+1, profileName))
			}
			continue
		}

		if handler, ok := activeCommands[cmd]; ok {
			handler(tokens, &c)
		} else {
			// Extension system: unrecognized command → Extra[cmd]
			if len(tokens) == 0 {
				c.AddWarning(fmt.Sprintf("Line %d: command '%s' has no arguments", lineNum+1, cmd))
				continue
			}
			ParseExtension(cmd, tokens, &c)
		}
	}

	// Auto-generate ID if not provided in the .cln file
	if c.Patient.Id == "" {
		c.Patient.Id = GenerateId()
	}

	// Run abnormal value checks (no crashes, only flags)
	CheckAbnormals(&c)

	// Globally expand abbreviations so they appear in JSON
	c.CC = ExpandAbbreviations(c.CC)
	c.HPI = ExpandAbbreviations(c.HPI)
	c.PMH = ExpandAbbreviations(c.PMH)
	c.DX = ExpandAbbreviations(c.DX)
	c.DDX = ExpandAbbreviations(c.DDX)
	for i := range c.Symptoms {
		c.Symptoms[i].Name = ExpandAbbreviations(c.Symptoms[i].Name)
	}

	return c
}

// ParseFile reads a .cln file from disk and calls ParseString.
// Used by the local CLI. Apps should use ParseString directly.
func ParseFile(path string) (Case, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Case{}, fmt.Errorf("cannot read file: %w", err)
	}
	return ParseString(string(data)), nil
}
