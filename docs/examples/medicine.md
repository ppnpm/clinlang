# Medicine: Acute Coronary Syndrome (ACS)

### Input
```text
pt 58M wt82
cc severe crushing chest pain 
hpi started 2 hours ago while walking, radiating to left arm. associated with profound sweating and nausea.
pmh dm2 htn ihd
sx chest_pain+++2h diaphoresis++ nausea+ sob+
vitals bp160/100 hr105 spo294 temp98.6 rr22
pe cvs:s1s2_normal no_murmur rs:clear
ix hb13.5 wbc11000 na138 k4.1 cr1.2
ix ecg:anterior stemi
ix trop+
dx Anterior STEMI
rx aspirin 300mg stat po
rx clopidogrel 300mg stat po
rx gtn sl prn
```

### Output
```text
──────────────────────────────────────────────────
Patient: 58Y/M | Wt: 82kg 

S — SUBJECTIVE
─────────────────────────
Chief Complaint: severe crushing chest pain
HPI            : started 2 hours ago while walking, radiating to left arm. associated with profound sweating and nausea.
PMH            : Type 2 Diabetes Mellitus Hypertension Ischaemic Heart Disease
Symptoms       : chest_pain (very severe, 2 hours); diaphoresis (severe); nausea (mild); sob (mild)

O — OBJECTIVE
─────────────────────────
Vitals         : BP: 160/100 | HR: 105 bpm | SpO2: 94% | Temp: 98.6 | RR: 22 /min
Physical Exam  : cvs:s1s2_normal no_murmur rs:clear
Imaging/Rad    : ECG anterior stemi
Labs           : HB 13.5 | WBC 11000 | NA 138 | K 4.1 | CR 1.2 | TROP +

A — ASSESSMENT
─────────────────────────
Diagnosis      : Anterior STEMI

P — PLAN
─────────────────────────
  ▸ Rx: Aspirin 300mg Orally, Immediately
  ▸ Rx: Clopidogrel 300mg Orally, Immediately
  ▸ Rx: Gtn Sublingual, As needed
──────────────────────────────────────────────────
```
