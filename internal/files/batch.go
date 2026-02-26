package files

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dl-alexandre/gdrv/internal/safety"
	"github.com/dl-alexandre/gdrv/internal/types"
	"github.com/dl-alexandre/gdrv/internal/utils"
)

// BatchUploadOptions configures batch upload operations
type BatchUploadOptions struct {
	ParentID      string
	Workers       int
	ContinueOnErr bool
	ProgressFunc  func(index, total int, path string, success bool, err error)
}

// BatchUploadResult represents the result of a batch upload operation
type BatchUploadResult struct {
	SuccessCount int
	FailedCount  int
	TotalCount   int
	Files        []*types.DriveFile
	Errors       []BatchError
}

// BatchDownloadOptions configures batch download operations
type BatchDownloadOptions struct {
	OutputDir     string
	Workers       int
	ContinueOnErr bool
	MimeType      string
	KeepName      bool
	NamePattern   string
	ProgressFunc  func(index, total int, id, name string, success bool, err error)
}

// BatchDownloadResult represents the result of a batch download operation
type BatchDownloadResult struct {
	SuccessCount int
	FailedCount  int
	TotalCount   int
	Files        []DownloadedFile
	Errors       []BatchError
}

// DownloadedFile represents a downloaded file info
type DownloadedFile struct {
	ID           string
	Name         string
	OriginalName string
	Path         string
}

// BatchDeleteOptions configures batch delete operations
type BatchDeleteOptions struct {
	Permanent     bool
	Workers       int
	ContinueOnErr bool
	SafetyOpts    safety.SafetyOptions
	DryRun        bool
	ProgressFunc  func(index, total int, id string, success bool, err error)
}

// BatchDeleteResult represents the result of a batch delete operation
type BatchDeleteResult struct {
	SuccessCount int
	FailedCount  int
	TotalCount   int
	DeletedIDs   []string
	Errors       []BatchError
}

// BatchError represents an error for a specific item in a batch operation
type BatchError struct {
	Index int
	ID    string
	Path  string
	Error string
}

// BatchOperationRecorder records batch operations for dry-run mode
type BatchOperationRecorder struct {
	recorder safety.DryRunRecorder
	mu       sync.Mutex
}

// NewBatchOperationRecorder creates a new batch operation recorder
func NewBatchOperationRecorder() *BatchOperationRecorder {
	return &BatchOperationRecorder{
		recorder: safety.NewDryRunRecorder(),
	}
}

// RecordUpload records a planned upload operation
func (r *BatchOperationRecorder) RecordUpload(path, name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recorder.RecordOperation(safety.PlannedOperation{
		Type:         safety.OpTypeUpdate,
		ResourceID:   "",
		ResourceName: name,
		Description:  fmt.Sprintf("Upload: %s", path),
		Parameters: map[string]interface{}{
			"path": path,
			"name": name,
		},
	})
}

// RecordDelete records a planned delete operation
func (r *BatchOperationRecorder) RecordDelete(id, name string, permanent bool) {
	safety.RecordDelete(r.recorder, id, name, permanent)
}

// GetResult returns the dry-run result
func (r *BatchOperationRecorder) GetResult() *safety.DryRunResult {
	return safety.NewDryRunResult(r.recorder.GetOperations(), nil)
}

// BatchUpload uploads multiple files concurrently
func (m *Manager) BatchUpload(ctx context.Context, reqCtx *types.RequestContext, paths []string, opts BatchUploadOptions) (*BatchUploadResult, error) {
	result := &BatchUploadResult{
		TotalCount: len(paths),
		Files:      make([]*types.DriveFile, 0, len(paths)),
		Errors:     make([]BatchError, 0),
	}

	if len(paths) == 0 {
		return result, nil
	}

	// Validate workers
	workers := opts.Workers
	if workers <= 0 {
		workers = 1
	}
	if workers > 10 {
		workers = 10 // Cap at 10 concurrent uploads
	}

	// Create work queue
	type workItem struct {
		index int
		path  string
	}

	workQueue := make(chan workItem, len(paths))
	for i, path := range paths {
		workQueue <- workItem{index: i, path: path}
	}
	close(workQueue)

	// Results channel
	type workResult struct {
		index int
		path  string
		file  *types.DriveFile
		err   error
	}

	resultChan := make(chan workResult, len(paths))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range workQueue {
				uploadOpts := UploadOptions{
					ParentID: opts.ParentID,
				}

				file, err := m.Upload(ctx, reqCtx, item.path, uploadOpts)

				res := workResult{
					index: item.index,
					path:  item.path,
					file:  file,
					err:   err,
				}
				resultChan <- res

				if opts.ProgressFunc != nil {
					success := err == nil
					opts.ProgressFunc(item.index+1, len(paths), item.path, success, err)
				}
			}
		}()
	}

	// Close result channel when all workers done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for res := range resultChan {
		if res.err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, BatchError{
				Index: res.index,
				Path:  res.path,
				Error: res.err.Error(),
			})

			if !opts.ContinueOnErr {
				return result, res.err
			}
		} else {
			result.SuccessCount++
			result.Files = append(result.Files, res.file)
		}
	}

	return result, nil
}

