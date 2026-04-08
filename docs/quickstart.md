# Quickstart

Write your very first clinical note in under 2 minutes.

## How to use it
1.  Open ClinLang on your device.
2.  Type your commands into the editor.
3.  Click "Generate Note" (or press Enter).
That's it. Your note is ready to copy into the hospital system.

---

## Your First Case
Let's write a standard pneumonia admission. Just type exactly what you would say during a case presentation.

**You type:**
```text
pt 65M wt75
cc fever and cough
hpi fever for 3 days, productive cough with yellow sputum
pmh htn dm2
sx fever++ cough+++ sob+
vitals bp110/70 hr110 spo292 temp101.4 rr24
pe cvs:rr s1s2 rs:creps_right_base
ix hb11.2 wbc18000 cr1.0 na135 
ix cxr:right lower lobe consolidation
dx Community Acquired Pneumonia
rx amoxicillin 500mg tds po
```

**The Final Note You Get:**
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
Vitals         : BP: 110/70 | HR: 110 bpm | SpO2: 92% | Temp: 101.4 | RR: 24 /min
Physical Exam  : cvs:rr s1s2 rs:creps_right_base
Imaging/Rad    : CXR right lower lobe consolidation
Labs           : HB 11.2 | WBC 18000 | CR 1.0 | NA 135
⚠ Abnormals    : ⚠ HR 110 bpm | ⚠ SpO2 92% | ⚠ RR 24 /min | ⚠ WBC 18000 /μL

A — ASSESSMENT
─────────────────────────
Diagnosis      : Community Acquired Pneumonia

P — PLAN
─────────────────────────
  ▸ Rx: Amoxicillin 500mg Orally, Three times daily
──────────────────────────────────────────────────
```

Notice how ClinLang automatically translated `dm2` into "Type 2 Diabetes Mellitus", expanded the prescription, and flagged the high heart rate and low oxygen for you!
