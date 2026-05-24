package engine

import (
	"strings"
	"testing"
)

func TestGenerateId_NoCollisionsIn1000(t *testing.T) {
	seen := make(map[string]struct{}, 1000)
	for i := 0; i < 1000; i++ {
		id := GenerateId()
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate ID at iteration %d: %s", i, id)
		}
		seen[id] = struct{}{}
	}
}

func TestGenerateId_Format(t *testing.T) {
	id := GenerateId()
	if !strings.HasPrefix(id, "PT-") {
		t.Errorf("ID must start with PT-, got %q", id)
	}
	// "PT" + "-" + 8-digit date + "-" + 6-digit time + "-" + 6-hex = 25 chars
	if len(id) != 25 {
		t.Errorf("ID length: want 25, got %d (%q)", len(id), id)
	}
	parts := strings.Split(id, "-")
	if len(parts) != 4 {
		t.Fatalf("ID parts: want 4 (PT/date/time/suffix), got %d (%q)", len(parts), id)
	}
	if len(parts[1]) != 8 {
		t.Errorf("date part: want 8 chars, got %d", len(parts[1]))
	}
	if len(parts[2]) != 6 {
		t.Errorf("time part: want 6 chars, got %d", len(parts[2]))
	}
	if len(parts[3]) != 6 {
		t.Errorf("random suffix: want 6 chars, got %d", len(parts[3]))
	}
}
