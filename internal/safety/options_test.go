package safety

import (
	"testing"
)

func TestDefault(t *testing.T) {
	opts := Default()

	if opts.DryRun {
		t.Error("Default() should have DryRun=false")
	}
	if opts.Force {
		t.Error("Default() should have Force=false")
	}
	if !opts.Yes {
		t.Error("Default() should have Yes=true (auto-confirm for agent-friendly CLI)")
	}
	if opts.Quiet {
		t.Error("Default() should have Quiet=false")
	}
	if opts.Interactive {
		t.Error("Default() should have Interactive=false (non-interactive for agent-friendly CLI)")
	}
}

func TestInteractive(t *testing.T) {
	opts := Interactive()

	if opts.DryRun {
		t.Error("Interactive() should have DryRun=false")
	}
	if opts.Force {
		t.Error("Interactive() should have Force=false")
	}
	if opts.Yes {
		t.Error("Interactive() should have Yes=false")
	}
	if opts.Quiet {
		t.Error("Interactive() should have Quiet=false")
	}
	if !opts.Interactive {
		t.Error("Interactive() should have Interactive=true")
	}
}

func TestNonInteractive(t *testing.T) {
	opts := NonInteractive()

	if opts.Interactive {
		t.Error("NonInteractive() should have Interactive=false")
	}
	if opts.DryRun {
		t.Error("NonInteractive() should have DryRun=false")
	}
	if opts.Force {
		t.Error("NonInteractive() should have Force=false")
	}
	if opts.Yes {
		t.Error("NonInteractive() should have Yes=false")
	}
}

func TestDryRunMode(t *testing.T) {
	opts := DryRunMode()

	if !opts.DryRun {
		t.Error("DryRunMode() should have DryRun=true")
	}
	if opts.Force {
		t.Error("DryRunMode() should have Force=false")
	}
	if opts.Yes {
		t.Error("DryRunMode() should have Yes=false")
	}
}

func TestShouldConfirm(t *testing.T) {
	tests := []struct {
		name     string
		opts     SafetyOptions
		expected bool
	}{
		{
			name:     "Default options should not confirm (auto-yes)",
			opts:     Default(),
			expected: false,
		},
		{
			name:     "Interactive options should confirm",
			opts:     Interactive(),
			expected: true,
		},
		{
			name:     "Force flag should skip confirmation",
			opts:     SafetyOptions{Force: true, Interactive: true},
			expected: false,
		},
		{
			name:     "Yes flag should skip confirmation",
			opts:     SafetyOptions{Yes: true, Interactive: true},
			expected: false,
		},
		{
			name:     "DryRun should skip confirmation",
			opts:     SafetyOptions{DryRun: true, Interactive: true},
			expected: false,
		},
		{
			name:     "Non-interactive should skip confirmation",
			opts:     SafetyOptions{Interactive: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.opts.ShouldConfirm()
			if result != tt.expected {
				t.Errorf("Expected ShouldConfirm()=%v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAutoConfirm(t *testing.T) {
	tests := []struct {
		name     string
		opts     SafetyOptions
		expected bool
	}{
		{
			name:     "Default options should auto-confirm (Yes=true)",
			opts:     Default(),
			expected: true,
		},
		{
			name:     "Interactive options should not auto-confirm",
			opts:     Interactive(),
			expected: false,
		},
		{
			name:     "Force flag should auto-confirm",
			opts:     SafetyOptions{Force: true},
			expected: true,
		},
		{
			name:     "Yes flag should auto-confirm",
			opts:     SafetyOptions{Yes: true},
			expected: true,
		},
		{
			name:     "DryRun should auto-confirm",
			opts:     SafetyOptions{DryRun: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.opts.AutoConfirm()
			if result != tt.expected {
				t.Errorf("Expected AutoConfirm()=%v, got %v", tt.expected, result)
			}
		})
	}
}

func TestShouldExecute(t *testing.T) {
	tests := []struct {
		name     string
		opts     SafetyOptions
		expected bool
	}{
		{
			name:     "Default options should execute",
			opts:     Default(),
			expected: true,
		},
		{
			name:     "Force flag should execute",
			opts:     SafetyOptions{Force: true},
			expected: true,
		},
		{
			name:     "DryRun should not execute",
			opts:     SafetyOptions{DryRun: true},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.opts.ShouldExecute()
			if result != tt.expected {
				t.Errorf("Expected ShouldExecute()=%v, got %v", tt.expected, result)
			}
		})
	}
}
