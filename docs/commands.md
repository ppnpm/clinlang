# Commands Directory

Each line in a ClinLang note must start with a command keyword. This tells the system how to format and categorize that line of information.

---

## 1. Patient Context & Details

| Command | Category / Name | Description & Usage | Clinical Example |
|:---|:---|:---|:---|
| **`pt`** | Patient particulars | Age, biological sex, weight, height, bed, and ward unit. | `pt 45M wt80kg ht175cm bed4 unit:ER` |
| **`id`** | Patient ID | The hospital identifier or admission record number. | `id MRN-88421-A` |
| **`day`** | Day of Round / Timeline | Tracks the hospital timeline or post-operative day. | `day Day 3 of admission (Post-op Day 1)` |
| **`alg`** / **`allergy`** | Allergies | Highlights active allergies in a prominent block at the top. | `alg Penicillin, sulfa drugs` |

---

## 2. History & Subjective

| Command | Category / Name | Description & Usage | Clinical Example |
|:---|:---|:---|:---|
| **`cc`** | Chief Complaint | The primary reason the patient is seeking care. | `cc shortness of breath and chest tightness` |
| **`hpi`** | History of Present Illness | The narrative chronological story of the illness. | `hpi sudden onset chest pain radiating to jaw` |
| **`pmh`** | Past Medical History | Chronic diseases and past surgeries (uses auto-expansions). | `pmh dm2 htn ihd copd` |
| **`sh`** | Social History | Smoking, alcohol intake, diet, living situation, or employment. | `sh smoking 20 pack-years, alcohol socially` |
| **`fh`** | Family History | Medical conditions present in first-degree relatives. | `fh premature coronary artery disease in father` |

---

## 3. Symptoms & Objective Findings

| Command | Category / Name | Description & Usage | Clinical Example |
|:---|:---|:---|:---|
| **`sx`** | Symptoms | Quick symptom tracking with severity (`+`/`-`) and duration. | `sx chest_pain+++2h fever++3d nausea-` |
| **`vitals`** | Vital Signs | Numeric vital signs. Supports BP, HR, RR, SpO2, and Temp. | `vitals bp130/80 hr95 rr20 temp98.6f spo298` |
| **`pe`** / **`oe`** | Physical / Objective Exam | Document clinical exam findings. Use colons for systems. | `pe cvs:s1s2 rs:clear lungs abd:soft` |

*Note: In `vitals`, temperature requires an explicit unit suffix (`C` or `F`). A bare number like `temp98.6` defaults to Fahrenheit.*

---

## 4. Investigations, Assessment & Plan

| Command | Category / Name | Description & Usage | Clinical Example |
|:---|:---|:---|:---|
| **`ix`** | Investigations | Unified command for labs and imaging. Auto-routes by name. | `ix hb12.5 wbc14000 crp+ cxr:clear` |
| **`lab`** / **`labs`** | Lab Results | Forces findings directly into the Labs section. | `lab na135 k4.0 cr1.1 troponin+` |
| **`rad`** | Radiology / Imaging | Forces findings directly into the Imaging/Radiology section. | `rad ct_brain:no acute hemorrhage` |
| **`dx`** | Diagnosis | The primary provisional or confirmed diagnosis. | `dx Acute Coronary Syndrome (NSTEMI)` |
| **`ddx`** | Differential Diagnosis | Alternatives or other diagnoses being ruled out. | `ddx Pulmonary Embolism, Aortic Dissection` |
| **`rx`** | Medication Prescriptions | Orders medications with dose, frequency, route, and duration. | `rx amoxicillin 500mg tds po x7d` |
| **`img`** / **`image`** | Image Attachment | References a photo or scan. Supports drag-and-drop in UI. | `img images/171540124-wound_photo.png` |

---

## What If I Use an Unknown Command?

ClinLang is completely safe. If you type a command keyword that the system does not recognize (e.g. `gcs` or `eyes`), it will **not** fail. Instead, it will treat it as a custom structured category and list it under a separate section:

*   **You type**: `gcs e3v4m6`
*   **ClinLang outputs**: `GCS: e3v4m6` in your final note.
