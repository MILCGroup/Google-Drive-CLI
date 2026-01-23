package safety

import (
	"testing"
)

func TestConfirmWithForceFlag(t *testing.T) {
	opts := SafetyOptions{Force: true, Quiet: true}

	confirmed, err := Confirm("Test message", opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !confirmed {
		t.Error("Confirm should return true with Force flag")
	}
}

func TestConfirmWithYesFlag(t *testing.T) {
	opts := SafetyOptions{Yes: true, Quiet: true}

	confirmed, err := Confirm("Test message", opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !confirmed {
		t.Error("Confirm should return true with Yes flag")
	}
}

func TestConfirmWithDryRun(t *testing.T) {
	opts := SafetyOptions{DryRun: true, Quiet: true}

	confirmed, err := Confirm("Test message", opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !confirmed {
		t.Error("Confirm should return true in DryRun mode")
	}
}

func TestConfirmNonInteractiveWithoutAutoConfirm(t *testing.T) {
	opts := SafetyOptions{Interactive: false}

	_, err := Confirm("Test message", opts)
	if err == nil {
		t.Error("Expected error in non-interactive mode without auto-confirm flag")
	}
}

func TestConfirmBulkOperationZeroItems(t *testing.T) {
	opts := SafetyOptions{Interactive: true}

	confirmed, err := ConfirmBulkOperation(0, "delete", opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if confirmed {
		t.Error("ConfirmBulkOperation should return false for zero items")
	}
}

func TestConfirmBulkOperationWithForce(t *testing.T) {
	opts := SafetyOptions{Force: true, Quiet: true}

	confirmed, err := ConfirmBulkOperation(5, "delete", opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !confirmed {
		t.Error("ConfirmBulkOperation should return true with Force flag")
	}
}

func TestConfirmDestructiveZeroItems(t *testing.T) {
	opts := SafetyOptions{Interactive: true}

	confirmed, err := ConfirmDestructive([]string{}, "delete", opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if confirmed {
		t.Error("ConfirmDestructive should return false for empty list")
	}
}

func TestConfirmDestructiveWithForce(t *testing.T) {
	opts := SafetyOptions{Force: true, Quiet: true}
	items := []string{"file1.txt", "file2.txt", "file3.txt"}

	confirmed, err := ConfirmDestructive(items, "permanently delete", opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !confirmed {
		t.Error("ConfirmDestructive should return true with Force flag")
	}
}

func TestConfirmDestructiveNonInteractiveWithoutAutoConfirm(t *testing.T) {
	opts := SafetyOptions{Interactive: false}
	items := []string{"file1.txt"}

	_, err := ConfirmDestructive(items, "delete", opts)
	if err == nil {
		t.Error("Expected error in non-interactive mode without auto-confirm flag")
	}
}

func TestConfirmWithDefaultForce(t *testing.T) {
	opts := SafetyOptions{Force: true, Quiet: true}

	confirmed, err := ConfirmWithDefault("Test message", false, opts)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !confirmed {
		t.Error("ConfirmWithDefault should return true with Force flag")
	}
}

func TestConfirmWithDefaultNonInteractive(t *testing.T) {
	opts := SafetyOptions{Interactive: false}

	_, err := ConfirmWithDefault("Test message", true, opts)
	if err == nil {
		t.Error("Expected error in non-interactive mode without auto-confirm flag")
	}
}

// Note: Testing interactive prompts that require stdin is difficult in unit tests.
// These tests cover the auto-confirm cases and error paths.
// Integration tests would be needed to test actual user input scenarios.
