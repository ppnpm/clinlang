package engine

import "testing"

func TestSymptomParser_DurationWithoutIntensity(t *testing.T) {
	input := "sx cough3d fever++2h"
	result := ParseString(input)

	if len(result.Symptoms) != 2 {
		t.Fatalf("Expected 2 symptoms, got %d", len(result.Symptoms))
	}

	// First symptom: cough3d (no intensity, duration 3d)
	sym1 := result.Symptoms[0]
	if sym1.Name != "cough" {
		t.Errorf("Expected first symptom name to be 'cough', got %q", sym1.Name)
	}
	if sym1.Intensity != "" {
		t.Errorf("Expected first symptom intensity to be empty, got %q", sym1.Intensity)
	}
	if sym1.Duration != "3d" {
		t.Errorf("Expected first symptom duration to be '3d', got %q", sym1.Duration)
	}

	// Second symptom: fever++2h (intensity severe, duration 2h)
	sym2 := result.Symptoms[1]
	if sym2.Name != "fever" {
		t.Errorf("Expected second symptom name to be 'fever', got %q", sym2.Name)
	}
	if sym2.Intensity != "severe" {
		t.Errorf("Expected second symptom intensity to be 'severe', got %q", sym2.Intensity)
	}
	if sym2.Duration != "2h" {
		t.Errorf("Expected second symptom duration to be '2h', got %q", sym2.Duration)
	}
}
