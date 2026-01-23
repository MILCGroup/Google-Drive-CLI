package safety

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestNewOperationTracker(t *testing.T) {
	tracker := NewOperationTracker()
	if tracker == nil {
		t.Fatal("NewOperationTracker() returned nil")
	}

	if tracker.Count() != 0 {
		t.Errorf("Expected Count()=0, got %d", tracker.Count())
	}
}

func TestMarkOperationCompleted(t *testing.T) {
	tracker := NewOperationTracker()

	opID := "test-operation-1"
	tracker.MarkOperationCompleted(opID)

	if !tracker.IsOperationCompleted(opID) {
		t.Error("Expected operation to be marked as completed")
	}

	if tracker.Count() != 1 {
		t.Errorf("Expected Count()=1, got %d", tracker.Count())
	}
}

func TestIsOperationCompleted(t *testing.T) {
	tracker := NewOperationTracker()

	opID := "test-operation-1"

	if tracker.IsOperationCompleted(opID) {
		t.Error("Operation should not be completed before marking")
	}

	tracker.MarkOperationCompleted(opID)

	if !tracker.IsOperationCompleted(opID) {
		t.Error("Operation should be completed after marking")
	}
}

func TestGetCompletionTime(t *testing.T) {
	tracker := NewOperationTracker()

	opID := "test-operation-1"

	// Before marking as completed
	_, exists := tracker.GetCompletionTime(opID)
	if exists {
		t.Error("GetCompletionTime should return false for non-existent operation")
	}

	// After marking as completed
	before := time.Now()
	tracker.MarkOperationCompleted(opID)
	after := time.Now()

	completionTime, exists := tracker.GetCompletionTime(opID)
	if !exists {
		t.Error("GetCompletionTime should return true for completed operation")
	}

	if completionTime.Before(before) || completionTime.After(after) {
		t.Errorf("Completion time %v is not between %v and %v", completionTime, before, after)
	}
}

func TestClearOperationTracker(t *testing.T) {
	tracker := NewOperationTracker()

	tracker.MarkOperationCompleted("op1")
	tracker.MarkOperationCompleted("op2")

	if tracker.Count() != 2 {
		t.Errorf("Expected Count()=2, got %d", tracker.Count())
	}

	tracker.Clear()

	if tracker.Count() != 0 {
		t.Errorf("Expected Count()=0 after Clear(), got %d", tracker.Count())
	}

	if tracker.IsOperationCompleted("op1") {
		t.Error("Operation should not be completed after Clear()")
	}
}

func TestClearExpired(t *testing.T) {
	tracker := NewOperationTracker()

	// Mark operations as completed
	tracker.MarkOperationCompleted("old-op")
	time.Sleep(10 * time.Millisecond)
	tracker.MarkOperationCompleted("new-op")

	// Clear operations older than 5ms (should only remove old-op)
	tracker.ClearExpired(5 * time.Millisecond)

	if tracker.IsOperationCompleted("old-op") {
		t.Error("Old operation should be cleared")
	}

	if !tracker.IsOperationCompleted("new-op") {
		t.Error("New operation should still exist")
	}
}

func TestRetryCount(t *testing.T) {
	tracker := NewOperationTracker()

	opID := "test-operation"

	// First completion
	tracker.MarkOperationCompleted(opID)

	// Mark again (simulating retry)
	tracker.MarkOperationCompleted(opID)
	tracker.MarkOperationCompleted(opID)

	// The operation should still be marked as completed
	if !tracker.IsOperationCompleted(opID) {
		t.Error("Operation should remain completed after multiple marks")
	}
}

func TestConcurrentTracking(t *testing.T) {
	tracker := NewOperationTracker()
	var wg sync.WaitGroup

	// Mark 100 operations concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			opID := GenerateOperationID("delete", string(rune(id)))
			tracker.MarkOperationCompleted(opID)
		}(i)
	}

	wg.Wait()

	if tracker.Count() != 100 {
		t.Errorf("Expected Count()=100, got %d", tracker.Count())
	}
}

func TestGenerateOperationID(t *testing.T) {
	// Same parameters should generate same ID
	id1 := GenerateOperationID("delete", "file123", "permanent")
	id2 := GenerateOperationID("delete", "file123", "permanent")

	if id1 != id2 {
		t.Error("Same parameters should generate same operation ID")
	}

	// Different parameters should generate different IDs
	id3 := GenerateOperationID("delete", "file456", "permanent")

	if id1 == id3 {
		t.Error("Different parameters should generate different operation IDs")
	}
}

