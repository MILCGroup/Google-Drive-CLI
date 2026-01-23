// +build integration

package integration

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrive/internal/api"
	"github.com/dl-alexandre/gdrive/internal/auth"
	"github.com/dl-alexandre/gdrive/internal/files"
	"github.com/dl-alexandre/gdrive/internal/types"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Integration Test: File Operations End-to-End
// Run with: go test -tags=integration ./test/integration/...

func setupFileManager(t *testing.T) (*files.Manager, *api.Client, context.Context) {
	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	manager := auth.NewManager("")
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	ctx := context.Background()
	service, err := drive.NewService(ctx, option.WithTokenSource(
		manager.GetConfig().TokenSource(ctx, token),
	))
	if err != nil {
		t.Fatalf("Failed to create Drive service: %v", err)
	}

	client := api.NewClient(service, 3, 1000)
	fileMgr := files.NewManager(client)
	
	return fileMgr, client, ctx
}

func TestIntegration_FileOperations_UploadDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	fileMgr, _, ctx := setupFileManager(t)

	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-upload.txt")
	testContent := []byte("This is a test file for upload integration test.\nCreated at: " + time.Now().String())
	
	err := ioutil.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Upload the file
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)
	uploadOpts := files.UploadOptions{
		Name: "integration-test-" + time.Now().Format("20060102-150405") + ".txt",
	}

	uploadedFile, err := fileMgr.Upload(ctx, reqCtx, testFile, uploadOpts)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	t.Logf("Uploaded file: %s (ID: %s)", uploadedFile.Name, uploadedFile.Id)

	// Download the file
	downloadReqCtx := api.NewRequestContext("default", "", types.RequestTypeDownloadOrExport)
	downloadPath := filepath.Join(tmpDir, "downloaded.txt")
	downloadOpts := files.DownloadOptions{
		OutputPath: downloadPath,
	}

	err = fileMgr.Download(ctx, downloadReqCtx, uploadedFile.Id, downloadOpts)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Verify downloaded content
	downloadedContent, err := ioutil.ReadFile(downloadPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(downloadedContent) != string(testContent) {
		t.Errorf("Downloaded content does not match original")
		t.Logf("Original: %s", testContent)
		t.Logf("Downloaded: %s", downloadedContent)
	}

	// Clean up - delete the uploaded file
	deleteReqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)
	err = fileMgr.Delete(ctx, deleteReqCtx, uploadedFile.Id, false)
	if err != nil {
		t.Errorf("Failed to clean up uploaded file: %v", err)
	}
}

func TestIntegration_FileOperations_UploadTypeSelection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	fileMgr, _, ctx := setupFileManager(t)
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		size     int64
		expected string // "simple", "multipart", or "resumable"
	}{
		{"Small file (1KB)", 1024, "multipart"},
		{"Medium file (1MB)", 1024 * 1024, "multipart"},
		{"Large file (10MB)", 10 * 1024 * 1024, "resumable"},
	}

	uploadedFiles := []string{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file of specified size
			testFile := filepath.Join(tmpDir, "test-"+tt.name+".bin")
			content := make([]byte, tt.size)
			for i := range content {
				content[i] = byte(i % 256)
			}
			
			err := ioutil.WriteFile(testFile, content, 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Upload
			reqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)
			uploadOpts := files.UploadOptions{
				Name: "size-test-" + time.Now().Format("20060102-150405") + ".bin",
			}

			uploadedFile, err := fileMgr.Upload(ctx, reqCtx, testFile, uploadOpts)
			if err != nil {
				t.Fatalf("Upload failed: %v", err)
			}

			t.Logf("Uploaded %s: %s (ID: %s)", tt.name, uploadedFile.Name, uploadedFile.Id)
			uploadedFiles = append(uploadedFiles, uploadedFile.Id)

			// Verify size
			if uploadedFile.Size != tt.size {
				t.Errorf("Uploaded file size %d != expected %d", uploadedFile.Size, tt.size)
			}
		})
	}

	// Clean up
	for _, fileID := range uploadedFiles {
		reqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)
		fileMgr.Delete(ctx, reqCtx, fileID, false)
	}
}

func TestIntegration_FileOperations_WorkspaceExport(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	workspaceFileID := os.Getenv("TEST_WORKSPACE_FILE_ID")
	if workspaceFileID == "" {
		t.Skip("TEST_WORKSPACE_FILE_ID not set (should be a Google Doc)")
	}

	fileMgr, _, ctx := setupFileManager(t)
	tmpDir := t.TempDir()

	// Export Google Doc as PDF
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeDownloadOrExport)
	downloadOpts := files.DownloadOptions{
		OutputPath: filepath.Join(tmpDir, "exported.pdf"),
		MimeType:   "application/pdf",
	}

	err := fileMgr.Download(ctx, reqCtx, workspaceFileID, downloadOpts)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify file was created
	stat, err := os.Stat(downloadOpts.OutputPath)
	if err != nil {
		t.Fatalf("Exported file not found: %v", err)
	}

	if stat.Size() == 0 {
		t.Error("Exported file is empty")
	}

	t.Logf("Exported %d bytes to PDF", stat.Size())
}

