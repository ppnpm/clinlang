package main

import (
	"clinlang/pkg/engine"
	_ "clinlang/pkg/engine/plugins/obgyn"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// =============================================================================
// OUTPUT FORMATTERS
// =============================================================================

// PrintClinicalNote prints a human-readable structured clinical note.
func PrintClinicalNote(c engine.ClinicalCase) {
	sep := strings.Repeat("─", 50)
	fmt.Println(sep)
	fmt.Println("CLINICAL NOTE")
	fmt.Println(sep)

	fmt.Printf("ID     : %s\n", c.Patient.Id)

	//Accessing Patient's Age
	age := fmt.Sprintf("%d", c.Patient.Age)
	if c.Patient.Age == 0 {
		age = "?"
	}


	fmt.Printf("Patient: %s/%s", age, c.Patient.Sex)
	if c.Patient.Weight > 0 {
		fmt.Printf("  Wt: %gkg", c.Patient.Weight)
	}
	if c.Patient.Height > 0 {
		fmt.Printf("  Ht: %gcm", c.Patient.Height)
	}
	fmt.Println()
	fmt.Println(sep)

	printField("Chief Complaint", c.CC)
	printField("HPI", c.HPI)
	printField("Past Medical History", c.PMH)

	fmt.Printf("Vitals : %s\n", engine.FormatVitals(c.Vitals))

	// Abnormal flags
	if len(c.AbnormalFlags) > 0 {
		fmt.Println()
		for _, f := range c.AbnormalFlags {
			icon := "⚠"
			if f.Severity == engine.SeverityCritical {
				icon = "🔴 CRITICAL"
			}
			fmt.Printf("  %s  %s: %s — %s\n", icon, f.Field, f.Value, f.Message)
		}
	}

	// Symptoms
	if len(c.Symptoms) > 0 {
		fmt.Println("Symptoms:")
		for _, s := range c.Symptoms {
			label := engine.IntensityLabel(s.Intensity)
			dur := ""
			if s.Duration != "" {
				dur = " × " + s.Duration
			}
			fmt.Printf("  ▸ %-20s [%s%s]\n", s.Name, label, dur)
		}
	}

	// Extension data
	if len(c.Extra) > 0 {
		fmt.Println(sep)
		fmt.Println("Additional Data:")
		for cmd, kv := range c.Extra {
			fmt.Printf("  [%s]\n", strings.ToUpper(cmd))
			for k, v := range kv {
				if v == "true" {
					fmt.Printf("    ▸ %s\n", k)
				} else {
					fmt.Printf("    ▸ %-12s: %s\n", k, v)
				}
			}
		}
	}

	// Specialty Plugin Data
	if c.SpecialtyData != nil {
		fmt.Println(sep)
		fmt.Printf("Specialty Profile: %s\n", strings.ToUpper(c.Profile))
		b, _ := json.MarshalIndent(c.SpecialtyData, "  ", "  ")
		str := string(b)
		str = strings.TrimPrefix(str, "{\n")
		str = strings.TrimSuffix(str, "\n}")
		fmt.Println(str)
	}

	// Prescriptions
	if len(c.Prescriptions) > 0 {
		fmt.Println(sep)
		fmt.Print(engine.FormatPrescriptions(c.Prescriptions))
	}

	fmt.Println(sep)
	printField("Diagnosis", c.DX)
	fmt.Println(sep)

	if len(c.Warnings) > 0 {
		fmt.Println("⚠  Warnings:")
		for _, w := range c.Warnings {
			fmt.Println("   •", w)
		}
		fmt.Println(sep)
	}
}

func printField(label, value string) {
	if value != "" {
		fmt.Printf("%-22s: %s\n", label, value)
	}
}

// PrintJSON outputs the case as indented JSON.
func PrintJSON(c engine.ClinicalCase) {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}
	fmt.Println(string(b))
}

// PrintValidation runs all checks and reports warnings + abnormal flags.
func PrintValidation(c engine.ClinicalCase) {
	fmt.Println("=== Validation Report ===")
	if len(c.Warnings) == 0 && len(c.AbnormalFlags) == 0 {
		fmt.Println("✔  No issues found.")
		return
	}
	for _, w := range c.Warnings {
		fmt.Println("⚠ ", w)
	}
	for _, f := range c.AbnormalFlags {
		icon := "⚠ "
		if f.Severity == engine.SeverityCritical {
			icon = "🔴"
		}
		fmt.Printf("%s [%s] %s: %s — %s\n", icon, strings.ToUpper(f.Severity), f.Field, f.Value, f.Message)
	}
}

// =============================================================================
// MAIN
// =============================================================================

func main() {
	args := os.Args
	if len(args) < 2 {
		printUsage()
		return
	}

	subcommand := strings.ToLower(args[1])

	//## Server command

	// `clinlang server` takes an optional --port flag, no file needed, 'server' subcommand requires only 2 arguments to run
	if subcommand == "server" {
		port := "8080"
		for i, a := range args {
			if a == "--port" && i+1 < len(args) {
				port = args[i+1] // checking --port from the arguments and updating default port 8080 to custom value from the arguments : ./clinlang.exe server --port 9090
			}
		}
		StartServer(port)
		return
	}

	//## File based commands to read .cln file

	//ensures that filepath is provided for file based commands: ./clinlang.exe run <file.cln>
	if len(args) < 3 {
		printUsage()
		return
	}

	filePath := args[2]
	c, err := engine.ParseFile(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	switch subcommand {
	case "run":
		PrintClinicalNote(c)
	case "json":
		PrintJSON(c)
	case "soap":
		fmt.Print(engine.FormatSOAP(c))
	case "validate":
		PrintValidation(c)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`ClinLang — Clinical Shorthand DSL

Usage:
  clinlang run      <file.cln>         Formatted clinical note
  clinlang soap     <file.cln>         SOAP-format note
  clinlang json     <file.cln>         Structured JSON outputclin
  clinlang validate <file.cln>         Validation + abnormal flag report
  clinlang server   [--port 8080]      Start HTTP JSON API server`)
}
