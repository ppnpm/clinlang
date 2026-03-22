# ClinLang — Developer Reference

Complete reference for the ClinLang engine internals, architecture, and API.

---

## Project Structure

```
clinscript/
├── main.go                  ← CLI dispatcher + output formatters + ParseString()
├── models.go                ← Shared types: Patient, Vitals, Symptom, Case, etc.
├── patient_parser.go        ← `pt` command
├── symptom_parser.go        ← `sx` command with intensity + duration
├── vitals_parser.go         ← `vitals` command with structured output
├── extension_parser.go      ← Generic key-value parser for any unknown command
├── prescription_parser.go   ← `rx` command (drug, dose, frequency, route, duration)
├── abbreviations.go         ← Medical shorthand → full term expansion (output-only)
├── abnormal.go              ← Vital/lab range checks → AbnormalFlag[]
├── soap_formatter.go        ← SOAP note generator
├── server.go                ← HTTP JSON API server (stdlib only)
└── docs/
    ├── documentation.md     ← Doctor-facing guide
    ├── language.md          ← This file
    └── SPEC.md              ← Original v0.1 spec
```

---

## Core Architecture

### Library Entry Point

```go
// In your app, import and call:
c := ParseString(rawCLNText)
```

`ParseString(input string) Case` is the single function your app calls.
It never reads from disk. The caller passes the raw `.cln` string.

`ParseFile(path string)` is a thin CLI wrapper that reads the file and calls `ParseString`.

### Command Dispatch

```go
// Local copy of active commands
var activeCommands = make(map[string]ParserFunc)
```

Unknown commands fall through to `ParseExtension(cmd, tokens, &c)` — stored in `Case.Extra[cmd][key]=value`.

### Specialty Plugin Registry

ClinLang supports a modular Plugin Architecture for different medical specialties.
Users declare their specialty context using the `@profile` tag.

```text
@profile obgyn
pt 28F
gpal G3P1A1L1
```

When `@profile` is parsed, the engine dynamically looks up the plugin in the registry and injects the specialty-specific commands into the parser. The resulting custom data is bundled inside the `Case.SpecialtyData` field and serialized safely to JSON/SOAP.

## Data Models (`models.go`)

```go
type Case struct {
    Patient       Patient
    CC            string
    HPI           string
    PMH           string
    DX            string
    Vitals        Vitals
    Symptoms      []Symptom
    Prescriptions []Prescription
    Extra         map[string]map[string]string  // extension data
    AbnormalFlags []AbnormalFlag
    Warnings      []string
}

type Vitals struct {
    BP   string  // "140/90"
    HR   int
    SpO2 int
    Temp string
    RR   int
}

type Symptom struct {
    Name      string
    Intensity string  // "mild", "severe", "very severe", "improving", etc.
    Duration  string  // "3d", "2h", "1w" — empty if not specified
}

type Prescription struct {
    Drug      string
    Dose      string    // "25mg"
    Frequency string    // "BD", "OD" etc.
    Route     string    // "IV", "PO" etc (optional)
    Duration  string    // "7d" (optional)
}

type AbnormalFlag struct {
    Field    string  // "HR", "SpO2", "Hb"
    Value    string  // "102 bpm"
    Message  string  // "Tachycardia (>100 bpm)"
    Severity string  // "warning" or "critical"
}
```

---

## Parser Files

### `patient_parser.go`
**Command:** `pt`

Token recognition (order-independent):
- `34M` or `34F` → age + sex combined
- `34` (pure number) → age
- `M` / `F` → sex
- `wt<n>` → weight in kg
- `ht<n>` → height in cm

### `symptom_parser.go`
**Command:** `sx`

Token format: `<name><intensity><duration?>`

Intensity is matched right-to-left (longest match first):
- `+++` → very severe
- `++` → severe
- `+` → mild
- `---` → resolved
- `--` → resolving
- `-` → improving

Duration: a number followed by `d`, `h`, `w`, or `m` appended immediately after intensity.

```go
// Example: "fever++3d"
// → Name="fever", Intensity="severe", Duration="3d"
```

### `vitals_parser.go`
**Command:** `vitals`

Prefix-based token parsing (all lowercased):
- `bp<sys/dia>` → BP
- `hr<n>` → HR
- `spo2<n>` or `spo<n>` → SpO2
- `temp<val>` → Temp
- `rr<n>` → RR

### `extension_parser.go`
**Command:** any unknown keyword

Splits each token at the first digit or `.` character:
```go
"hb13.5"  → key="hb", value="13.5"
"fasting" → stored as: Extra[cmd]["fasting"] = "true"
```

### `prescription_parser.go`
**Command:** `rx`

