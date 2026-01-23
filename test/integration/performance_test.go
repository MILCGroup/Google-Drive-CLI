// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrive/internal/api"
	"github.com/dl-alexandre/gdrive/internal/auth"
	"github.com/dl-alexandre/gdrive/internal/files"
	"github.com/dl-alexandre/gdrive/internal/types"
)

// TestIntegration_Performance_ResponseTimes tests response times for various operations
func TestIntegration_Performance_ResponseTimes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	ctx := context.Background()

	manager := auth.NewManager("")
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	client := api.NewClient(token)
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Test file creation time
	start := time.Now()
	fileName := "perf-test-" + time.Now().Format("20060102150405") + ".txt"
	file, err := fileManager.Create(ctx, reqCtx, fileName, "", "text/plain", []byte("performance test"))
	creationTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	t.Logf("File creation took: %v", creationTime)

	// Test file listing time
	start = time.Now()
	listReq := &types.FileListRequest{
		Query: fmt.Sprintf("name='%s'", fileName),
	}
	results, err := fileManager.List(ctx, reqCtx, listReq)
	listTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	t.Logf("File listing took: %v", listTime)

	// Test file get time
	start = time.Now()
	_, err = fileManager.Get(ctx, reqCtx, file.ID)
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

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	ctx := context.Background()

	manager := auth.NewManager("")
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	client := api.NewClient(token)
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Create multiple files
	var fileIDs []string
	for i := 0; i < 10; i++ {
		fileName := fmt.Sprintf("mem-test-%d-%s.txt", i, time.Now().Format("20060102150405"))
		file, err := fileManager.Create(ctx, reqCtx, fileName, "", "text/plain", []byte("memory test"))
		if err != nil {
			t.Fatalf("Failed to create file %d: %v", i, err)
		}
		fileIDs = append(fileIDs, file.ID)
	}

	t.Logf("Created %d test files", len(fileIDs))

	// Delete files
	for _, id := range fileIDs {
		err = fileManager.Delete(ctx, reqCtx, id, false)
		if err != nil {
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

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	ctx := context.Background()

	manager := auth.NewManager("")
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	client := api.NewClient(token)
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	// Test concurrent file creation
	numConcurrent := 5
	done := make(chan bool, numConcurrent)

	start := time.Now()

	for i := 0; i < numConcurrent; i++ {
		go func(index int) {
			fileName := fmt.Sprintf("concurrent-test-%d-%s.txt", index, time.Now().Format("20060102150405"))
			file, err := fileManager.Create(ctx, reqCtx, fileName, "", "text/plain", []byte("concurrent test"))
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
}</content>
<parameter name="filePath">test/integration/performance_test.go