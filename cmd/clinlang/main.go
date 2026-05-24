package main

import (
	"clinlang/pkg/engine"
	_ "clinlang/pkg/engine/plugins/obgyn"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// disclaimer is shown in the CLI banner and the HTTP /health endpoint.
// Keep this in sync with DISCLAIMER.md at the repository root.
const disclaimer = "ClinLang is a personal note-taking and templating tool — not a medical device. No diagnosis, dosing, or decision support."

// =============================================================================
// OUTPUT FORMATTERS
// =============================================================================

// PrintJSON outputs the case as indented JSON.
func PrintJSON(c engine.ClinicalCase) {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}
	fmt.Println(string(b))
}

// PrintLintReport prints parser warnings and out-of-range markers in
// neutral form. There is no pass/fail and no severity.
func PrintLintReport(c engine.ClinicalCase) {
	fmt.Println("=== Lint Report ===")
	if len(c.Warnings) == 0 && len(c.RangeMarkers) == 0 {
		fmt.Println("No parser warnings. No values outside reference ranges.")
		return
	}
	for _, w := range c.Warnings {
		fmt.Println("warning:", w)
	}
	for _, m := range c.RangeMarkers {
		fmt.Printf("out-of-range: %s %s [ref %s, %s]\n",
			m.Field, m.Value, m.ReferenceRange, m.Source)
	}
}

// =============================================================================
// ARG HELPERS
// =============================================================================

// extractFlag walks args, removes any token equal to flag, and returns
// the remaining tokens plus whether the flag was present.
func extractFlag(args []string, flag string) ([]string, bool) {
	out := make([]string, 0, len(args))
	found := false
	for _, a := range args {
		if a == flag {
			found = true
			continue
		}
		out = append(out, a)
	}
	return out, found
}

// =============================================================================
// MAIN
// =============================================================================

func main() {
	// Optional override of the embedded reference ranges. Set
	// CLINLANG_REFERENCE_RANGES to the path of a JSON file with the
	// same schema as pkg/engine/reference_ranges.json. See
	// docs/reference-ranges.md.
	if path := os.Getenv("CLINLANG_REFERENCE_RANGES"); path != "" {
		if err := engine.LoadReferenceRanges(path); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to load reference ranges from %s: %v\n", path, err)
		}
	}

	args := os.Args

	// No args: launch the server in the configured mode and, in
	// local mode, open the user's browser to the bind URL. This is
	// the "double-click the binary" experience.
	if len(args) < 2 {
		StartServer("", "", true)
		return
	}

	subcommand := strings.ToLower(args[1])

	// `clinlang server [--mode local|hosted] [--port N]` — explicit
	// server start with no browser auto-launch.
	if subcommand == "server" {
		port := ""
		mode := ""
		for i, a := range args {
			if a == "--port" && i+1 < len(args) {
				port = args[i+1]
			}
			if a == "--mode" && i+1 < len(args) {
				mode = args[i+1]
			}
		}
		StartServer(mode, port, false)
		return
	}

	// File-based subcommands: parse flags, then resolve the first
	// remaining positional as the file path.
	rest := args[2:]
	rest, markersFlag := extractFlag(rest, "--markers")

	if len(rest) < 1 {
		printUsage()
		return
	}

	filePath := rest[0]
	c, err := engine.ParseFile(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	opts := engine.FormatOptions{ShowRangeMarkers: markersFlag}

	switch subcommand {
	case "run":
		fmt.Print(engine.FormatPlainNote(c))
	case "json":
		PrintJSON(c)
	case "soap":
		fmt.Print(engine.FormatSOAPWithOptions(c, opts))
	case "markdown", "md":
		fmt.Print(engine.FormatMarkdownWithOptions(c, opts))
	case "lint":
		PrintLintReport(c)
	case "validate":
		fmt.Fprintln(os.Stderr, "Deprecation: 'validate' is renamed to 'lint' and will be removed in a future release.")
		PrintLintReport(c)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(disclaimer)
	fmt.Println()
	fmt.Println(`ClinLang — Personal Clinical Shorthand & Templating Tool

Usage:
  clinlang run      <file.cln>             Formatted clinical note
  clinlang soap     <file.cln> [--markers] SOAP-format note (markers off by default)
  clinlang markdown <file.cln> [--markers] Markdown export (markers off by default)
  clinlang json     <file.cln>             Structured JSON output
  clinlang lint     <file.cln>             Parser warnings + reference-range markers
  clinlang server   [--mode local|hosted] [--port 8080]
                                           Start HTTP JSON API server

Env:
  CLINLANG_MODE=local|hosted               Deployment mode (default: local)
  CLINLANG_BIND=host:port                  Override bind address
  CLINLANG_WORKSPACE=<path>                Note storage root (required in hosted)
  CLINLANG_REFERENCE_RANGES=<path>         Override embedded reference ranges
                                            (see docs/reference-ranges.md)`)
}
