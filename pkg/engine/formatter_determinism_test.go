package engine

import (
	"regexp"
	"strings"
	"testing"
)

// TestParseString_JSONByteIdentical50x asserts that parsing the same
// input N times produces byte-identical JSON output. This is the
// catch-net for any future map-iteration or ordering bug.
//
// The Patient.Id field is generated with a random suffix per parse, so
// we strip it from both sides before comparing.
func TestParseString_JSONByteIdentical50x(t *testing.T) {
	input := `pt 40M wt70 ht172
cc fever cough
hpi onset 3 days, productive cough
pmh dm2 htn
sx fever++ cough+++
vitals bp130/85 hr96 spo295 rr18
ix hb12 wbc8000 na138 k4.1 cr1.0
ix cxr:clear ecg:nsr
pe chest:clear cvs:normal
obs ga34w
rx amoxicillin 500mg tds po
dx URI`

	idRe := regexp.MustCompile(`"id":\s*"[^"]*"`)
	normalize := func(s string) string {
		return idRe.ReplaceAllString(s, `"id":"<stripped>"`)
	}

	first := normalize(FormatJSON(ParseString(input)))
	for i := 1; i < 50; i++ {
		got := normalize(FormatJSON(ParseString(input)))
		if got != first {
			t.Fatalf("iteration %d differs from iteration 0\n--- first ---\n%s\n--- got ---\n%s",
				i, first, got)
		}
	}
}

func TestFormatSOAP_DeterministicOrdering(t *testing.T) {
	c := NewClinicalCase()
	c.Patient.Id = "PT-TEST"
	c.Imaging = map[string]string{
		"xray": "ordered",
		"ct":   "head",
	}
	c.Labs = map[string]string{
		"wbc": "12000",
		"hb":  "9",
	}
	c.Extra = map[string]map[string]string{
		"noteb": {
			"z": "1",
			"a": "2",
		},
		"notea": {
			"k": "true",
		},
	}

	out := FormatSOAP(c)

	labsIdx := strings.Index(out, "Labs           : HB 9 | WBC 12000")
	if labsIdx == -1 {
		t.Fatalf("expected sorted labs in SOAP output, got:\n%s", out)
	}

	imagingIdx := strings.Index(out, "Imaging/Rad    : CT head | XRAY ordered")
	if imagingIdx == -1 {
		t.Fatalf("expected sorted imaging in SOAP output, got:\n%s", out)
	}

	extraOrderOK := strings.Index(out, "NOTEA") < strings.Index(out, "NOTEB")
	if !extraOrderOK {
		t.Fatalf("expected sorted extra command sections, got:\n%s", out)
	}
}
