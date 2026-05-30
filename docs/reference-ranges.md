# Reference Ranges & Out-of-Range Markers

ClinLang can automatically annotate numeric vitals and lab values in your notes when they fall outside a configured reference range. 

> [!IMPORTANT]
> **Transcription Aid Only**
> Out-of-range markers are designed strictly as a transcription aid to help clinicians double-check values. They are **not** clinical decision support, severity indicators, or recommendations. The clinician is always the final authority and must interpret all findings in context.

---

## What it Looks Like

When markers are enabled, the SOAP note will flag out-of-range values like this:

```text
Vitals: BP: 160/100 | HR: 105 bpm | SpO2: 94% | Temp: 98.6 | RR: 22 /min
        - HR 105 bpm (outside ref 60-100, AHA adult defaults)
        - BP 160/100 mmHg (outside ref 90-140 / 60-90, JNC-8 defaults)
```

In the ClinLang web UI, these markers are highlighted on hover over the vitals or lab lines, and are displayed in the SOAP preview.

---

## Default Reference Ranges

ClinLang comes with standard adult reference ranges built-in. Below is the list of values checked:

| Field | Key | Default Range | Unit | Source Citation |
|:---|:---|:---|:---|:---|
| **Heart Rate** | `vitals.hr` | 60 – 100 | bpm | user-default adult ranges |
| **Systolic BP** | `vitals.bp.systolic` | 90 – 140 | mmHg | user-default adult ranges |
| **Diastolic BP** | `vitals.bp.diastolic` | 60 – 90 | mmHg | user-default adult ranges |
| **Oxygen Saturation** | `vitals.spo2` | 94 – 100 | % | user-default adult ranges |
| **Respiratory Rate** | `vitals.rr` | 12 – 20 | breaths/min | user-default adult ranges |
| **Temperature** | `vitals.temp` | 97.0 – 100.0 | °F | user-default adult ranges |
| **Haemoglobin (Male)** | `labs.hb.M` | 13.0 – 17.0 | g/dL | user-default adult ranges |
| **Haemoglobin (Female)** | `labs.hb.F` | 11.0 – 17.0 | g/dL | user-default adult ranges |
| **White Blood Cells** | `labs.wbc` | 4,000 – 11,000 | /µL | user-default adult ranges |
| **Creatinine** | `labs.creatinine` | 0.6 – 1.2 | mg/dL | user-default adult ranges |
| **Sodium (Na)** | `labs.na` | 135 – 145 | mEq/L | user-default adult ranges |
| **Potassium (K)** | `labs.k` | 3.5 – 5.0 | mEq/L | user-default adult ranges |
| **Glucose** | `labs.glucose` | 3.9 – 7.8 | mmol/L | user-default adult ranges |

---

## Customizing Reference Ranges

You can override any default range to match your local laboratory guidelines or hospital policy.

### Editing in the Web UI (Easiest)
1. Open the **Settings Drawer** (gear icon in top right).
2. Scroll to the **Shorthand & Terms** section.
3. Select **`reference_ranges.json`** from the configuration files dropdown.
4. Modify the values in the JSON text box and click **Save Configuration**.

### Reference Range Schema Format
Each entry in `reference_ranges.json` looks like this:
```json
"vitals.hr": {
  "low": 60,
  "high": 100,
  "unit": "bpm",
  "source": "AHA Adult Defaults"
}
```
*   **`low`** (number, optional): The lower limit. Values below this are flagged.
*   **`high`** (number, optional): The upper limit. Values above this are flagged.
*   **`unit`** (string, optional): A text label showing the unit.
*   **`source`** (string, required): The citation source shown in the warning. Always name this so you (and other readers) can audit where the range came from (e.g. "Hospital X Lab Guideline 2026").

---

## Special Range Rules

### 1. Haemoglobin (Sex-Specific)
Haemoglobin ranges are split into `labs.hb.M` and `labs.hb.F`. 
*   If a patient is registered as male (`pt 58M`), the system uses `labs.hb.M`.
*   If a patient is registered as female (`pt 58F`), the system uses `labs.hb.F`.
*   If you omit the sex in the `pt` line (e.g. `pt 58`), **no Haemoglobin markers will be generated**. The engine does not guess.

### 2. Temperature (Unit Suffixes)
ClinLang does **not** perform automatic unit conversion between Fahrenheit (°F) and Celsius (°C). 
*   If you type your temperature in Celsius (`temp37.5c`), it must match a range configured with `"unit": "C"`.
*   If you type your temperature in Fahrenheit (`temp98.6f`), it must match a range configured with `"unit": "F"`.
*   If the units do not match, the range check is skipped. If you want to use Celsius defaults, edit your `reference_ranges.json` to reflect your desired Celsius limits and set `"unit": "C"`.
