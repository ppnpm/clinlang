# Reference Ranges

ClinLang can annotate numeric values in a note when they fall outside a configured reference band. These annotations are **transcription aids only** — they are not clinical decision support, not severity classifications, and not recommendations. The clinician supplies all clinical interpretation. See [DISCLAIMER.md](../DISCLAIMER.md).

## What you get

When markers are enabled, output looks like this:

```
HR: 110 bpm  [outside ref 60-100, user-default adult ranges]
BP: 160/100  [outside ref 90-140 / 60-90, user-default adult ranges]
```

Each marker carries:
- the value as it was recorded,
- the reference range it was compared against,
- the source that supplied that range (so the clinician can judge whether the source is appropriate).

There is no severity classification. There is no named clinical condition. The clinician decides what, if anything, to do.

## When markers appear

| Output                              | Markers shown? |
|-------------------------------------|----------------|
| `clinlang run <file.cln>`           | yes (always)   |
| `clinlang lint <file.cln>`          | yes (always)   |
| `clinlang json <file.cln>`          | yes — in the `range_markers` field of the JSON |
| `clinlang soap <file.cln>`          | no (default)   |
| `clinlang soap --markers <file>`    | yes — as a "Notes (out of ref)" subsection inside Objective |
| `clinlang markdown <file>`          | no (default)   |
| `clinlang markdown --markers <file>`| yes — as a "Notes (out of ref)" section |

Exports (SOAP, Markdown) default to **off** because exported notes are often shared with others; the clinician opts in when they want annotations included.

## The reference-range file

Default ranges are embedded in the binary at build time from `pkg/engine/reference_ranges.json`. The schema is a flat map of dotted keys to range entries:

```json
{
  "vitals.hr":           {"low": 60,   "high": 100,  "unit": "bpm",   "source": "user-default adult ranges"},
  "vitals.bp.systolic":  {"low": 90,   "high": 140,  "unit": "mmHg",  "source": "user-default adult ranges"},
  "vitals.bp.diastolic": {"low": 60,   "high": 90,   "unit": "mmHg",  "source": "user-default adult ranges"},
  "vitals.spo2":         {"low": 94,   "high": 100,  "unit": "%",     "source": "user-default adult ranges"},
  "vitals.temp":         {"low": 97,   "high": 100,  "unit": "F",     "source": "user-default adult ranges"},
  "vitals.rr":           {"low": 12,   "high": 20,   "unit": "/min",  "source": "user-default adult ranges"},
  "labs.hb.M":           {"low": 13,   "high": 17,   "unit": "g/dL",  "source": "user-default adult ranges"},
  "labs.hb.F":           {"low": 11,   "high": 17,   "unit": "g/dL",  "source": "user-default adult ranges"},
  "labs.wbc":            {"low": 4000, "high": 11000,"unit": "/uL",   "source": "user-default adult ranges"},
  "labs.creatinine":     {"low": 0.6,  "high": 1.2,  "unit": "mg/dL", "source": "user-default adult ranges"},
  "labs.na":             {"low": 135,  "high": 145,  "unit": "mEq/L", "source": "user-default adult ranges"},
  "labs.k":              {"low": 3.5,  "high": 5.0,  "unit": "mEq/L", "source": "user-default adult ranges"},
  "labs.glucose":        {"low": 3.9,  "high": 7.8,  "unit": "mmol/L","source": "user-default adult ranges"}
}
```

### Field-by-field

- **Key**: dotted path identifying the field being ranged. The engine looks up these exact keys, so don't rename them unless you are also editing the engine.
- **`low`** (optional number): lower bound, inclusive. Values below `low` are out of range.
- **`high`** (optional number): upper bound, inclusive. Values above `high` are out of range.
- **`unit`** (optional string): the unit the bounds are expressed in. Informational — not used by the engine; the displayed value still carries the unit that the engine recorded from the input.
- **`source`** (required string): cited next to the marker. Use something the clinician can audit ("hospital X policy 2025", "WHO 2023", "department default").

A key may have only `low`, only `high`, or both. A key with neither is ignored.

### Haemoglobin uses sex-specific keys

`labs.hb.M` is used when the patient sex is `M`. `labs.hb.F` is used when the patient sex is `F`. If sex is unrecorded, **no Hb marker is emitted** — the engine deliberately refuses to guess.

### Blood pressure uses two keys

`vitals.bp.systolic` and `vitals.bp.diastolic` are separate entries. The engine combines them into a single marker per BP reading, with the display `90-140 / 60-90`.

### Temperature requires explicit units; default is Fahrenheit

ClinLang does **not** convert between °F and °C. The clinician chooses the unit and the engine uses it verbatim:

- `vitals temp98.6F` — explicit Fahrenheit
- `vitals temp37C` — explicit Celsius
- `vitals temp98.6` — no suffix, defaults to **Fahrenheit**

The `vitals.temp` reference range has a `unit` field. A temperature marker is emitted **only when the recorded unit matches the range's configured unit**. If they don't match, the engine silently skips the check rather than guessing a conversion. To switch to Celsius defaults, override the range and record your temps with `C` suffix:

```json
{
  "vitals.temp": {"low": 36, "high": 38, "unit": "C", "source": "your-clinic ranges"}
}
```

This explicit-unit rule is deliberate: a software-driven F→C conversion is the kind of "the program decided what your number meant" behaviour that the medicolegal posture of this project avoids.

## Overriding the defaults

There are two ways to use a custom range file.

### 1. Environment variable (recommended)

Set `CLINLANG_REFERENCE_RANGES` to the path of your JSON file:

```bash
export CLINLANG_REFERENCE_RANGES=/path/to/your/ranges.json
clinlang run examples/mi.cln
```

```powershell
$env:CLINLANG_REFERENCE_RANGES = "C:\path\to\your\ranges.json"
clinlang run examples/mi.cln
```

If the file is missing or invalid, the CLI prints a warning to stderr and falls back to the embedded defaults.

### 2. Programmatic (Go library users)

```go
import "clinlang/pkg/engine"

if err := engine.LoadReferenceRanges("/path/to/your/ranges.json"); err != nil {
    log.Fatal(err)
}
// All subsequent ParseString / ParseFile calls use the override.
```

The override is process-wide. In hosted deployments where multiple users share one process, per-user range sets will be introduced in a later phase.

## A note on the defaults

The values shipped in `reference_ranges.json` are labelled `"user-default adult ranges"`. They are intended as a **starting point** for a single adult user to override. They are not endorsed by any clinical body, are not jurisdiction-specific, and should be replaced before use in any real workflow. The intent of shipping defaults at all is so the feature works out of the box for demonstration, not so the defaults function as authoritative ranges.

The clinician is the source of truth. ClinLang is the notebook.
