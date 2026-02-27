package safety_test

import (
	"testing"

	"github.com/milcgroup/gdrv/internal/safety"
)

// TestIntegrationDryRunWorkflow tests a complete dry-run workflow
func TestIntegrationDryRunWorkflow(t *testing.T) {
	// Create recorder
	recorder := safety.NewDryRunRecorder()

	// Simulate multiple operations
	safety.RecordDelete(recorder, "file1", "document.pdf", false)
	safety.RecordDelete(recorder, "file2", "spreadsheet.xlsx", true)
	safety.RecordMove(recorder, "file3", "image.png", "folder1", "Archive")
	safety.RecordPermissionUpdate(recorder, "file4", "data.csv", "perm1", "writer")
	safety.RecordPermissionDelete(recorder, "file5", "report.doc", "perm2")

	// Verify all operations recorded
	if recorder.Count() != 5 {
		t.Errorf("Expected 5 operations, got %d", recorder.Count())
	}

	// Get operations and create result
	ops := recorder.GetOperations()
	warnings := []string{
		"File file1 is large and may take time to process",
		"Folder1 has restricted permissions",
	}
	result := safety.NewDryRunResult(ops, warnings)

	// Verify result
	if result.TotalCount != 5 {
		t.Errorf("Expected TotalCount=5, got %d", result.TotalCount)
	}

	if result.Summary[safety.OpTypeTrash] != 1 {
		t.Errorf("Expected 1 trash operation, got %d", result.Summary[safety.OpTypeTrash])
	}

	if result.Summary[safety.OpTypeDelete] != 1 {
		t.Errorf("Expected 1 delete operation, got %d", result.Summary[safety.OpTypeDelete])
	}

	if result.Summary[safety.OpTypeMove] != 1 {
		t.Errorf("Expected 1 move operation, got %d", result.Summary[safety.OpTypeMove])
	}

	if len(result.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(result.Warnings))
	}

	// Print result (mainly for visual verification during test runs)
	safety.PrintDryRunResult(result)
}

// TestIntegrationIdempotentRetry tests idempotent retry workflow
func TestIntegrationIdempotentRetry(t *testing.T) {
	tracker := safety.NewOperationTracker()

	// Simulate multiple operations
	operations := []struct {
		id   string
		name string
	}{
		{"file1", "document.pdf"},
		{"file2", "spreadsheet.xlsx"},
		{"file3", "image.png"},
	}

	// Execute operations with idempotency
	for _, op := range operations {
		opID := safety.GenerateDeleteOperationID(op.id, true)

		executed := false
		err := safety.SafeExecute(tracker, opID, func() error {
			executed = true
			// Simulate operation
			return nil
		})

		if err != nil {
			t.Errorf("Unexpected error for %s: %v", op.name, err)
		}

		if !executed {
			t.Errorf("Operation for %s should have been executed", op.name)
		}

		// Verify tracked
		if !tracker.IsOperationCompleted(opID) {
			t.Errorf("Operation %s should be tracked", op.name)
		}
	}

	// Verify all operations tracked
	if tracker.Count() != 3 {
		t.Errorf("Expected 3 tracked operations, got %d", tracker.Count())
	}

	// Try to re-execute - should be skipped
	for _, op := range operations {
		opID := safety.GenerateDeleteOperationID(op.id, true)

		executed := false
		err := safety.SafeExecute(tracker, opID, func() error {
			executed = true
			return nil
		})

		if err != nil {
			t.Errorf("Unexpected error on retry for %s: %v", op.name, err)
		}

		if executed {
			t.Errorf("Operation for %s should have been skipped (already completed)", op.name)
		}
	}
}

