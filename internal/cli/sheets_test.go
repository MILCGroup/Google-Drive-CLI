package cli

import (
	"os"
	"testing"
)

func TestReadSheetValues_EmptyInput(t *testing.T) {
	if _, err := readSheetValuesFrom("", ""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadSheetValues_FromJSON(t *testing.T) {
	values, err := readSheetValuesFrom(`[[1,2],[3,4]]`, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(values) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(values))
	}
}

func TestReadSheetValues_FromFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "values-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(tmp.Name())
	})
	if _, err := tmp.WriteString(`[[1,2]]`); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	values, err := readSheetValuesFrom("", tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(values) != 1 {
		t.Fatalf("expected 1 row, got %d", len(values))
	}
}

func TestReadSheetValues_InvalidJSON(t *testing.T) {
	if _, err := readSheetValuesFrom(`invalid`, ""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadSheetValues_FileNotFound(t *testing.T) {
	if _, err := readSheetValuesFrom("", "/nonexistent/file.json"); err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadSheetsBatchRequests_EmptyInput(t *testing.T) {
	if _, err := readSheetsBatchRequestsFrom("", ""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadSheetsBatchRequests_FromJSON(t *testing.T) {
	requests, err := readSheetsBatchRequestsFrom(`[{"updateSpreadsheetProperties":{"properties":{"title":"New Title"},"fields":"title"}}]`, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}
}

func TestReadSheetsBatchRequests_FromFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "batch-requests-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(tmp.Name())
	})
	if _, err := tmp.WriteString(`[{"updateSpreadsheetProperties":{"properties":{"title":"New Title"},"fields":"title"}}]`); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}
	requests, err := readSheetsBatchRequestsFrom("", tmp.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(requests))
	}
}

func TestReadSheetsBatchRequests_InvalidJSON(t *testing.T) {
	if _, err := readSheetsBatchRequestsFrom(`invalid`, ""); err == nil {
		t.Fatalf("expected error")
	}
}

func TestReadSheetsBatchRequests_FileNotFound(t *testing.T) {
	if _, err := readSheetsBatchRequestsFrom("", "/nonexistent/file.json"); err == nil {
		t.Fatalf("expected error")
	}
}
