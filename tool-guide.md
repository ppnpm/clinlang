# ClinLang - Complete Tool & Developer Guide

Welcome to the comprehensive guide for **ClinLang**. This document explains the entire architecture of the project, teaches you the Go programming concepts used within it, provides a step-by-step tutorial on how to remake this project from scratch, and offers a deep dive explaining the code file by file.

---

## 1. Project Overview & Architecture

ClinLang is a high-performance **parsing engine** for clinical shorthand. Doctors can type simple commands (like `pt 34M cc fever`), and ClinLang converts them into structured JSON or formatted SOAP notes.

### Directory Structure
This project strictly follows Standard Go Project Layout:
* **`cmd/`**: Contains the executable applications.
  * **`clinlang/`**: The CLI program and HTTP Server. Reads files, calls the engine, and formats the output.
  * **`wasm/`**: WebAssembly wrappers to run the engine inside a web browser without a backend.
* **`pkg/`**: Contains the library code that other projects (or your commands) can import.
  * **`engine/`**: The pure parsing logic. This is where the magic happens (`parser.go`, `models.go`, etc.).
* **`web/`**: Code for a front-end web interface (Progressive Web App).

---

## 2. Core Go Concepts Used

If you are learning Go, ClinLang is a perfect repository to study. Here are the main concepts used with examples from the codebase:

### A. Structs and Struct Tags
Go uses `structs` (similar to Classes or Objects in other languages) to define the shape of your data.
```go
type Patient struct {
	Id      string  `json:"id"`
	Age     int     `json:"age"`
	Sex     string  `json:"sex"`
}
```
* **Struct Tags**: The part in backticks `` `json:"id"` `` tells Go's JSON encoder exactly what name to use when converting this struct into JSON text.

### B. Maps as Function Registries
Instead of massive `if/else` statements, ClinLang uses a `map` (a dictionary/hashmap) linking string commands to functions.
```go
var coreCommands = map[string]func(tokens []string, c *Case){
	"pt": ParsePatient,
	"vitals": ParseVitals,
}
```
This makes it highly **extensible**. Adding a new command is as simple as adding one line to this map.

### C. Pointers vs Values
Notice the asterisk `*` in `c *Case`. This means we are passing a **Pointer** (memory address) to the `Case` struct, rather than copying it.
* If we passed `(c Case)`, any modifications (like `c.Patient.Age = 34`) would only happen on a local copy and be lost.
* Passing `c *Case` ensures all parsers modify the **same** central struct in memory.

### D. Slices
Slices are dynamically-sized arrays. 
```go
func ParseString(input string) {
    parts := strings.Fields(line) // Splits into a Slice of strings based on spaces
    cmd := parts[0]               // First element
    tokens := parts[1:]           // Slice of everything from index 1 to the end
}
```

---

## 3. Deep Dive: Codebase File by File

Let's open up the hood and look exactly at what each file does.

### A. `pkg/engine/models.go` - The Data Structures
This file is the single source of truth for the shape of the data. 

**What it does:**
It defines strict data types (`Vitals`, `Symptom`, `Patient`, `Prescription`) and the master `Case` structure that holds everything together.

**Key Code Explained:**
```go
// Case is the root structure representing a full clinical encounter.
type Case struct {
	Patient       Patient   `json:"patient"`
	CC            string    `json:"cc,omitempty"`
	Vitals        Vitals    `json:"vitals"`
	Symptoms      []Symptom `json:"symptoms"`
	// ... other fields
}
```
* `omitempty` tells the JSON formatter: "If this field is empty (like an empty string `""`), do not include it in the final JSON output at all."
```go
// NewCase creates a properly initialized Case.
func NewCase() Case {
	return Case{
		Extra:         make(map[string]map[string]string),
		Symptoms:      []Symptom{},
	}
}
```
* `NewCase()` is a constructor. In Go, if you declare a map or slice but don't initialize it, it is `nil`. Trying to add data to it will cause a program crash (panic). `make()` initializes the map and empty brackets `[]` initialize the slices so they are safe to use immediately.

### B. `pkg/engine/parser.go` - The Brain of the Engine
This file reads raw text and coordinates the parsing.

**What it does:**
It breaks the input text line-by-line, figures out what command the user typed, and passes control to specific handlers.

