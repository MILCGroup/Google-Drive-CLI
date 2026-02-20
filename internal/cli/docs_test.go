package cli

import (
	"os"
	"testing"
)

func TestParseDocsRequests_EmptyInput(t *testing.T) {
	if _, err := parseDocsRequests("", ""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseDocsRequests_FromJSON(t *testing.T) {
	requests, err := parseDocsRequests(`[{"insertText":{"location":{"index":1},"text":"Hello"}}]`, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}
}

func TestParseDocsRequests_FromFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "requests-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(tmp.Name())
	})
	if _, err := tmp.WriteString(`[{"insertText":{"location":{"index":1},"text":"Hello"}}]`); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	requests, err := parseDocsRequests("", tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}
}

func TestParseDocsRequests_InvalidJSON(t *testing.T) {
	if _, err := parseDocsRequests(`invalid`, ""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseDocsRequests_FileNotFound(t *testing.T) {
	if _, err := parseDocsRequests("", "/nonexistent/file.json"); err == nil {
		t.Fatalf("expected error")
	}
}
