package discovery

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMethod_IsListMethod(t *testing.T) {
	tests := []struct {
		name     string
		method   Method
		expected bool
	}{
		{
			name: "list method with pageToken",
			method: Method{
				ID:         "drive.files.list",
				Parameters: map[string]Parameter{"pageToken": {Location: "query"}},
			},
			expected: true,
		},
		{
			name: "list method with pageSize",
			method: Method{
				ID:         "drive.files.list",
				Parameters: map[string]Parameter{"pageSize": {Location: "query"}},
			},
			expected: true,
		},
		{
			name: "list method with maxResults",
			method: Method{
				ID:         "gmail.users.messages.list",
				Parameters: map[string]Parameter{"maxResults": {Location: "query"}},
			},
			expected: true,
		},
		{
			name: "non-list method",
			method: Method{
				ID:         "drive.files.get",
				Parameters: map[string]Parameter{},
			},
			expected: false,
		},
		{
			name: "method without list in name",
			method: Method{
				ID:         "drive.files.create",
				Parameters: map[string]Parameter{"pageToken": {Location: "query"}},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.method.IsListMethod()
			if got != tt.expected {
				t.Errorf("IsListMethod() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMethod_HasRequestBody(t *testing.T) {
	tests := []struct {
		name     string
		method   Method
		expected bool
	}{
		{
			name: "POST with request body",
			method: Method{
				HTTPMethod: "POST",
				Request:    &TypeRef{Ref: "File"},
			},
			expected: true,
		},
		{
			name: "PUT with request body",
			method: Method{
				HTTPMethod: "PUT",
				Request:    &TypeRef{Ref: "File"},
			},
			expected: true,
		},
		{
			name: "GET without request body",
			method: Method{
				HTTPMethod: "GET",
				Request:    &TypeRef{Ref: "File"},
			},
			expected: false,
		},
		{
			name: "POST without request body ref",
			method: Method{
				HTTPMethod: "POST",
				Request:    nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.method.HasRequestBody()
			if got != tt.expected {
				t.Errorf("HasRequestBody() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCachedDiscoveryDoc_Serialization(t *testing.T) {
	original := &CachedDiscoveryDoc{
		Document: &DiscoveryDocument{
			Name:    "drive",
			Version: "v3",
			Title:   "Google Drive API",
		},
		FetchedAt: time.Now().UTC().Truncate(time.Second),
		ETag:      "abc123",
	}

	// Serialize
	data, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Deserialize
	restored, err := CachedDiscoveryDocFromJSON(data)
	if err != nil {
		t.Fatalf("CachedDiscoveryDocFromJSON() error = %v", err)
	}

	// Verify
	if restored.Document.Name != original.Document.Name {
		t.Errorf("Document.Name = %v, want %v", restored.Document.Name, original.Document.Name)
	}
	if restored.Document.Version != original.Document.Version {
		t.Errorf("Document.Version = %v, want %v", restored.Document.Version, original.Document.Version)
	}
	if restored.ETag != original.ETag {
		t.Errorf("ETag = %v, want %v", restored.ETag, original.ETag)
	}
	if !restored.FetchedAt.Equal(original.FetchedAt) {
		t.Errorf("FetchedAt = %v, want %v", restored.FetchedAt, original.FetchedAt)
	}
}

func TestResolveMethod(t *testing.T) {
	client := NewClient(ClientOptions{})

	doc := &DiscoveryDocument{
		Name: "drive",
		Resources: map[string]Resource{
			"files": {
				Methods: map[string]Method{
					"list": {
						ID:         "drive.files.list",
						Path:       "/files",
						HTTPMethod: "GET",
					},
					"get": {
						ID:         "drive.files.get",
						Path:       "/files/{fileId}",
						HTTPMethod: "GET",
					},
				},
			},
		},
		Methods: map[string]Method{
			"about": {
				ID:         "drive.about.get",
				Path:       "/about",
				HTTPMethod: "GET",
			},
		},
	}

	tests := []struct {
		name        string
		path        string
		expectError bool
		wantMethod  string
		wantPath    string
	}{
		{
			name:        "top-level method",
			path:        "about",
			expectError: false,
			wantMethod:  "about",
			wantPath:    "/about",
		},
		{
			name:        "nested resource method",
			path:        "files.list",
			expectError: false,
			wantMethod:  "list",
			wantPath:    "/files",
		},
		{
			name:        "deep nested resource method",
			path:        "files.get",
			expectError: false,
			wantMethod:  "get",
			wantPath:    "/files/{fileId}",
		},
		{
			name:        "non-existent resource",
			path:        "permissions.list",
			expectError: true,
		},
		{
			name:        "non-existent method",
			path:        "files.deleteAll",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := client.ResolveMethod(doc, tt.path)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if resolved.MethodName != tt.wantMethod {
				t.Errorf("MethodName = %v, want %v", resolved.MethodName, tt.wantMethod)
			}
			if resolved.FullPath != tt.wantPath {
				t.Errorf("FullPath = %v, want %v", resolved.FullPath, tt.wantPath)
			}
		})
	}
}

func TestBuildRequestURL(t *testing.T) {
	client := NewClient(ClientOptions{})

	tests := []struct {
		name       string
		doc        *DiscoveryDocument
		resolved   *ResolvedMethod
		pathParams map[string]string
		wantURL    string
		wantError  bool
	}{
		{
			name: "simple URL without params",
			doc: &DiscoveryDocument{
				BaseURL: "https://www.googleapis.com/drive/v3",
			},
			resolved: &ResolvedMethod{
				FullPath: "/files",
			},
			pathParams: map[string]string{},
			wantURL:    "https://www.googleapis.com/drive/v3/files",
			wantError:  false,
		},
		{
			name: "URL with path parameter using ServicePath",
			doc: &DiscoveryDocument{
				RootURL:     "https://www.googleapis.com/",
				ServicePath: "drive/v3/",
			},
			resolved: &ResolvedMethod{
				FullPath: "/files/{fileId}",
			},
			pathParams: map[string]string{
				"fileId": "abc123",
			},
			wantURL:   "https://www.googleapis.com/drive/v3/files/abc123",
			wantError: false,
		},
		{
			name: "URL with missing path parameter",
			doc: &DiscoveryDocument{
				BaseURL: "https://www.googleapis.com/drive/v3",
			},
			resolved: &ResolvedMethod{
				FullPath: "/files/{fileId}",
			},
			pathParams: map[string]string{},
			wantURL:    "",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.BuildRequestURL(tt.doc, tt.resolved, tt.pathParams)
			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if got != tt.wantURL {
				t.Errorf("BuildRequestURL() = %v, want %v", got, tt.wantURL)
			}
		})
	}
}

func TestAPIDirectoryEntry_Serialization(t *testing.T) {
	entry := APIDirectoryEntry{
		Kind:             "discovery#directoryItem",
		ID:               "drive:v3",
		Name:             "drive",
		Version:          "v3",
		Title:            "Google Drive API",
		Description:      "Manages files in Drive including uploading, downloading, searching, detecting changes, and updating sharing permissions.",
		DiscoveryRestURL: "https://www.googleapis.com/discovery/v1/apis/drive/v3/rest",
		Preferred:        true,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var restored APIDirectoryEntry
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if restored.ID != entry.ID {
		t.Errorf("ID = %v, want %v", restored.ID, entry.ID)
	}
	if restored.Name != entry.Name {
		t.Errorf("Name = %v, want %v", restored.Name, entry.Name)
	}
	if restored.Preferred != entry.Preferred {
		t.Errorf("Preferred = %v, want %v", restored.Preferred, entry.Preferred)
	}
}
