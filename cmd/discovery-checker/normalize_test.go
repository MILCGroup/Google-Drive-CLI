package main

import (
	"encoding/json"
	"testing"

	"github.com/dl-alexandre/gdrv/internal/discovery"
)

// TestGolden_Normalization tests snapshot normalization with known inputs
func TestGolden_Normalization(t *testing.T) {
	// Create a discovery document with volatile and stable fields
	doc := &discovery.DiscoveryDocument{
		Name:        "test",
		Version:     "v1",
		Title:       "Test API",
		Description: "A test API",
		BaseURL:     "https://test.googleapis.com/",
		RootURL:     "https://test.googleapis.com/",
		ServicePath: "test/v1/",
		Resources: map[string]discovery.Resource{
			"files": {
				Methods: map[string]discovery.Method{
					"list": {
						ID:                    "test.files.list",
						Path:                  "/files",
						HTTPMethod:            "GET",
						SupportsMediaUpload:   false,
						SupportsMediaDownload: true,
						Parameters: map[string]discovery.Parameter{
							"pageToken": {
								Type:     "string",
								Location: "query",
								Required: false,
							},
						},
					},
				},
			},
		},
		Schemas: map[string]discovery.Schema{
			"File": {
				ID:          "File",
				Type:        "object",
				Description: "Represents a file",
				Properties: map[string]discovery.Schema{
					"id": {
						Type: "string",
					},
					"name": {
						Type: "string",
					},
				},
				Required: []string{"id"},
			},
		},
		Auth: &discovery.AuthInfo{
			OAuth2: struct {
				Scopes map[string]struct {
					Description string `json:"description"`
				} `json:"scopes"`
			}{
				Scopes: map[string]struct {
					Description string `json:"description"`
				}{
					"https://www.googleapis.com/auth/test.readonly": {
						Description: "Read test data",
					},
				},
			},
		},
	}

	// Normalize
	data, err := NormalizeSnapshot(doc)
	if err != nil {
		t.Fatalf("NormalizeSnapshot failed: %v", err)
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Normalized output is not valid JSON: %v", err)
	}

	// Verify stable fields are present
	if result["name"] != "test" {
		t.Errorf("Expected name='test', got %v", result["name"])
	}
	if result["version"] != "v1" {
		t.Errorf("Expected version='v1', got %v", result["version"])
	}

	// Compact snapshot
	compact, err := CompactSnapshot(doc, false)
	if err != nil {
		t.Fatalf("CompactSnapshot failed: %v", err)
	}

	// Verify compact is valid JSON
	var compactResult map[string]interface{}
	if err := json.Unmarshal(compact, &compactResult); err != nil {
		t.Fatalf("Compact output is not valid JSON: %v", err)
	}

	// Media flags should be preserved in compact
	resources, ok := compactResult["resources"].(map[string]interface{})
	if !ok {
		t.Fatal("resources not found in compact output")
	}
	files, ok := resources["files"].(map[string]interface{})
	if !ok {
		t.Fatal("files resource not found")
	}
	methods, ok := files["methods"].(map[string]interface{})
	if !ok {
		t.Fatal("methods not found")
	}
	list, ok := methods["list"].(map[string]interface{})
	if !ok {
		t.Fatal("list method not found")
	}

	// Check media flags are preserved
	if download, ok := list["supportsMediaDownload"].(bool); !ok || !download {
		t.Errorf("supportsMediaDownload should be true, got %v", download)
	}

	// Descriptions should be excluded when keepDescriptions=false
	if _, hasDesc := list["description"]; hasDesc {
		t.Error("description should be excluded when keepDescriptions=false")
	}
}

