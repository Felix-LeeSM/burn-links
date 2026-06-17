package events

import (
	"strings"
	"testing"
)

func TestNewJobIDGeneratesURLSafeToken(t *testing.T) {
	first, err := NewJobID()
	if err != nil {
		t.Fatalf("new job id: %v", err)
	}
	second, err := NewJobID()
	if err != nil {
		t.Fatalf("new second job id: %v", err)
	}

	if len(first) != 24 {
		t.Fatalf("job id length = %d, want 24", len(first))
	}
	if strings.Contains(first, "=") {
		t.Fatalf("job id contains padding: %q", first)
	}
	if strings.ContainsAny(first, "+/") {
		t.Fatalf("job id is not URL-safe: %q", first)
	}
	if first == second {
		t.Fatalf("job ids are equal: %q", first)
	}
}
