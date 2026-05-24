// Package obgyn provides the OB/GYN specialty plugin for ClinScript.
//
// It demonstrates the full plugin contract:
//   - GetCommands()      → new standalone commands (lmp, edd, gpal, fhs, ctx)
//   - GetCommandTokens() → inline token extensions for core commands
//                          "pt"     → ga:<weeks>w   (gestational age)
//                          "vitals" → fhr:<bpm>     (fetal heart rate)
//
// Usage in a .cln file:
//
//	@profile obgyn
//	pt 28F wt65 ga:34w          ← ga handled inline by this plugin
//	vitals bp120/75 fhr:142     ← fhr handled inline by this plugin
//	lmp 2025-06-15              ← standalone command
//	edd 2026-03-22              ← standalone command
//	gpal G2P1A0L1               ← standalone command
//	fhs 142 regular, reactive   ← standalone command
//	ctx 3 in 10min lasting 45s  ← standalone command
package obgyn

import (
	"clinlang/pkg/engine"
	"regexp"
	"strconv"
	"strings"
)

// gaPattern matches gestational age values such as "34", "34w", "34W".
// Anything else (e.g. "34www", "34ww", "wweek") fails to match and the
// caller emits a warning.
var gaPattern = regexp.MustCompile(`^(?i)(\d+)w?$`)

// fhrPattern matches fetal heart rate values such as "142", "142bpm",
// "142BPM". Anything else fails to match.
var fhrPattern = regexp.MustCompile(`^(?i)(\d+)(?:bpm)?$`)

// ─────────────────────────────────────────────────────────────────────────────
// Plugin struct
// ─────────────────────────────────────────────────────────────────────────────

type ObGynPlugin struct{}

func (p *ObGynPlugin) GetName() string        { return "obgyn" }
func (p *ObGynPlugin) GetDescription() string { return "OB/GYN Specialty — obstetric and gynaecologic fields" }

func (p *ObGynPlugin) GetCommandSummary() map[string]string {
	return map[string]string{
		// Standalone commands
		"lmp":  "Last menstrual period date (e.g. lmp 2025-06-15)",
		"edd":  "Estimated date of delivery (e.g. edd 2026-03-22)",
		"gpal": "Gravida/Para/Abortus/Living (e.g. gpal G2P1A0L1)",
		"fhs":  "Fetal heart sounds (e.g. fhs 142 regular reactive)",
		"ctx":  "Contractions (e.g. ctx 3in10min lasting45s)",
		// Inline pt tokens
		"[pt] ga:": "Gestational age inside pt command (e.g. pt 28F ga:34w)",
		// Inline vitals tokens
		"[vitals] fhr:": "Fetal heart rate inside vitals command (e.g. vitals bp120/75 fhr:142)",
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Plugin data struct — all OB/GYN-specific fields
// ─────────────────────────────────────────────────────────────────────────────

// ObGynData holds all obstetric/gynaecologic structured data for the case.
// It lives in ClinicalCase.SpecialtyData and is never mixed into Patient.
type ObGynData struct {
	// Set via inline pt token extension: pt 28F ga:34w
	GA     int    `json:"ga_weeks,omitempty"` // gestational age in weeks
	GAUnit string `json:"ga_unit,omitempty"`  // always "w" for now

	// Set via inline vitals token extension: vitals bp120/75 fhr:142
	FHR int `json:"fhr_bpm,omitempty"` // fetal heart rate (bpm)

	// Set via standalone commands
	LMP  string `json:"lmp,omitempty"`  // last menstrual period
	EDD  string `json:"edd,omitempty"`  // estimated date of delivery
	GPAL string `json:"gpal,omitempty"` // gravida/para/abortus/living
	FHS  string `json:"fhs,omitempty"`  // fetal heart sounds (full description)
	CTX  string `json:"ctx,omitempty"`  // contractions (full description)
}

func (p *ObGynPlugin) InitData() any { return &ObGynData{} }

// ─────────────────────────────────────────────────────────────────────────────
// Standalone commands (new keywords added by this plugin)
// ─────────────────────────────────────────────────────────────────────────────

func (p *ObGynPlugin) GetCommands() map[string]engine.ParserFunc {
	return map[string]engine.ParserFunc{
		"lmp": func(tokens []string, c *engine.ClinicalCase) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.LMP = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
		"edd": func(tokens []string, c *engine.ClinicalCase) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.EDD = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
		"gpal": func(tokens []string, c *engine.ClinicalCase) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.GPAL = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
		"fhs": func(tokens []string, c *engine.ClinicalCase) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.FHS = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
		"ctx": func(tokens []string, c *engine.ClinicalCase) {
			data, ok := c.SpecialtyData.(*ObGynData)
			if ok {
				data.CTX = engine.ExpandAbbreviations(strings.Join(tokens, " "))
			}
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Inline token extensions for core commands (CommandExtendable interface)
// ─────────────────────────────────────────────────────────────────────────────

// GetCommandTokens implements engine.CommandExtendable.
//
// This is where OB/GYN-specific inline tokens are registered for core commands.
// The engine calls this when @profile obgyn is encountered and populates the
// per-parse CommandTokenRegistry with these handlers.
//
// Plugin authors: never modify core parsers — add tokens here instead.
func (p *ObGynPlugin) GetCommandTokens() map[string]map[string]engine.TokenExtFunc {
	return map[string]map[string]engine.TokenExtFunc{

		// ── Extensions for the `pt` command ──────────────────────────────────
		"pt": {
			// ga:<n>w — gestational age in weeks
			// Accepted formats: ga:34w  ga:34  (weeks assumed)
			// Writes into ObGynData.GA and ObGynData.GAUnit
			"ga:": func(token string, c *engine.ClinicalCase) bool {
				data, ok := c.SpecialtyData.(*ObGynData)
				if !ok {
					return false
				}
				raw := strings.TrimSpace(token[3:]) // strip "ga:"
				m := gaPattern.FindStringSubmatch(raw)
				if m == nil {
					c.AddWarning("obgyn: invalid gestational age value: " + token)
					return true // we claimed it, but it was malformed
				}
				val, _ := strconv.Atoi(m[1])
				data.GA = val
				data.GAUnit = "w"
				return true
			},
		},

		// ── Extensions for the `vitals` command ──────────────────────────────
		"vitals": {
			// fhr:<bpm> — fetal heart rate
			// Accepted formats: fhr:142  fhr:145bpm
			// Writes into ObGynData.FHR
			"fhr:": func(token string, c *engine.ClinicalCase) bool {
				data, ok := c.SpecialtyData.(*ObGynData)
				if !ok {
					return false
				}
				raw := strings.TrimSpace(token[4:]) // strip "fhr:"
				m := fhrPattern.FindStringSubmatch(raw)
				if m == nil {
					c.AddWarning("obgyn: invalid fetal heart rate value: " + token)
					return true
				}
				val, _ := strconv.Atoi(m[1])
				data.FHR = val
				return true
			},
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Self-registration
// ─────────────────────────────────────────────────────────────────────────────

func init() {
	engine.RegisterPlugin(&ObGynPlugin{})
}