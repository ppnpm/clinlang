# Command Reference

Start every line with one of these short commands to tell ClinLang what you are writing.

## Patient Context
*   **`pt`** (Patient): Basic demographics.
    *   `pt 45M wt80`
    *   `pt 12F ht120 wt30`
*   **`day`** (Hospital Day / Timeline): For daily ward rounds.
    *   `day Day 3 of admission`
    *   `day Day 1 post-op`
*   **`alg`** (Allergies): Puts a massive warning at the top of the note.
    *   `alg Penicillin`
    *   `alg Sulfa, peanuts`

## Clinical History
*   **`cc`** (Chief Complaint): Why they are here.
    *   `cc chest pain for 3 hours`
*   **`hpi`** (History of Present Illness): The story.
    *   `hpi sudden onset, radiating to jaw`
*   **`pmh`** (Past Medical History): Use abbreviations!
    *   `pmh dm2 htn ihd`
*   **`sh`** (Social History):
    *   `sh smoker 20py`
*   **`fh`** (Family History):
    *   `fh cad in mother`

## Clinical Exam
*   **`sx`** (Symptoms): Use `+` for severity.
    *   `sx chestpain+++ sob++ fever-`
*   **`vitals`**: Attach the numbers directly to the vital signs.
    *   `vitals bp120/80 hr90 rr18 temp98.6F spo299`
    *   Temperature requires an explicit unit suffix (`F` or `C`); a bare `temp98.6` defaults to Fahrenheit. The engine does not convert between units — see [reference-ranges.md](reference-ranges.md).
*   **`pe`** (Physical Exam): Use colons for body systems.
    *   `pe chest:clear cvs:normal abd:soft`

## Investigations & Plan
*   **`ix`** (Investigations): Works for labs, imaging, and cultures.
    *   `ix hb12.1 wbc14000 na135`
    *   `ix cxr:wnl ecg:stemi`
    *   `ix hiv- dengue+`
*   **`dx`** (Diagnosis): 
    *   `dx Dengue Hemorrhagic Fever`
*   **`rx`** (Prescriptions): Write the drug, dose, and frequency directly.
    *   `rx paracetamol 1g qds po`
    *   `rx ceftriaxone 1g od iv`

## Specialized Commands
If you use a command ClinLang doesn't know, it prints it exactly as structured data.
*   **`obs`** (Obstetrics): `obs ga34w parity2`
*   **`neuro`** (Neurology): `neuro gcs15 pupils:equal`
