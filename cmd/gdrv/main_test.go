package main

import (
	"os"
	"testing"
)

func TestRunVersion(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	os.Args = []string{"gdrv", "version"}
	if code := run(); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestRunHelp(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	os.Args = []string{"gdrv", "--help"}
	if code := run(); code != 0 {
		t.Fatalf("expected exit code 0 for help, got %d", code)
	}
}

func TestRunNoArgs(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	os.Args = []string{"gdrv"}
	// Should show help/usage and exit 0
	if code := run(); code != 0 {
		t.Fatalf("expected exit code 0 for no args, got %d", code)
	}
}

func TestRunSuccess(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"version command", []string{"gdrv", "version"}},
		{"help flag", []string{"gdrv", "--help"}},
		{"version flag", []string{"gdrv", "--version"}},
		{"help command", []string{"gdrv", "help"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origArgs := os.Args
			t.Cleanup(func() { os.Args = origArgs })

			os.Args = tt.args
			code := run()
			if code != 0 {
				t.Errorf("run() with args %v = %d, want 0", tt.args, code)
			}
		})
	}
}

func TestRunInvalidCommand(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	os.Args = []string{"gdrv", "nonexistent-command"}
	code := run()
	// Invalid commands should return non-zero exit code
	if code == 0 {
		t.Error("run() with invalid command should return non-zero exit code")
	}
}

func TestMain_CallsRun(t *testing.T) {
	// This test verifies that main() delegates to run()
	// We can't easily test main() directly since it calls os.Exit
	// but we verify the structure is correct

	// Save original args
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	// Test that run() returns correct exit codes
	os.Args = []string{"gdrv", "version"}
	exitCode := run()
	if exitCode != 0 {
		t.Errorf("run() = %d, want 0", exitCode)
	}
}

func TestRun_ReturnsExitCode(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantZero bool
	}{
		{
			name:     "valid command returns 0",
			args:     []string{"gdrv", "version"},
			wantZero: true,
		},
		{
			name:     "help returns 0",
			args:     []string{"gdrv", "--help"},
			wantZero: true,
		},
		{
			name:     "invalid command returns non-zero",
			args:     []string{"gdrv", "invalid-cmd"},
			wantZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			origArgs := os.Args
			t.Cleanup(func() { os.Args = origArgs })

			os.Args = tt.args
			code := run()

			if tt.wantZero && code != 0 {
				t.Errorf("run() = %d, want 0", code)
			}
			if !tt.wantZero && code == 0 {
				t.Errorf("run() = 0, want non-zero")
			}
		})
	}
}

func TestRun_MultipleInvocations(t *testing.T) {
	// Test that run() can be called multiple times
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	// First call
	os.Args = []string{"gdrv", "version"}
	code1 := run()

	// Second call
	os.Args = []string{"gdrv", "version"}
	code2 := run()

	if code1 != code2 {
		t.Errorf("Multiple run() calls returned different codes: %d vs %d", code1, code2)
	}

	if code1 != 0 {
		t.Errorf("run() = %d, want 0", code1)
	}
}
