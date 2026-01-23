package safety

import (
	"fmt"
	"sync"
	"time"
)

// DryRunRecorder records planned operations during dry-run mode.
// Implementations should be thread-safe.
//
// Requirements:
//   - Requirement 13.1: Support --dry-run mode for destructive operations
type DryRunRecorder interface {
	// RecordOperation records a planned operation
	RecordOperation(op PlannedOperation)

	// GetOperations returns all recorded operations
	GetOperations() []PlannedOperation

	// Clear clears all recorded operations
	Clear()

	// Count returns the number of recorded operations
	Count() int
}

// OperationType represents the type of operation
type OperationType string

const (
	// OpTypeDelete represents a delete operation
	OpTypeDelete OperationType = "delete"

	// OpTypeTrash represents a trash operation
	OpTypeTrash OperationType = "trash"

	// OpTypeMove represents a move operation
	OpTypeMove OperationType = "move"

	// OpTypeUpdatePermission represents a permission update operation
	OpTypeUpdatePermission OperationType = "update_permission"

	// OpTypeDeletePermission represents a permission deletion operation
	OpTypeDeletePermission OperationType = "delete_permission"

	// OpTypeUpdate represents a metadata update operation
	OpTypeUpdate OperationType = "update"

	// OpTypeCopy represents a copy operation
	OpTypeCopy OperationType = "copy"
)

// PlannedOperation represents a planned operation in dry-run mode.
//
// Requirements:
//   - Requirement 13.1: Support --dry-run mode for destructive operations
type PlannedOperation struct {
	// Type is the type of operation
	Type OperationType

	// ResourceID is the ID of the resource being operated on
	ResourceID string

	// ResourceName is the name of the resource (optional, for display)
	ResourceName string

	// Description is a human-readable description of the operation
	Description string

	// Parameters contains operation-specific parameters
	Parameters map[string]interface{}

	// Timestamp is when the operation was recorded
	Timestamp time.Time
}

// DryRunResult stores the results of a dry-run operation.
//
// Requirements:
//   - Requirement 13.1: Support --dry-run mode for destructive operations
type DryRunResult struct {
	// Operations is the list of operations that would be performed
	Operations []PlannedOperation

	// TotalCount is the total number of operations
	TotalCount int

	// Summary provides operation type counts
	Summary map[OperationType]int

	// Warnings contains any warnings generated during planning
	Warnings []string
}

// NewDryRunResult creates a new DryRunResult from a list of operations
func NewDryRunResult(operations []PlannedOperation, warnings []string) *DryRunResult {
	summary := make(map[OperationType]int)
	for _, op := range operations {
		summary[op.Type]++
	}

	return &DryRunResult{
		Operations: operations,
		TotalCount: len(operations),
		Summary:    summary,
		Warnings:   warnings,
	}
}

// DefaultDryRunRecorder is a thread-safe in-memory implementation of DryRunRecorder
type DefaultDryRunRecorder struct {
	mu         sync.RWMutex
	operations []PlannedOperation
}

// NewDryRunRecorder creates a new default dry-run recorder
func NewDryRunRecorder() *DefaultDryRunRecorder {
	return &DefaultDryRunRecorder{
		operations: make([]PlannedOperation, 0),
	}
}

// RecordOperation records a planned operation
func (r *DefaultDryRunRecorder) RecordOperation(op PlannedOperation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if op.Timestamp.IsZero() {
		op.Timestamp = time.Now()
	}

	r.operations = append(r.operations, op)
}

// GetOperations returns all recorded operations
func (r *DefaultDryRunRecorder) GetOperations() []PlannedOperation {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]PlannedOperation, len(r.operations))
	copy(result, r.operations)
	return result
}

// Clear clears all recorded operations
func (r *DefaultDryRunRecorder) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.operations = make([]PlannedOperation, 0)
}

// Count returns the number of recorded operations
func (r *DefaultDryRunRecorder) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.operations)
}

// RecordDelete records a planned delete operation
func RecordDelete(recorder DryRunRecorder, resourceID, resourceName string, permanent bool) {
	opType := OpTypeTrash
	desc := fmt.Sprintf("Trash: %s", resourceName)
	if permanent {
		opType = OpTypeDelete
		desc = fmt.Sprintf("Permanently delete: %s", resourceName)
	}

	recorder.RecordOperation(PlannedOperation{
		Type:         opType,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Description:  desc,
		Parameters: map[string]interface{}{
			"permanent": permanent,
		},
	})
}

// RecordMove records a planned move operation
func RecordMove(recorder DryRunRecorder, resourceID, resourceName, targetParentID, targetParentName string) {
	recorder.RecordOperation(PlannedOperation{
		Type:         OpTypeMove,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Description:  fmt.Sprintf("Move '%s' to '%s'", resourceName, targetParentName),
		Parameters: map[string]interface{}{
			"targetParentID":   targetParentID,
			"targetParentName": targetParentName,
		},
	})
}

// RecordPermissionUpdate records a planned permission update operation
func RecordPermissionUpdate(recorder DryRunRecorder, fileID, fileName, permissionID, newRole string) {
	recorder.RecordOperation(PlannedOperation{
		Type:         OpTypeUpdatePermission,
		ResourceID:   fileID,
		ResourceName: fileName,
		Description:  fmt.Sprintf("Update permission on '%s' to role '%s'", fileName, newRole),
		Parameters: map[string]interface{}{
			"permissionID": permissionID,
			"newRole":      newRole,
		},
	})
}

// RecordPermissionDelete records a planned permission deletion operation
func RecordPermissionDelete(recorder DryRunRecorder, fileID, fileName, permissionID string) {
	recorder.RecordOperation(PlannedOperation{
		Type:         OpTypeDeletePermission,
		ResourceID:   fileID,
		ResourceName: fileName,
		Description:  fmt.Sprintf("Remove permission from '%s'", fileName),
		Parameters: map[string]interface{}{
			"permissionID": permissionID,
		},
	})
}

// RecordUpdate records a planned update operation
func RecordUpdate(recorder DryRunRecorder, resourceID, resourceName string, fields map[string]interface{}) {
	recorder.RecordOperation(PlannedOperation{
		Type:         OpTypeUpdate,
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Description:  fmt.Sprintf("Update '%s'", resourceName),
		Parameters:   fields,
	})
}

// PrintDryRunResult prints a formatted dry-run result to stdout
func PrintDryRunResult(result *DryRunResult) {
	fmt.Println("\n=== DRY RUN RESULTS ===")
	fmt.Printf("Total operations planned: %d\n\n", result.TotalCount)

	if len(result.Summary) > 0 {
		fmt.Println("Operations by type:")
		for opType, count := range result.Summary {
			fmt.Printf("  - %s: %d\n", opType, count)
		}
		fmt.Println()
	}

	if len(result.Operations) > 0 {
		fmt.Println("Planned operations:")
		for i, op := range result.Operations {
			fmt.Printf("  %d. [%s] %s (ID: %s)\n", i+1, op.Type, op.Description, op.ResourceID)
		}
		fmt.Println()
	}

	if len(result.Warnings) > 0 {
		fmt.Println("⚠️  Warnings:")
		for _, warning := range result.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
		fmt.Println()
	}

	fmt.Println("NOTE: This was a dry run. No actual changes were made.")
	fmt.Println("Remove --dry-run flag to execute these operations.")
}
