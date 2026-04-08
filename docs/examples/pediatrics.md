# Pediatrics: Viral Fever

### Input
```text
pt 4Y M wt16 
alg nil
cc high grade fever for 2 days
hpi associated with runny nose, mild dry cough, and reduced appetite. no vomiting or loose stools.
pmh fully immunized
sx fever++2d cough+ runny_nose+ lethargy+
vitals hr120 spo298 temp102.1 rr26
pe general:active well_hydrated throat:mild_erythema chest:clear
ix hb11.5 wbc8000
ix dengue-
dx Viral Upper Respiratory Tract Infection
rx paracetamol 250mg tds po
rx cetirizine 5mg od po nocte
```

### Output
```text
──────────────────────────────────────────────────
Patient: 4Y/M | Wt: 16kg 

[!] ALLERGIES: NIL [!]
──────────────────────────────────────────────────

S — SUBJECTIVE
─────────────────────────
Chief Complaint: high grade fever for 2 days
HPI            : associated with runny nose, mild dry cough, and reduced appetite. no vomiting or loose stools.
PMH            : fully immunized
Symptoms       : fever (severe, 2 days); cough (mild); runny_nose (mild); lethargy (mild)

O — OBJECTIVE
─────────────────────────
Vitals         : HR: 120 bpm | SpO2: 98% | Temp: 102.1 | RR: 26 /min
Physical Exam  : general:active well_hydrated throat:mild_erythema chest:clear
Labs           : HB 11.5 | WBC 8000 | DENGUE -
⚠ Abnormals    : ⚠ HR 120 bpm | ⚠ RR 26 /min 

A — ASSESSMENT
─────────────────────────
Diagnosis      : Viral Upper Respiratory Tract Infection

P — PLAN
─────────────────────────
  ▸ Rx: Paracetamol 250mg Orally, Three times daily
  ▸ Rx: Cetirizine 5mg Orally, Once daily nocte
──────────────────────────────────────────────────
```
