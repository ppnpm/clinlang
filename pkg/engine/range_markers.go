package engine

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// RangeMarker is a neutral, transcription-only annotation that a value
// recorded in a ClinicalCase falls outside a user-configurable reference
// range. It is NOT a clinical interpretation. The clinician decides
// whether and how the marker is meaningful.
//
// See DISCLAIMER.md at the repository root.
type RangeMarker struct {
	Field          string `json:"field"`
	Value          string `json:"value"`
	ReferenceRange string `json:"reference_range"`
	Source         string `json:"source"`
	OutOfRange     bool   `json:"out_of_range"`
}

// Range describes one reference band loaded from reference_ranges.json.
// Low and High are pointers so a one-sided range (only an upper or lower
// bound) is expressible.
type Range struct {
	Low    *float64 `json:"low,omitempty"`
	High   *float64 `json:"high,omitempty"`
	Unit   string   `json:"unit,omitempty"`
	Source string   `json:"source"`
}

// ReferenceRanges is keyed by dotted path, e.g. "vitals.hr" or "labs.hb.M".
type ReferenceRanges map[string]Range

// Contains reports whether v falls inside the (inclusive) range.
// A missing bound is treated as unbounded on that side.
func (r Range) Contains(v float64) bool {
	if r.Low != nil && v < *r.Low {
		return false
	}
	if r.High != nil && v > *r.High {
		return false
	}
	return true
}

// Display returns a short ASCII-friendly representation of the bounds,
// without the unit (the unit is already implied by the marker's Field).
func (r Range) Display() string {
	switch {
	case r.Low != nil && r.High != nil:
		return fmt.Sprintf("%s-%s", trimFloat(*r.Low), trimFloat(*r.High))
	case r.High != nil:
		return "<=" + trimFloat(*r.High)
	case r.Low != nil:
		return ">=" + trimFloat(*r.Low)
	default:
		return ""
	}
}

func trimFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

//go:embed reference_ranges.json
var defaultRangesJSON []byte

var (
	activeRanges   ReferenceRanges
	activeRangesMu sync.RWMutex
)

func init() {
	var r ReferenceRanges
	if err := json.Unmarshal(defaultRangesJSON, &r); err != nil {
		r = ReferenceRanges{}
	}
	activeRanges = r
}

// LoadReferenceRanges replaces the active reference set with the contents
// of the file at path. The file must be JSON matching the same schema as
// the embedded default (see docs/reference-ranges.md).
func LoadReferenceRanges(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var r ReferenceRanges
	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}
	activeRangesMu.Lock()
	defer activeRangesMu.Unlock()
	activeRanges = r
	return nil
}

// DefaultReferenceRanges returns the embedded baseline. Useful for UI
// reset-to-default flows. Returns a copy so callers cannot mutate the
// internal state.
func DefaultReferenceRanges() ReferenceRanges {
	var r ReferenceRanges
	if err := json.Unmarshal(defaultRangesJSON, &r); err != nil {
		return ReferenceRanges{}
	}
	return r
}

func currentRanges() ReferenceRanges {
	activeRangesMu.RLock()
	defer activeRangesMu.RUnlock()
	return activeRanges
}

// CheckRangeMarkers inspects vitals and known lab values and appends
// RangeMarker entries to the case for every value that falls outside
// its configured reference band.
//
// This is a transcription aid, not clinical decision support. It does
// not interpret the values, does not assign severity, and does not
// recommend any action.
func CheckRangeMarkers(c *ClinicalCase) {
	r := currentRanges()
	checkVitalMarkers(c, r)
	checkLabMarkers(c, r)
}