**Key Code Explained:**
```go
func ParseString(input string) Case {
	c := NewCase() // 1. Create a fresh memory state
	lines := strings.Split(input, "\n") // 2. Break block of text into lines

	for lineNum, line := range lines { // 3. Loop over every line
		// 4. Clean the text: remove carriage returns (Windows) and trim edge spaces.
		line = strings.TrimSpace(strings.ReplaceAll(line, "\r", ""))

		// 5. Ignore blank lines or lines starting with comments (# or //)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		parts := strings.Fields(line) // 6. Split line by spaces: "pt 34M" -> ["pt", "34M"]
		cmd := strings.ToLower(parts[0]) // 7. The first word is the command (e.g. "pt")
		tokens := parts[1:] // 8. The rest of the words are arguments (e.g. ["34M"])

		// 9. Check if the command exists in our map
		if handler, ok := coreCommands[cmd]; ok {
			handler(tokens, &c) // 10. Call the matched function, passing the address of 'c'
		} else {
             // 11. Custom Extension handling if command is unknown...
		}
	}
    
    // 12. Post-processing
	CheckAbnormals(&c)      // Flags HR > 100 as abnormal, etc.
	c.CC = ExpandAbbreviations(c.CC) // Convert "dm2" -> "Type 2 Diabetes Mellitus"

	return c
}
```

### C. `pkg/engine/patient_parser.go` - Specific Component Logic
When `parser.go` sees the command `pt`, it hands control over to this file.

**What it does:**
It reads an array of words like `["34M", "wt80kg"]` and populates the `Patient` struct.

**Key Code Explained:**
```go
func ParsePatient(tokens []string, c *Case) {
	for _, rawTok := range tokens {
		tok := strings.ToLower(strings.TrimSpace(rawTok))
		
		// Checking for Weight
		if strings.HasPrefix(tok, "wt") { // e.g. "wt80kg"
			numPart := tok[2:] // Slice string starting at index 2 to remove "wt" -> "80kg"
			numPart = strings.TrimRight(numPart, "kg") // Remove "kg" -> "80"
			
            // ParseFloat converts string "80" to an actual math decimal 80.0
			val, err := strconv.ParseFloat(numPart, 64) 
			if err == nil { // If no error occurred (it was a valid number)
				c.Patient.Weight = val // Save it to memory
				continue // Skip the rest of the loop and move to next token!
			}
		}

		// Extracting Biological Sex
		if len(tok) > 0 {
			lastChar := tok[len(tok)-1] // Get the very last character of "34M" -> 'M'
			if lastChar == 'M' || lastChar == 'F' { 
				c.Patient.Sex = string(lastChar) // Save sex
				tok = tok[:len(tok)-1] // Remove 'M' from string -> "34"
			}
		}

		// Extracting integer Age 
		if isAllDigits(tok) && tok != "" { 
            // Atoi means "ASCII to Integer". "34" -> 34
			c.Patient.Age, _ = strconv.Atoi(tok) 
			c.Patient.AgeUnit = "Y" // Default everything strictly numeric to Years
			continue
		}
	}
}
```

### D. `cmd/clinlang/main.go` - The Command Line Interface (CLI)
This is not library text; this is the actual `main()` method program that compiles into `clinlang.exe`.

**What it does:**
It reads what the user typed in their command prompt/terminal, passes the requested file to the engine, and formats the output directly onto the screen.

**Key Code Explained:**
```go
func main() {
	args := os.Args // Reads arguments from terminal (e.g., ["clinlang", "run", "file.cln"])
	if len(args) < 2 {
		printUsage()
		return // Exit early if the user didn't type enough
	}

	subcommand := strings.ToLower(args[1]) // extract "run" or "json"

	filePath := args[2] // extract "file.cln"
    // ParseFile is a helper in parser.go that reads a file and calls ParseString()
	c, err := engine.ParseFile(filePath) 
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err) // Print red error text to console
		os.Exit(1) // Crash the app safely
	}

    // Switch statement executes code depending on what the user asked for
	switch subcommand {
	case "run":
		PrintClinicalNote(c) // Custom function to print beautiful ASCII text
	case "json":
		PrintJSON(c) // Converts 'c' into JSON and prints it
	case "soap":
		fmt.Print(engine.FormatSOAP(c)) // Prints a SOAP format string
	}
}

// How PrintJSON works:
func PrintJSON(c engine.Case) {
    // MarshalIndent turns our Go 'Case' structure into heavily formatted JSON strings
	b, err := json.MarshalIndent(c, "", "  ") 
	if err != nil { return }
	fmt.Println(string(b)) // Cast standard bytes `b` into a readable string and print to console
}
```

---

## 4. Step-by-Step Guide: How to Remake This

If you want to build this from scratch for learning purposes, follow these steps:

### Step 1: Initialize the Project
Create a folder, open your terminal, and run:
`go mod init your_project_name`
This tracks your dependencies.

### Step 2: Define Data Models
Create `pkg/engine/models.go`.
Define exactly what the final output should look like using Structs. (Create `Case`, `Patient`, `Vitals`). Use `json` struct tags to define exact JSON keys.

### Step 3: Build the Main Parsing Loop
Create `pkg/engine/parser.go`.
1. Write a function `Parse(input string) *Case`
2. Use `strings.Split(input, "\n")` to get lines.
3. Loop over lines `for _, line := range lines`.
4. Use `strings.Fields(line)` to get words. Identify the command (word 0).

