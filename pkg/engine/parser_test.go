package engine

import (
	"strings"
	"testing"
)

type testAdditivePlugin struct{}

func (p *testAdditivePlugin) GetName() string { return "testadd" }
func (p *testAdditivePlugin) GetDescription() string { return "test additive plugin" }
func (p *testAdditivePlugin) GetCommandSummary() map[string]string {
	return map[string]string{"testadd.note": "test command"}
}
func (p *testAdditivePlugin) InitData() any { return map[string]string{} }
func (p *testAdditivePlugin) GetCommands() map[string]ParserFunc {
	return map[string]ParserFunc{
		"testadd.note": func(tokens []string, c *ClinicalCase) {
			c.SetExtra("testadd", "note", strings.Join(tokens, " "))
		},
	}
}

type testSecondPlugin struct{}

func (p *testSecondPlugin) GetName() string { return "testsecond" }
func (p *testSecondPlugin) GetDescription() string { return "test second plugin" }
func (p *testSecondPlugin) GetCommandSummary() map[string]string {
	return map[string]string{"testsecond.flag": "second command"}
}
func (p *testSecondPlugin) InitData() any { return map[string]string{"plugin": "second"} }
func (p *testSecondPlugin) GetCommands() map[string]ParserFunc {
	return map[string]ParserFunc{
		"testsecond.flag": func(tokens []string, c *ClinicalCase) {
			c.SetExtra("testsecond", "flag", "on")
		},
	}
}

type testOverridePlugin struct{}

func (p *testOverridePlugin) GetName() string { return "testoverride" }
func (p *testOverridePlugin) GetDescription() string { return "tries overriding core command" }
func (p *testOverridePlugin) GetCommandSummary() map[string]string {
	return map[string]string{"cc": "attempted override"}
}
func (p *testOverridePlugin) InitData() any { return nil }
func (p *testOverridePlugin) GetCommands() map[string]ParserFunc {
	return map[string]ParserFunc{
		"cc": func(tokens []string, c *ClinicalCase) {
			c.CC = "OVERRIDDEN"
		},
	}
}

// TestParseString_Basic is a simple, standard unit test.
// It follows the Arrange, Act, Assert pattern.
func TestParseString_Basic(t *testing.T) {
	// 1. ARRANGE: Set up the scenario
	input := "pt 40M wt60 ht165"

	// 2. ACT: Call the function you want to test
	result := ParseString(input)

	// 3. ASSERT: Verify the outcome is exactly what you expect
	if result.Patient.Age != 40 {
		// t.Errorf logs a failure but continues running the rest of the test
		t.Errorf("Expected Age to be 40, got %d", result.Patient.Age)
	}
	
	if result.Patient.Sex != "M" {
		t.Errorf("Expected Sex to be 'M', got '%s'", result.Patient.Sex)
	}

	if result.Patient.Weight != 60 {
		t.Errorf("Expected weight should be 60, got %f", result.Patient.Weight)
	}

	if result.Patient.Height != 165 {
		t.Errorf("Expected height should be 165, got %f", result.Patient.Height)
	}
}

// TestParseString_TableDriven is the idiomatic "Go way" to write tests.
// It allows you to test dozens of scenarios with very little code repetition.
func TestParseString_TableDriven(t *testing.T) {
	// We define a slice of anonymous structs (our test cases)
	tests := []struct {
		name     string // Name of the test
		input    string // The fake user input
		validate func(t *testing.T, c ClinicalCase) // A custom function to check results
	}{
		{
			name:  "Patient age, sex, weight, and height",
			input: "pt 25F wt60 ht165",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Age != 25 {
					t.Errorf("Expected age 25, got %d", c.Patient.Age)
				}
				if c.Patient.Sex != "F" {
					t.Errorf("Expected sex 'F', got '%s'", c.Patient.Sex)
				}
				if c.Patient.Weight != 60 {
					t.Errorf("Expected weight 60, got %f", c.Patient.Weight)
				}
			},
		},
		{
			name:  "Chief complaint parsing",
			input: "cc chest pain",
			validate: func(t *testing.T, c ClinicalCase) {
				// Note: If you have an abbreviation expansion for 'chest pain', 
				// you would test for the expanded version here!
				if c.CC == "" {
					t.Errorf("Expected CC to be populated, got empty string")
				}
			},
		},
	}

	// Loop over all the test cases
	for _, tc := range tests {
		// t.Run creates a sub-test, so if one fails, you know exactly which one.
		t.Run(tc.name, func(t *testing.T) {
			result := ParseString(tc.input)
			
			// Call the custom validation function for this specific test case
			tc.validate(t, result)
		})
	}
}

