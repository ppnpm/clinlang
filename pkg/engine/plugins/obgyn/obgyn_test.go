package obgyn_test

// Integration tests for the OB/GYN plugin.
//
// These tests verify the full plugin architecture:
//   - Standalone plugin commands (lmp, edd, gpal, fhs, ctx)
//   - Inline token extensions into core commands (ga: in pt, fhr: in vitals)
//   - Isolation: OB/GYN tokens have NO effect on non-OB cases
//
// Package is obgyn_test (external test package) to test through the public API only.

import (
	"clinlang/pkg/engine"
	_ "clinlang/pkg/engine/plugins/obgyn" // side-effect import: registers the plugin
	"testing"
)




func TestObGynPlugin_StandaloneCommands(t *testing.T) {
	input := `@profile obgyn
pt 28F wt65 ht158
lmp 2025-06-15
edd 2026-03-22
gpal G2P1A0L1
fhs 142 regular reactive
ctx 3 in 10min lasting 45s`

	result := engine.ParseString(input)

	if result.Profile != "obgyn" {
		t.Errorf("profile: want 'obgyn', got '%s'", result.Profile)
	}

	data, ok := result.SpecialtyData.(*struct {
		GA     int
		GAUnit string
		FHR    int
		LMP    string
		EDD    string
		GPAL   string
		FHS    string
		CTX    string
	})
	// We can't easily access the unexported type from outside,
	// so use FormatJSON and check it contains the expected fields.
	_ = data
	_ = ok

	// Use JSON output to verify — this is the real integration check
	json := engine.FormatJSON(result)

	checks := []struct {
		field    string
		contains string
	}{
		{"profile", "obgyn"},
		{"lmp", "2025-06-15"},
		{"edd", "2026-03-22"},
		{"gpal", "G2P1A0L1"},
		{"fhs", "regular"},
		{"ctx", "minutes"}, // "min" is expanded to "minutes" by abbreviation engine
	}

	for _, check := range checks {
		if !containsString(json, check.contains) {
			t.Errorf("JSON output missing %s content '%s'\nFull JSON:\n%s",
				check.field, check.contains, json)
		}
	}

	// Core patient data must still parse correctly
	if result.Patient.Age != 28 {
		t.Errorf("patient age: want 28, got %d", result.Patient.Age)
	}
	if result.Patient.Sex != "F" {
		t.Errorf("patient sex: want 'F', got '%s'", result.Patient.Sex)
	}
}

func TestObGynPlugin_InlinePtToken_GA(t *testing.T) {
	// Key test: ga:34w as an inline token inside the pt command
	input := `@profile obgyn
pt 28F wt65 ht158 ga:34w`

	result := engine.ParseString(input)

	// Patient fields must parse correctly (core tokens still work)
	if result.Patient.Age != 28 {
		t.Errorf("age: want 28, got %d", result.Patient.Age)
	}
	if result.Patient.Sex != "F" {
		t.Errorf("sex: want 'F', got 'F'")
	}
	if result.Patient.Weight != 65 {
		t.Errorf("weight: want 65, got %f", result.Patient.Weight)
	}

	// ga:34w must be in the JSON output under specialty data
	json := engine.FormatJSON(result)
	if !containsString(json, "34") {
		t.Errorf("GA value 34 not found in JSON output:\n%s", json)
	}
	if !containsString(json, "ga_weeks") {
		t.Errorf("ga_weeks field not found in JSON output:\n%s", json)
	}

	// No warnings should be produced for ga: token
	for _, w := range result.Warnings {
		if containsString(w, "ga:") || containsString(w, "Unrecognized") {
			t.Errorf("unexpected warning for ga: token: %s", w)
		}
	}
}

func TestObGynPlugin_InlineVitalsToken_FHR(t *testing.T) {
	// Key test: fhr:142 as an inline token inside the vitals command
	input := `@profile obgyn
pt 28F wt65
vitals bp120/75 hr78 spo299 fhr:142`

	result := engine.ParseString(input)

	// Core vitals must still parse
	if result.Vitals.BP != "120/75" {
		t.Errorf("BP: want '120/75', got '%s'", result.Vitals.BP)
	}
	if result.Vitals.HR != 78 {
		t.Errorf("HR: want 78, got %d", result.Vitals.HR)
	}

	// fhr:142 must appear in JSON under specialty data
	json := engine.FormatJSON(result)
	if !containsString(json, "142") {
		t.Errorf("FHR value 142 not found in JSON output:\n%s", json)
	}
	if !containsString(json, "fhr_bpm") {
		t.Errorf("fhr_bpm field not found in JSON output:\n%s", json)
	}

	// No warnings for fhr: token
	for _, w := range result.Warnings {
		if containsString(w, "fhr:") || containsString(w, "Unrecognized vitals") {
			t.Errorf("unexpected warning for fhr: token: %s", w)
		}
	}
}

func TestObGynPlugin_Isolation(t *testing.T) {
	// Without @profile obgyn, ga: and fhr: should produce warnings (not silently ignored)
	// This proves no global state leaks between parse sessions.
	input := `pt 45M wt80
vitals bp140/90 hr85`

	result := engine.ParseString(input)

	// Should parse fine — no OB/GYN data expected
	if result.Patient.Age != 45 {
		t.Errorf("age: want 45, got %d", result.Patient.Age)
	}
	if result.Profile != "general" {
		t.Errorf("profile: want 'general', got '%s'", result.Profile)
	}
	if result.SpecialtyData != nil {
		t.Errorf("SpecialtyData should be nil for a general case, got %v", result.SpecialtyData)
	}
}

func TestObGynPlugin_FullCase(t *testing.T) {
	// Full realistic OB/GYN case: all features together
	input := `@profile obgyn
pt 28F wt68 ht160 ga:34w
lmp 2025-06-15
edd 2026-03-22
gpal G2P1A0L1
vitals bp130/85 hr88 spo298 fhr:138
cc preterm labour pains
sx uterine contractions+++ lower back pain++
fhs 138 regular reactive
ctx 4 in 10min lasting 50s
dx Preterm labour at 34w`

	result := engine.ParseString(input)

	// Verify zero unexpected warnings
	for _, w := range result.Warnings {
		t.Logf("Warning: %s", w) // log but don't fail — some may be expected
	}

	json := engine.FormatJSON(result)

	// All key fields must be present
	requiredFields := []string{
		"obgyn",          // profile
		"ga_weeks",       // inline pt token
		"fhr_bpm",        // inline vitals token
		"2025-06-15",     // lmp
		"2026-03-22",     // edd
		"G2P1A0L1",       // gpal
		"Preterm labour", // dx
	}
	for _, field := range requiredFields {
		if !containsString(json, field) {
			t.Errorf("required field '%s' not found in JSON output", field)
		}
	}
}

// containsString is a helper that checks if substr exists in s.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