### Step 4: Write Specific Parsers
Create `patient_parser.go`.
Write logic for parsing things like "34M". 
* Use `strings.HasSuffix("M")` to extract sex.
* Use `strconv.Atoi` on the remainder to get age.
Update the `Case` pointer.

### Step 5: Build CLI wrapper
Create `cmd/cli/main.go`.
Use `os.Args` to read the file path the user provides in the terminal.
Pass the file content to your `engine.Parse()` function.
Use `fmt.Println()` or `encoding/json` to print the result to the console!

---

## 5. Suggested Features & Improvements for Further Development

As requested, here is an analysis of areas where the project can be significantly improved or expanded:

### A. Advanced Parsing Features
1. **NLP / AI Fallback Integration:**
   Currently, unrecognized commands go to the `Extra` map. You could integrate a local LLM or OpenAI API to attempt to cleanly classify unrecognized sentences into the correct `Case` fields.
2. **Context-Aware Abbreviations:**
   "BS" could mean "Blood Sugar" in Vitals, but "Bowel Sounds" in an exam. Updating the abbreviation engine to expand based on the *current command section*.
   
### B. Infrastructure Improvements
1. **Concurrency for Batch Processing:**
   If a hospital uploads 10,000 `.cln` files, you could use Go's powerful `goroutines` and `channels` to parse them parallelly rather than one by one.
2. **Database Integration layer:**
   Add a `clinlang save <file>` command that directly connects to MongoDB or PostgreSQL using `pgx` or `gorm` to instantly push parsed data into a production database.

### C. Tooling & User Experience
1. **Interactive REPL Mode:**
   Build an interactive CLI mode where a doctor types `pt 35M` and hits Enter, and the terminal instantly shows a live-updating JSON preview on the right side of the screen. (You could use the `charmbracelet/bubbletea` package for a beautiful terminal UI).
2. **Better Error/Warning Checking Systems:**
   Instead of just string warnings, define strict `Warning` structs with line numbers and character positions to help syntax-highlighting in an IDE extension.
3. **FHIR Export Standardization:**
   Right now it exports custom JSON. A massive upgrade would be adding an output formatter that converts the `Case` struct into strict *HL7 FHIR* (Fast Healthcare Interoperability Resources) JSON format, making it instantly compatible with global hospital software (Epic, Cerner).

---

## 6. Tutorial: How to Build Your Own Specialty Profile (Plugin)

You don't need AI to add new medical specialties! ClinLang's Plugin Registry makes it incredibly easy. Here is exactly how to add a new profile (for example, `ortho`):

### Step 1: Create the Folder
Inside `pkg/engine/plugins/`, create a new folder for your specialty:
```bash
mkdir pkg/engine/plugins/ortho
```

### Step 2: Create the Plugin File
Create `pkg/engine/plugins/ortho/ortho.go`. 
At the top, declare your package and imports. Then, define the JSON structure you want to hold your custom data:

```go
package ortho

import (
	"clinlang/pkg/engine"
	"strings"
)

// 1. Define your data structure
type OrthoData struct {
	ROM  string `json:"rom,omitempty"`
	Cast string `json:"cast,omitempty"`
}

// 2. Define the Plugin struct
type OrthoPlugin struct{}

// 3. Set the name that users will type after @profile
func (p *OrthoPlugin) GetName() string { return "ortho" }

// 4. Return an empty instance of your data structure
func (p *OrthoPlugin) InitData() interface{} {
	return &OrthoData{}
}
```

### Step 3: Define Your Commands
In the same file, write the `GetCommands` function. This tells the parser what to do when it sees commands like `rom` or `cast`:

```go
func (p *OrthoPlugin) GetCommands() map[string]engine.ParserFunc {
	return map[string]engine.ParserFunc{ // returns a map of functions
		"rom": func(tokens []string, c *engine.Case) {
            // First, cast the generic SpecialtyData back to your OrthoData
			data := c.SpecialtyData.(*OrthoData)
            // Save the tokens as a single string into your struct
			data.ROM = strings.Join(tokens, " ")
		},
		"cast": func(tokens []string, c *engine.Case) {
			data := c.SpecialtyData.(*OrthoData)
			data.Cast = strings.Join(tokens, " ")
		},
	}
}
```

### Step 4: Register it on Startup
At the very bottom of the file, use Go's magic `init()` function to automatically register the plugin when the program starts:

```go
func init() {
	engine.RegisterPlugin(&OrthoPlugin{})
}
```

### Step 5: Import it into the CLI
Open `cmd/clinlang/main.go` and add the "blank import" to the top to ensure your `init()` function runs:

```go
import (
	"clinlang/pkg/engine"
	_ "clinlang/pkg/engine/plugins/obgyn"
	_ "clinlang/pkg/engine/plugins/ortho" // Add your new plugin here!
)
```

**That's it!** You can now process `.cln` files that look like this:
```txt
@profile ortho
pt 25M
rom Left knee flexion limited to 90deg
cast Plaster applied
```
