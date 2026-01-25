// +build integration

package integration

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrv/internal/api"
	"github.com/dl-alexandre/gdrv/internal/auth"
	"github.com/dl-alexandre/gdrv/internal/files"
	"github.com/dl-alexandre/gdrv/internal/types"
)

// TestIntegration_CLIIntegration_CommandExecution tests CLI commands with various flag combinations
func TestIntegration_CLIIntegration_CommandExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	// Test help command
	cmd := exec.Command("./gdrv", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Help command failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "gdrv") {
		t.Error("Help output doesn't contain expected content")
	}

	// Test version/info command if available
	cmd = exec.Command("./gdrv", "about")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("About command failed (might not be implemented): %v", err)
	} else {
		t.Logf("About command output: %s", output)
	}
}

// TestIntegration_CLIIntegration_AuthCommands tests auth-related CLI commands
func TestIntegration_CLIIntegration_AuthCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	// Test auth status
	cmd := exec.Command("./gdrv", "auth", "status", "--profile", profile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Auth status failed: %v\nOutput: %s", err, output)
	}

	t.Logf("Auth status output: %s", output)
}

// TestIntegration_CLIIntegration_FileCommands tests file-related CLI commands
func TestIntegration_CLIIntegration_FileCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	ctx := context.Background()

	// First create a test file using the API directly
	manager := auth.NewManager("")
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	client := api.NewClient(token)
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	fileName := "cli-test-" + time.Now().Format("20060102150405") + ".txt"
	testFile, err := fileManager.Create(ctx, reqCtx, fileName, "", "text/plain", []byte("CLI test content"))
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test files list command
	cmd := exec.Command("./gdrv", "files", "list", "--profile", profile, "--query", "name='"+fileName+"'")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Files list failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), fileName) {
		t.Errorf("File list output doesn't contain expected file: %s", output)
	}

	// Test files get command
	cmd = exec.Command("./gdrv", "files", "get", testFile.ID, "--profile", profile)
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Files get failed: %v\nOutput: %s", err, output)
	}

	// Clean up
	err = fileManager.Delete(ctx, reqCtx, testFile.ID, false)
	if err != nil {
		t.Errorf("Failed to clean up test file: %v", err)
	}
}

// TestIntegration_CLIIntegration_OutputFormats tests output format options
func TestIntegration_CLIIntegration_OutputFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	// Test JSON output
	cmd := exec.Command("./gdrv", "files", "list", "--profile", profile, "--json", "--page-size", "1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("JSON output failed: %v\nOutput: %s", err, output)
	}

	if !strings.Contains(string(output), "{") || !strings.Contains(string(output), "}") {
		t.Error("JSON output doesn't look like JSON")
	}

	// Test table output
	cmd = exec.Command("./gdrv", "files", "list", "--profile", profile, "--table", "--page-size", "1")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Table output failed: %v\nOutput: %s", err, output)
	}

	t.Logf("Table output: %s", output)
}

// TestIntegration_CLIIntegration_SafetyControls tests safety controls via CLI
func TestIntegration_CLIIntegration_SafetyControls(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	ctx := context.Background()

	// Create test file
	manager := auth.NewManager("")
	token, err := manager.GetToken(profile)
	if err != nil {
		t.Fatalf("Failed to get token: %v", err)
	}

	client := api.NewClient(token)
	fileManager := files.NewManager(client)
	reqCtx := &types.RequestContext{}

	fileName := "safety-test-" + time.Now().Format("20060102150405") + ".txt"
	testFile, err := fileManager.Create(ctx, reqCtx, fileName, "", "text/plain", nil)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test dry-run delete
	cmd := exec.Command("./gdrv", "files", "delete", testFile.ID, "--profile", profile, "--dry-run")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Dry-run delete failed: %v\nOutput: %s", err, output)
	}

	t.Logf("Dry-run output: %s", output)

	// Verify file still exists
	_, err = fileManager.Get(ctx, reqCtx, testFile.ID)
	if err != nil {
		t.Fatalf("File should still exist after dry-run: %v", err)
	}

	// Real delete
	cmd = exec.Command("./gdrv", "files", "delete", testFile.ID, "--profile", profile, "--yes")
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Real delete failed: %v\nOutput: %s", err, output)
	}

	t.Log("Safety controls test completed")
}

// TestIntegration_CLIIntegration_ConfigurationLoading tests configuration loading from CLI
func TestIntegration_CLIIntegration_ConfigurationLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	profile := os.Getenv("TEST_PROFILE")
	if profile == "" {
		t.Skip("TEST_PROFILE not set")
	}

	// Test with verbose flag
	cmd := exec.Command("./gdrv", "files", "list", "--profile", profile, "--verbose", "--page-size", "1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Verbose command failed: %v\nOutput: %s", err, output)
	}

	t.Logf("Verbose output length: %d", len(output))
}</content>
<parameter name="filePath">test/integration/cli_integration_test.go