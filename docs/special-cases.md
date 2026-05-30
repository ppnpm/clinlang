# Specialty Workflows (Pediatrics & Obstetrics)

ClinLang accommodates specialty-specific workflows out of the box, with built-in rules for pediatric age reporting and an extensible plugin system for obstetrics and gynaecology.

---

## 1. Pediatrics & Neonates

For neonates and pediatric cases, reporting age in years is insufficient. ClinLang parses age numbers followed by specific duration units:

*   **Days**: Append `d` to the age number (e.g., `pt 10d M` $\rightarrow$ 10-day-old male).
*   **Weeks**: Append `w` to the age number (e.g., `pt 3w F` $\rightarrow$ 3-week-old female).
*   **Months**: Append `m` or `mo` to the age number (e.g., `pt 18m M` $\rightarrow$ 18-month-old male).

### Full Pediatric Admission Example:
```text
pt 18m M wt11.5
cc fever and ear pulling for 2 days
hpi irritable, poor feeding since yesterday. no vomiting.
sx fever++2d ear_pain+ irritable++
vitals hr125 rr28 temp101.8f spo298
pe general:miserable but hydrated ent:right tympanic membrane bulging and erythematous
dx Acute Otitis Media (Right)
rx amoxicillin 250mg tds po x5d
```

---

## 2. Obstetrics & Gynaecology (`obgyn` profile)

For pregnant patients, you can activate the **OB/GYN specialty profile** by adding `@profile obgyn` at the very top of your file. 

This profile unlocks obstetric-specific commands and allows you to specify gestational age and fetal heart rates in-line with your normal patient details and vitals.

### Standalone Obstetric Commands

When `@profile obgyn` is active, you can use these five dedicated keywords:

| Keyword | Description | Clinical Example |
|:---|:---|:---|
| **`lmp`** | Last Menstrual Period date | `lmp 2025-06-15` |
| **`edd`** | Estimated Date of Delivery | `edd 2026-03-22` |
| **`gpal`** | Gravida, Para, Abortus, Living | `gpal G2P1A0L1` |
| **`fhs`** | Fetal Heart Sounds description | `fhs 140 bpm, regular, reactive` |
| **`ctx`** | Uterine Contractions description | `ctx 3 in 10 minutes, lasting 40 seconds` |

### Inline Obstetric Details

The `obgyn` profile also allows you to insert obstetric numbers directly into the standard `pt` and `vitals` commands:

1.  **Gestational Age inside `pt`**: Append `ga:<weeks>w` or `ga:<weeks>` to the patient command:
    *   `pt 28F ga:34w` $\rightarrow$ 28-year-old female at 34 weeks gestation.
2.  **Fetal Heart Rate inside `vitals`**: Append `fhr:<bpm>` or `fhr:<bpm>bpm` to the vitals command:
    *   `vitals bp110/70 fhr:142` $\rightarrow$ Maternal blood pressure 110/70 and Fetal Heart Rate of 142 bpm.

---

## Full Obstetric Case Example

Copy and paste the following note into the editor with the obgyn profile active:

```text
@profile obgyn
pt 28F wt68 ga:38w
gpal G2P1A0L1
lmp 2025-09-08
edd 2026-06-15
cc lower abdominal pains and leaking fluid
hpi contractions started 4 hours ago. clear fluid leaking since 1 hour.
sx labor_pains++4h leaking+1h bleeding-
vitals bp120/80 hr85 temp98.6f fhr:140
pe abd:uterus term cephalic ctx:3in10min
pe pv:cervix 4cm dilated 80% effaced station-1
ix hb11.0 wbc12000
dx Active Phase of Labor (Term Gestation)
```
