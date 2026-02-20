package cli

import (
	"os"
	"testing"
)

func TestParseSlidesRequests_EmptyInput(t *testing.T) {
	if _, err := parseSlidesRequests("", ""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseSlidesRequests_FromJSON(t *testing.T) {
	requests, err := parseSlidesRequests(`[{"insertText":{"objectId":"slide1","insertionIndex":0,"text":"Hello"}}]`, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}
}

func TestParseSlidesRequests_FromFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "slides-requests-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(tmp.Name())
	})
	if _, err := tmp.WriteString(`[{"insertText":{"objectId":"slide1","insertionIndex":0,"text":"Hello"}}]`); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	requests, err := parseSlidesRequests("", tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}
}

func TestParseSlidesRequests_InvalidJSON(t *testing.T) {
	if _, err := parseSlidesRequests(`invalid`, ""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseSlidesRequests_FileNotFound(t *testing.T) {
	if _, err := parseSlidesRequests("", "/nonexistent/file.json"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseSlidesReplacements_EmptyInput(t *testing.T) {
	if _, err := parseSlidesReplacements("", ""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseSlidesReplacements_FromJSON(t *testing.T) {
	replacements, err := parseSlidesReplacements(`{"{{name}}":"John","{{date}}":"2024-01-01"}`, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replacements) != 2 {
		t.Fatalf("expected 2 replacements, got %d", len(replacements))
	}
}

func TestParseSlidesReplacements_FromFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "slides-replacements-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(tmp.Name())
	})
	if _, err := tmp.WriteString(`{"{{name}}":"John"}`); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	replacements, err := parseSlidesReplacements("", tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replacements) != 1 {
		t.Fatalf("expected 1 replacement, got %d", len(replacements))
	}
}

func TestParseSlidesReplacements_InvalidJSON(t *testing.T) {
	if _, err := parseSlidesReplacements(`invalid`, ""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseSlidesReplacements_FileNotFound(t *testing.T) {
	if _, err := parseSlidesReplacements("", "/nonexistent/file.json"); err == nil {
		t.Fatalf("expected error")
	}
}
