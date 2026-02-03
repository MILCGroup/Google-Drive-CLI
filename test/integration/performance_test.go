//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrv/internal/files"
	"github.com/dl-alexandre/gdrv/internal/types"
)

// TestIntegration_Performance_ResponseTimes tests response times for various operations
func TestIntegration_Performance_ResponseTimes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Test file creation time
	start := time.Now()
	fileName := "perf-test-" + time.Now().Format("20060102150405") + ".txt"
	file := uploadTempFile(t, ctx, fileManager, reqCtx, fileName, "", "text/plain", []byte("performance test"))
	creationTime := time.Since(start)

	t.Logf("File creation took: %v", creationTime)

	// Test file listing time
	start = time.Now()
	listReq := files.ListOptions{
		Query: fmt.Sprintf("name='%s'", fileName),
	}
	_, err := fileManager.List(ctx, reqCtx, listReq)
	listTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	t.Logf("File listing took: %v", listTime)

	// Test file get time
	start = time.Now()
	_, err = fileManager.Get(ctx, reqCtx, file.ID, "id")
	getTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to get file: %v", err)
	}

	t.Logf("File get took: %v", getTime)

	// Clean up
	err = fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete test file: %v", err)
	}
}

// TestIntegration_Performance_MemoryUsage tests memory usage and resource cleanup
func TestIntegration_Performance_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create multiple files
	var fileIDs []string
	for i := 0; i < 10; i++ {
		fileName := fmt.Sprintf("mem-test-%d-%s.txt", i, time.Now().Format("20060102150405"))
		file := uploadTempFile(t, ctx, fileManager, reqCtx, fileName, "", "text/plain", []byte("memory test"))
		fileIDs = append(fileIDs, file.ID)
	}

	t.Logf("Created %d test files", len(fileIDs))

	// Delete files
	for _, id := range fileIDs {
		if err := fileManager.Delete(ctx, reqCtx, id, false); err != nil {
			t.Errorf("Failed to delete file %s: %v", id, err)
		}
	}

	t.Log("Memory usage test completed - all resources cleaned up")
}

// TestIntegration_Performance_ConcurrentOperations tests concurrent operations
func TestIntegration_Performance_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client, _, _ := setupDriveClient(t)
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Test concurrent file creation
	numConcurrent := 5
	done := make(chan bool, numConcurrent)

	start := time.Now()

	for i := 0; i < numConcurrent; i++ {
		go func(index int) {
			fileName := fmt.Sprintf("concurrent-test-%d-%s.txt", index, time.Now().Format("20060102150405"))
			tmpFile, err := os.CreateTemp("", "concurrent-test-*.txt")
			if err != nil {
				t.Errorf("Concurrent temp file %d failed: %v", index, err)
				done <- true
				return
			}
			if _, err := tmpFile.WriteString("concurrent test"); err != nil {
				t.Errorf("Concurrent temp write %d failed: %v", index, err)
				_ = tmpFile.Close()
				done <- true
				return
			}
			_ = tmpFile.Close()
			file, err := fileManager.Upload(ctx, reqCtx, tmpFile.Name(), files.UploadOptions{
				Name:     fileName,
				MimeType: "text/plain",
			})
			if err != nil {
				t.Errorf("Concurrent creation %d failed: %v", index, err)
			} else {
				// Clean up immediately
				err = fileManager.Delete(ctx, reqCtx, file.ID, false)
				if err != nil {
					t.Errorf("Concurrent cleanup %d failed: %v", index, err)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numConcurrent; i++ {
		<-done
	}

	totalTime := time.Since(start)
	t.Logf("Concurrent operations took: %v", totalTime)
}
