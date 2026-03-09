package discovery

import (
	"encoding/json"
	"time"
)

// APIDirectoryEntry represents an API in the discovery directory
type APIDirectoryEntry struct {
	Kind             string `json:"kind"`
	ID               string `json:"id"`
	Name             string `json:"name"`
	Version          string `json:"version"`
	Title            string `json:"title"`
	Description      string `json:"description"`
	DiscoveryRestURL string `json:"discoveryRestUrl"`
	Icons            struct {
		X16 string `json:"x16"`
		X32 string `json:"x32"`
	} `json:"icons"`
	DocumentationLink string `json:"documentationLink"`
	Preferred         bool   `json:"preferred"`
}

// APIDirectoryList represents the directory listing response
type APIDirectoryList struct {
	Kind string              `json:"kind"`
	APIs []APIDirectoryEntry `json:"items"`
	etag string
}

// DiscoveryDocument represents the full discovery document for an API
type DiscoveryDocument struct {
	Kind                         string               `json:"kind"`
	ID                           string               `json:"id"`
	Name                         string               `json:"name"`
	Version                      string               `json:"version"`
	Title                        string               `json:"title"`
	Description                  string               `json:"description"`
	Icons                        map[string]string    `json:"icons"`
	Protocol                     string               `json:"protocol"`
	BaseURL                      string               `json:"baseUrl"`
	BasePath                     string               `json:"basePath"`
	RootURL                      string               `json:"rootUrl"`
	ServicePath                  string               `json:"servicePath"`
	BatchPath                    string               `json:"batchPath"`
	DocumentationLink            string               `json:"documentationLink"`
	Labels                       []string             `json:"labels"`
	CanonicalName                string               `json:"canonicalName"`
	MTLSRootURL                  string               `json:"mtlsRootUrl"`
	FullyEncodeReservedExpansion bool                 `json:"fullyEncodeReservedExpansion"`
	Resources                    map[string]Resource  `json:"resources"`
	Schemas                      map[string]Schema    `json:"schemas"`
	Methods                      map[string]Method    `json:"methods"`
	Parameters                   map[string]Parameter `json:"parameters"`
	Auth                         *AuthInfo            `json:"auth,omitempty"`
}

// AuthInfo contains OAuth2 scopes for the API
type AuthInfo struct {
	OAuth2 struct {
		Scopes map[string]struct {
			Description string `json:"description"`
		} `json:"scopes"`
	} `json:"oauth2"`
}

// Resource represents a REST resource in the API
type Resource struct {
	Methods   map[string]Method   `json:"methods"`
	Resources map[string]Resource `json:"resources"`
}

// Method represents an API method
type Method struct {
	ID                    string               `json:"id"`
	Path                  string               `json:"path"`
	HTTPMethod            string               `json:"httpMethod"`
	Description           string               `json:"description"`
	Parameters            map[string]Parameter `json:"parameters"`
	Request               *TypeRef             `json:"request,omitempty"`
	Response              *TypeRef             `json:"response,omitempty"`
	Scopes                []string             `json:"scopes"`
	SupportsMediaUpload   bool                 `json:"supportsMediaUpload"`
	SupportsMediaDownload bool                 `json:"supportsMediaDownload"`
	MediaUpload           *MediaUpload         `json:"mediaUpload,omitempty"`
	FlatPath              string               `json:"flatPath"`
	Etag                  string               `json:"etag"`
}

// Parameter represents a method parameter
type Parameter struct {
	ID               string      `json:"id"`
	Type             string      `json:"type"`
	Ref              string      `json:"$ref"`
	Description      string      `json:"description"`
	Default          interface{} `json:"default"`
	Required         bool        `json:"required"`
	Pattern          string      `json:"pattern"`
	Format           string      `json:"format"`
	Minimum          string      `json:"minimum"`
	Maximum          string      `json:"maximum"`
	Repeated         bool        `json:"repeated"`
	Location         string      `json:"location"`
	Enum             []string    `json:"enum"`
	EnumDescriptions []string    `json:"enumDescriptions"`
}

// TypeRef represents a reference to a type (request or response)
type TypeRef struct {
	Ref  string `json:"$ref"`
	Type string `json:"type"`
}

