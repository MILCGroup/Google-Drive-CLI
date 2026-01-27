package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()

	if info == nil {
		t.Fatal("Get() returned nil")
	}

	// Test that all fields are populated
	if info.Version == "" {
		t.Error("Version should not be empty")
	}

	if info.GitCommit == "" {
		t.Error("GitCommit should not be empty")
	}

	if info.BuildTime == "" {
		t.Error("BuildTime should not be empty")
	}

	if info.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}

	if info.Platform == "" {
		t.Error("Platform should not be empty")
	}

	// Verify GoVersion matches runtime
	if info.GoVersion != runtime.Version() {
		t.Errorf("GoVersion = %s, want %s", info.GoVersion, runtime.Version())
	}

	// Verify Platform format
	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if info.Platform != expectedPlatform {
		t.Errorf("Platform = %s, want %s", info.Platform, expectedPlatform)
	}
}

func TestGet_UsesGlobalVariables(t *testing.T) {
	// Store original values
	origVersion := Version
	origGitCommit := GitCommit
	origBuildTime := BuildTime

	// Temporarily modify globals
	Version = "test-version"
	GitCommit = "abc123"
	BuildTime = "2024-01-01"

	info := Get()

	if info.Version != "test-version" {
		t.Errorf("Version = %s, want test-version", info.Version)
	}

	if info.GitCommit != "abc123" {
		t.Errorf("GitCommit = %s, want abc123", info.GitCommit)
	}

	if info.BuildTime != "2024-01-01" {
		t.Errorf("BuildTime = %s, want 2024-01-01", info.BuildTime)
	}

	// Restore original values
	Version = origVersion
	GitCommit = origGitCommit
	BuildTime = origBuildTime
}

func TestInfo_String(t *testing.T) {
	tests := []struct {
		name      string
		info      *Info
		wantParts []string
	}{
		{
			name: "Complete info",
			info: &Info{
				Version:   "1.0.0",
				GitCommit: "abc123",
				BuildTime: "2024-01-01T00:00:00Z",
				GoVersion: "go1.21.0",
				Platform:  "linux/amd64",
			},
			wantParts: []string{"gdrv", "1.0.0", "abc123", "built", "2024-01-01T00:00:00Z"},
		},
		{
			name: "Dev version",
			info: &Info{
				Version:   "dev",
				GitCommit: "unknown",
				BuildTime: "unknown",
				GoVersion: "go1.21.0",
				Platform:  "darwin/arm64",
			},
			wantParts: []string{"gdrv", "dev", "unknown", "built", "unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.info.String()

			if got == "" {
				t.Error("String() returned empty string")
			}

			for _, part := range tt.wantParts {
				if !strings.Contains(got, part) {
					t.Errorf("String() = %q, want to contain %q", got, part)
				}
			}
		})
	}
}

func TestInfo_String_Format(t *testing.T) {
	info := &Info{
		Version:   "1.0.0",
		GitCommit: "abc123",
		BuildTime: "2024-01-01",
		GoVersion: "go1.21.0",
		Platform:  "linux/amd64",
	}

	got := info.String()
	want := "gdrv 1.0.0 (abc123) built 2024-01-01"

	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestInfo_Short(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"Release version", "1.0.0"},
		{"Dev version", "dev"},
		{"Prerelease version", "1.0.0-beta.1"},
		{"Commit version", "v1.2.3-45-gabcdef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &Info{
				Version:   tt.version,
				GitCommit: "abc123",
				BuildTime: "2024-01-01",
				GoVersion: "go1.21.0",
				Platform:  "linux/amd64",
			}

			got := info.Short()
			if got != tt.version {
				t.Errorf("Short() = %q, want %q", got, tt.version)
			}
		})
	}
}

func TestInfo_Short_OnlyReturnsVersion(t *testing.T) {
	info := &Info{
		Version:   "1.0.0",
		GitCommit: "abc123",
		BuildTime: "2024-01-01",
		GoVersion: "go1.21.0",
		Platform:  "linux/amd64",
	}

	got := info.Short()

	// Should only contain version, not commit or build time
	if strings.Contains(got, "abc123") {
		t.Error("Short() should not contain git commit")
	}

	if strings.Contains(got, "2024-01-01") {
		t.Error("Short() should not contain build time")
	}

	if strings.Contains(got, "gdrv") {
		t.Error("Short() should not contain 'gdrv' prefix")
	}

	if got != "1.0.0" {
		t.Errorf("Short() = %q, want %q", got, "1.0.0")
	}
}

func TestDefaultValues(t *testing.T) {
	// Test that default values are set (these may be overridden at build time)
	if Version == "" {
		t.Error("Version should have a default value")
	}

	if GitCommit == "" {
		t.Error("GitCommit should have a default value")
	}

	if BuildTime == "" {
		t.Error("BuildTime should have a default value")
	}
}

func TestInfo_AllFieldsPopulated(t *testing.T) {
	info := Get()

	// Use reflection to ensure all string fields are non-empty
	if info.Version == "" {
		t.Error("Info.Version is empty")
	}
	if info.GitCommit == "" {
		t.Error("Info.GitCommit is empty")
	}
	if info.BuildTime == "" {
		t.Error("Info.BuildTime is empty")
	}
	if info.GoVersion == "" {
		t.Error("Info.GoVersion is empty")
	}
	if info.Platform == "" {
		t.Error("Info.Platform is empty")
	}
}

func TestInfo_PlatformFormat(t *testing.T) {
	info := Get()

	// Platform should be in format "os/arch"
	parts := strings.Split(info.Platform, "/")
	if len(parts) != 2 {
		t.Errorf("Platform format invalid: %q, want os/arch", info.Platform)
	}

	// First part should be a valid OS
	validOS := map[string]bool{
		"darwin": true, "linux": true, "windows": true,
		"freebsd": true, "openbsd": true, "netbsd": true,
		"android": true, "ios": true,
	}
	if !validOS[parts[0]] && parts[0] != runtime.GOOS {
		t.Logf("Platform OS: %q (may be valid but uncommon)", parts[0])
	}

	// Second part should be a valid arch
	validArch := map[string]bool{
		"amd64": true, "arm64": true, "386": true,
		"arm": true, "ppc64le": true, "s390x": true,
	}
	if !validArch[parts[1]] && parts[1] != runtime.GOARCH {
		t.Logf("Platform arch: %q (may be valid but uncommon)", parts[1])
	}
}

func TestInfo_ConsistencyBetweenMethods(t *testing.T) {
	info := Get()

	// String() should contain Short()
	shortVersion := info.Short()
	fullString := info.String()

	if !strings.Contains(fullString, shortVersion) {
		t.Errorf("String() = %q does not contain Short() = %q", fullString, shortVersion)
	}
}

func TestGet_ReturnsNewInstance(t *testing.T) {
	info1 := Get()
	info2 := Get()

	// Should return different instances
	if info1 == info2 {
		t.Error("Get() should return a new instance each time")
	}

	// But with same values
	if info1.Version != info2.Version {
		t.Error("Get() instances should have same Version")
	}
	if info1.GitCommit != info2.GitCommit {
		t.Error("Get() instances should have same GitCommit")
	}
	if info1.BuildTime != info2.BuildTime {
		t.Error("Get() instances should have same BuildTime")
	}
}
