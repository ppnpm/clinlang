# Special Cases

ClinLang works for every specialty. Here is how to format notes for specific situations.

## Neonates
For babies less than a month old, use `d` (days) instead of years. The system handles it automatically.
```text
pt 10d M wt3.2
```

For infants, use `m` (months).
```text
pt 6m F wt6.5
```

## Pregnancy / Obstetrics
For pregnant patients, just use the `obs` command to capture gravidity, parity, and gestational age.
```text
pt 26F wt70
obs g2p1l1 
obs ga32w edd2024-10-12
```

## Complex Lab Tests
For tests like titers or arterial blood gases, just use a colon to tell ClinLang exactly what the test is.
```text
ix widal:1/160
ix pco2:45
ix po2:90
```