// MediaUpload represents media upload configuration
type MediaUpload struct {
	Accept    []string `json:"accept"`
	MaxSize   string   `json:"maxSize"`
	Protocols struct {
		Simple struct {
			Multipart bool `json:"multipart"`
		} `json:"simple"`
		Resumable struct {
			Multipart bool   `json:"multipart"`
			Path      string `json:"path"`
		} `json:"resumable"`
	} `json:"protocols"`
}

// Schema represents a data schema
type Schema struct {
	ID                   string            `json:"id"`
	Type                 string            `json:"type"`
	Description          string            `json:"description"`
	Default              interface{}       `json:"default"`
	Properties           map[string]Schema `json:"properties"`
	Items                *Schema           `json:"items,omitempty"`
	Ref                  string            `json:"$ref"`
	Required             []string          `json:"required"`
	Format               string            `json:"format"`
	Pattern              string            `json:"pattern"`
	Enum                 []string          `json:"enum"`
	EnumDescriptions     []string          `json:"enumDescriptions"`
	AdditionalProperties interface{}       `json:"additionalProperties"`
}

// ResolvedMethod contains the full method information after resolving path
type ResolvedMethod struct {
	ServiceName  string
	ResourcePath []string
	MethodName   string
	Method       Method
	FullPath     string
	HTTPMethod   string
}

// IsListMethod returns true if the method appears to be a list method
func (m *Method) IsListMethod() bool {
	// Heuristic: method name contains "list" and has pageToken/pageSize parameters
	if !containsIgnoreCase(m.ID, "list") {
		return false
	}
	_, hasPageToken := m.Parameters["pageToken"]
	_, hasPageSize := m.Parameters["pageSize"]
	_, hasMaxResults := m.Parameters["maxResults"]
	return hasPageToken || hasPageSize || hasMaxResults
}

// HasRequestBody returns true if the method requires a request body
func (m *Method) HasRequestBody() bool {
	return m.Request != nil && m.HTTPMethod != "GET" && m.HTTPMethod != "DELETE"
}

// GetScopes returns the OAuth scopes required for this method
func (m *Method) GetScopes(doc *DiscoveryDocument) []string {
	if len(m.Scopes) > 0 {
		return m.Scopes
	}
	if doc.Auth != nil {
		scopes := make([]string, 0, len(doc.Auth.OAuth2.Scopes))
		for scope := range doc.Auth.OAuth2.Scopes {
			scopes = append(scopes, scope)
		}
		return scopes
	}
	return nil
}

// CachedDiscoveryDoc wraps a discovery document with cache metadata
type CachedDiscoveryDoc struct {
	Document  *DiscoveryDocument `json:"document"`
	FetchedAt time.Time          `json:"fetched_at"`
	ETag      string             `json:"etag"`
}

// ToJSON serializes the cached document to JSON
func (c *CachedDiscoveryDoc) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

// CachedDiscoveryDocFromJSON deserializes a cached document from JSON
func CachedDiscoveryDocFromJSON(data []byte) (*CachedDiscoveryDoc, error) {
	var cached CachedDiscoveryDoc
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}
	return &cached, nil
}

// APIListResult represents a single API in the list command output
type APIListResult struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Preferred   bool   `json:"preferred"`
}

// MethodListResult represents a method in the list output
type MethodListResult struct {
	Path        string `json:"path"`
	HTTPMethod  string `json:"httpMethod"`
	Description string `json:"description"`
	ID          string `json:"id"`
}

// APIResponse represents the generic API response structure
type APIResponse struct {
	Data          interface{} `json:"data"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
	PageCount     int         `json:"pageCount,omitempty"`
	TotalItems    int         `json:"totalItems,omitempty"`
}

func containsIgnoreCase(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	lowerS := make([]rune, len(s))
	lowerSubstr := make([]rune, len(substr))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			lowerS[i] = r + ('a' - 'A')
		} else {
			lowerS[i] = r
		}
	}
	for i, r := range substr {
		if r >= 'A' && r <= 'Z' {
			lowerSubstr[i] = r + ('a' - 'A')
		} else {
			lowerSubstr[i] = r
		}
	}

	for i := 0; i <= len(lowerS)-len(lowerSubstr); i++ {
		match := true
		for j := 0; j < len(lowerSubstr); j++ {
			if lowerS[i+j] != lowerSubstr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