// TestGolden_Classifier tests change classification with known diffs
func TestGolden_Classifier(t *testing.T) {
	tests := []struct {
		name         string
		old          *discovery.DiscoveryDocument
		new          *discovery.DiscoveryDocument
		wantAdditive int
		wantRisky    int
		wantBreaking int
	}{
		{
			name: "no changes",
			old: &discovery.DiscoveryDocument{
				Name: "test",
				Resources: map[string]discovery.Resource{
					"files": {
						Methods: map[string]discovery.Method{
							"list": {
								ID:         "test.files.list",
								Path:       "/files",
								HTTPMethod: "GET",
							},
						},
					},
				},
			},
			new: &discovery.DiscoveryDocument{
				Name: "test",
				Resources: map[string]discovery.Resource{
					"files": {
						Methods: map[string]discovery.Method{
							"list": {
								ID:         "test.files.list",
								Path:       "/files",
								HTTPMethod: "GET",
							},
						},
					},
				},
			},
			wantAdditive: 0,
			wantRisky:    0,
			wantBreaking: 0,
		},
		{
			name: "new optional method - additive",
			old: &discovery.DiscoveryDocument{
				Name: "test",
				Resources: map[string]discovery.Resource{
					"files": {
						Methods: map[string]discovery.Method{
							"list": {
								ID:         "test.files.list",
								Path:       "/files",
								HTTPMethod: "GET",
							},
						},
					},
				},
			},
			new: &discovery.DiscoveryDocument{
				Name: "test",
				Resources: map[string]discovery.Resource{
					"files": {
						Methods: map[string]discovery.Method{
							"list": {
								ID:         "test.files.list",
								Path:       "/files",
								HTTPMethod: "GET",
							},
							"get": {
								ID:         "test.files.get",
								Path:       "/files/{id}",
								HTTPMethod: "GET",
							},
						},
					},
				},
			},
			wantAdditive: 1,
			wantRisky:    0,
			wantBreaking: 0,
		},
		{
			name: "method removed - breaking",
			old: &discovery.DiscoveryDocument{
				Name: "test",
				Resources: map[string]discovery.Resource{
					"files": {
						Methods: map[string]discovery.Method{
							"list": {
								ID:         "test.files.list",
								Path:       "/files",
								HTTPMethod: "GET",
							},
						},
					},
				},
			},
			new: &discovery.DiscoveryDocument{
				Name: "test",
				Resources: map[string]discovery.Resource{
					"files": {
						Methods: map[string]discovery.Method{},
					},
				},
			},
			wantAdditive: 0,
			wantRisky:    0,
			wantBreaking: 1,
		},
		{
			name: "required param added - risky",
			old: &discovery.DiscoveryDocument{
				Name: "test",
				Resources: map[string]discovery.Resource{
					"files": {
						Methods: map[string]discovery.Method{
							"list": {
								ID:         "test.files.list",
								Path:       "/files",
								HTTPMethod: "GET",
								Parameters: map[string]discovery.Parameter{},
							},
						},
					},
				},
			},
			new: &discovery.DiscoveryDocument{
				Name: "test",
				Resources: map[string]discovery.Resource{
					"files": {
						Methods: map[string]discovery.Method{
							"list": {
								ID:         "test.files.list",
								Path:       "/files",
								HTTPMethod: "GET",
								Parameters: map[string]discovery.Parameter{
									"orderBy": {
										Type:     "string",
										Location: "query",
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			wantAdditive: 0,
			wantRisky:    1,
			wantBreaking: 0,
		},
		{
			name: "enum value added in property - risky by default",
			old: &discovery.DiscoveryDocument{
				Name: "test",
				Schemas: map[string]discovery.Schema{
					"File": {
						Type: "object",
						Properties: map[string]discovery.Schema{
							"color": {
								Type: "string",
								Enum: []string{"red", "green"},
							},
						},
					},
				},
			},
			new: &discovery.DiscoveryDocument{
				Name: "test",
				Schemas: map[string]discovery.Schema{
					"File": {
						Type: "object",
						Properties: map[string]discovery.Schema{
							"color": {
								Type: "string",
								Enum: []string{"red", "green", "blue"},
							},
						},
					},
				},
			},
			wantAdditive: 0, // Conservative default: enum expansion is risky
			wantRisky:    1,
			wantBreaking: 0,
		},
		{
			name: "enum value removed from property - breaking",
			old: &discovery.DiscoveryDocument{
				Name: "test",
				Schemas: map[string]discovery.Schema{
					"File": {
						Type: "object",
						Properties: map[string]discovery.Schema{
							"color": {
								Type: "string",
								Enum: []string{"red", "green", "blue"},
							},
						},
					},
				},
			},
			new: &discovery.DiscoveryDocument{
				Name: "test",
				Schemas: map[string]discovery.Schema{
					"File": {
						Type: "object",
						Properties: map[string]discovery.Schema{
							"color": {
								Type: "string",
								Enum: []string{"red", "green"},
							},
						},
					},
				},
			},
			wantAdditive: 0,
			wantRisky:    0,
			wantBreaking: 1,
		},
		{
			name: "HTTP method change - breaking",
			old: &discovery.DiscoveryDocument{
				Name: "test",
				Resources: map[string]discovery.Resource{
					"files": {
						Methods: map[string]discovery.Method{
							"create": {
								ID:         "test.files.create",
								Path:       "/files",
								HTTPMethod: "POST",
							},
						},
					},
				},
			},
			new: &discovery.DiscoveryDocument{
				Name: "test",
				Resources: map[string]discovery.Resource{
					"files": {
						Methods: map[string]discovery.Method{
							"create": {
								ID:         "test.files.create",
								Path:       "/files",
								HTTPMethod: "PUT",
							},
						},
					},
				},
			},
			wantAdditive: 0,
			wantRisky:    0,
			wantBreaking: 1,
		},
	}

	classifier := NewRefinedClassifier()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := classifier.ClassifyChanges("test", tt.old, tt.new)

			if report.AdditiveChanges != tt.wantAdditive {
				t.Errorf("AdditiveChanges = %d, want %d", report.AdditiveChanges, tt.wantAdditive)
			}
			if report.RiskyChanges != tt.wantRisky {
				t.Errorf("RiskyChanges = %d, want %d", report.RiskyChanges, tt.wantRisky)
			}
			if report.BreakingChanges != tt.wantBreaking {
				t.Errorf("BreakingChanges = %d, want %d", report.BreakingChanges, tt.wantBreaking)
			}
		})
	}
}

// TestGolden_RequiredSemantics tests the distinction between request/response required fields
func TestGolden_RequiredSemantics(t *testing.T) {
	// Required field added to request schema - risky
	oldRequest := &discovery.DiscoveryDocument{
		Name: "test",
		Schemas: map[string]discovery.Schema{
			"CreateRequest": {
				Type:     "object",
				Required: []string{},
				Properties: map[string]discovery.Schema{
					"name": {Type: "string"},
				},
			},
		},
	}

	newRequest := &discovery.DiscoveryDocument{
		Name: "test",
		Schemas: map[string]discovery.Schema{
			"CreateRequest": {
				Type:     "object",
				Required: []string{"name"},
				Properties: map[string]discovery.Schema{
					"name": {Type: "string"},
				},
			},
		},
	}

	classifier := NewRefinedClassifier()

	// Required field added to request is risky (clients may not be sending it)
	report := classifier.ClassifyChanges("test", oldRequest, newRequest)
	if report.RiskyChanges != 1 {
		t.Errorf("Request required field addition should be risky, got additive=%d, risky=%d, breaking=%d",
			report.AdditiveChanges, report.RiskyChanges, report.BreakingChanges)
	}
}

// TestGolden_HashStability ensures consistent hashing
func TestGolden_HashStability(t *testing.T) {
	doc := &discovery.DiscoveryDocument{
		Name:    "test",
		Version: "v1",
	}

	data1, err := NormalizeSnapshot(doc)
	if err != nil {
		t.Fatalf("First normalization failed: %v", err)
	}

	data2, err := NormalizeSnapshot(doc)
	if err != nil {
		t.Fatalf("Second normalization failed: %v", err)
	}

	hash1 := SnapshotHash(data1)
	hash2 := SnapshotHash(data2)

	if hash1 != hash2 {
		t.Errorf("Hash should be stable: got %s vs %s", hash1, hash2)
	}

	// Even with different map iteration orders, hash should be same
	doc2 := &discovery.DiscoveryDocument{
		Name:    "test",
		Version: "v1",
		Resources: map[string]discovery.Resource{
			"b": {Methods: map[string]discovery.Method{"z": {ID: "z", Path: "/z", HTTPMethod: "GET"}}},
			"a": {Methods: map[string]discovery.Method{"y": {ID: "y", Path: "/y", HTTPMethod: "GET"}}},
		},
	}

	data3, err := NormalizeSnapshot(doc2)
	if err != nil {
		t.Fatalf("Third normalization failed: %v", err)
	}

	hash3 := SnapshotHash(data3)
	t.Logf("Hash with different order: %s", hash3)
	// Note: we can't directly compare hash1 and hash3 since the docs have different content
	// but we verify the normalization produces valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(data3, &result); err != nil {
		t.Errorf("Normalization with different order produced invalid JSON: %v", err)
	}
}