// TestIntegrationSafetyOptionsFlow tests safety options decision flow
func TestIntegrationSafetyOptionsFlow(t *testing.T) {
	tests := []struct {
		name          string
		opts          safety.SafetyOptions
		shouldConfirm bool
		autoConfirm   bool
		shouldExecute bool
	}{
		{
			name:          "Default (non-interactive, auto-yes)",
			opts:          safety.Default(),
			shouldConfirm: false,
			autoConfirm:   true,
			shouldExecute: true,
		},
		{
			name:          "Interactive mode",
			opts:          safety.Interactive(),
			shouldConfirm: true,
			autoConfirm:   false,
			shouldExecute: true,
		},
		{
			name:          "Force flag",
			opts:          safety.SafetyOptions{Force: true, Interactive: true},
			shouldConfirm: false,
			autoConfirm:   true,
			shouldExecute: true,
		},
		{
			name:          "Yes flag",
			opts:          safety.SafetyOptions{Yes: true, Interactive: true},
			shouldConfirm: false,
			autoConfirm:   true,
			shouldExecute: true,
		},
		{
			name:          "Dry-run mode",
			opts:          safety.DryRunMode(),
			shouldConfirm: false,
			autoConfirm:   true,
			shouldExecute: false,
		},
		{
			name:          "Non-interactive strict",
			opts:          safety.NonInteractive(),
			shouldConfirm: false,
			autoConfirm:   false,
			shouldExecute: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opts.ShouldConfirm() != tt.shouldConfirm {
				t.Errorf("Expected ShouldConfirm()=%v, got %v", tt.shouldConfirm, tt.opts.ShouldConfirm())
			}

			if tt.opts.AutoConfirm() != tt.autoConfirm {
				t.Errorf("Expected AutoConfirm()=%v, got %v", tt.autoConfirm, tt.opts.AutoConfirm())
			}

			if tt.opts.ShouldExecute() != tt.shouldExecute {
				t.Errorf("Expected ShouldExecute()=%v, got %v", tt.shouldExecute, tt.opts.ShouldExecute())
			}
		})
	}
}

// TestIntegrationOperationIDUniqueness tests that operation IDs are unique and deterministic
func TestIntegrationOperationIDUniqueness(t *testing.T) {
	// Generate operation IDs for different scenarios
	deleteID1 := safety.GenerateDeleteOperationID("file123", true)
	deleteID2 := safety.GenerateDeleteOperationID("file123", true)
	deleteID3 := safety.GenerateDeleteOperationID("file123", false)
	deleteID4 := safety.GenerateDeleteOperationID("file456", true)

	moveID1 := safety.GenerateMoveOperationID("file123", "parent1")
	moveID2 := safety.GenerateMoveOperationID("file123", "parent1")
	moveID3 := safety.GenerateMoveOperationID("file123", "parent2")

	permUpdateID1 := safety.GeneratePermissionUpdateOperationID("file123", "perm1", "writer")
	permUpdateID2 := safety.GeneratePermissionUpdateOperationID("file123", "perm1", "writer")
	permUpdateID3 := safety.GeneratePermissionUpdateOperationID("file123", "perm1", "reader")

	// Same parameters should generate same ID
	if deleteID1 != deleteID2 {
		t.Error("Same delete operation should generate same ID")
	}

	if moveID1 != moveID2 {
		t.Error("Same move operation should generate same ID")
	}

	if permUpdateID1 != permUpdateID2 {
		t.Error("Same permission update should generate same ID")
	}

	// Different parameters should generate different IDs
	if deleteID1 == deleteID3 {
		t.Error("Different permanent flag should generate different ID")
	}

	if deleteID1 == deleteID4 {
		t.Error("Different file ID should generate different ID")
	}

	if moveID1 == moveID3 {
		t.Error("Different target parent should generate different ID")
	}

	if permUpdateID1 == permUpdateID3 {
		t.Error("Different role should generate different ID")
	}

	// Different operation types should generate different IDs
	if deleteID1 == moveID1 {
		t.Error("Different operation types should generate different IDs")
	}

	// Verify all IDs are non-empty
	ids := []string{deleteID1, moveID1, permUpdateID1}
	for i, id := range ids {
		if id == "" {
			t.Errorf("Operation ID %d is empty", i)
		}
	}
}

// TestIntegrationConcurrentSafety tests thread-safety of safety components
func TestIntegrationConcurrentSafety(t *testing.T) {
	recorder := safety.NewDryRunRecorder()
	tracker := safety.NewOperationTracker()

	// Launch multiple goroutines
	done := make(chan bool)
	numGoroutines := 10
	opsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			for j := 0; j < opsPerGoroutine; j++ {
				fileID := string(rune(routineID*100 + j))

				// Record operation
				safety.RecordDelete(recorder, fileID, "file.txt", false)

				// Track operation
				opID := safety.GenerateDeleteOperationID(fileID, false)
				tracker.MarkOperationCompleted(opID)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify counts
	expectedCount := numGoroutines * opsPerGoroutine
	if recorder.Count() != expectedCount {
		t.Errorf("Expected recorder count=%d, got %d", expectedCount, recorder.Count())
	}

	if tracker.Count() != expectedCount {
		t.Errorf("Expected tracker count=%d, got %d", expectedCount, tracker.Count())
	}
}
