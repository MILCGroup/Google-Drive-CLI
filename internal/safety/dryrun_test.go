package safety

import (
	"sync"
	"testing"
	"time"
)

func TestNewDryRunRecorder(t *testing.T) {
	recorder := NewDryRunRecorder()
	if recorder == nil {
		t.Fatal("NewDryRunRecorder() returned nil")
	}

	if recorder.Count() != 0 {
		t.Errorf("Expected Count()=0, got %d", recorder.Count())
	}
}

func TestRecordOperation(t *testing.T) {
	recorder := NewDryRunRecorder()

	op := PlannedOperation{
		Type:         OpTypeDelete,
		ResourceID:   "file123",
		ResourceName: "test.txt",
		Description:  "Delete test.txt",
		Parameters: map[string]interface{}{
			"permanent": true,
		},
	}

	recorder.RecordOperation(op)

	if recorder.Count() != 1 {
		t.Errorf("Expected Count()=1, got %d", recorder.Count())
	}

	ops := recorder.GetOperations()
	if len(ops) != 1 {
		t.Fatalf("Expected 1 operation, got %d", len(ops))
	}

	if ops[0].Type != OpTypeDelete {
		t.Errorf("Expected Type=%s, got %s", OpTypeDelete, ops[0].Type)
	}
	if ops[0].ResourceID != "file123" {
		t.Errorf("Expected ResourceID=file123, got %s", ops[0].ResourceID)
	}
}

func TestRecordOperationSetsTimestamp(t *testing.T) {
	recorder := NewDryRunRecorder()

	op := PlannedOperation{
		Type:       OpTypeDelete,
		ResourceID: "file123",
	}

	before := time.Now()
	recorder.RecordOperation(op)
	after := time.Now()

	ops := recorder.GetOperations()
	if ops[0].Timestamp.IsZero() {
		t.Error("Expected Timestamp to be set")
	}
	if ops[0].Timestamp.Before(before) || ops[0].Timestamp.After(after) {
		t.Errorf("Timestamp %v is not between %v and %v", ops[0].Timestamp, before, after)
	}
}

func TestClear(t *testing.T) {
	recorder := NewDryRunRecorder()

	recorder.RecordOperation(PlannedOperation{Type: OpTypeDelete, ResourceID: "1"})
	recorder.RecordOperation(PlannedOperation{Type: OpTypeTrash, ResourceID: "2"})

	if recorder.Count() != 2 {
		t.Errorf("Expected Count()=2, got %d", recorder.Count())
	}

	recorder.Clear()

	if recorder.Count() != 0 {
		t.Errorf("Expected Count()=0 after Clear(), got %d", recorder.Count())
	}
}

func TestGetOperationsReturnsCopy(t *testing.T) {
	recorder := NewDryRunRecorder()

	recorder.RecordOperation(PlannedOperation{Type: OpTypeDelete, ResourceID: "1"})

	ops1 := recorder.GetOperations()
	ops2 := recorder.GetOperations()

	// Modify first copy
	ops1[0].ResourceID = "modified"

	// Second copy should be unchanged
	if ops2[0].ResourceID != "1" {
		t.Error("GetOperations() should return a copy, not a reference")
	}
}

func TestConcurrentRecording(t *testing.T) {
	recorder := NewDryRunRecorder()
	var wg sync.WaitGroup

	// Record 100 operations concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			recorder.RecordOperation(PlannedOperation{
				Type:       OpTypeDelete,
				ResourceID: string(rune(id)),
			})
		}(i)
	}

	wg.Wait()

	if recorder.Count() != 100 {
		t.Errorf("Expected Count()=100, got %d", recorder.Count())
	}
}

