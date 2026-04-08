# Obstetrics: Labor Admission

### Input
```text
pt 28F wt68
obs g2p1l1 ga38w edd2026-10-14
cc lower abdominal pains and leaking pv
hpi pain started 4 hours ago, occurring every 5 minutes. clear fluid leaking since 1 hour.
sx labor_pains++4h leaking+1h bleeding-
vitals bp120/80 hr85 spo299 temp98.6 rr16
pe cvs:rr s1s2 rs:clear abd:uterus_term cephalic fhs140 regular_contractions+
pe pv:cx_fully_dilated fully_effaced station+1 clear_liquor
ix hb11.0 wbc12000
ix usg:single live intrauterine fetus term cephalic
dx Active Phase of Labor (Term Gestation)
```

### Output
```text
──────────────────────────────────────────────────
Patient: 28Y/F | Wt: 68kg 

S — SUBJECTIVE
─────────────────────────
Chief Complaint: lower abdominal pains and leaking pv
HPI            : pain started 4 hours ago, occurring every 5 minutes. clear fluid leaking since 1 hour.
Symptoms       : labor_pains (severe, 4 hours); leaking (mild, 1 hours); bleeding (improving)

O — OBJECTIVE
─────────────────────────
Vitals         : BP: 120/80 | HR: 85 bpm | SpO2: 99% | Temp: 98.6 | RR: 16 /min
Physical Exam  : cvs:rr s1s2 rs:clear abd:uterus_term cephalic fhs140 regular_contractions+ pv:cx_fully_dilated fully_effaced station+1 clear_liquor
Imaging/Rad    : USG single live intrauterine fetus term cephalic
Labs           : HB 11.0 | WBC 12000
OBS            : g2p1l1 ga38w edd2026-10-14

A — ASSESSMENT
─────────────────────────
Diagnosis      : Active Phase of Labor (Term Gestation)

P — PLAN
─────────────────────────
──────────────────────────────────────────────────
```
