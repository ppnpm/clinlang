package engine

import (
	"fmt"
	"strconv"
	"strings"
)

// Severity levels for abnormal flags.
const (
	SeverityWarning  = "warning"
	SeverityCritical = "critical"
)

// AbnormalFlag describes one out-of-range value detected in a case.
type AbnormalFlag struct {
	Field    string `json:"field"`
	Value    string `json:"value"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "warning" or "critical"
}

// =============================================================================
// VITALS RANGES
// (Based on standard adult reference ranges)
// =============================================================================

// CheckAbnormals inspects vitals and known lab values in Extra and
// appends AbnormalFlag entries to the ClinicalCase.
func CheckAbnormals(c *ClinicalCase) {
	checkVitals(c)
	checkLabs(c)
	checkSerology(c)
}

func checkVitals(c *ClinicalCase) {
	v := c.Vitals

	// Blood pressure: parse systolic from "140/90"
	if v.BP != "" {
		sys, dia := parseBP(v.BP)
		if sys > 0 {
			switch {
			case sys >= 180:
				c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"BP", v.BP, "Hypertensive Crisis (SBP ≥ 180)", SeverityCritical})
			case sys >= 140:
				c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"BP", v.BP, "Hypertension (SBP ≥ 140)", SeverityWarning})
			case sys < 90:
				c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"BP", v.BP, "Hypotension (SBP < 90)", SeverityCritical})
			}
		}
		if dia > 0 && dia >= 110 {
			c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"BP", v.BP, "Severely elevated diastolic (DBP ≥ 110)", SeverityCritical})
		}
	}

	// Heart rate
	if v.HR > 0 {
		switch {
		case v.HR > 150:
			c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"HR", fmt.Sprintf("%d bpm", v.HR), "Severe tachycardia (>150)", SeverityCritical})
		case v.HR > 100:
			c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"HR", fmt.Sprintf("%d bpm", v.HR), "Tachycardia (>100 bpm)", SeverityWarning})
		case v.HR < 40:
			c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"HR", fmt.Sprintf("%d bpm", v.HR), "Severe bradycardia (<40 bpm)", SeverityCritical})
		case v.HR < 60:
			c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"HR", fmt.Sprintf("%d bpm", v.HR), "Bradycardia (<60 bpm)", SeverityWarning})
		}
	}

	// SpO2
	if v.SpO2 > 0 {
		switch {
		case v.SpO2 < 85:
			c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"SpO2", fmt.Sprintf("%d%%", v.SpO2), "Severe hypoxia (SpO2 < 85%)", SeverityCritical})
		case v.SpO2 < 94:
			c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"SpO2", fmt.Sprintf("%d%%", v.SpO2), "Hypoxia (SpO2 < 94%)", SeverityWarning})
		}
	}

	//Temprature
	if v.Temp > 0 {
		switch {
			case v.Temp > 40:
				c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"Temp", fmt.Sprintf("%.1f °C", v.Temp), "Severe hyperthermia (>40 °C)", SeverityCritical})
			case v.Temp > 39:
				c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"Temp", fmt.Sprintf("%.1f °C", v.Temp), "Hyperthermia (>39 °C)", SeverityWarning})
			case v.Temp < 35:
				c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"Temp", fmt.Sprintf("%.1f °C", v.Temp), "Severe hypothermia (<35 °C)", SeverityCritical})
			case v.Temp < 36:
				c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"Temp", fmt.Sprintf("%.1f °C", v.Temp), "Hypothermia (<36 °C)", SeverityWarning})
		}
	}

	// Respiratory rate
	if v.RR > 0 {
		
		switch {
		case v.RR > 30:
			c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"RR", fmt.Sprintf("%d /min", v.RR), "Severe tachypnoea (>30/min)", SeverityCritical})
		case v.RR > 20:
			c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"RR", fmt.Sprintf("%d /min", v.RR), "Tachypnoea (>20/min)", SeverityWarning})
		case v.RR < 10:
			c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{"RR", fmt.Sprintf("%d /min", v.RR), "Bradypnoea (<10/min)", SeverityCritical})
		}
	}
}

func checkLabs(c *ClinicalCase) {
	labs := c.Labs
	if len(labs) == 0 {
		return
	}
	
	p := c.Patient

	checkLabFloat(c, labs, "hb", func(v float64) *AbnormalFlag {
		lowerBound := 11.0
		if p.Sex == "M" {
			lowerBound = 13.0
		}
		switch {
		case v < 7:
			return &AbnormalFlag{"Hb", fmt.Sprintf("%.1f g/dL", v), "Severe anaemia (Hb < 7)", SeverityCritical}
		case v < lowerBound:
			return &AbnormalFlag{"Hb", fmt.Sprintf("%.1f g/dL", v), fmt.Sprintf("Anaemia (Hb < %.1f)", lowerBound), SeverityWarning}
		case v > 17:
			return &AbnormalFlag{"Hb", fmt.Sprintf("%.1f g/dL", v), "Polycythaemia (Hb > 17)", SeverityWarning}
		}
		return nil
	})

	checkLabFloat(c, labs, "wbc", func(v float64) *AbnormalFlag {
		switch {
		case v > 20000:
			return &AbnormalFlag{"WBC", fmt.Sprintf("%.0f /μL", v), "Severe leukocytosis (>20,000)", SeverityCritical}
		case v > 11000:
			return &AbnormalFlag{"WBC", fmt.Sprintf("%.0f /μL", v), "Leukocytosis (>11,000)", SeverityWarning}
		case v < 4000:
			return &AbnormalFlag{"WBC", fmt.Sprintf("%.0f /μL", v), "Leukopenia (<4,000)", SeverityWarning}
		}
		return nil
	})

	checkLabFloat(c, labs, "creatinine", func(v float64) *AbnormalFlag {
		switch {
		case v > 5:
			return &AbnormalFlag{"Creatinine", fmt.Sprintf("%.1f mg/dL", v), "Severely elevated creatinine (>5)", SeverityCritical}
		case v > 1.2:
			return &AbnormalFlag{"Creatinine", fmt.Sprintf("%.1f mg/dL", v), "Elevated creatinine (>1.2)", SeverityWarning}
		}
		return nil
	})

	checkLabFloat(c, labs, "na", func(v float64) *AbnormalFlag {
		switch {
		case v < 125:
			return &AbnormalFlag{"Na+", fmt.Sprintf("%.0f mEq/L", v), "Severe hyponatraemia (<125)", SeverityCritical}
		case v < 135:
			return &AbnormalFlag{"Na+", fmt.Sprintf("%.0f mEq/L", v), "Hyponatraemia (<135)", SeverityWarning}
		case v > 150:
			return &AbnormalFlag{"Na+", fmt.Sprintf("%.0f mEq/L", v), "Hypernatraemia (>150)", SeverityWarning}
		}
		return nil
	})

	checkLabFloat(c, labs, "k", func(v float64) *AbnormalFlag {
		switch {
		case v < 2.5:
			return &AbnormalFlag{"K+", fmt.Sprintf("%.1f mEq/L", v), "Severe hypokalaemia (<2.5)", SeverityCritical}
		case v < 3.5:
			return &AbnormalFlag{"K+", fmt.Sprintf("%.1f mEq/L", v), "Hypokalaemia (<3.5)", SeverityWarning}
		case v > 6:
			return &AbnormalFlag{"K+", fmt.Sprintf("%.1f mEq/L", v), "Severe hyperkalaemia (>6.0)", SeverityCritical}
		case v > 5:
			return &AbnormalFlag{"K+", fmt.Sprintf("%.1f mEq/L", v), "Hyperkalaemia (>5.0)", SeverityWarning}
		}
		return nil
	})

	checkLabFloat(c, labs, "glucose", func(v float64) *AbnormalFlag {
		switch {
		case v > 20:
			return &AbnormalFlag{"Glucose", fmt.Sprintf("%.1f mmol/L", v), "Severely elevated glucose (>20)", SeverityCritical}
		case v > 11:
			return &AbnormalFlag{"Glucose", fmt.Sprintf("%.1f mmol/L", v), "Hyperglycaemia (>11)", SeverityWarning}
		case v < 3:
			return &AbnormalFlag{"Glucose", fmt.Sprintf("%.1f mmol/L", v), "Hypoglycaemia (<3.0)", SeverityCritical}
		}
		return nil
	})
}

// checkSerology scans all c.Labs for common positive/negative test outcomes
func checkSerology(c *ClinicalCase) {
	viralPanel := map[string]bool{"hiv": true, "hbv": true, "hcv": true, "dengue": true, "trop": true, "crp": true}
	for k, v := range c.Labs {
		lowerK := strings.ToLower(k)
		if viralPanel[lowerK] {
			if strings.Contains(v, "+") { // picks up +, ++, +++
				sev := SeverityWarning
				if lowerK == "trop" || lowerK == "hiv" {
					sev = SeverityCritical
				}
				c.AbnormalFlags = append(c.AbnormalFlags, AbnormalFlag{
					Field: strings.ToUpper(k), Value: v, 
					Message: fmt.Sprintf("Positive %s marker", strings.ToUpper(k)), 
					Severity: sev,
				})
			}
		}
	}
}

// checkLabFloat parses a lab value as float64 and calls the checker fn.
func checkLabFloat(c *ClinicalCase, labs map[string]string, key string, fn func(float64) *AbnormalFlag) {
	val, ok := labs[key]
	if !ok {
		return
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return
	}
	if flag := fn(f); flag != nil {
		c.AbnormalFlags = append(c.AbnormalFlags, *flag)
	}
}

// parseBP splits "140/90" → (140, 90). Returns (0,0) on failure.
func parseBP(bp string) (int, int) {
	for i, ch := range bp {
		if ch == '/' {
			sys, err1 := strconv.Atoi(bp[:i])
			dia, err2 := strconv.Atoi(bp[i+1:])
			if err1 != nil || err2 != nil {
				return 0, 0
			}
			return sys, dia
		}
	}
	return 0, 0
}
