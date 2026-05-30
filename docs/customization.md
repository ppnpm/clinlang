# Customizing Shorthand & Clinic Settings

ClinLang is fully customizable. You can adapt the abbreviation engine, prescription vocabulary, and symptom labels to match your specific medical department or specialty without writing code.

---

## The Workspace `.config/` Directory

All customization is stored inside a hidden folder called `.config/` located at the root of your active workspace directory. 

You can edit these settings files directly in the ClinLang Web UI:
1.  Click the **Settings Drawer** (gear icon in the top right).
2.  Scroll down to the **Shorthand & Terms** section.
3.  Select the configuration file you want to edit from the dropdown list.
4.  Modify the JSON content and click **Save Configuration**.

---

## Customizable Configuration Files

The following 8 JSON configuration files can be overridden:

### 1. Abbreviations (`abbreviations.json`)
Maps clinical shorthand keys to their full professional expansions. 
*   **Default maps**: `dm2` $\rightarrow$ "Type 2 Diabetes Mellitus", `htn` $\rightarrow$ "Hypertension", `copd` $\rightarrow$ "Chronic Obstructive Pulmonary Disease".
*   **Example customization**:
    ```json
    {
      "ckd": "Chronic Kidney Disease",
      "pvd": "Peripheral Vascular Disease",
      "cabg": "Coronary Artery Bypass Graft"
    }
    ```

### 2. Prescription Frequencies (`frequencies.json`)
Normalizes frequency instructions in your `rx` prescriptions.
*   **Default maps**: `tds` $\rightarrow$ "Three times daily", `bd` $\rightarrow$ "Twice daily", `nocte` $\rightarrow$ "Once daily nocte".
*   **Example customization**:
    ```json
    {
      "aliases": {
        "tds": "TDS",
        "tid": "TDS"
      },
      "expansions": {
        "TDS": "Three times daily",
        "BD": "Twice daily"
      }
    }
    ```

### 3. Medication Routes (`routes.json`)
Determines how prescription administration routes are expanded.
*   **Default maps**: `po` $\rightarrow$ "Orally", `iv` $\rightarrow$ "Intravenously", `sc` $\rightarrow$ "Subcutaneously".
*   **Example customization**:
    ```json
    {
      "aliases": {
        "po": "PO",
        "iv": "IV"
      },
      "expansions": {
        "PO": "Orally",
        "IV": "Intravenously"
      }
    }
    ```

### 4. Symptoms Severity (`symptoms.json`)
Changes the clinical text used for symptom intensity levels (`+` and `-`).
*   **Default maps**: `+++` $\rightarrow$ "very severe", `++` $\rightarrow$ "severe", `+` $\rightarrow$ "mild".
*   **Example customization**:
    ```json
    {
      "+++": "critical severity",
      "++": "marked",
      "+": "slight"
    }
    ```

### 5. Custom Drugs Catalog (`drugs.json`)
Allows ClinLang to parse multi-word drugs (e.g. "Vitamin B12" or "Ferrous Sulfate") as a single name instead of confusing the dose. List all your clinic's common drugs as a simple JSON array:
```json
[
  "Aspirin",
  "Atorvastatin",
  "Metformin",
  "Vitamin B12",
  "Ferrous Sulfate"
]
```

### 6. Radiology Routing Keys (`rad_keys.json`)
Determines which keywords typed under the general investigation (`ix`) command are automatically routed to the **Imaging/Radiology** block instead of the **Labs** block.
*   **Defaults**: `cxr`, `xray`, `ct`, `mri`, `usg`, `echo`, `ecg`, `ekg`, `ultrasound`.
*   **Example customization**:
    ```json
    [
      "cxr",
      "hrct",
      "oct",
      "mammogram"
    ]
    ```

### 7. Time Durations (`durations.json`)
Defines the time suffixes (like `d` or `h`) recognized when reading patient age or symptom durations.
*   **Defaults**: `y`/`yr` (Years), `m`/`mo` (Months), `w`/`wk` (Weeks), `d`/`dy` (Days).

### 8. Reference Ranges (`reference_ranges.json`)
Sets high and low threshold markers for vitals and lab results. See **[Reference Ranges Guide](reference-ranges.md)** for a complete schema guide.
