package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// goldenCases are pairs of (.cln input, .expected.txt golden) under
// testdata/. The .cln files use an "id PT-FIXED-NNN" line so the
// generated Patient.Id is deterministic and the golden is stable.
//
// We deliberately do not include a plugin-loaded case here: blank-
// importing a plugin from the engine package would create an import
// cycle (plugin → engine). The plugin packages have their own tests
// that cover specialty-data rendering.
var goldenCases = []string{
	"note_basic",
	"note_full",
}

// TestFormatPlainNote_Golden compares FormatPlainNote(ParseString(input))
// byte-for-byte against a golden file. Set UPDATE_GOLDEN=1 to regenerate
// the golden files in place — use that flag when the rendering changes
// intentionally and you want to refresh the expected output.
func TestFormatPlainNote_Golden(t *testing.T) {
	update := os.Getenv("UPDATE_GOLDEN") == "1"
	for _, name := range goldenCases {
		t.Run(name, func(t *testing.T) {
			inputPath := filepath.Join("testdata", name+".cln")
			goldenPath := filepath.Join("testdata", "golden", name+".expected.txt")

			raw, err := os.ReadFile(inputPath)
			if err != nil {
				t.Fatalf("read %s: %v", inputPath, err)
			}

			got := FormatPlainNote(ParseString(string(raw)))

			if update {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
					t.Fatalf("mkdir golden dir: %v", err)
				}
				if err := os.WriteFile(goldenPath, []byte(got), 0644); err != nil {
					t.Fatalf("write golden %s: %v", goldenPath, err)
				}
				t.Logf("regenerated %s", goldenPath)
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden %s: %v (run with UPDATE_GOLDEN=1 to create)", goldenPath, err)
			}
			// Normalize carriage returns for Windows compatibility
			wantStr := strings.ReplaceAll(string(want), "\r\n", "\n")
			if got != wantStr {
				t.Errorf("output mismatch for %s\n--- want ---\n%s\n--- got ---\n%s",
					name, wantStr, got)
			}
		})
	}
}
