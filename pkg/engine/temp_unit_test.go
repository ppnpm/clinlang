package engine

import (
	"strings"
	"testing"
)

func TestVitals_TempUnitDefaultsToF(t *testing.T) {
	c := ParseString("pt 40M\nvitals temp98.6")
	if c.Vitals.Temp != 98.6 {
		t.Errorf("Temp value: want 98.6, got %v", c.Vitals.Temp)
	}
	if c.Vitals.TempUnit != "F" {
		t.Errorf("TempUnit: want 'F' (default), got %q", c.Vitals.TempUnit)
	}
}

func TestVitals_TempUnitExplicitF(t *testing.T) {
	c := ParseString("pt 40M\nvitals temp98.6F")
	if c.Vitals.Temp != 98.6 {
		t.Errorf("Temp value: want 98.6, got %v", c.Vitals.Temp)
	}
	if c.Vitals.TempUnit != "F" {
		t.Errorf("TempUnit: want 'F', got %q", c.Vitals.TempUnit)
	}
}

func TestVitals_TempUnitExplicitC(t *testing.T) {
	c := ParseString("pt 40M\nvitals temp37C")
	if c.Vitals.Temp != 37 {
		t.Errorf("Temp value: want 37, got %v", c.Vitals.Temp)
	}
	if c.Vitals.TempUnit != "C" {
		t.Errorf("TempUnit: want 'C', got %q", c.Vitals.TempUnit)
	}
}

func TestRangeMarkers_TempFiresOnUnitMatch(t *testing.T) {
	// Default range is F (97-100). Input 101F is out of range → marker.
	c := ParseString("pt 40M\nvitals temp101F")
	m := findMarker(c.RangeMarkers, "Temp")
	if m == nil {
		t.Fatalf("expected Temp marker for 101F, got %+v", c.RangeMarkers)
	}
	if !strings.Contains(m.Value, "F") {
		t.Errorf("Value should carry unit F, got %q", m.Value)
	}
}

func TestRangeMarkers_TempSuppressedOnUnitMismatch(t *testing.T) {
	// Default range is F. Input 37C — units don't match the range's
	// configured unit. Engine refuses to convert and emits no marker.
	c := ParseString("pt 40M\nvitals temp37C")
	if m := findMarker(c.RangeMarkers, "Temp"); m != nil {
		t.Errorf("expected NO Temp marker when units differ from range (got %+v)", m)
	}
}

func TestRangeMarkers_TempInRangeF(t *testing.T) {
	// 98.6F is inside the F range (97-100) → no marker.
	c := ParseString("pt 40M\nvitals temp98.6F")
	if m := findMarker(c.RangeMarkers, "Temp"); m != nil {
		t.Errorf("expected no Temp marker for 98.6F (in range), got %+v", m)
	}
}

func TestVitals_FormatVitalsCarriesUnit(t *testing.T) {
	c := ParseString("pt 40M\nvitals temp37C")
	out := FormatVitals(c.Vitals)
	if !strings.Contains(out, "37.0 C") {
		t.Errorf("FormatVitals output should contain '37.0 C', got %q", out)
	}
}
