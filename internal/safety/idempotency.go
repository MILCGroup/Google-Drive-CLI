package safety

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// OperationTracker tracks completed operations to support idempotent retries.
// Operations are tracked by a unique operation ID that should be deterministic
// based on the operation parameters.
//
// Requirements:
//   - Requirement 13.4: Add idempotent behavior for retry operations
type OperationTracker interface {
	// IsOperationCompleted checks if an operation has been completed
	IsOperationCompleted(operationID string) bool

	// MarkOperationCompleted marks an operation as completed
	MarkOperationCompleted(operationID string)

	// GetCompletionTime returns when the operation was completed
	GetCompletionTime(operationID string) (time.Time, bool)

	// Clear removes all tracked operations
	Clear()

	// ClearExpired removes operations older than the specified duration
	ClearExpired(maxAge time.Duration)

	// Count returns the number of tracked operations
	Count() int
}

// CompletedOperation represents a completed operation
type CompletedOperation struct {
	OperationID   string
	CompletedAt   time.Time
	RetryCount    int
	LastAttemptAt time.Time
}

// DefaultOperationTracker is an in-memory implementation of OperationTracker
type DefaultOperationTracker struct {
	mu         sync.RWMutex
	operations map[string]*CompletedOperation
}

// NewOperationTracker creates a new default operation tracker
func NewOperationTracker() *DefaultOperationTracker {
	return &DefaultOperationTracker{
		operations: make(map[string]*CompletedOperation),
	}
}

// IsOperationCompleted checks if an operation has been completed
func (t *DefaultOperationTracker) IsOperationCompleted(operationID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, exists := t.operations[operationID]
	return exists
}

// MarkOperationCompleted marks an operation as completed
func (t *DefaultOperationTracker) MarkOperationCompleted(operationID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if existing, exists := t.operations[operationID]; exists {
		existing.RetryCount++
		existing.LastAttemptAt = time.Now()
	} else {
		t.operations[operationID] = &CompletedOperation{
			OperationID:   operationID,
			CompletedAt:   time.Now(),
			RetryCount:    0,
			LastAttemptAt: time.Now(),
		}
	}
}

// GetCompletionTime returns when the operation was completed
func (t *DefaultOperationTracker) GetCompletionTime(operationID string) (time.Time, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if op, exists := t.operations[operationID]; exists {
		return op.CompletedAt, true
	}
	return time.Time{}, false
}

// Clear removes all tracked operations
func (t *DefaultOperationTracker) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.operations = make(map[string]*CompletedOperation)
}

// ClearExpired removes operations older than the specified duration
func (t *DefaultOperationTracker) ClearExpired(maxAge time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for id, op := range t.operations {
		if op.CompletedAt.Before(cutoff) {
			delete(t.operations, id)
		}
	}
}

// Count returns the number of tracked operations
func (t *DefaultOperationTracker) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.operations)
}

// GenerateOperationID generates a deterministic operation ID based on operation parameters.
// This allows retries of the same operation to be detected as duplicates.
//
// Parameters:
//   - operationType: Type of operation (e.g., "delete", "move", "update_permission")
//   - resourceID: ID of the resource being operated on
//   - params: Additional parameters that uniquely identify the operation
//
// Returns a SHA-256 hash of the operation parameters.
//
// Requirements:
//   - Requirement 13.4: Add idempotent behavior for retry operations
func GenerateOperationID(operationType string, resourceID string, params ...string) string {
	hash := sha256.New()

	// Hash operation type
	hash.Write([]byte(operationType))
	hash.Write([]byte(":"))

	// Hash resource ID
	hash.Write([]byte(resourceID))
	hash.Write([]byte(":"))

	// Hash additional parameters
	for i, param := range params {
		if i > 0 {
			hash.Write([]byte(":"))
		}
		hash.Write([]byte(param))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// GenerateDeleteOperationID generates an operation ID for delete operations
func GenerateDeleteOperationID(resourceID string, permanent bool) string {
	permanentStr := "trash"
	if permanent {
		permanentStr = "permanent"
	}
	return GenerateOperationID("delete", resourceID, permanentStr)
}

// GenerateMoveOperationID generates an operation ID for move operations
func GenerateMoveOperationID(resourceID, targetParentID string) string {
	return GenerateOperationID("move", resourceID, targetParentID)
}

// GeneratePermissionUpdateOperationID generates an operation ID for permission update operations
func GeneratePermissionUpdateOperationID(fileID, permissionID, newRole string) string {
	return GenerateOperationID("update_permission", fileID, permissionID, newRole)
}

// GeneratePermissionDeleteOperationID generates an operation ID for permission deletion operations
func GeneratePermissionDeleteOperationID(fileID, permissionID string) string {
	return GenerateOperationID("delete_permission", fileID, permissionID)
}

// SafeExecute executes an operation with idempotency support.
// If the operation has already been completed (based on operationID), it skips execution.
// Otherwise, it executes the operation and marks it as completed.
//
// Parameters:
//   - tracker: Operation tracker to check for completion
//   - operationID: Unique identifier for the operation
//   - operation: Function to execute
//
// Returns the result of the operation or nil if already completed.
//
// Requirements:
//   - Requirement 13.4: Add idempotent behavior for retry operations
func SafeExecute(tracker OperationTracker, operationID string, operation func() error) error {
	// Check if operation already completed
	if tracker.IsOperationCompleted(operationID) {
		return nil
	}

	// Execute operation
	err := operation()
	if err != nil {
		return err
	}

	// Mark as completed
	tracker.MarkOperationCompleted(operationID)
	return nil
}

// SafeExecuteWithRetry executes an operation with retry logic and idempotency support.
// If the operation has already been completed, it skips execution.
// Otherwise, it retries the operation up to maxRetries times on failure.
//
// Parameters:
//   - tracker: Operation tracker to check for completion
//   - operationID: Unique identifier for the operation
//   - maxRetries: Maximum number of retry attempts
//   - operation: Function to execute
//
// Returns the result of the operation or an error if all retries fail.
//
// Requirements:
//   - Requirement 13.4: Add idempotent behavior for retry operations
func SafeExecuteWithRetry(tracker OperationTracker, operationID string, maxRetries int, operation func() error) error {
	// Check if operation already completed
	if tracker.IsOperationCompleted(operationID) {
		return nil
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := operation()
		if err == nil {
			// Success - mark as completed
			tracker.MarkOperationCompleted(operationID)
			return nil
		}

		lastErr = err

		// Don't retry if it's a non-retryable error
		if !isRetryableError(err) {
			break
		}

		// Wait before retry (exponential backoff could be added here)
		if attempt < maxRetries {
			time.Sleep(time.Second * time.Duration(attempt+1))
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", maxRetries+1, lastErr)
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	// This is a simplified implementation
	// In a real implementation, you would check for specific error types
	// (network errors, temporary failures, rate limits, etc.)
	if err == nil {
		return false
	}

	// For now, we'll retry on any error
	// This should be enhanced to only retry on transient errors
	return true
}