First token (or multiple tokens if matching a known drug in `DrugsList`) is the drug name. Remaining tokens are identified by:
- **Dose:** contains a digit (e.g., `25mg`, `500mg`, `1g`)
- **Frequency:** matched against `frequencyAliases` map (od, bd, tds, qds, stat, prn…)
- **Route:** matched against `routeAliases` map (iv, im, po, sc, sl, neb…)
- **Duration:** starts with `x` + number + unit OR bare `7d`, `5d`, etc.

*Note: It is recommended to use `rx <drug> <dose> <route>` instead of prepending the drug with `tab` or `inj`, as this keeps the extracted `<drug>` clean.*

### `abbreviations.go`
**Used by:** output formatters, SOAP formatter

`ExpandAbbreviations(string) string` — word-by-word lookup in the `Abbreviations` map.
Applied at output time only. The `Case` always stores the original abbreviated text.

### `abnormal.go`
**Called by:** `ParseString()` after all commands are parsed.

`CheckAbnormals(c *Case)` — checks `c.Vitals` and `c.Extra["lab"]`.
Appends to `c.AbnormalFlags` — never panics, never modifies existing parsed data.

Ranges implemented:
| Field | Warning | Critical |
|---|---|---|
| SBP | ≥140 | ≥180 or <90 |
| HR | >100 or <60 | >150 or <40 |
| SpO2 | <94% | <85% |
| RR | >20 | >30 or <10 |
| Hb | <11 or >17 | <7 |
| WBC | >11k or <4k | >20k |
| Creatinine | >1.2 | >5 |
| Na+ | <135 or >150 | <125 |
| K+ | <3.5 or >5 | <2.5 or >6 |
| Glucose | >11 or <3 | >20 |

### `soap_formatter.go`
**Function:** `FormatSOAP(c Case) string`

Sections:
- **S (Subjective):** CC, HPI, PMH (with abbrev expansion), Symptoms
- **O (Objective):** Vitals, AbnormalFlags inline, Extra data (labs, exam)
- **A (Assessment):** DX
- **P (Plan):** Prescriptions from `rx` commands (automatically expands route and frequency abbreviations)

---

## HTTP API Server (`server.go`)

### Start
```bash
clinlang server --port 8080
```

### Endpoints

All endpoints accept:
```
POST /<endpoint>
Content-Type: application/json
Body: { "input": "<raw .cln text>" }
```

| Endpoint | Returns |
|---|---|
| `POST /parse` | Full `Case` object as JSON |
| `POST /note` | `{"note": "...", "warnings": [], "abnormal_flags": []}` |
| `POST /soap` | `{"soap": "...", "warnings": [], "abnormal_flags": []}` |
| `POST /validate` | `{"valid": bool, "warnings": [], "abnormal_flags": []}` |
| `GET /health` | `{"status": "ok"}` |

### CORS
All endpoints include `Access-Control-Allow-Origin: *`. A web frontend can call the API directly with no proxy needed.

### Example cURL

```bash
curl -X POST http://localhost:8080/parse \
  -H "Content-Type: application/json" \
  -d '{"input": "pt 34M\ncc fever 3d\nvitals hr110 bp160/100 spo290\ndx sepsis"}'
```

### Example JS Fetch (for web app)
```js
const response = await fetch('http://localhost:8080/note', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ input: clinText })
});
const data = await response.json();
console.log(data.note);
console.log(data.abnormal_flags);
```

---

## Adding a New Core Command

1. (Optional) create `<command>_parser.go`
2. Add to `coreCommands` map in `main.go`:

```go
"newcmd": func(tokens []string, c *Case) {
    // parse tokens, write to c
},
```

That's it. No other files need to change.

## Adding a New Extension (No Code Change)

Just use any new prefix in the `.cln` file:

```
obs ga32w parity2 edd2026-06-01
neuro gcs14 pupils_equal+ babinski-
```

These are stored in `Case.Extra["obs"]`, `Case.Extra["neuro"]` automatically.

---

## CLI Subcommands

```bash
clinlang run      file.cln         # Formatted clinical note
clinlang soap     file.cln         # SOAP note
clinlang json     file.cln         # Full JSON output
clinlang validate file.cln         # Warnings + abnormal flag report
clinlang server   [--port 8080]    # Start HTTP API
```

---

## Extension Roadmap (Not Yet Implemented)

| Feature | Effort |
|---|---|
| PDF export | Medium |
| Watch mode (`clinlang watch`) | Low |
| Clinical score calculators (CURB-65, HEART, qSOFA) | Medium |
| Template generator (`clinlang new acs`) | Low |
| Color terminal output (ANSI) | Low |
| VS Code extension | High |

* Vim

all have:

```text
strict grammar → flexible usage
```

---

# Next Step

We now upgrade your system to:

> **error handling + validation layer**

This will make your tool feel “professional”.

---

Say **“validate”** and we implement it.
