package engine

import (
	"maps"
	"fmt"
	"os"
	"strings"
)

// CORE COMMAND REGISTRY

// ParserFunc is the type for all command handler functions, it would require to have 2 arguments one of which is the tokens slice and the other is the pointer to the ClinicalCase object

type ParserFunc func(tokens []string, c *ClinicalCase)

// coreCommands is a map that has a string command and a ParserFunc as its value. The parserFunc could be anything but it must take 2 arguments one of which is the tokens((a slice of strings after cmd in the line)) and the other is the pointer to the ClinicalCase object.

// Any command NOT in this map is passed to the extension system.

var coreCommands = map[string]ParserFunc{
	"pt": func(tokens []string, c *ClinicalCase) {
		ParsePatient(tokens, c)
	},
	"cc": func(tokens []string, c *ClinicalCase) {
		// since cc is a free text, all the tokens in the slice would be joined and make a long string. Hence for now let it be like that later, I think we will built a parser for CC as well. I think CC would be written in order of time.

		c.CC = strings.Join(tokens, " ")

	},
	"hpi": func(tokens []string, c *ClinicalCase) {
		c.HPI = strings.Join(tokens, " ")
	},
	"pmh": func(tokens []string, c *ClinicalCase) {
		c.PMH = strings.Join(tokens, " ")
	},
	"dx": func(tokens []string, c *ClinicalCase) {
		c.DX = strings.Join(tokens, " ")
	},
	"ddx": func(tokens []string, c *ClinicalCase) {
		c.DDX = strings.Join(tokens, " ")
	},
	"sx": func(tokens []string, c *ClinicalCase) {
		ParseSymptoms(tokens, c)
	},
	"vitals": func(tokens []string, c *ClinicalCase) {
		ParseVitals(tokens, c)
	},
	// New prescription
	"rx": func(tokens []string, c *ClinicalCase) {
		ParsePrescription(tokens, c)
	},
	"rhx": func(tokens []string, c *ClinicalCase) {
		c.AddWarning("rhx (past treatment) not yet natively parsed, treated as general token")
	},
	"day": func(tokens []string, c *ClinicalCase) {
		c.Day = strings.Join(tokens, " ")
	},
	"alg": func(tokens []string, c *ClinicalCase) {
		c.Allergies = strings.Join(tokens, " ")
	},
	"allergy": func(tokens []string, c *ClinicalCase) {
		c.Allergies = strings.Join(tokens, " ")
	},
	"sh": func(tokens []string, c *ClinicalCase) {
		c.SH = strings.Join(tokens, " ")
	},
	"fh": func(tokens []string, c *ClinicalCase) {
		c.FH = strings.Join(tokens, " ")
	},
	"pe": func(tokens []string, c *ClinicalCase) {
		c.PE = strings.Join(tokens, " ")
	},
	"oe": func(tokens []string, c *ClinicalCase) {
		c.PE = strings.Join(tokens, " ")
	},
	"lab": func(tokens []string, c *ClinicalCase) {
		ParseLab(tokens, c)
	},
	"labs": func(tokens []string, c *ClinicalCase) {
		ParseLab(tokens, c)
	},
	"rad": func(tokens []string, c *ClinicalCase) {
		ParseRad(tokens, c)
	},
	"ix": func(tokens []string, c *ClinicalCase) {
		ParseIx(tokens, c)
	},

	"id": func(tokens []string, c *ClinicalCase) {
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

// LIBRARY ENTRY POINT

// ParseString is the primary library function. It takes raw .cln text

// (not a file path) and returns a fully parsed ClinicalCase.

// This is the function your app/server should call.

// It never reads from disk — the caller passes the string content.

func ParseString(input string) ClinicalCase {
	c := NewClinicalCase()
	c.Profile = "general"

	// Local copy of commands to allow plugin injection.
	activeCommands := make(map[string]ParserFunc)
	maps.Copy(activeCommands, coreCommands)

	lines := strings.Split(input, "\n")

	for lineNum, line := range lines {
		// Normalize: strip Windows \r, trim whitespace.
		line = strings.TrimSpace(strings.ReplaceAll(line, "\r", ""))

		// Skip blanks, comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// Skip the old .cln terminator dot (if someone uses it).
		if line == "." {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		// Normalize command to lowercase.
		cmd := strings.ToLower(parts[0])
		tokens := parts[1:]

		// Handle @profile plugin loading - check for @prefix directly
		if strings.HasPrefix(cmd, "@") {
			profileName := strings.TrimPrefix(cmd, "@")
			if profileName == "profile" && len(tokens) > 0 {
				profileName = strings.ToLower(tokens[0])
			}
			
			if profileName != "" {
				c.Profile = profileName
				plugin := GetPlugin(profileName)
				if plugin != nil {
					c.SpecialtyData = plugin.InitData()
					maps.Copy(activeCommands, plugin.GetCommands())
				} else {
					c.AddWarning(fmt.Sprintf("Line %d: Unknown profile '%s'", lineNum+1, profileName))
				}
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
	c.SH = ExpandAbbreviations(c.SH)
	c.FH = ExpandAbbreviations(c.FH)
	c.PE = ExpandAbbreviations(c.PE)
	c.Allergies = ExpandAbbreviations(c.Allergies)
	c.DX = ExpandAbbreviations(c.DX)
	c.DDX = ExpandAbbreviations(c.DDX)
	for i := range c.Symptoms {
		c.Symptoms[i].Name = ExpandAbbreviations(c.Symptoms[i].Name)
	}

	return c
}

// ParseFile reads a .cln file from disk and calls ParseString.

// Used by the local CLI. Apps should use ParseString directly.

//this function is an example of multiple return value function where it return ClinicalCase{} object and error

func ParseFile(path string) (ClinicalCase, error) {
	data, err := os.ReadFile(path)
	// fmt.Println("Data: ", string(data))
	if err != nil {
		return ClinicalCase{}, fmt.Errorf("cannot read file: %w", err)
	}
	return ParseString(string(data)), nil
}
