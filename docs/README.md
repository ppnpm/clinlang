# ClinLang Documentation

ClinLang is a fast way to write structured clinical cases using short commands instead of long sentences. 

Instead of typing repetitive paragraphs in your hospital notes, you type quick clinical shorthand. ClinLang instantly turns it into a clear, formatted ward note. Optional reference-range markers can be enabled to annotate values that fall outside user-configurable ranges (transcription aid, not decision support).

ClinLang is a personal note-taking and templating tool — not a medical device. It does not provide diagnosis, treatment, dosing, or clinical decision support. See [DISCLAIMER.md](../DISCLAIMER.md) at the repository root.

### Quick Example

**You type:**
```text
pt 58M wt82
cc chest pain for 1h
sx pain+++ sob++
vitals bp160/100 hr98
ix ecg:stemi trop+
```

**ClinLang gives you:**
```text
Patient: 58Y/M | Wt: 82kg

S — SUBJECTIVE
Chief Complaint: chest pain for 1 hour
Symptoms       : pain (very severe); sob (severe)

O — OBJECTIVE
Vitals         : BP: 160/100 | HR: 98 bpm
Imaging/Rad    : ECG stemi
Labs           : TROP +
```

### Table of Contents
*   [Quickstart](quickstart.md) - Write your first note in 2 minutes.
*   [How to Write Fast](how-to-write-fast.md) - Learn the shorthand tricks.
*   [Commands Reference](commands.md) - A simple list of all commands.
*   [Reference Ranges](reference-ranges.md) - Optional out-of-range markers; how to override.
*   [Special Cases](special-cases.md) - Pediatrics and Obstetrics.
*   [Examples Directory](examples/README.md) - Real cases by specialty.
*   [Templates](templates/README.md) - Copy-paste templates for daily use.
*   [FAQ](faq.md)