// TestPtCommand covers every feature of the `pt` command and patient_parser.go.
// Each test case targets one specific behaviour so failures are easy to pinpoint.
func TestPtCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, c ClinicalCase)
	}{

		// ── Age (plain digits) ──────────────────────────────────────────────────

		{
			// "34" → age 34, unit defaults to "Y" (years)
			name:  "Plain age defaults to years",
			input: "pt 34M",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Age != 34 {
					t.Errorf("age: want 34, got %d", c.Patient.Age)
				}
				if c.Patient.AgeUnit != "Y" {
					t.Errorf("age unit: want 'Y', got '%s'", c.Patient.AgeUnit)
				}
			},
		},

		// ── Sex ─────────────────────────────────────────────────────────────────

		{
			// Standalone "M" token (separate from age)
			name:  "Sex male standalone token",
			input: "pt 30 M",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Sex != "M" {
					t.Errorf("sex: want 'M', got '%s'", c.Patient.Sex)
				}
			},
		},
		{
			name:  "Sex female standalone token",
			input: "pt 25 F",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Sex != "F" {
					t.Errorf("sex: want 'F', got '%s'", c.Patient.Sex)
				}
			},
		},

		// ── Age + Sex combined in one token ─────────────────────────────────────

		{
			// "45M" → age 45 years, sex Male
			name:  "Age and sex combined: 45M",
			input: "pt 45M",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Age != 45 {
					t.Errorf("age: want 45, got %d", c.Patient.Age)
				}
				if c.Patient.Sex != "M" {
					t.Errorf("sex: want 'M', got '%s'", c.Patient.Sex)
				}
			},
		},
		{
			// "6moF" → age 6 months, sex Female
			name:  "Age with unit and sex combined: 6moF",
			input: "pt 6moF",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Age != 6 {
					t.Errorf("age: want 6, got %d", c.Patient.Age)
				}
				if c.Patient.AgeUnit != "mo" {
					t.Errorf("age unit: want 'mo', got '%s'", c.Patient.AgeUnit)
				}
				if c.Patient.Sex != "F" {
					t.Errorf("sex: want 'F', got '%s'", c.Patient.Sex)
				}
			},
		},

		// ── Age Units ────────────────────────────────────────────────────────────

		{
			// "30min" → newborn/neonate age in minutes
			name:  "Age in minutes",
			input: "pt 30minM",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Age != 30 {
					t.Errorf("age: want 30, got %d", c.Patient.Age)
				}
				if c.Patient.AgeUnit != "min" {
					t.Errorf("age unit: want 'min', got '%s'", c.Patient.AgeUnit)
				}
			},
		},
		{
			// "3d" → neonate age in days
			name:  "Age in days",
			input: "pt 3dM",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Age != 3 {
					t.Errorf("age: want 3, got %d", c.Patient.Age)
				}
				if c.Patient.AgeUnit != "d" {
					t.Errorf("age unit: want 'd', got '%s'", c.Patient.AgeUnit)
				}
			},
		},
		{
			// "2w" → infant age in weeks
			name:  "Age in weeks",
			input: "pt 2wM",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Age != 2 {
					t.Errorf("age: want 2, got %d", c.Patient.Age)
				}
				if c.Patient.AgeUnit != "w" {
					t.Errorf("age unit: want 'w', got '%s'", c.Patient.AgeUnit)
				}
			},
		},
		{
			// "8mo" → infant age in months
			name:  "Age in months",
			input: "pt 8moM",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Age != 8 {
					t.Errorf("age: want 8, got %d", c.Patient.Age)
				}
				if c.Patient.AgeUnit != "mo" {
					t.Errorf("age unit: want 'mo', got '%s'", c.Patient.AgeUnit)
				}
			},
		},

		// ── Weight ───────────────────────────────────────────────────────────────

		{
			name:  "Weight without suffix",
			input: "pt 30M wt75",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Weight != 75 {
					t.Errorf("weight: want 75, got %f", c.Patient.Weight)
				}
			},
		},
		{
			// Weight with explicit kg suffix — parser strips it, result is still numeric
			name:  "Weight with kg suffix",
			input: "pt 45M wt82kg",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Weight != 82 {
					t.Errorf("weight: want 82, got %f", c.Patient.Weight)
				}
			},
		},
		{
			// Decimal weight (e.g. for paediatrics or precise dosing)
			name:  "Decimal weight",
			input: "pt 2wM wt3.5",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Weight != 3.5 {
					t.Errorf("weight: want 3.5, got %f", c.Patient.Weight)
				}
			},
		},

		// ── Height ───────────────────────────────────────────────────────────────

		{
			name:  "Height without suffix",
			input: "pt 30M ht172",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Height != 172 {
					t.Errorf("height: want 172, got %f", c.Patient.Height)
				}
			},
		},
		{
			// Height with explicit cm suffix — parser strips it
			name:  "Height with cm suffix",
			input: "pt 30M ht165cm",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Height != 165 {
					t.Errorf("height: want 165, got %f", c.Patient.Height)
				}
			},
		},
		{
			name:  "Decimal height",
			input: "pt 30M ht172.5",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Height != 172.5 {
					t.Errorf("height: want 172.5, got %f", c.Patient.Height)
				}
			},
		},

		// ── BMI and BSA auto-calculation ─────────────────────────────────────────

		{
			// BMI = wt / (ht_m)^2 → 70 / (1.70)^2 ≈ 24.22
			// BSA (Mosteller) = sqrt(ht * wt / 3600) → sqrt(170*70/3600) ≈ 1.82
			name:  "BMI and BSA auto-calculated when both wt and ht present",
			input: "pt 30M wt70 ht170",
			validate: func(t *testing.T, c ClinicalCase) {
				expectedBMI := 70.0 / (1.70 * 1.70)
				if c.Patient.BMI < expectedBMI-0.1 || c.Patient.BMI > expectedBMI+0.1 {
					t.Errorf("BMI: want ~%.2f, got %.2f", expectedBMI, c.Patient.BMI)
				}
				if c.Patient.BSA <= 0 {
					t.Errorf("BSA should be calculated, got %.2f", c.Patient.BSA)
				}
			},
		},
		{
			// When height is missing, BMI and BSA must NOT be calculated (stay 0)
			name:  "BMI not calculated when height missing",
			input: "pt 30M wt70",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.BMI != 0 {
					t.Errorf("BMI should be 0 when height missing, got %.2f", c.Patient.BMI)
				}
			},
		},

		// ── Bed number ───────────────────────────────────────────────────────────

		{
			name:  "Bed number parsed",
			input: "pt 45M bed7",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Bed != 7 {
					t.Errorf("bed: want 7, got %d", c.Patient.Bed)
				}
			},
		},

		// ── Hospital unit ────────────────────────────────────────────────────────
		{
			name:  "Unit name parsed",
			input: "pt 45M unit:ICU",
			validate: func(t *testing.T, c ClinicalCase) {
				if c.Patient.Unit != "ICU" {
					t.Errorf("unit: want 'ICU', got '%s'", c.Patient.Unit)
				}
			},
		},

		// ── Warning cases ────────────────────────────────────────────────────────

		{
			// Omitting sex should produce a warning
			name:  "Missing sex produces warning",
			input: "pt 40",
			validate: func(t *testing.T, c ClinicalCase) {
				found := false
				for _, w := range c.Warnings {
					if w == "Patient sex not specified" {
						found = true
					}
				}
				if !found {
					t.Errorf("expected 'sex not specified' warning, got: %v", c.Warnings)
				}
			},
		},
		{
			// Omitting age entirely should produce a warning
			name:  "Missing age produces warning",
			input: "pt M",
			validate: func(t *testing.T, c ClinicalCase) {
				found := false
				for _, w := range c.Warnings {
					if w == "Patient age not specified" {
						found = true
					}
				}
				if !found {
					t.Errorf("expected 'age not specified' warning, got: %v", c.Warnings)
				}
			},
		},
		{
			// A garbage token must NOT crash the parser — just warn and continue
			name:  "Unrecognized token does not crash, adds warning",
			input: "pt 30M xyz??",
			validate: func(t *testing.T, c ClinicalCase) {
				// Valid tokens must still parse correctly
				if c.Patient.Age != 30 {
					t.Errorf("age: want 30, got %d", c.Patient.Age)
				}
				// The bad token should have triggered a warning
				found := false
				for _, w := range c.Warnings {
					if w == "Unrecognized patient token: xyz??" {
						found = true
					}
				}
				if !found {
					t.Errorf("expected unrecognized token warning, got: %v", c.Warnings)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ParseString(tc.input)
			tc.validate(t, result)
		})
	}
}