func TestGenerateDeleteOperationID(t *testing.T) {
	id1 := GenerateDeleteOperationID("file123", true)
	id2 := GenerateDeleteOperationID("file123", true)
	id3 := GenerateDeleteOperationID("file123", false)

	if id1 != id2 {
		t.Error("Same delete operation should generate same ID")
	}

	if id1 == id3 {
		t.Error("Different permanent flag should generate different ID")
	}
}

func TestGenerateMoveOperationID(t *testing.T) {
	id1 := GenerateMoveOperationID("file123", "parent456")
	id2 := GenerateMoveOperationID("file123", "parent456")
	id3 := GenerateMoveOperationID("file123", "parent789")

	if id1 != id2 {
		t.Error("Same move operation should generate same ID")
	}

	if id1 == id3 {
		t.Error("Different target parent should generate different ID")
	}
}

func TestGeneratePermissionUpdateOperationID(t *testing.T) {
	id1 := GeneratePermissionUpdateOperationID("file123", "perm456", "writer")
	id2 := GeneratePermissionUpdateOperationID("file123", "perm456", "writer")
	id3 := GeneratePermissionUpdateOperationID("file123", "perm456", "reader")

	if id1 != id2 {
		t.Error("Same permission update should generate same ID")
	}

	if id1 == id3 {
		t.Error("Different role should generate different ID")
	}
}

func TestGeneratePermissionDeleteOperationID(t *testing.T) {
	id1 := GeneratePermissionDeleteOperationID("file123", "perm456")
	id2 := GeneratePermissionDeleteOperationID("file123", "perm456")
	id3 := GeneratePermissionDeleteOperationID("file123", "perm789")

	if id1 != id2 {
		t.Error("Same permission delete should generate same ID")
	}

	if id1 == id3 {
		t.Error("Different permission ID should generate different ID")
	}
}

func TestSafeExecute(t *testing.T) {
	tracker := NewOperationTracker()
	opID := "test-operation"

	executed := false
	operation := func() error {
		executed = true
		return nil
	}

	// First execution should run
	err := SafeExecute(tracker, opID, operation)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !executed {
		t.Error("Operation should have been executed")
	}

	if !tracker.IsOperationCompleted(opID) {
		t.Error("Operation should be marked as completed")
	}

	// Second execution should skip
	executed = false
	err = SafeExecute(tracker, opID, operation)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if executed {
		t.Error("Operation should not have been executed again")
	}
}

func TestSafeExecuteWithError(t *testing.T) {
	tracker := NewOperationTracker()
	opID := "test-operation"

	expectedErr := errors.New("operation failed")
	operation := func() error {
		return expectedErr
	}

	err := SafeExecute(tracker, opID, operation)
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// Operation should not be marked as completed on error
	if tracker.IsOperationCompleted(opID) {
		t.Error("Operation should not be marked as completed on error")
	}
}

func TestSafeExecuteWithRetry(t *testing.T) {
	tracker := NewOperationTracker()
	opID := "test-operation"

	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := SafeExecuteWithRetry(tracker, opID, 5, operation)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	if !tracker.IsOperationCompleted(opID) {
		t.Error("Operation should be marked as completed")
	}
}

func TestSafeExecuteWithRetryExhaustsAttempts(t *testing.T) {
	tracker := NewOperationTracker()
	opID := "test-operation"

	attempts := 0
	operation := func() error {
		attempts++
		return errors.New("persistent failure")
	}

	maxRetries := 2
	err := SafeExecuteWithRetry(tracker, opID, maxRetries, operation)
	if err == nil {
		t.Error("Expected error after exhausting retries")
	}

	// Should have tried maxRetries+1 times (initial + retries)
	expectedAttempts := maxRetries + 1
	if attempts != expectedAttempts {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attempts)
	}

	// Operation should not be marked as completed
	if tracker.IsOperationCompleted(opID) {
		t.Error("Operation should not be marked as completed after failure")
	}
}

func TestSafeExecuteWithRetrySkipsIfCompleted(t *testing.T) {
	tracker := NewOperationTracker()
	opID := "test-operation"

	// Mark as completed
	tracker.MarkOperationCompleted(opID)

	executed := false
	operation := func() error {
		executed = true
		return nil
	}

	err := SafeExecuteWithRetry(tracker, opID, 3, operation)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if executed {
		t.Error("Operation should not have been executed (already completed)")
	}
}
