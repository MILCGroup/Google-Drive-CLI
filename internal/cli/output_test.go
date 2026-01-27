package cli

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dl-alexandre/gdrv/internal/types"
)

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe error: %v", err)
	}
	os.Stderr = w
	fn()
	if err := w.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}
	os.Stderr = orig
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	return string(data)
}

func TestTruncate(t *testing.T) {
	if got := truncate("short", 10); got != "short" {
		t.Fatalf("expected short, got %q", got)
	}
	if got := truncate("1234567890", 7); got != "1234..." {
		t.Fatalf("expected 1234..., got %q", got)
	}
}

func TestFormatSize(t *testing.T) {
	if got := formatSize(0); got != "0 B" {
		t.Fatalf("expected 0 B, got %q", got)
	}
	if got := formatSize(1024); got != "1.0 KB" {
		t.Fatalf("expected 1.0 KB, got %q", got)
	}
	if got := formatSize(1536); got != "1.5 KB" {
		t.Fatalf("expected 1.5 KB, got %q", got)
	}
}

func TestOutputWriterAddWarning(t *testing.T) {
	w := NewOutputWriter(types.OutputFormatJSON, false, false)
	w.AddWarning("code", "msg", "low")
	if len(w.warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(w.warnings))
	}
}

func TestOutputWriterLogQuiet(t *testing.T) {
	w := NewOutputWriter(types.OutputFormatJSON, true, false)
	out := captureStderr(t, func() {
		w.Log("hello %s", "world")
	})
	if out != "" {
		t.Fatalf("expected no output, got %q", out)
	}
}

func TestOutputWriterLog(t *testing.T) {
	w := NewOutputWriter(types.OutputFormatJSON, false, false)
	out := captureStderr(t, func() {
		w.Log("hello %s", "world")
	})
	if !strings.Contains(out, "hello world") {
		t.Fatalf("expected log output, got %q", out)
	}
}

func TestOutputWriterVerbose(t *testing.T) {
	w := NewOutputWriter(types.OutputFormatJSON, false, true)
	out := captureStderr(t, func() {
		w.Verbose("detail %d", 1)
	})
	if !strings.Contains(out, "[VERBOSE] detail 1") {
		t.Fatalf("expected verbose output, got %q", out)
	}
}
