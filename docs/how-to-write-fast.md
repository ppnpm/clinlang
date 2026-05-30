# How to Write Fast

ClinLang is built for speed. Its core engine uses smart rules so you don't have to type spaces, colons, or punctuation for simple entries. Follow these clinical shorthand tricks to write notes in seconds.

---

## 1. Patient Details (No Spaces)

Avoid unnecessary typing. Combine age, biological sex, weight, and height into single tokens:

*   **Age and Sex**: Type age followed immediately by `M` or `F` (e.g., `58M`, `32F`, `6mF` for a 6-month-old female).
*   **Weight**: Attach `wt` to the front of the weight (e.g., `wt70` or `wt70kg`).
*   **Height**: Attach `ht` to the front of the height (e.g., `ht170` or `ht170cm`).
*   **Bed & Unit**: Type `bed12` or `unit:ICU`.

**Example:**
*   Instead of typing: `Patient is a 58 year old male weighing 70kg on bed 12`
*   Type this: `pt 58M wt70 bed12`

---

## 2. Symptom Tracking with Plus & Minus (`+` & `-`)

Instead of writing sentences about symptom severity or progression, append `+` or `-` to the symptom name:

| Suffix | Meaning | Output Translation |
|:---|:---|:---|
| `+` | Mild or present | mild |
| `++` | Severe | severe |
| `+++` | Very severe | very severe |
| `-` | Improving | improving |
| `--` | Resolving / Better | resolving |
| `---` | Resolved / Gone | resolved |

### Appending Durations to Symptoms
You can add duration directly after the plus/minus signs using `h` (hours) or `d` (days):
*   `pain+++4h` $\rightarrow$ Very severe pain for 4 hours
*   `cough++3d` $\rightarrow$ Severe cough for 3 days
*   `nausea+` $\rightarrow$ Mild nausea

---

## 3. Using Colons (`:`) for Descriptive Findings

For unstructured text or findings that require sentences/words, separate the body system (or test) from the description with a colon (`:`).

*   `pe chest:clear bilaterally` $\rightarrow$ Lungs clear bilaterally under Physical Exam.
*   `pe cvs:normal s1s2 no murmurs` $\rightarrow$ Normal heart sounds under Physical Exam.
*   `ix cxr:patchy consolidation left lower lobe` $\rightarrow$ Chest X-Ray findings grouped under Imaging.

---

## 4. Blood Tests & Lab Values (No Punctuation)

For simple numeric lab values, you don't need spaces or colons. Type the abbreviation followed by the number:

*   `hb12.1` $\rightarrow$ Hemoglobin 12.1
*   `wbc14200` $\rightarrow$ White Blood Cell count 14,200
*   `cr1.1` $\rightarrow$ Creatinine 1.1

### Serology & Positive/Negative Flags
For positive/negative tests, append `+` or `-` to the test abbreviation:
*   `hiv-` $\rightarrow$ HIV negative
*   `dengue+` $\rightarrow$ Dengue positive
*   `crp+++` $\rightarrow$ CRP strongly positive

---

## 5. Writing Prescriptions (Rx)

Write prescriptions naturally in a single line using standard medical shorthand. The engine recognizes the drug, dose, frequency, route, and duration automatically:

```text
rx <drug_name> <dose> <frequency> <route> <duration>
```

### Examples:
*   `rx paracetamol 1g qds po x5d` $\rightarrow$ Paracetamol 1g Orally, Four times daily for 5 days
*   `rx ceftriaxone 1g od iv` $\rightarrow$ Ceftriaxone 1g Intravenously, Once daily
*   `rx gtn sl prn` $\rightarrow$ GTN Sublingual, As needed
