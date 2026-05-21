package engine

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
)

// CORE COMMAND REGISTRY

// ParserFunc is the type for all command handler functions, it would require to have 2 arguments one of which is the tokens slice and the other is the pointer to the ClinicalCase object

type ParserFunc func(tokens []string, c *ClinicalCase)

// coreCommands is a map that has a string command and a ParserFunc as its value. The parserFunc could be anything but it must take 2 arguments one of which is the tokens((a slice of strings after cmd in the line)) and the other is the pointer to the ClinicalCase object.

// Any command NOT in this map is passed to the extension system.

var coreCommands = map[string]ParserFunc{
	"pt": func(tokens []string, c *ClinicalCase) {
		ParsePatient(tokens, c, nil) // nil = no extensions (overridden in ParseString)
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
		ParseSymptoms(tokens, c, nil) // nil = no extensions (overridden in ParseString)
	},
	"vitals": func(tokens []string, c *ClinicalCase) {
		ParseVitals(tokens, c, nil) // nil = no extensions (overridden in ParseString)
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
	var pluginDataStore map[string]any

	// Local copy of commands to allow plugin injection.
	activeCommands := make(map[string]ParserFunc)
	maps.Copy(activeCommands, coreCommands)

	// ── CommandTokenRegistry ──────────────────────────────────────────────────
	// One fresh registry per parse session — zero global state.
	// Starts empty; populated when a plugin with CommandExtendable is loaded.
	// Re-register the three extensible core commands so they all share this reg.
	reg := NewCommandTokenRegistry()
	activeCommands["pt"] = func(tokens []string, c *ClinicalCase) {
		ParsePatient(tokens, c, reg)
	}
	activeCommands["vitals"] = func(tokens []string, c *ClinicalCase) {
		ParseVitals(tokens, c, reg)
	}
	activeCommands["sx"] = func(tokens []string, c *ClinicalCase) {
		ParseSymptoms(tokens, c, reg)
	}

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
			profileSpec := strings.TrimPrefix(cmd, "@")
			if profileSpec == "profile" && len(tokens) > 0 {
				profileSpec = strings.ToLower(tokens[0])
			}

			if profileSpec != "" {
				profiles := parseProfileList(profileSpec)
				if len(profiles) == 0 {
					c.AddWarning(fmt.Sprintf("Line %d: Invalid profile declaration", lineNum+1))
					continue
				}

				c.Profile = strings.Join(profiles, "+")
				if len(profiles) > 1 {
					pluginDataStore = make(map[string]any, len(profiles))
					c.SpecialtyData = pluginDataStore
				}

				for _, profileName := range profiles {
					plugin := GetPlugin(profileName)
					if plugin == nil {
						c.AddWarning(fmt.Sprintf("Line %d: Unknown profile '%s'", lineNum+1, profileName))
						continue
					}

					data := plugin.InitData()
					if len(profiles) == 1 {
						c.SpecialtyData = data
					} else {
						pluginDataStore[profileName] = data
					}

					for pluginCmd, pluginHandler := range plugin.GetCommands() {
						normalizedCmd := strings.ToLower(pluginCmd)
						if _, exists := activeCommands[normalizedCmd]; exists {
							c.AddWarning(fmt.Sprintf("Line %d: Plugin '%s' command '%s' ignored (additive-only mode)", lineNum+1, profileName, normalizedCmd))
							continue
						}
						activeCommands[normalizedCmd] = wrapPluginCommand(data, pluginHandler)
					}

					if extPlugin, ok := plugin.(CommandExtendable); ok {
						for extCmd, tokenMap := range extPlugin.GetCommandTokens() {
							for prefix, fn := range tokenMap {
								if !reg.RegisterUnique(extCmd, prefix, wrapTokenExt(data, fn)) {
									c.AddWarning(fmt.Sprintf("Line %d: Plugin '%s' token extension '%s' for command '%s' ignored (already registered)", lineNum+1, profileName, strings.ToLower(prefix), strings.ToLower(extCmd)))
								}
							}
						}
					}
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

func parseProfileList(spec string) []string {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(spec)), "+")
	unique := make(map[string]struct{}, len(parts))
	var out []string
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		if _, exists := unique[name]; exists {
			continue
		}
		unique[name] = struct{}{}
		out = append(out, name)
	}
	slices.Sort(out)
	return out
}

func wrapPluginCommand(data any, handler ParserFunc) ParserFunc {
	return func(tokens []string, c *ClinicalCase) {
		original := c.SpecialtyData
		c.SpecialtyData = data
		defer func() { c.SpecialtyData = original }()
		handler(tokens, c)
	}
}

func wrapTokenExt(data any, fn TokenExtFunc) TokenExtFunc {
	return func(token string, c *ClinicalCase) bool {
		original := c.SpecialtyData
		c.SpecialtyData = data
		defer func() { c.SpecialtyData = original }()
		return fn(token, c)
	}
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
