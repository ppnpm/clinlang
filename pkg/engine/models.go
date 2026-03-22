package engine

import (
	"fmt"
	"time"
)

// Vitals holds structured vital signs.
type Vitals struct {
	BP   string `json:"BP"`
	HR   int    `json:"HR"`
	SpO2 int    `json:"SpO2"`
	Temp string `json:"Temp"`
	RR   int    `json:"RR"`
}

// Symptom holds a single symptom with optional intensity and duration.
type Symptom struct {
	Name      string `json:"name"`
	Intensity string `json:"intensity"`
	Duration  string `json:"duration"`
}

// Patient holds demographic information.
type Patient struct {
	Id      string `json:"id"`
	Age     int    `json:"age"`
	AgeUnit string `json:"age_unit,omitempty"`
	Sex     string `json:"sex"`
	Weight  float64 `json:"weight_kg"`
	Height  float64 `json:"height_cm"`
	GPAL    string `json:"gpal,omitempty"`
	MOA     string `json:"moa,omitempty"`
}

// Case is the root structure representing a full clinical encounter.
type Case struct {
	Profile       string                       `json:"profile,omitempty"`
	Patient       Patient                      `json:"patient"`
	CC            string                       `json:"cc,omitempty"`
	HPI           string                       `json:"hpi,omitempty"`
	PMH           string                       `json:"pmh,omitempty"`
	DX            string                       `json:"dx,omitempty"`
	DDX           string                       `json:"ddx,omitempty"`
	Vitals        Vitals                       `json:"vitals"`
	Symptoms      []Symptom                    `json:"symptoms"`
	Prescriptions []Prescription               `json:"prescriptions,omitempty"`
	Extra         map[string]map[string]string `json:"extra,omitempty"`
	SpecialtyData any                  `json:"specialty_data,omitempty"`
	AbnormalFlags []AbnormalFlag               `json:"abnormal_flags,omitempty"`
	Warnings      []string                     `json:"warnings,omitempty"`
}

// NewCase creates a properly initialized Case.
func NewCase() Case {
	return Case{
		Extra:         make(map[string]map[string]string),
		Symptoms:      []Symptom{},
		Prescriptions: []Prescription{},
		AbnormalFlags: []AbnormalFlag{},
		Warnings:      []string{},
	}
}

// AddWarning appends a non-fatal warning message.
func (c *Case) AddWarning(msg string) {
	c.Warnings = append(c.Warnings, msg)
}

// SetExtra stores a key-value pair under an extension command namespace.
func (c *Case) SetExtra(cmd, key, value string) {
	if c.Extra[cmd] == nil {
		c.Extra[cmd] = make(map[string]string)
	}
	c.Extra[cmd][key] = value
}

// GenerateId creates a timestamp-based unique patient ID.
func GenerateId() string {
	now := time.Now()
	return fmt.Sprintf("PT-%04d%02d%02d-%02d%02d%02d",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
	)
}
