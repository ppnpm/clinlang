package engine

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatSOAP generates a SOAP-structured clinical note from a parsed ClinicalCase.
//
// SOAP = Subjective / Objective / Assessment / Plan
//
// This is the universal documentation format used in:
//   - Ward rounds
//   - ER documentation
//   - Clinic notes
//   - Referral letters
func FormatSOAP(c ClinicalCase) string {
	sb := strings.Builder{}

	sep := strings.Repeat("─", 50)
	sb.WriteString(sep + "\n")
	sb.WriteString(fmt.Sprintf("Patient: %s  |  ID: %s\n", formatPatientLine(c.Patient), c.Patient.Id))
	if c.Day != "" {
		sb.WriteString(fmt.Sprintf("Context: %s\n", c.Day))
	}
	if c.Allergies != "" {
		sb.WriteString(fmt.Sprintf("\n[!] ALLERGIES: %s [!]\n", strings.ToUpper(c.Allergies)))
	}
	sb.WriteString(sep + "\n\n")

	// ── S: Subjective ─────────────────────────────────────
	sb.WriteString("S — SUBJECTIVE\n")
	sb.WriteString(strings.Repeat("─", 25) + "\n")

	if c.CC != "" {
		sb.WriteString(fmt.Sprintf("Chief Complaint: %s\n", c.CC))
	}
	if c.HPI != "" {
		sb.WriteString(fmt.Sprintf("HPI            : %s\n", c.HPI))
	}
	if c.PMH != "" {
		sb.WriteString(fmt.Sprintf("PMH            : %s\n", c.PMH))
	}
	if c.SH != "" {
		sb.WriteString(fmt.Sprintf("Social History : %s\n", c.SH))
	}
	if c.FH != "" {
		sb.WriteString(fmt.Sprintf("Family History : %s\n", c.FH))
	}

	// Symptoms listed under subjective
	if len(c.Symptoms) > 0 {
		sb.WriteString("Symptoms       : ")
		symParts := []string{}
		for _, s := range c.Symptoms {
			part := s.Name
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
		sb.WriteString(fmt.Sprintf("Physical Exam  : %s\n", c.PE))
	}

	if len(c.Imaging) > 0 {
		sb.WriteString("Imaging/Rad    : ")
		pairs := []string{}
		for k, v := range c.Imaging {
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
		for k, v := range c.Labs {
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

	// Abnormal flags inline under objective
	if len(c.AbnormalFlags) > 0 {
		sb.WriteString("⚠ Abnormals    : ")
		flagParts := []string{}
		for _, f := range c.AbnormalFlags {
			icon := "⚠"
			if f.Severity == SeverityCritical {
				icon = "🔴"
			}
			flagParts = append(flagParts, fmt.Sprintf("%s %s %s", icon, f.Field, f.Value))
		}
		sb.WriteString(strings.Join(flagParts, " | ") + "\n")
	}

	// Extension data (labs, exam, etc.) under Objective
	if len(c.Extra) > 0 {
		for cmd, kv := range c.Extra {
			sb.WriteString(fmt.Sprintf("%-15s: ", strings.ToUpper(cmd)))
			pairs := []string{}
			for k, v := range kv {
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
		sb.WriteString(fmt.Sprintf("Diagnosis      : %s\n", c.DX))
	} else {
		sb.WriteString("Diagnosis      : [Pending]\n")
	}

	if c.DDX != "" {
		sb.WriteString(fmt.Sprintf("Differential   : %s\n", c.DDX))
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
			
			routeStr := p.Route
			if expanded, ok := expandedRoutes[routeStr]; ok {
				routeStr = expanded
			}
			if routeStr != "" {
				line += " " + routeStr
			}

			freqStr := p.Frequency
			if expanded, ok := expandedFrequencies[freqStr]; ok {
				freqStr = expanded
			}
			if freqStr != "" {
				line += ", " + freqStr
			}
			
			if p.Duration != "" {
				line += " for " + ExpandDuration(p.Duration)
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
	if p.GPAL != "" {
		line += fmt.Sprintf(" | %s", p.GPAL)
	}
	if p.MOA != "" {
		line += fmt.Sprintf(" | MOA: %s", p.MOA)
	}
	return line
}
