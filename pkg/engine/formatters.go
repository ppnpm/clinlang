package engine

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatOptions controls how exported formats render optional sections.
// The zero value is the safe default: out-of-range markers are suppressed.
// Callers opt in to marker rendering by setting ShowRangeMarkers to true.
type FormatOptions struct {
	ShowRangeMarkers bool
}

// displayText returns the user's input with abbreviations expanded for
// human-friendly display. The parser deliberately stores user input
// verbatim; expansion is applied here at format time only. Empty input
// returns empty output.
func displayText(s string) string {
	if s == "" {
		return ""
	}
	return ExpandAbbreviations(s)
}

// FormatJSON produces a standard JSON representation of the ClinicalCase.
func FormatJSON(c ClinicalCase) string {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "{ \"error\": \"Failed to format JSON\" }"
	}
	return string(b)
}

// FormatPlainNote returns the canonical human-readable plain-note
// rendering of a ClinicalCase. This is the single source of truth used
// by the CLI's `run` subcommand and the HTTP server's /note endpoint.
//
// Out-of-range markers are always shown (in neutral form — no severity
// icons, no clinical interpretation). Abbreviation expansion happens
// here at render time only; the parsed case carries verbatim user input.
func FormatPlainNote(c ClinicalCase) string {
	sb := strings.Builder{}
	sep := strings.Repeat("─", 50)

	writeLine := func(s string) { sb.WriteString(s + "\n") }
	writeField := func(label, value string) {
		if value != "" {
			sb.WriteString(fmt.Sprintf("%-22s: %s\n", label, value))
		}
	}

	writeLine(sep)
	writeLine("CLINICAL NOTE")
	writeLine(sep)

	writeLine(fmt.Sprintf("ID     : %s", c.Patient.Id))

	age := fmt.Sprintf("%d", c.Patient.Age)
	if c.Patient.Age == 0 {
		age = "?"
	}
	patLine := fmt.Sprintf("Patient: %s/%s", age, c.Patient.Sex)
	if c.Patient.Weight > 0 {
		patLine += fmt.Sprintf("  Wt: %gkg", c.Patient.Weight)
	}
	if c.Patient.Height > 0 {
		patLine += fmt.Sprintf("  Ht: %gcm", c.Patient.Height)
	}
	writeLine(patLine)
	writeLine(sep)

	writeField("Chief Complaint", displayText(c.CC))
	writeField("HPI", displayText(c.HPI))
	writeField("Past Medical History", displayText(c.PMH))

	writeLine(fmt.Sprintf("Vitals : %s", FormatVitals(c.Vitals)))

	// Out-of-range markers: neutral, transcription-only annotations.
	if len(c.RangeMarkers) > 0 {
		writeLine("")
		for _, m := range c.RangeMarkers {
			writeLine(fmt.Sprintf("  %s: %s  [outside ref %s, %s]",
				m.Field, m.Value, m.ReferenceRange, m.Source))
		}
	}

	if len(c.Symptoms) > 0 {
		writeLine("Symptoms:")
		for _, s := range c.Symptoms {
			label := IntensityLabel(s.Intensity)
			dur := ""
			if s.Duration != "" {
				dur = " × " + s.Duration
			}
			writeLine(fmt.Sprintf("  ▸ %-20s [%s%s]", displayText(s.Name), label, dur))
		}
	}

	if len(c.Extra) > 0 {
		writeLine(sep)
		writeLine("Additional Data:")
		for _, cmd := range SortedMapKeys(c.Extra) {
			kv := c.Extra[cmd]
			writeLine(fmt.Sprintf("  [%s]", strings.ToUpper(cmd)))
			for _, k := range SortedMapKeys(kv) {
				v := kv[k]
				if v == "true" {
					writeLine(fmt.Sprintf("    ▸ %s", k))
				} else {
					writeLine(fmt.Sprintf("    ▸ %-12s: %s", k, v))
				}
			}
		}
	}

	if c.SpecialtyData != nil {
		writeLine(sep)
		writeLine(fmt.Sprintf("Specialty Profile: %s", strings.ToUpper(c.Profile)))
		b, _ := json.MarshalIndent(c.SpecialtyData, "  ", "  ")
		str := string(b)
		str = strings.TrimPrefix(str, "{\n")
		str = strings.TrimSuffix(str, "\n}")
		writeLine(str)
	}

	if len(c.Prescriptions) > 0 {
		writeLine(sep)
		sb.WriteString(FormatPrescriptions(c.Prescriptions))
	}

	writeLine(sep)
	writeField("Diagnosis", displayText(c.DX))
	writeLine(sep)

	if len(c.Warnings) > 0 {
		writeLine("Warnings:")
		for _, w := range c.Warnings {
			writeLine("   • " + w)
		}
		writeLine(sep)
	}

	return sb.String()
}

