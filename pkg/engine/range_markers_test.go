package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// findMarker returns the first RangeMarker matching field, or nil.
func findMarker(markers []RangeMarker, field string) *RangeMarker {
	for i := range markers {
		if markers[i].Field == field {
			return &markers[i]
		}
	}
	return nil
}

func TestRangeMarkers_HROutOfRange(t *testing.T) {
	c := ParseString("pt 40M\nvitals hr120")
	m := findMarker(c.RangeMarkers, "HR")
	if m == nil {
		t.Fatalf("expected HR marker, got %+v", c.RangeMarkers)
	}
	if !m.OutOfRange {
		t.Errorf("expected OutOfRange=true, got false")
	}
	if m.ReferenceRange != "60-100" {
		t.Errorf("ReferenceRange: want '60-100', got %q", m.ReferenceRange)
	}
	if m.Source == "" {
		t.Errorf("Source must be populated")
	}
	if !strings.Contains(m.Value, "120") {
		t.Errorf("Value should contain '120', got %q", m.Value)
	}
}

func TestRangeMarkers_HRInRangeNoMarker(t *testing.T) {
	c := ParseString("pt 40M\nvitals hr75")
	if m := findMarker(c.RangeMarkers, "HR"); m != nil {
		t.Errorf("expected no HR marker for in-range value, got %+v", m)
	}
}

func TestRangeMarkers_HbSexKeyed(t *testing.T) {
	// 12 g/dL — below male threshold (13) but at/above female threshold (11).
	cM := ParseString("pt 40M\nlab hb12.0")
	if m := findMarker(cM.RangeMarkers, "Hb"); m == nil {
		t.Errorf("male: expected Hb marker for 12.0 g/dL, got %+v", cM.RangeMarkers)
	}

	cF := ParseString("pt 40F\nlab hb12.0")
	if m := findMarker(cF.RangeMarkers, "Hb"); m != nil {
		t.Errorf("female: expected NO Hb marker for 12.0 g/dL, got %+v", m)
	}
}

func TestRangeMarkers_HbNoSexNoMarker(t *testing.T) {
	c := ParseString("lab hb7.0")
	if m := findMarker(c.RangeMarkers, "Hb"); m != nil {
		t.Errorf("expected no Hb marker when sex unknown, got %+v", m)
	}
}

func TestRangeMarkers_BPCombined(t *testing.T) {
	c := ParseString("pt 40M\nvitals bp180/110")
	m := findMarker(c.RangeMarkers, "BP")
	if m == nil {
		t.Fatalf("expected BP marker, got %+v", c.RangeMarkers)
	}
	if m.Value != "180/110" {
		t.Errorf("Value: want '180/110', got %q", m.Value)
	}
	if m.ReferenceRange != "90-140 / 60-90" {
		t.Errorf("ReferenceRange: want '90-140 / 60-90', got %q", m.ReferenceRange)
	}
}

func TestRangeMarkers_NoSerologyInterpretation(t *testing.T) {
	// Phase 2 deliberately removes checkSerology — a positive
	// transcribed lab result must NOT produce a clinical marker.
	c := ParseString("pt 40M\nix trop+")
	for _, m := range c.RangeMarkers {
		if strings.EqualFold(m.Field, "TROP") {
			t.Errorf("checkSerology should be gone but emitted TROP marker: %+v", m)
		}
	}
}

func TestRangeMarkers_LoadOverride(t *testing.T) {
	// Reset to defaults at the end so other tests are unaffected.
	t.Cleanup(func() {
		activeRangesMu.Lock()
		activeRanges = DefaultReferenceRanges()
		activeRangesMu.Unlock()
	})

	dir := t.TempDir()
	overridePath := filepath.Join(dir, "ranges.json")
	override := `{
		"vitals.hr": {"low": 70, "high": 90, "unit": "bpm", "source": "custom"}
	}`
	if err := os.WriteFile(overridePath, []byte(override), 0644); err != nil {
		t.Fatalf("write override: %v", err)
	}

	if err := LoadReferenceRanges(overridePath); err != nil {
		t.Fatalf("LoadReferenceRanges: %v", err)
	}

	// HR 95 — out of override (>90) but in default (60-100).
	c := ParseString("pt 40M\nvitals hr95")
	m := findMarker(c.RangeMarkers, "HR")
	if m == nil {
		t.Fatalf("expected HR marker after override, got %+v", c.RangeMarkers)
	}
	if m.ReferenceRange != "70-90" {
		t.Errorf("ReferenceRange: want '70-90' from override, got %q", m.ReferenceRange)
	}
	if m.Source != "custom" {
		t.Errorf("Source: want 'custom', got %q", m.Source)
	}
}

func TestRange_OneSidedBounds(t *testing.T) {
	high := 5.0
	r := Range{High: &high, Source: "test"}
	if !r.Contains(4.9) {
		t.Errorf("4.9 should be in range with only high bound")
	}
	if r.Contains(5.1) {
		t.Errorf("5.1 should be out of range with high=5")
	}
	if got := r.Display(); got != "<=5" {
		t.Errorf("Display: want '<=5', got %q", got)
	}

	low := 7.0
	r2 := Range{Low: &low, Source: "test"}
	if got := r2.Display(); got != ">=7" {
		t.Errorf("Display: want '>=7', got %q", got)
	}
}
