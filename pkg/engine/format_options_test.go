package engine

import (
	"strings"
	"testing"
)

func TestFormatSOAP_MarkersOffByDefault(t *testing.T) {
	c := ParseString("pt 40M\nvitals hr150 bp180/110")
	if len(c.RangeMarkers) == 0 {
		t.Fatalf("setup: expected RangeMarkers populated, got 0")
	}

	out := FormatSOAP(c)
	if strings.Contains(out, "Notes (out of ref)") {
		t.Errorf("default FormatSOAP must not include 'Notes (out of ref)' block; got:\n%s", out)
	}
	if strings.Contains(out, "outside ref") {
		t.Errorf("default FormatSOAP must not include 'outside ref' text; got:\n%s", out)
	}
}

func TestFormatSOAP_MarkersOptIn(t *testing.T) {
	c := ParseString("pt 40M\nvitals hr150 bp180/110")
	out := FormatSOAPWithOptions(c, FormatOptions{ShowRangeMarkers: true})

	if !strings.Contains(out, "Notes (out of ref)") {
		t.Errorf("opt-in FormatSOAP must include 'Notes (out of ref)' block; got:\n%s", out)
	}
	if !strings.Contains(out, "outside ref 60-100") {
		t.Errorf("opt-in FormatSOAP must include HR ref range; got:\n%s", out)
	}
	if strings.Contains(out, "critical") || strings.Contains(out, "Critical") {
		t.Errorf("opt-in FormatSOAP must not use severity wording; got:\n%s", out)
	}
}

func TestFormatMarkdown_MarkersOffByDefault(t *testing.T) {
	c := ParseString("pt 40M\nvitals hr150 bp180/110")
	out := FormatMarkdown(c)
	if strings.Contains(out, "Notes (out of ref)") {
		t.Errorf("default FormatMarkdown must not include marker section; got:\n%s", out)
	}
	if strings.Contains(out, "Clinical Alerts") {
		t.Errorf("'Clinical Alerts' section must be removed; got:\n%s", out)
	}
}

func TestFormatMarkdown_MarkersOptIn(t *testing.T) {
	c := ParseString("pt 40M\nvitals hr150")
	out := FormatMarkdownWithOptions(c, FormatOptions{ShowRangeMarkers: true})
	if !strings.Contains(out, "Notes (out of ref)") {
		t.Errorf("opt-in FormatMarkdown must include marker section; got:\n%s", out)
	}
}