// FormatMarkdown produces a markdown document of the encounter using the
// default options (no out-of-range markers in output).
func FormatMarkdown(c ClinicalCase) string {
	return FormatMarkdownWithOptions(c, FormatOptions{})
}

// FormatMarkdownWithOptions produces the same document with explicit
// rendering options.
func FormatMarkdownWithOptions(c ClinicalCase, opts FormatOptions) string {
	var sb strings.Builder

	// Title & Patient Header
	sb.WriteString(fmt.Sprintf("# Clinical Summary: %s\n\n", c.Patient.Id))
	
	sb.WriteString("## Patient Information\n")
	sb.WriteString(fmt.Sprintf("- **Profile**: %s\n", formatPatientLine(c.Patient)))
	if c.Day != "" {
		sb.WriteString(fmt.Sprintf("- **Timeline**: %s\n", c.Day))
	}
	if c.Allergies != "" {
		sb.WriteString(fmt.Sprintf("- **Allergies**: %s\n", displayText(c.Allergies)))
	}
	sb.WriteString("\n")

	// Subjective
	sb.WriteString("## Subjective\n")
	if c.CC != "" {
		sb.WriteString(fmt.Sprintf("- **Chief Complaint**: %s\n", displayText(c.CC)))
	}
	if c.HPI != "" {
		sb.WriteString(fmt.Sprintf("- **HPI**: %s\n", displayText(c.HPI)))
	}
	if c.PMH != "" {
		sb.WriteString(fmt.Sprintf("- **PMH**: %s\n", displayText(c.PMH)))
	}

	if len(c.Symptoms) > 0 {
		sb.WriteString("- **Active Symptoms**:\n")
		for _, s := range c.Symptoms {
			sb.WriteString(fmt.Sprintf("  - %s", displayText(s.Name)))
			if s.Intensity != "" || s.Duration != "" {
				sb.WriteString(" (")
				parts := []string{}
				if s.Intensity != "" { parts = append(parts, s.Intensity) }
				if s.Duration != "" { parts = append(parts, ExpandDuration(s.Duration)) }
				sb.WriteString(strings.Join(parts, ", "))
				sb.WriteString(")")
			}
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")

	// Objective
	sb.WriteString("## Objective\n")
	sb.WriteString(fmt.Sprintf("- **Vital Signs**: %s\n", FormatVitals(c.Vitals)))
	if c.PE != "" {
		sb.WriteString(fmt.Sprintf("- **Examination**: %s\n", displayText(c.PE)))
	}
	
	if len(c.Labs) > 0 {
		sb.WriteString("- **Investigations**:\n")
		for _, k := range SortedMapKeys(c.Labs) {
			v := c.Labs[k]
			val := v
			if val == "true" { val = "Ordered" }
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", strings.ToUpper(k), val))
		}
	}
	
	if opts.ShowRangeMarkers && len(c.RangeMarkers) > 0 {
		sb.WriteString("\n### Notes (out of ref)\n")
		for _, m := range c.RangeMarkers {
			sb.WriteString(fmt.Sprintf("- **%s**: %s [outside ref %s, %s]\n",
				m.Field, m.Value, m.ReferenceRange, m.Source))
		}
	}
	sb.WriteString("\n")

	// Assessment
	sb.WriteString("## Assessment\n")
	diag := displayText(c.DX)
	if diag == "" { diag = "TBD" }
	sb.WriteString(fmt.Sprintf("- **Diagnosis**: %s\n", diag))
	if c.DDX != "" {
		sb.WriteString(fmt.Sprintf("- **Differentials**: %s\n", displayText(c.DDX)))
	}
	sb.WriteString("\n")

	// Plan
	sb.WriteString("## Plan\n")
	if len(c.Prescriptions) > 0 {
		sb.WriteString("### Prescriptions\n")
		for _, p := range c.Prescriptions {
			sb.WriteString(fmt.Sprintf("- **%s** %s %s, %s for %s\n", 
				titleCase(p.Drug), p.Dose, p.Route, p.Frequency, ExpandDuration(p.Duration)))
		}
	} else {
		sb.WriteString("*No prescriptions documented.*\n")
	}

	sb.WriteString("\n---\n")
	sb.WriteString("*Generated by ClinLang Engine*")

	return sb.String()
}
