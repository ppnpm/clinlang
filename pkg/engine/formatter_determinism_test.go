package engine

import (
	"strings"
	"testing"
)

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
