package engine

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Vitals holds structured vital signs.
//
// Temperature carries an explicit TempUnit ("F" or "C"). Inputs without
// a unit suffix default to "F". The engine never converts between
// units — out-of-range markers fire only when the recorded TempUnit
// matches the reference range's configured unit.
type Vitals struct {
	BP       string  `json:"BP"`
	HR       int     `json:"HR"`
	SpO2     int     `json:"SpO2"`
	Temp     float64 `json:"Temp"`
	TempUnit string  `json:"temp_unit,omitempty"`
	RR       int     `json:"RR"`
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
	Sex     string  `json:"sex"`       //M/F/O
	Weight  float64 `json:"weight_kg"` //wt68
	Height  float64 `json:"height_cm"` //ht170
	Bed     int     `json:"bed"`       //bed number
	Unit    string  `json:"unit"`      //Unit type
	BMI     float64 `json:"bmi"`       //Body mass index auto-calculated
	BSA     float64 `json:"bsa"`       //body surface area auto-calculated
	BloodGroup	string	`json:"blood_group,omitempty"` //Blood Group
}

type ClinicalFlags struct {
	Pregnant  bool `json:"is_pregnant"`  //Pregnancy Indicator
	Lactation bool `json:"is_lactating"` // Lactation indicator
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
	RangeMarkers  []RangeMarker                `json:"range_markers,omitempty"`  //Out-of-range markers derived from user-configurable reference ranges (transcription aid, not clinical decision support)
	ClinicalFlags ClinicalFlags                `json:"clinical_flags"`           //Clinical flags patient already have such as pregnancy, lactation, etc
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
		RangeMarkers:  []RangeMarker{},
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

// GenerateId creates a unique placeholder ID for a parsed case. Format:
//
//	PT-YYYYMMDD-HHMMSS-XXXXXX
//
// where XXXXXX is six hex characters from crypto/rand. The random suffix
// makes the ID collision-resistant within a single second (the time
// stamp by itself was second-resolution and could collide for cases
// parsed in rapid succession).
//
// Note: this is a non-cryptographic identifier intended only to give
// each parsed case a stable label. It is not a record-of-patient ID.
func GenerateId() string {
	now := time.Now()
	var b [3]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand should not fail on a healthy host. Fall back to
		// the nanosecond timestamp so we still emit a unique-enough ID.
		nano := now.UnixNano()
		return fmt.Sprintf("PT-%04d%02d%02d-%02d%02d%02d-%06x",
			now.Year(), now.Month(), now.Day(),
			now.Hour(), now.Minute(), now.Second(),
			uint32(nano)&0xFFFFFF,
		)
	}
	return fmt.Sprintf("PT-%04d%02d%02d-%02d%02d%02d-%s",
		now.Year(), now.Month(), now.Day(),
		now.Hour(), now.Minute(), now.Second(),
		hex.EncodeToString(b[:]),
	)
}