func checkVitalMarkers(c *ClinicalCase, r ReferenceRanges) {
	v := c.Vitals

	if v.BP != "" {
		if sys, dia, ok := parseBP(v.BP); ok {
			sysR, sysOK := r["vitals.bp.systolic"]
			diaR, diaOK := r["vitals.bp.diastolic"]
			sysOut := sysOK && !sysR.Contains(float64(sys))
			diaOut := diaOK && !diaR.Contains(float64(dia))
			if sysOut || diaOut {
				combinedRange := ""
				switch {
				case sysOK && diaOK:
					combinedRange = sysR.Display() + " / " + diaR.Display()
				case sysOK:
					combinedRange = sysR.Display()
				case diaOK:
					combinedRange = diaR.Display()
				}
				source := ""
				if sysOK {
					source = sysR.Source
				} else if diaOK {
					source = diaR.Source
				}
				c.RangeMarkers = append(c.RangeMarkers, RangeMarker{
					Field:          "BP",
					Value:          v.BP,
					ReferenceRange: combinedRange,
					Source:         source,
					OutOfRange:     true,
				})
			}
		}
	}

	if v.HR > 0 {
		appendIfOut(c, r, "vitals.hr", "HR", fmt.Sprintf("%d bpm", v.HR), float64(v.HR))
	}
	if v.SpO2 > 0 {
		appendIfOut(c, r, "vitals.spo2", "SpO2", fmt.Sprintf("%d%%", v.SpO2), float64(v.SpO2))
	}
	if v.Temp > 0 {
		// Temperature unit safety: emit a marker only when the recorded
		// TempUnit (F or C) matches the reference range's configured Unit.
		// The engine never converts between units. If a clinician records
		// in C but the range is configured in F (or vice versa), no marker
		// is emitted — they must align the unit themselves.
		tempUnit := v.TempUnit
		if tempUnit == "" {
			tempUnit = "F"
		}
		if band, ok := r["vitals.temp"]; ok && strings.EqualFold(band.Unit, tempUnit) {
			appendIfOut(c, r, "vitals.temp", "Temp",
				fmt.Sprintf("%.1f %s", v.Temp, tempUnit), v.Temp)
		}
	}
	if v.RR > 0 {
		appendIfOut(c, r, "vitals.rr", "RR", fmt.Sprintf("%d /min", v.RR), float64(v.RR))
	}
}

func checkLabMarkers(c *ClinicalCase, r ReferenceRanges) {
	if len(c.Labs) == 0 {
		return
	}

	checkLabKey(c, r, "hb", hbKeyForSex(c.Patient.Sex), "Hb", "g/dL")
	checkLabKey(c, r, "wbc", "labs.wbc", "WBC", "/uL")
	checkLabKey(c, r, "creatinine", "labs.creatinine", "Creatinine", "mg/dL")
	checkLabKey(c, r, "na", "labs.na", "Na+", "mEq/L")
	checkLabKey(c, r, "k", "labs.k", "K+", "mEq/L")
	checkLabKey(c, r, "glucose", "labs.glucose", "Glucose", "mmol/L")
}

// hbKeyForSex returns the reference-range key for haemoglobin.
// Empty string means no marker should be emitted (sex unknown).
func hbKeyForSex(sex string) string {
	switch sex {
	case "M":
		return "labs.hb.M"
	case "F":
		return "labs.hb.F"
	default:
		return ""
	}
}

// checkLabKey looks up labs[labKey] as a float, then compares against
// the range at rangeKey. labKey is the user-facing lab name in c.Labs;
// rangeKey is the dotted path in the reference set. If rangeKey is "",
// no check is performed (e.g. Hb when sex is unknown).
func checkLabKey(c *ClinicalCase, r ReferenceRanges, labKey, rangeKey, displayField, unit string) {
	if rangeKey == "" {
		return
	}
	raw, ok := c.Labs[labKey]
	if !ok {
		return
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return
	}
	band, ok := r[rangeKey]
	if !ok || band.Contains(f) {
		return
	}
	c.RangeMarkers = append(c.RangeMarkers, RangeMarker{
		Field:          displayField,
		Value:          fmt.Sprintf("%s %s", trimFloat(f), unit),
		ReferenceRange: band.Display(),
		Source:         band.Source,
		OutOfRange:     true,
	})
}

// appendIfOut is a small helper for the integer-valued vitals
// (HR, SpO2, RR, Temp via float). It appends a marker only when the
// range exists in r and the value is outside it.
func appendIfOut(c *ClinicalCase, r ReferenceRanges, rangeKey, displayField, displayValue string, v float64) {
	band, ok := r[rangeKey]
	if !ok || band.Contains(v) {
		return
	}
	c.RangeMarkers = append(c.RangeMarkers, RangeMarker{
		Field:          displayField,
		Value:          displayValue,
		ReferenceRange: band.Display(),
		Source:         band.Source,
		OutOfRange:     true,
	})
}

// parseBP splits "140/90" into (140, 90, true). Returns ok=false on any
// malformed input.
func parseBP(bp string) (int, int, bool) {
	for i, ch := range bp {
		if ch == '/' {
			sys, err1 := strconv.Atoi(bp[:i])
			dia, err2 := strconv.Atoi(bp[i+1:])
			if err1 != nil || err2 != nil {
				return 0, 0, false
			}
			return sys, dia, true
		}
	}
	return 0, 0, false
}