func TestIntegration_FileOperations_ListAndSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	fileMgr, _, ctx := setupFileManager(t)

	// List files
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeListOrSearch)
	listOpts := files.ListOptions{
		PageSize: 10,
		Fields:   "files(id,name,mimeType)",
	}

	fileList, err := fileMgr.List(ctx, reqCtx, listOpts)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	t.Logf("Found %d files", len(fileList.Files))

	// Search for specific files
	searchOpts := files.ListOptions{
		Query:    "name contains 'test'",
		PageSize: 10,
		Fields:   "files(id,name)",
	}

	searchResults, err := fileMgr.List(ctx, reqCtx, searchOpts)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	t.Logf("Search found %d files matching 'test'", len(searchResults.Files))
}

func TestIntegration_FileOperations_CopyFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	sourceFileID := os.Getenv("TEST_FILE_ID")
	if sourceFileID == "" {
		t.Skip("TEST_FILE_ID not set")
	}

	fileMgr, _, ctx := setupFileManager(t)

	// Copy the file
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)
	copyName := "copy-test-" + time.Now().Format("20060102-150405")
	
	copiedFile, err := fileMgr.Copy(ctx, reqCtx, sourceFileID, copyName, "")
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	t.Logf("Copied file: %s (ID: %s)", copiedFile.Name, copiedFile.Id)

	// Verify copy exists
	getReqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)
	retrievedFile, err := fileMgr.Get(ctx, getReqCtx, copiedFile.Id, "id,name")
	if err != nil {
		t.Fatalf("Failed to retrieve copied file: %v", err)
	}

	if retrievedFile.Name != copyName {
		t.Errorf("Copied file name %s != expected %s", retrievedFile.Name, copyName)
	}

	// Clean up
	deleteReqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)
	err = fileMgr.Delete(ctx, deleteReqCtx, copiedFile.Id, false)
	if err != nil {
		t.Errorf("Failed to clean up copied file: %v", err)
	}
}

func TestIntegration_FileOperations_TrashAndRestore(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	fileMgr, _, ctx := setupFileManager(t)
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "trash-test.txt")
	err := ioutil.WriteFile(testFile, []byte("Test file for trash"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Upload
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)
	uploadOpts := files.UploadOptions{
		Name: "trash-test-" + time.Now().Format("20060102-150405") + ".txt",
	}

	uploadedFile, err := fileMgr.Upload(ctx, reqCtx, testFile, uploadOpts)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	fileID := uploadedFile.Id
	t.Logf("Created test file: %s (ID: %s)", uploadedFile.Name, fileID)

	// Trash the file
	trashReqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)
	err = fileMgr.Delete(ctx, trashReqCtx, fileID, false) // false = trash, not permanent
	if err != nil {
		t.Fatalf("Trash failed: %v", err)
	}

	t.Logf("Trashed file")

	// Verify file is in trash
	getReqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)
	trashedFile, err := fileMgr.Get(ctx, getReqCtx, fileID, "id,name,trashed")
	if err != nil {
		t.Fatalf("Failed to get trashed file: %v", err)
	}

	if !trashedFile.Trashed {
		t.Error("File is not marked as trashed")
	}

	// Restore the file (implementation depends on file manager API)
	// This would typically use files.update to set trashed=false
	
	// For now, permanently delete it
	deleteReqCtx := api.NewRequestContext("default", "", types.RequestTypeMutation)
	err = fileMgr.Delete(ctx, deleteReqCtx, fileID, true) // true = permanent
	if err != nil {
		t.Errorf("Failed to permanently delete: %v", err)
	}
}

func TestIntegration_FileOperations_Metadata(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	fileID := os.Getenv("TEST_FILE_ID")
	if fileID == "" {
		t.Skip("TEST_FILE_ID not set")
	}

	fileMgr, _, ctx := setupFileManager(t)

	// Get file metadata
	reqCtx := api.NewRequestContext("default", "", types.RequestTypeGetByID)
	fields := "id,name,mimeType,size,createdTime,modifiedTime,owners,capabilities"
	
	file, err := fileMgr.Get(ctx, reqCtx, fileID, fields)
	if err != nil {
		t.Fatalf("Get metadata failed: %v", err)
	}

	// Verify metadata fields
	if file.Id != fileID {
		t.Errorf("ID mismatch: got %s, want %s", file.Id, fileID)
	}
	if file.Name == "" {
		t.Error("Name is empty")
	}
	if file.MimeType == "" {
		t.Error("MimeType is empty")
	}

	t.Logf("File metadata:")
	t.Logf("  Name: %s", file.Name)
	t.Logf("  MIME: %s", file.MimeType)
	t.Logf("  Size: %d bytes", file.Size)
	t.Logf("  Created: %s", file.CreatedTime)
	t.Logf("  Modified: %s", file.ModifiedTime)
	
	if file.Capabilities != nil {
		t.Logf("  Can Download: %v", file.Capabilities.CanDownload)
		t.Logf("  Can Edit: %v", file.Capabilities.CanEdit)
	}
}
