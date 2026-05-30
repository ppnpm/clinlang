# Quickstart Guide

Learn how to write and format your very first structured clinical note in under two minutes.

---

## How It Works in 3 Steps

1. **Type Shorthand**: Open the ClinLang editor and write patient details using short, natural abbreviations (e.g., `htn` for hypertension, `rx` for prescriptions).
2. **Review Live**: Watch the structured SOAP note render instantly in the preview panel on the right.
3. **Copy and Save**: Click the copy icon in the preview panel to copy the formatted text and paste it directly into your hospital's electronic medical record (EMR).

---

## Try Writing Your First Note

Copy and paste the following shorthand block into the editor to see how ClinLang processes an admission for community-acquired pneumonia:

```text
pt 65M wt75
cc fever and cough
hpi fever for 3 days, productive cough with yellow sputum
pmh htn dm2
sx fever++ cough+++ sob+
vitals bp110/70 hr110 spo292 temp101.4f rr24
pe cvs:rr s1s2 rs:creps_right_base
ix hb11.2 wbc18000 na135 
ix cxr:right lower lobe consolidation
dx Community Acquired Pneumonia
rx amoxicillin 500mg tds po
```

### What ClinLang Generates:

```text
──────────────────────────────────────────────────
Patient: 65Y/M | Wt: 75kg  

S — SUBJECTIVE
─────────────────────────
Chief Complaint: fever and cough
HPI            : fever for 3 days, productive cough with yellow sputum
PMH            : Hypertension Type 2 Diabetes Mellitus
Symptoms       : fever (severe); cough (very severe); sob (mild)

O — OBJECTIVE
─────────────────────────
Vitals         : BP: 110/70 | HR: 110 bpm | SpO2: 92% | Temp: 101.4 F | RR: 24 /min
Physical Exam  : cvs:rr s1s2 rs:creps_right_base
Imaging/Rad    : CXR right lower lobe consolidation
Labs           : HB 11.2 | WBC 18000 | NA 135

A — ASSESSMENT
─────────────────────────
Diagnosis      : Community Acquired Pneumonia

P — PLAN
─────────────────────────
  ▸ Rx: Amoxicillin 500mg Orally, Three times daily
──────────────────────────────────────────────────
```

---

## Things to Notice
* **Automatic Abbreviations**: ClinLang expanded `dm2` to "Type 2 Diabetes Mellitus" and `htn` to "Hypertension".
* **Prescription Translation**: The prescription line `rx amoxicillin 500mg tds po` was translated to clear patient-facing instructions: "Amoxicillin 500mg Orally, Three times daily".
* **Auto-Categorization**: You do not have to write the SOAP headings. ClinLang reads your shorthand commands (like `ix` and `vitals`) and puts them into the correct **Subjective**, **Objective**, **Assessment**, or **Plan** sections.
* **Typing Flexibility**: You can type commands in any order. The generator will always output a clean, standardized note.

---

## Next Steps
* Learn the shortcuts to write even faster: **[How to Write Fast](how-to-write-fast.md)**
* Learn the full list of commands: **[Commands Directory](commands.md)**
