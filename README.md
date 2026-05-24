# ClinLang

 ClinLang is a personal shorthand and templating tool for clinicians' own notes. It expands abbreviations, formats notes in SOAP/Markdown/JSON, and helps you write faster. It is not a
  ▎ medical device, does not provide diagnosis, treatment, dosing, or clinical decision support, and must not be used as the basis for any clinical decision. All clinical content is authored
  ▎ by the user; ClinLang only transcribes and formats it.
  

Built purely in Go, it runs anywhere: from the terminal, as an HTTP API, and seamlessly compiled to WebAssembly for browser-based offline Progressive Web Apps.

## Features
* **Zero Punctuation Shorthand:** Type as fast as you think (`pt 34M cc fever 3d vitals hr110`).
* **Abbreviation Expansion:** Automatically expands short-hand like `dm2` to "Type 2 Diabetes Mellitus" and prescription frequencies like `bd` to "Twice daily" on output.
* **Clinical Decision Support:** Flags abnormal vitals and lab values natively.
* **Ward Ready:** Supports `day`, `sh` (Social History), `fh` (Family), `alg` (Allergies), `pe` (Physical Exam) and deeply parses `lab hb9.8 wbc12000`.
* **Extensible System:** Any unrecognized command becomes a structured key-value pair (`obs ga34w`).
* **Output Agnostic:** Export to JSON, raw Clinical Note, or SOAP-format.

## Architecture

This strictly follows standard Go library architecture:
* `pkg/engine/` - The pure parsing logic.
* `cmd/clinlang/` - The CLI and HTTP Server implementations.

## Installation

```bash
go get github.com/yourname/clinlang/pkg/engine
go install ./cmd/clinlang
```

## Quick Start as CLI

```bash
clinlang run examples/mi.cln       # Outputs a clean clinical note
clinlang soap examples/mi.cln      # Outputs a SOAP note
clinlang json examples/mi.cln      # Outputs JSON
clinlang server --port 8080      # Starts the HTTP API
```

## Import as Library

```go
package main

import (
	"fmt"
	"clinlang/pkg/engine"
)

func main() {
	rawText := "pt 40M wt90\ncc chest pain\nlab hb9.2"
	caseData := engine.ParseString(rawText)

	fmt.Printf("Age: %d\n", caseData.Patient.Age)
	fmt.Printf("Hb Warning: %v\n", caseData.AbnormalFlags)
	fmt.Println(engine.FormatSOAP(caseData))
}
```

## Contributing
Pull requests, bug reports, and new abbreviation maps are welcome! Please ensure any new features do not require changes to existing core syntax, to maintain backwards compatibility.

## License

MIT License. See `LICENSE` for details.
