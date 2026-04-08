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
	Frequency int    `json:"frequency"`
}

// Patient holds demographic information.
type Patient struct {
	Id      string  `json:"id"`
	Age     int     `json:"age"`
	AgeUnit string  `json:"age_unit,omitempty"`
	Sex     string  `json:"sex"`            //M/F/O
	Weight  float64 `json:"weight_kg"`      //wt68
	Height  float64 `json:"height_cm"`      //ht170
	GPAL    string  `json:"gpal,omitempty"` //Gravida, Para, Abortus, Living
	MOA     string  `json:"moa,omitempty"`  //months of amenorrhoea
}

// ClinicalCase is the root structure representing a full clinical summary.

type ClinicalCase struct {
	Profile       string                       `json:"profile,omitempty"`        //ClinicalCase profile
	Patient       Patient                      `json:"patient"`                  //Patient Information
	Day           string                       `json:"day,omitempty"`            //Rounding day/event timeline
	Allergies     string                       `json:"allergies,omitempty"`      //Allergies
	CC            string                       `json:"cc,omitempty"`             //Chief complaint
	HPI           string                       `json:"hpi,omitempty"`            //History of presenting illness
	PMH           string                       `json:"pmh,omitempty"`            //Past medical history
	SH            string                       `json:"sh,omitempty"`             //Social history
	FH            string                       `json:"fh,omitempty"`             //Family history
	PE            string                       `json:"pe,omitempty"`             //Physical/Objective exam
	DX            string                       `json:"dx,omitempty"`             //Provisional Diagnosis
	DDX           string                       `json:"ddx,omitempty"`            //Differential Diagnosis
	Vitals        Vitals                       `json:"vitals"`                   //Vital signs
	Symptoms      []Symptom                    `json:"symptoms"`                 //Symptoms slice of Symtoms struct
	Labs          map[string]string            `json:"labs,omitempty"`           //Structured labs dictionary
	Imaging       map[string]string            `json:"imaging,omitempty"`        //Structured radiology/imaging
	Prescriptions []Prescription               `json:"prescriptions,omitempty"`  //Prescriptions
	Extra         map[string]map[string]string `json:"extra,omitempty"`          //Extra information
	SpecialtyData any                          `json:"specialty_data,omitempty"` //Specialty data
	AbnormalFlags []AbnormalFlag               `json:"abnormal_flags,omitempty"` //Abnormal flags
	Warnings      []string                     `json:"warnings,omitempty"`       //Warnings
}

// NewClinicalCase creates a properly initialized ClinicalCase.
func NewClinicalCase() ClinicalCase {
	return ClinicalCase{
		Extra:         make(map[string]map[string]string),
		Labs:          make(map[string]string),
		Imaging:       make(map[string]string),
		Symptoms:      []Symptom{},
		Prescriptions: []Prescription{},
		AbnormalFlags: []AbnormalFlag{},
		Warnings:      []string{},
	}
}



// AddWarning is a method that appends a non-fatal warning message to the 'Warnings' slice array.
func (c *ClinicalCase) AddWarning(msg string) {
	c.Warnings = append(c.Warnings, msg)
}

// SetExtra stores a key-value pair under an extension command namespace. Here Final value could be anything such as int, float, string, bool etc hence added any
func (c *ClinicalCase) SetExtra(cmd, key, value string) {
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