func TestParseString_MultiProfileAdditive(t *testing.T) {
	original := pluginsRegistry
	pluginsRegistry = map[string]SpecialtyPlugin{
		"testadd":    &testAdditivePlugin{},
		"testsecond": &testSecondPlugin{},
	}
	t.Cleanup(func() { pluginsRegistry = original })

	input := "@testsecond+testadd\ntestadd.note hello world\ntestsecond.flag"
	result := ParseString(input)

	if result.Profile != "testadd+testsecond" {
		t.Fatalf("profile: want testadd+testsecond, got %q", result.Profile)
	}

	dataMap, ok := result.SpecialtyData.(map[string]any)
	if !ok {
		t.Fatalf("expected specialty_data map for multi profile, got %T", result.SpecialtyData)
	}
	if _, ok := dataMap["testadd"]; !ok {
		t.Fatalf("expected testadd data in specialty_data")
	}
	if _, ok := dataMap["testsecond"]; !ok {
		t.Fatalf("expected testsecond data in specialty_data")
	}

	if got := result.Extra["testadd"]["note"]; got != "hello world" {
		t.Fatalf("testadd.note not parsed, got %q", got)
	}
	if got := result.Extra["testsecond"]["flag"]; got != "on" {
		t.Fatalf("testsecond.flag not parsed, got %q", got)
	}
}

func TestParseString_PluginCannotOverrideCoreCommand(t *testing.T) {
	original := pluginsRegistry
	pluginsRegistry = map[string]SpecialtyPlugin{
		"testoverride": &testOverridePlugin{},
	}
	t.Cleanup(func() { pluginsRegistry = original })

	result := ParseString("@testoverride\ncc chest pain")

	if result.CC != "chest pain" {
		t.Fatalf("core cc should remain active, got %q", result.CC)
	}

	found := false
	for _, w := range result.Warnings {
		if strings.Contains(w, "additive-only mode") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected additive-only warning, got %v", result.Warnings)
	}
}
