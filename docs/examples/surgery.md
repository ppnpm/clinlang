# Surgery: Acute Abdomen

### Input
```text
pt 24F wt55
alg sulfa
cc right lower quadrant pain 
hpi started 12 hours ago around umbilicus, migrated to RIF. associated with vomiting x2.
pmh nil
sx abd_pain+++12h vomiting+ fever+
vitals bp110/70 hr98 spo299 temp100.2 rr18
pe abd:tender_rif rebound_tenderness+ rovsings_sign+
ix hb12.2 wbc16000 plt250000 
ix usg:inflamed appendix 9mm with free fluid
dx Acute Appendicitis
rx paracetamol 1g stat iv
rx ceftriaxone 1g stat iv
rx metronidazole 500mg stat iv
```

### Output
```text
──────────────────────────────────────────────────
Patient: 24Y/F | Wt: 55kg 

[!] ALLERGIES: SULFA [!]
──────────────────────────────────────────────────

S — SUBJECTIVE
─────────────────────────
Chief Complaint: right lower quadrant pain
HPI            : started 12 hours ago around umbilicus, migrated to RIF. associated with vomiting x2.
PMH            : nil
Symptoms       : abd_pain (very severe, 12 hours); vomiting (mild); fever (mild)

O — OBJECTIVE
─────────────────────────
Vitals         : BP: 110/70 | HR: 98 bpm | SpO2: 99% | Temp: 100.2 | RR: 18 /min
Physical Exam  : abd:tender_rif rebound_tenderness+ rovsings_sign+
Imaging/Rad    : USG inflamed appendix 9mm with free fluid
Labs           : HB 12.2 | WBC 16000 | PLT 250000

A — ASSESSMENT
─────────────────────────
Diagnosis      : Acute Appendicitis

P — PLAN
─────────────────────────
  ▸ Rx: Paracetamol 1g Intravenously, Immediately
  ▸ Rx: Ceftriaxone 1g Intravenously, Immediately
  ▸ Rx: Metronidazole 500mg Intravenously, Immediately
──────────────────────────────────────────────────
```
