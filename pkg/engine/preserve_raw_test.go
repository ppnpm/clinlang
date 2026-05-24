package engine

import (
	"strings"
	"testing"
)

// TestParseString_PreservesRawText pins down the Phase 3 decision that
// the parser stores user-typed text verbatim and abbreviation expansion
// happens only at format time.
//
// Why this matters: the JSON layer is the data layer. If `dx dm2` were
// silently rewritten to "Type 2 Diabetes Mellitus" at parse time, the
// software would be the one supplying the diagnostic term — that is a
// medicolegal posture we don't want. The clinician supplies the term;
// the formatter expands it only for human-friendly display.
func TestParseString_PreservesRawText(t *testing.T) {
	input := "pt 40M\npmh dm2 htn\ndx dm2"
	c := ParseString(input)

	if !strings.Contains(c.PMH, "dm2") {
		t.Errorf("PMH should preserve raw 'dm2', got %q", c.PMH)
	}
	if strings.Contains(c.PMH, "Type 2 Diabetes Mellitus") {
		t.Errorf("PMH should NOT be expanded at parse time, got %q", c.PMH)
	}
	if c.DX != "dm2" {
		t.Errorf("DX should be raw 'dm2', got %q", c.DX)
	}
}

// TestFormatSOAP_ExpandsAtFormatTime confirms the other side: when we
// render for display, the abbreviations expand.
func TestFormatSOAP_ExpandsAtFormatTime(t *testing.T) {
	c := ParseString("pt 40M\npmh dm2 htn")
	out := FormatSOAP(c)
	if !strings.Contains(out, "Type 2 Diabetes Mellitus") {
		t.Errorf("SOAP output should contain expanded form, got:\n%s", out)
	}
}

// TestFormatJSON_PreservesRawText confirms the JSON serialisation reflects
// what the user actually typed.
func TestFormatJSON_PreservesRawText(t *testing.T) {
	c := ParseString("pt 40M\ndx dm2")
	out := FormatJSON(c)
	if !strings.Contains(out, "\"dx\": \"dm2\"") {
		t.Errorf("JSON should contain raw \"dx\": \"dm2\", got:\n%s", out)
	}
}