func TestRecordDelete(t *testing.T) {
	recorder := NewDryRunRecorder()

	RecordDelete(recorder, "file123", "test.txt", false)

	ops := recorder.GetOperations()
	if len(ops) != 1 {
		t.Fatalf("Expected 1 operation, got %d", len(ops))
	}

	if ops[0].Type != OpTypeTrash {
		t.Errorf("Expected Type=%s, got %s", OpTypeTrash, ops[0].Type)
	}

	// Test permanent delete
	recorder.Clear()
	RecordDelete(recorder, "file456", "test2.txt", true)

	ops = recorder.GetOperations()
	if ops[0].Type != OpTypeDelete {
		t.Errorf("Expected Type=%s for permanent delete, got %s", OpTypeDelete, ops[0].Type)
	}
}

func TestRecordMove(t *testing.T) {
	recorder := NewDryRunRecorder()

	RecordMove(recorder, "file123", "test.txt", "parent456", "New Folder")

	ops := recorder.GetOperations()
	if len(ops) != 1 {
		t.Fatalf("Expected 1 operation, got %d", len(ops))
	}

	if ops[0].Type != OpTypeMove {
		t.Errorf("Expected Type=%s, got %s", OpTypeMove, ops[0].Type)
	}

	targetParentID := ops[0].Parameters["targetParentID"]
	if targetParentID != "parent456" {
		t.Errorf("Expected targetParentID=parent456, got %v", targetParentID)
	}
}

func TestRecordPermissionUpdate(t *testing.T) {
	recorder := NewDryRunRecorder()

	RecordPermissionUpdate(recorder, "file123", "test.txt", "perm456", "writer")

	ops := recorder.GetOperations()
	if len(ops) != 1 {
		t.Fatalf("Expected 1 operation, got %d", len(ops))
	}

	if ops[0].Type != OpTypeUpdatePermission {
		t.Errorf("Expected Type=%s, got %s", OpTypeUpdatePermission, ops[0].Type)
	}
}

func TestRecordPermissionDelete(t *testing.T) {
	recorder := NewDryRunRecorder()

	RecordPermissionDelete(recorder, "file123", "test.txt", "perm456")

	ops := recorder.GetOperations()
	if len(ops) != 1 {
		t.Fatalf("Expected 1 operation, got %d", len(ops))
	}

	if ops[0].Type != OpTypeDeletePermission {
		t.Errorf("Expected Type=%s, got %s", OpTypeDeletePermission, ops[0].Type)
	}
}

func TestNewDryRunResult(t *testing.T) {
	ops := []PlannedOperation{
		{Type: OpTypeDelete, ResourceID: "1"},
		{Type: OpTypeDelete, ResourceID: "2"},
		{Type: OpTypeTrash, ResourceID: "3"},
		{Type: OpTypeMove, ResourceID: "4"},
	}

	warnings := []string{"Warning 1", "Warning 2"}

	result := NewDryRunResult(ops, warnings)

	if result.TotalCount != 4 {
		t.Errorf("Expected TotalCount=4, got %d", result.TotalCount)
	}

	if result.Summary[OpTypeDelete] != 2 {
		t.Errorf("Expected 2 delete operations, got %d", result.Summary[OpTypeDelete])
	}
	if result.Summary[OpTypeTrash] != 1 {
		t.Errorf("Expected 1 trash operation, got %d", result.Summary[OpTypeTrash])
	}
	if result.Summary[OpTypeMove] != 1 {
		t.Errorf("Expected 1 move operation, got %d", result.Summary[OpTypeMove])
	}

	if len(result.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(result.Warnings))
	}
}

func TestPrintDryRunResult(t *testing.T) {
	// This is mainly for code coverage
	// In a real scenario, you'd capture stdout and verify output
	ops := []PlannedOperation{
		{Type: OpTypeDelete, ResourceID: "1", ResourceName: "file1.txt", Description: "Delete file1.txt"},
		{Type: OpTypeTrash, ResourceID: "2", ResourceName: "file2.txt", Description: "Trash file2.txt"},
	}

	result := NewDryRunResult(ops, []string{"Warning: File is large"})

	// This should not panic
	PrintDryRunResult(result)
}
