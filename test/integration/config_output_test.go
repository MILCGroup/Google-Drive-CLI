// +build integration

package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dl-alexandre/gdrive/internal/api"
	"github.com/dl-alexandre/gdrive/internal/auth"
	"github.com/dl-alexandre/gdrive/internal/config"
	"github.com/dl-alexandre/gdrive/internal/files"
	"github.com/dl-alexandre/gdrive/internal/types"
)

// TestIntegration_ConfigOutput_ConfigurationLoading tests configuration loading and validation
func TestIntegration_ConfigOutput_ConfigurationLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test default config loading
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if cfg.DefaultOutputFormat == "" {
		t.Error("Default output format not set")
	}

	// Test environment variable override
	os.Setenv("GDRIVE_DEFAULT_PROFILE", "test-profile")
	defer os.Unsetenv("GDRIVE_DEFAULT_PROFILE")

	cfg2, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config with env vars: %v", err)
	}

	if cfg2.DefaultProfile != "test-profile" {
		t.Errorf("Expected default profile 'test-profile', got '%s'", cfg2.DefaultProfile)
	}
}

// TestIntegration_ConfigOutput_OutputFormats tests JSON vs table output formats
func TestIntegration_ConfigOutput_OutputFormats(t *testing.T) {
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

	// Create a test file
	fileName := "output-test-" + time.Now().Format("20060102150405") + ".txt"
	file, err := fileManager.Create(ctx, reqCtx, fileName, "", "text/plain", nil)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test JSON output
	jsonOutput := config.NewOutputFormatter(config.OutputOptions{
		Format: types.OutputFormatJSON,
	})
	err = jsonOutput.WriteSuccess("files.list", []*types.DriveFile{file})
	if err != nil {
		t.Fatalf("Failed to output in JSON format: %v", err)
	}

	// Test table output
	tableOutput := config.NewOutputFormatter(config.OutputOptions{
		Format: types.OutputFormatTable,
	})
	err = tableOutput.WriteSuccess("files.list", []*types.DriveFile{file})
	if err != nil {
		t.Fatalf("Failed to output in table format: %v", err)
	}

	// Clean up
	err = fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete test file: %v", err)
	}
}

// TestIntegration_ConfigOutput_FieldMasking tests field masking and export links
func TestIntegration_ConfigOutput_FieldMasking(t *testing.T) {
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

	// Create a test file
	fileName := "field-test-" + time.Now().Format("20060102150405") + ".txt"
	file, err := fileManager.Create(ctx, reqCtx, fileName, "", "text/plain", nil)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	minimalMask := config.GetFieldMask(config.FieldMaskMinimal, false)
	if minimalMask == "" {
		t.Fatal("Expected minimal field mask to be set")
	}
	if strings.Contains(minimalMask, "exportLinks") {
		t.Fatal("Expected minimal field mask to exclude exportLinks")
	}

	exportMask := config.GetFieldMask(config.FieldMaskStandard, true)
	if exportMask == "" {
		t.Fatal("Expected export field mask to be set")
	}
	if !strings.Contains(exportMask, "exportLinks") {
		t.Fatal("Expected export field mask to include exportLinks")
	}

	// Clean up
	err = fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete test file: %v", err)
	}
}

// TestIntegration_ConfigOutput_QuietVerboseModes tests quiet/verbose modes
func TestIntegration_ConfigOutput_QuietVerboseModes(t *testing.T) {
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

	// Create a test file
	fileName := "quiet-verbose-test-" + time.Now().Format("20060102150405") + ".txt"
	file, err := fileManager.Create(ctx, reqCtx, fileName, "", "text/plain", nil)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test quiet mode (no output)
	quietOutput := config.NewOutputFormatter(config.OutputOptions{
		Format: types.OutputFormatTable,
		Quiet:  true,
	})
	err = quietOutput.WriteSuccess("files.list", []*types.DriveFile{file})
	if err != nil {
		t.Fatalf("Failed to output in quiet mode: %v", err)
	}

	// Test verbose mode (with extra info)
	verboseOutput := config.NewOutputFormatter(config.OutputOptions{
		Format:  types.OutputFormatTable,
		Verbose: true,
	})
	err = verboseOutput.WriteSuccess("files.list", []*types.DriveFile{file})
	if err != nil {
		t.Fatalf("Failed to output in verbose mode: %v", err)
	}

	// Clean up
	err = fileManager.Delete(ctx, reqCtx, file.ID, false)
	if err != nil {
		t.Errorf("Failed to delete test file: %v", err)
	}
}</content>
<parameter name="filePath">test/integration/config_output_test.go