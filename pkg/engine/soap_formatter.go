package engine

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// FormatSOAP generates a SOAP-structured clinical note from a parsed
// ClinicalCase using the default options (no out-of-range markers).
//
// SOAP = Subjective / Objective / Assessment / Plan
func FormatSOAP(c ClinicalCase) string {
	return FormatSOAPWithOptions(c, FormatOptions{})
}

// FormatSOAPWithOptions produces the same SOAP note with explicit
// rendering options. Set opts.ShowRangeMarkers = true to append a
// "Notes (out of ref)" subsection at the end of the Objective block.
func FormatSOAPWithOptions(c ClinicalCase, opts FormatOptions) string {
	sb := strings.Builder{}

	sep := strings.Repeat("─", 50)
	sb.WriteString(sep + "\n")
	sb.WriteString(fmt.Sprintf("Patient: %s  |  ID: %s\n", formatPatientLine(c.Patient), c.Patient.Id))
	if c.Day != "" {
		sb.WriteString(fmt.Sprintf("Context: %s\n", c.Day))
	}
	if c.Allergies != "" {
		sb.WriteString(fmt.Sprintf("\n[!] ALLERGIES: %s [!]\n", strings.ToUpper(displayText(c.Allergies, c))))
	}
	sb.WriteString(sep + "\n\n")

	// ── S: Subjective ─────────────────────────────────────
	sb.WriteString("S — SUBJECTIVE\n")
	sb.WriteString(strings.Repeat("─", 25) + "\n")

	if c.CC != "" {
		sb.WriteString(fmt.Sprintf("Chief Complaint: %s\n", displayText(c.CC, c)))
	}
	if c.HPI != "" {
		sb.WriteString(fmt.Sprintf("HPI            : %s\n", displayText(c.HPI, c)))
	}
	if c.PMH != "" {
		sb.WriteString(fmt.Sprintf("PMH            : %s\n", displayText(c.PMH, c)))
	}
	if c.SH != "" {
		sb.WriteString(fmt.Sprintf("Social History : %s\n", displayText(c.SH, c)))
	}
	if c.FH != "" {
		sb.WriteString(fmt.Sprintf("Family History : %s\n", displayText(c.FH, c)))
	}

	// Symptoms listed under subjective
	if len(c.Symptoms) > 0 {
		sb.WriteString("Symptoms       : ")
		symParts := []string{}
		for _, s := range c.Symptoms {
			part := displayText(s.Name, c)
			if s.Intensity != "" {
				part += " (" + s.Intensity
				if s.Duration != "" {
					part += ", " + ExpandDuration(s.Duration)
				}
				part += ")"
			} else if s.Duration != "" {
				part += " (" + ExpandDuration(s.Duration) + ")"
			}
			symParts = append(symParts, part)
		}
		sb.WriteString(strings.Join(symParts, "; ") + "\n")
	}

	sb.WriteString("\n")

	// ── O: Objective ──────────────────────────────────────
	sb.WriteString("O — OBJECTIVE\n")
	sb.WriteString(strings.Repeat("─", 25) + "\n")

	sb.WriteString(fmt.Sprintf("Vitals         : %s\n", FormatVitals(c.Vitals)))

	if c.PE != "" {
		sb.WriteString(fmt.Sprintf("Physical Exam  : %s\n", displayText(c.PE, c)))
	}
	if len(c.Images) > 0 {
		sb.WriteString(fmt.Sprintf("Images/Attach  : %s\n", strings.Join(c.Images, ", ")))
	}

	if len(c.Imaging) > 0 {
		sb.WriteString("Imaging/Rad    : ")
		pairs := []string{}
		for _, k := range SortedMapKeys(c.Imaging) {
			v := c.Imaging[k]
			if v == "true" {
				pairs = append(pairs, strings.ToUpper(k))
			} else {
				pairs = append(pairs, strings.ToUpper(k)+" "+v)
			}
		}
		sb.WriteString(strings.Join(pairs, " | ") + "\n")
	}

	if len(c.Labs) > 0 {
		sb.WriteString("Labs           : ")
		pairs := []string{}
		for _, k := range SortedMapKeys(c.Labs) {
			v := c.Labs[k]
			if v == "true" {
				pairs = append(pairs, strings.ToUpper(k))
			} else {
				pairs = append(pairs, strings.ToUpper(k)+" "+v)
			}
		}
		sb.WriteString(strings.Join(pairs, " | ") + "\n")
	}

	// Specialty data (Plugin output) inline under objective
	if c.SpecialtyData != nil {
		sb.WriteString(fmt.Sprintf("Profile Data   : [%s]\n", strings.ToUpper(c.Profile)))
		b, _ := json.MarshalIndent(c.SpecialtyData, "", "  ")
		str := string(b)
		str = strings.TrimPrefix(str, "{\n")
		str = strings.TrimSuffix(str, "\n}")
		lines := strings.Split(str, "\n")
		for _, l := range lines {
			if strings.TrimSpace(l) != "" {
				l = strings.ReplaceAll(l, "\"", "")
				l = strings.ReplaceAll(l, ",", "")
				sb.WriteString(fmt.Sprintf("                 %s\n", strings.TrimSpace(l)))
			}
		}
	}

	// Out-of-range markers — opt-in via FormatOptions.ShowRangeMarkers.
	// Rendered as a neutral subsection at the end of the Objective block.
	// No icons, no severity, no clinical interpretation.
	if opts.ShowRangeMarkers && len(c.RangeMarkers) > 0 {
		sb.WriteString("Notes (out of ref):\n")
		for _, m := range c.RangeMarkers {
			sb.WriteString(fmt.Sprintf("  - %s %s [outside ref %s, %s]\n",
				m.Field, m.Value, m.ReferenceRange, m.Source))
		}
	}

	// Extension data (labs, exam, etc.) under Objective
	if len(c.Extra) > 0 {
		for _, cmd := range SortedMapKeys(c.Extra) {
			kv := c.Extra[cmd]
			sb.WriteString(fmt.Sprintf("%-15s: ", strings.ToUpper(cmd)))
			pairs := []string{}
			for _, k := range SortedMapKeys(kv) {
				v := kv[k]
				if v == "true" {
					pairs = append(pairs, k)
				} else {
					pairs = append(pairs, k+" "+v)
				}
			}
			sb.WriteString(strings.Join(pairs, " | ") + "\n")
		}
	}

	sb.WriteString("\n")

	// ── A: Assessment ─────────────────────────────────────
	sb.WriteString("A — ASSESSMENT\n")
	sb.WriteString(strings.Repeat("─", 25) + "\n")

	if c.DX != "" {
		sb.WriteString(fmt.Sprintf("Diagnosis      : %s\n", displayText(c.DX, c)))
	} else {
		sb.WriteString("Diagnosis      : [Pending]\n")
	}

	if c.DDX != "" {
		sb.WriteString(fmt.Sprintf("Differential   : %s\n", displayText(c.DDX, c)))
	}

	sb.WriteString("\n")

	// ── P: Plan ───────────────────────────────────────────
	sb.WriteString("P — PLAN\n")
	sb.WriteString(strings.Repeat("─", 25) + "\n")

	if len(c.Prescriptions) > 0 {
		for _, p := range c.Prescriptions {
			line := fmt.Sprintf("  ▸ Rx: %s", titleCase(p.Drug))
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
	} else {
		sb.WriteString("  [Plan not documented]\n")
	}

	sb.WriteString("\n" + sep + "\n")

	// Warnings at the bottom
	if len(c.Warnings) > 0 {
		sb.WriteString("⚠ Parser Warnings:\n")
		for _, w := range c.Warnings {
			sb.WriteString("   • " + w + "\n")
		}
	}

	return sb.String()
}

func formatPatientLine(p Patient) string {
	unit := p.AgeUnit
	if unit == "" {
		unit = "Y"
	}
	age := fmt.Sprintf("%d%s", p.Age, unit)
	if p.Age == 0 {
		age = "Age?"
	}
	sex := p.Sex
	if sex == "" {
		sex = "?"
	}
	line := age + "/" + sex
	if p.Weight > 0 {
		line += fmt.Sprintf(" | Wt: %gkg", p.Weight)
	}
	if p.Height > 0 {
		line += fmt.Sprintf(" | Ht: %gcm", p.Height)
	}
	
	return line
}

// SortedMapKeys returns the keys of m in lexicographically sorted order.
// Exported so external callers (CLI, server, frontends) can render
// deterministically — Go map iteration order is intentionally random.
func SortedMapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