// BatchDownload downloads multiple files concurrently
func (m *Manager) BatchDownload(ctx context.Context, reqCtx *types.RequestContext, fileIDs []string, opts BatchDownloadOptions) (*BatchDownloadResult, error) {
	result := &BatchDownloadResult{
		TotalCount: len(fileIDs),
		Files:      make([]DownloadedFile, 0, len(fileIDs)),
		Errors:     make([]BatchError, 0),
	}

	if len(fileIDs) == 0 {
		return result, nil
	}

	// Create output directory if specified and doesn't exist
	if opts.OutputDir != "" {
		if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
			return nil, utils.NewAppError(utils.NewCLIError(utils.ErrCodeInvalidArgument,
				fmt.Sprintf("Failed to create output directory: %s", err)).Build())
		}
	}

	// Validate workers
	workers := opts.Workers
	if workers <= 0 {
		workers = 1
	}
	if workers > 10 {
		workers = 10
	}

	// Create work queue
	type workItem struct {
		index  int
		fileID string
	}

	workQueue := make(chan workItem, len(fileIDs))
	for i, id := range fileIDs {
		workQueue <- workItem{index: i, fileID: id}
	}
	close(workQueue)

	// Results channel
	type workResult struct {
		index  int
		fileID string
		name   string
		path   string
		err    error
	}

	resultChan := make(chan workResult, len(fileIDs))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range workQueue {
				// Get file metadata first
				file, err := m.Get(ctx, reqCtx, item.fileID, "id,name,mimeType")
				if err != nil {
					resultChan <- workResult{
						index:  item.index,
						fileID: item.fileID,
						err:    err,
					}
					if opts.ProgressFunc != nil {
						opts.ProgressFunc(item.index+1, len(fileIDs), item.fileID, "", false, err)
					}
					continue
				}

				// Determine output path
				outputPath := m.determineOutputPath(file, opts)

				// Download
				downloadOpts := DownloadOptions{
					OutputPath: outputPath,
					MimeType:   opts.MimeType,
				}
				err = m.Download(ctx, reqCtx, item.fileID, downloadOpts)

				res := workResult{
					index:  item.index,
					fileID: item.fileID,
					name:   file.Name,
					path:   outputPath,
					err:    err,
				}
				resultChan <- res

				if opts.ProgressFunc != nil {
					success := err == nil
					opts.ProgressFunc(item.index+1, len(fileIDs), item.fileID, file.Name, success, err)
				}
			}
		}()
	}

	// Close result channel when all workers done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for res := range resultChan {
		if res.err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, BatchError{
				Index: res.index,
				ID:    res.fileID,
				Error: res.err.Error(),
			})

			if !opts.ContinueOnErr {
				return result, res.err
			}
		} else {
			result.SuccessCount++
			result.Files = append(result.Files, DownloadedFile{
				ID:           res.fileID,
				Name:         res.name,
				OriginalName: res.name,
				Path:         res.path,
			})
		}
	}

	return result, nil
}

