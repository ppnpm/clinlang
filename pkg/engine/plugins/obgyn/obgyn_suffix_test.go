package obgyn_test

// Boundary tests for the obgyn inline-token suffix parsers, exercising
// the regex strictness added in Phase 3.

import (
	"strings"
	"testing"

	"clinlang/pkg/engine"
	_ "clinlang/pkg/engine/plugins/obgyn"
)

func hasGAWarning(c engine.ClinicalCase) bool {
	for _, w := range c.Warnings {
		if strings.Contains(w, "invalid gestational age value") {
			return true
		}
	}
	return false
}

func hasFHRWarning(c engine.ClinicalCase) bool {
	for _, w := range c.Warnings {
		if strings.Contains(w, "invalid fetal heart rate value") {
			return true
		}
	}
	return false
}

func TestObGyn_GASuffix(t *testing.T) {
	tests := []struct {
		token       string
		expectWarn  bool
		expectValue int
	}{
		{"ga:34", false, 34},
		{"ga:34w", false, 34},
		{"ga:34W", false, 34},
		{"ga:34www", true, 0},
		{"ga:34ww", true, 0},
		{"ga:wweek", true, 0},
		{"ga:", true, 0},
	}

	for _, tc := range tests {
		t.Run(tc.token, func(t *testing.T) {
			input := "@profile obgyn\npt 28F " + tc.token
			c := engine.ParseString(input)
			gotWarn := hasGAWarning(c)
			if gotWarn != tc.expectWarn {
				t.Errorf("warning: want=%v got=%v warnings=%v", tc.expectWarn, gotWarn, c.Warnings)
			}
		})
	}
}

func TestObGyn_FHRSuffix(t *testing.T) {
	tests := []struct {
		token      string
		expectWarn bool
	}{
		{"fhr:142", false},
		{"fhr:142bpm", false},
		{"fhr:142BPM", false},
		{"fhr:142Bpm", false},
		{"fhr:145mbp", true},
		{"fhr:145bpmbpm", true},
		{"fhr:", true},
		{"fhr:abc", true},
	}

	for _, tc := range tests {
		t.Run(tc.token, func(t *testing.T) {
			input := "@profile obgyn\npt 28F\nvitals bp120/75 " + tc.token
			c := engine.ParseString(input)
			gotWarn := hasFHRWarning(c)
			if gotWarn != tc.expectWarn {
				t.Errorf("warning: want=%v got=%v warnings=%v", tc.expectWarn, gotWarn, c.Warnings)
			}
		})
	}
}
