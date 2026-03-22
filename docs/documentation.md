# ClinLang — Doctor's Guide

**ClinLang** lets you document a clinical case in seconds using shorthand. A doctor types shorthand — the app shows a structured note.

---

## How It Works

You write in a `.cln` file (or a text box in the app). Each line starts with a **command keyword**, followed by your clinical data.

```
pt 45M wt80 ht172
cc chest pain 3 hours
hpi sudden onset, radiated to left arm, diaphoresis
pmh dm2 htn
sx chestpain+++ sob++ diaphoresis+
vitals bp160/100 hr98 spo292 rr22
rx aspirin 300mg stat
rx metoprolol 25mg bd
dx inferior STEMI
```

That's the entire note.

---

## Core Commands

### `pt` — Patient

```
pt <age><sex>  wt<weight>  ht<height>
```

| What to type | Meaning |
|---|---|
| `pt 45M` | 45-year-old male |
| `pt 30F wt60 ht160` | 30 female, 60 kg, 160 cm |
| `pt wt70 34M` | Order doesn't matter |

---

### `cc` — Chief Complaint
```
cc chest pain 3 hours
cc shortness of breath on exertion
```

---

### `hpi` — History of Presenting Illness
```
hpi sudden onset CP radiating to jaw, associated nausea
```

---

### `pmh` — Past Medical History
Use abbreviations — they are automatically expanded in the output:

| You type | Shown as |
|---|---|
| `dm2` | Type 2 Diabetes Mellitus |
| `htn` | Hypertension |
| `ihd` | Ischaemic Heart Disease |
| `cad` | Coronary Artery Disease |
| `ckd3` | CKD Stage 3 |
| `copd` | COPD |
| `tb` | Tuberculosis |
| `cva` | Cerebrovascular Accident |
| `af` | Atrial Fibrillation |

```
pmh dm2 htn ckd3 ihd
```

---

### `sx` — Symptoms

Add intensity right after the symptom name (no space):

| Suffix | Meaning |
|---|---|
| `+` | Mild / present |
| `++` | Severe |
| `+++` | Very severe |
| `-` | Improving |
| `--` | Resolving |
| `---` | Resolved |

Add duration after intensity:

| Suffix | Unit |
|---|---|
| `3d` | 3 days |
| `2h` | 2 hours |
| `1w` | 1 week |

```
sx fever++3d cough+ sob++ chills-
sx chestpain+++4h diaphoresis+ nausea+
sx headache++2d vomiting+ photophobia+
```

---

### `vitals` — Vital Signs

```
vitals bp130/85 hr102 spo294 temp100.4 rr22
```

| Token | Example | Meaning |
|---|---|---|
| `bp` | `bp130/85` | Blood pressure |
| `hr` | `hr102` | Heart rate |
| `spo2` | `spo292` | SpO2 % |
| `temp` | `temp100.4` | Temperature |
| `rr` | `rr22` | Respiratory rate |

> **Automatic alerts:** The system will flag any value outside the normal range — you don't need to do anything extra.

---

### `rx` — Medications / Prescriptions

```
rx <drug> <dose> <frequency> [route] [duration]
```

> **Pro Tip:** Avoid using form prefixes like `tab`, `cap`, or `inj` before the drug name (e.g., avoid `rx tab paracetamol`). This ensures the parsed "Drug Name" stays perfectly clean. Instead, use the Route (`po`, `iv`) to implicitly specify the formulation.


| Frequency | Meaning |
|---|---|
| `od` | Once daily |
| `bd` | Twice daily |
| `tds` | Three times daily |
| `qds` | Four times daily |
| `stat` | Immediately |
| `prn` | As needed |
| `nocte` | At night |

| Route | Meaning |
|---|---|
| `iv` | Intravenous |
| `im` | Intramuscular |
| `po` | Oral |
| `neb` | Nebulised |
| `sl` | Sublingual |
| `sc` | Subcutaneous |

```
rx metoprolol 25mg bd
rx amoxicillin 500mg tds po x7d
rx lasix 40mg od iv
rx gtn sl stat prn
rx paracetamol 1g qds prn
```

> **Note:** Frequency and route acronyms (e.g., `bd`, `iv`) are preserved strictly in the JSON payload but are automatically expanded to patient-friendly text (e.g., `Twice daily`, `Intravenously`) in the generated clinical note!

---

### `dx` — Diagnosis

```
dx community acquired pneumonia
dx inferior STEMI
dx DKA with severe hypokalaemia
```

---

## Specialty Data — Extension Commands

Anything not listed above is automatically captured as structured data. No special setup needed.

### Labs
```
lab hb11.2 wbc14500 plt220000 crp48 na138 k3.8 glucose18
```

> Lab values are also automatically checked against normal ranges.

### Clinical Exam
```
exam crepitations+ consolidation+ bronchialbreathing+
exam jvp+ pedaloedema+ s3+
```

### Ophthalmology
```
va od6/6 os6/18
iop od14 os22
fundus cupping+ avm+
```

### Obstetrics
```
obs ga34w parity2 edd2026-06-01
```

### ECG
```
ecg stemi+ inferiorleads st_elevation+
```

### Neurology
```
neuro gcs15 pupils_equal+ plantar_flexor+
```

---

## Full Example — Inferior STEMI

```cln
# Inferior STEMI — Emergency Admission

pt 58M wt82 ht172
cc chest pain and breathlessness since 4 hours
hpi sudden onset central chest pain radiating to left arm and jaw, associated diaphoresis and nausea
pmh dm2 htn hyperlipidaemia

sx chestpain+++4h sob++ diaphoresis+ nausea+

vitals bp160/100 hr98 spo292 temp98.6 rr22

lab hb13.2 wbc11000 trop0.8 ckmb48 creatinine1.1 na138 k4.1
ecg stemi+ inferiorlead II_III_avF

rx aspirin 300mg stat po
rx clopidogrel 300mg stat po
rx heparin 5000u stat iv
rx metoprolol 25mg bd

dx Inferior STEMI — for primary PCI
```

---

## Abnormal Value Alerts

The system automatically flags out-of-range values. You'll see:

```
⚠  HR: 102 bpm — Tachycardia (>100 bpm)
🔴 CRITICAL  SpO2: 88% — Severe hypoxia (SpO2 < 90%)
⚠  Hb: 6.8 g/dL — Severe anaemia (Hb < 7)
```

These appear in all output formats automatically.

---

## Tips

- **Comments:** Start a line with `#` — it's ignored
- **Order doesn't matter** within a command: `pt 34M wt65` = `pt wt65 34M`
- **Abbreviations expand automatically** in the output — type short, read full
- **No punctuation needed** — spaces separate everything
- **Case insensitive** — `PT`, `Pt`, `pt` all work