// determineOutputPath determines the output path for a downloaded file
func (m *Manager) determineOutputPath(file *types.DriveFile, opts BatchDownloadOptions) string {
	var filename string

	if opts.KeepName {
		filename = file.Name
	} else if opts.NamePattern != "" {
		// Simple pattern replacement: {id}, {name}, {index}
		filename = opts.NamePattern
		filename = strings.ReplaceAll(filename, "{id}", file.ID)
		filename = strings.ReplaceAll(filename, "{name}", file.Name)
		// {index} would need to be passed in, so we skip it for now
	} else {
		filename = file.Name
	}

	// Sanitize filename
	filename = sanitizeFilename(filename)

	if opts.OutputDir != "" {
		return filepath.Join(opts.OutputDir, filename)
	}
	return filename
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	// Replace invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}

// BatchDelete deletes multiple files concurrently
func (m *Manager) BatchDelete(ctx context.Context, reqCtx *types.RequestContext, fileIDs []string, opts BatchDeleteOptions) (*BatchDeleteResult, error) {
	result := &BatchDeleteResult{
		TotalCount: len(fileIDs),
		DeletedIDs: make([]string, 0, len(fileIDs)),
		Errors:     make([]BatchError, 0),
	}

	if len(fileIDs) == 0 {
		return result, nil
	}

	var recorder safety.DryRunRecorder
	if opts.DryRun {
		recorder = safety.NewDryRunRecorder()
	}

	// Validate workers
	workers := opts.Workers
	if workers <= 0 {
		workers = 1
	}
	if workers > 10 {
		workers = 10
	}

	// Create work queue
	type workItem struct {
		index  int
		fileID string
	}

	workQueue := make(chan workItem, len(fileIDs))
	for i, id := range fileIDs {
		workQueue <- workItem{index: i, fileID: id}
	}
	close(workQueue)

	// Results channel
	type workResult struct {
		index  int
		fileID string
		err    error
	}

	resultChan := make(chan workResult, len(fileIDs))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range workQueue {
				var err error

				if opts.DryRun {
					// In dry-run mode, get file metadata and record
					file, err := m.Get(ctx, reqCtx, item.fileID, "id,name")
					if err == nil {
						safety.RecordDelete(recorder, item.fileID, file.Name, opts.Permanent)
					}
				} else {
					err = m.DeleteWithSafety(ctx, reqCtx, item.fileID, opts.Permanent, opts.SafetyOpts, recorder)
				}

				res := workResult{
					index:  item.index,
					fileID: item.fileID,
					err:    err,
				}
				resultChan <- res

				if opts.ProgressFunc != nil {
					success := err == nil
					opts.ProgressFunc(item.index+1, len(fileIDs), item.fileID, success, err)
				}
			}
		}()
	}

	// Close result channel when all workers done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for res := range resultChan {
		if res.err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, BatchError{
				Index: res.index,
				ID:    res.fileID,
				Error: res.err.Error(),
			})

			if !opts.ContinueOnErr {
				return result, res.err
			}
		} else {
			result.SuccessCount++
			result.DeletedIDs = append(result.DeletedIDs, res.fileID)
		}
	}

	return result, nil
}

// LoadFileList loads file paths from a JSON or text file
func LoadFileList(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file list: %w", err)
	}

	// Try JSON first
	var jsonPaths []string
	if err := json.Unmarshal(data, &jsonPaths); err == nil {
		return jsonPaths, nil
	}

	// Try JSON object with "files" field
	var jsonObj struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal(data, &jsonObj); err == nil && len(jsonObj.Files) > 0 {
		return jsonObj.Files, nil
	}

	// Fall back to line-by-line text file
	var paths []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			paths = append(paths, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse file list: %w", err)
	}

	return paths, nil
}

// LoadIDList loads file IDs from a JSON or text file
func LoadIDList(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ID list: %w", err)
	}

	// Try JSON first
	var jsonIDs []string
	if err := json.Unmarshal(data, &jsonIDs); err == nil {
		return jsonIDs, nil
	}

	// Try JSON object with "ids" field
	var jsonObj struct {
		IDs []string `json:"ids"`
	}
	if err := json.Unmarshal(data, &jsonObj); err == nil && len(jsonObj.IDs) > 0 {
		return jsonObj.IDs, nil
	}

	// Fall back to line-by-line text file
	var ids []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			ids = append(ids, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse ID list: %w", err)
	}

	return ids, nil
}

// DefaultProgressFunc creates a default progress function that prints to stdout
func DefaultProgressFunc(operation string) func(index, total int, idOrPath, name string, success bool, err error) {
	startTime := time.Now()
	var mu sync.Mutex
	var lastProgress int

	return func(index, total int, idOrPath, name string, success bool, err error) {
		mu.Lock()
		defer mu.Unlock()

		// Only print progress at certain intervals to avoid spam
		progress := (index * 100) / total
		if progress != lastProgress || index == total || index == 1 {
			lastProgress = progress

			elapsed := time.Since(startTime)
			var eta time.Duration
			if index > 0 {
				rate := elapsed / time.Duration(index)
				remaining := total - index
				eta = rate * time.Duration(remaining)
			}

			status := "✓"
			if !success {
				status = "✗"
			}

			display := idOrPath
			if name != "" {
				display = name
			}

			fmt.Printf("\r[%3d%%] %s %s (%d/%d) - ETA: %s",
				progress, status, display, index, total, formatDuration(eta))

			if index == total {
				fmt.Println() // New line at end
			}
		}
	}
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "<1s"
	}
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
