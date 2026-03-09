package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/dl-alexandre/gdrv/internal/logging"
)

const (
	// DiscoveryBaseURL is the base URL for the Google Discovery Service
	DiscoveryBaseURL = "https://www.googleapis.com/discovery/v1"

	// DiscoveryReadOnlyScope is the OAuth scope for reading discovery docs
	DiscoveryReadOnlyScope = "https://www.googleapis.com/auth/discovery.readonly"

	// Default retry configuration
	DefaultMaxAttempts  = 3
	DefaultRetryDelay   = 1 * time.Second
	DefaultRetryBackoff = 2.0

	// AllowedHosts restricts discovery fetches to trusted domains
	AllowedHosts = "www.googleapis.com,discovery.googleapis.com,drivelabels.googleapis.com"
)

// Client provides access to the Google Discovery Service
type Client struct {
	httpClient   *http.Client
	cache        *Cache
	logger       logging.Logger
	baseURL      string
	maxAttempts  int
	retryDelay   time.Duration
	retryBackoff float64
}

// ClientOptions contains configuration options for the discovery client
type ClientOptions struct {
	HTTPClient   *http.Client
	Cache        *Cache
	Logger       logging.Logger
	BaseURL      string
	MaxAttempts  int
	RetryDelay   time.Duration
	RetryBackoff float64
}

// NewClient creates a new Discovery Service client
func NewClient(opts ClientOptions) *Client {
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = DiscoveryBaseURL
	}

	maxAttempts := opts.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = DefaultMaxAttempts
	}

	retryDelay := opts.RetryDelay
	if retryDelay <= 0 {
		retryDelay = DefaultRetryDelay
	}

	retryBackoff := opts.RetryBackoff
	if retryBackoff <= 0 {
		retryBackoff = DefaultRetryBackoff
	}

	return &Client{
		httpClient:   httpClient,
		cache:        opts.Cache,
		logger:       opts.Logger,
		baseURL:      baseURL,
		maxAttempts:  maxAttempts,
		retryDelay:   retryDelay,
		retryBackoff: retryBackoff,
	}
}

// ListAPIs returns a list of all available APIs from the directory
func (c *Client) ListAPIs(ctx context.Context) (*APIDirectoryList, error) {
	// Check cache first
	if c.cache != nil {
		if cached, found := c.cache.GetDirectory(); found {
			c.logDebug("Using cached API directory")
			return cached, nil
		}
	}

	url := fmt.Sprintf("%s/apis", c.baseURL)
	c.logDebug("Fetching API directory from %s", url)

	// Validate host
	if err := c.validateHost(url); err != nil {
		return nil, fmt.Errorf("host validation failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch API directory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API directory request failed: %s (status %d)", string(body), resp.StatusCode)
	}

	var directory APIDirectoryList
	if err := json.NewDecoder(resp.Body).Decode(&directory); err != nil {
		return nil, fmt.Errorf("failed to decode API directory: %w", err)
	}

	// Cache the result
	if c.cache != nil {
		if err := c.cache.SetDirectory(&directory); err != nil {
			c.logDebug("Failed to cache API directory: %v", err)
		}
	}

	return &directory, nil
}

// GetDiscoveryDocument fetches the discovery document for a specific API
func (c *Client) GetDiscoveryDocument(ctx context.Context, serviceName, version string) (*DiscoveryDocument, error) {
	// Normalize service name
	serviceName = strings.ToLower(serviceName)

	// Check cache first
	if c.cache != nil {
		if cached, found := c.cache.GetDocument(serviceName, version); found {
			c.logDebug("Using cached discovery document for %s %s", serviceName, version)
			return cached.Document, nil
		}
	}

	// If no version specified, find the preferred one
	if version == "" {
		directory, err := c.ListAPIs(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list APIs to find preferred version: %w", err)
		}

		for _, api := range directory.APIs {
			if strings.EqualFold(api.Name, serviceName) && api.Preferred {
				version = api.Version
				c.logDebug("Found preferred version %s for %s", version, serviceName)
				break
			}
		}

		if version == "" {
			return nil, fmt.Errorf("could not find preferred version for service %s", serviceName)
		}
	}

	// Fetch the discovery document using Discovery API
	url := fmt.Sprintf("%s/apis/%s/%s/rest", c.baseURL, serviceName, version)
	c.logDebug("Fetching discovery document from %s", url)

	// Validate host
	if err := c.validateHost(url); err != nil {
		return nil, fmt.Errorf("host validation failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discovery document request failed: %s (status %d)", string(body), resp.StatusCode)
	}

	var doc DiscoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to decode discovery document: %w", err)
	}

	// Cache the result
	if c.cache != nil {
		cached := &CachedDiscoveryDoc{
			Document:  &doc,
			FetchedAt: time.Now(),
			ETag:      resp.Header.Get("ETag"),
		}
		if err := c.cache.SetDocument(serviceName, version, cached); err != nil {
			c.logDebug("Failed to cache discovery document: %v", err)
		}
	}

	return &doc, nil
}

// ResolveMethod finds a method in the discovery document by resource path
// resourcePath should be in the format "resource.subresource.method" (e.g., "files.list")
func (c *Client) ResolveMethod(doc *DiscoveryDocument, resourcePath string) (*ResolvedMethod, error) {
	parts := strings.Split(resourcePath, ".")
	if len(parts) < 1 {
		return nil, fmt.Errorf("invalid resource path: %s", resourcePath)
	}

	methodName := parts[len(parts)-1]
	resourceParts := parts[:len(parts)-1]

	// Navigate to the resource
	var currentResource *Resource
	if len(resourceParts) == 0 {
		// Top-level method
		if method, ok := doc.Methods[methodName]; ok {
			return &ResolvedMethod{
				ServiceName:  doc.Name,
				ResourcePath: []string{},
				MethodName:   methodName,
				Method:       method,
				FullPath:     method.Path,
				HTTPMethod:   method.HTTPMethod,
			}, nil
		}
		return nil, fmt.Errorf("top-level method not found: %s", methodName)
	}

	// Navigate nested resources
	resource, ok := doc.Resources[resourceParts[0]]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceParts[0])
	}
	currentResource = &resource

	for i := 1; i < len(resourceParts); i++ {
		nextResource, ok := currentResource.Resources[resourceParts[i]]
		if !ok {
			return nil, fmt.Errorf("nested resource not found: %s", resourceParts[i])
		}
		currentResource = &nextResource
	}

	// Find the method
	method, ok := currentResource.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("method not found: %s in resource %s", methodName, strings.Join(resourceParts, "."))
	}

	return &ResolvedMethod{
		ServiceName:  doc.Name,
		ResourcePath: resourceParts,
		MethodName:   methodName,
		Method:       method,
		FullPath:     method.Path,
		HTTPMethod:   method.HTTPMethod,
	}, nil
}

// doWithRetry performs an HTTP request with exponential backoff retry
func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt < c.maxAttempts; attempt++ {
		if attempt > 0 {
			delay := c.retryDelay * time.Duration(float64(attempt)*c.retryBackoff)
			c.logDebug("Retry attempt %d/%d after %v", attempt+1, c.maxAttempts, delay)
			time.Sleep(delay)
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			c.logDebug("Request failed (attempt %d): %v", attempt+1, err)
			continue
		}

		// Check for retryable status codes
		if resp.StatusCode >= 500 || resp.StatusCode == 429 {
			c.logDebug("Retryable status code %d (attempt %d)", resp.StatusCode, attempt+1)
			resp.Body.Close()
			continue
		}

		// Success or non-retryable error
		return resp, nil
	}

	if err != nil {
		return nil, fmt.Errorf("request failed after %d total attempts (including %d retries): %w", c.maxAttempts, c.maxAttempts-1, err)
	}
	return resp, nil
}

// validateHost ensures the URL is from an allowed host
func (c *Client) validateHost(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("parsing URL: %w", err)
	}

	if u.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS: %s", urlStr)
	}

	allowedHosts := strings.Split(AllowedHosts, ",")
	for _, host := range allowedHosts {
		if strings.EqualFold(u.Host, host) || strings.HasSuffix(u.Host, "."+host) {
			return nil
		}
	}

	return fmt.Errorf("host %s is not in allowed list: %s", u.Host, AllowedHosts)
}

// BuildRequestURL constructs the full request URL with path parameters substituted
func (c *Client) BuildRequestURL(doc *DiscoveryDocument, resolved *ResolvedMethod, pathParams map[string]string) (string, error) {
	baseURL := doc.BaseURL
	if baseURL == "" {
		baseURL = doc.RootURL + doc.ServicePath
	}

	path := resolved.FullPath

	// Substitute path parameters
	for key, value := range pathParams {
		placeholder := "{" + key + "}"
		path = strings.ReplaceAll(path, placeholder, url.PathEscape(value))
	}

	// Check for unsubstituted parameters
	if strings.Contains(path, "{") {
		return "", fmt.Errorf("unsubstituted path parameters in: %s", path)
	}

	// Ensure no double slashes when joining baseURL and path
	baseURL = strings.TrimSuffix(baseURL, "/")
	path = strings.TrimPrefix(path, "/")

	return baseURL + "/" + path, nil
}

// ListMethods returns all methods in a discovery document for a given resource path
func (c *Client) ListMethods(doc *DiscoveryDocument, resourcePath string) ([]MethodListResult, error) {
	var results []MethodListResult

	if resourcePath == "" {
		// List top-level methods
		for name, method := range doc.Methods {
			results = append(results, MethodListResult{
				ID:          name,
				Path:        method.Path,
				HTTPMethod:  method.HTTPMethod,
				Description: method.Description,
			})
		}
		return results, nil
	}

	parts := strings.Split(resourcePath, ".")
	resource, ok := doc.Resources[parts[0]]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", parts[0])
	}

	// Navigate to nested resource
	for i := 1; i < len(parts); i++ {
		nextResource, ok := resource.Resources[parts[i]]
		if !ok {
			return nil, fmt.Errorf("nested resource not found: %s", parts[i])
		}
		resource = nextResource
	}

	// List methods in the resource
	for name, method := range resource.Methods {
		results = append(results, MethodListResult{
			ID:          resourcePath + "." + name,
			Path:        method.Path,
			HTTPMethod:  method.HTTPMethod,
			Description: method.Description,
		})
	}

	return results, nil
}

// GetPreferredAPIVersion returns the preferred version for a given service
func (c *Client) GetPreferredAPIVersion(ctx context.Context, serviceName string) (string, error) {
	directory, err := c.ListAPIs(ctx)
	if err != nil {
		return "", err
	}

	serviceName = strings.ToLower(serviceName)
	for _, api := range directory.APIs {
		if strings.EqualFold(api.Name, serviceName) && api.Preferred {
			return api.Version, nil
		}
	}

	return "", fmt.Errorf("no preferred version found for service: %s", serviceName)
}

// GetAPIBasePath returns the base path for API-specific discovery URLs
func (c *Client) GetAPIBasePath(serviceName string) string {
	// For most APIs, the pattern is https://service.googleapis.com/$discovery/rest?version=X
	return fmt.Sprintf("https://%s.googleapis.com", serviceName)
}

// logDebug logs a debug message if a logger is configured
func (c *Client) logDebug(format string, args ...interface{}) {
	if c.logger != nil {
		c.logger.Debug(fmt.Sprintf(format, args...))
	}
}

// ExecuteRequest executes a resolved API request and returns the response
func (c *Client) ExecuteRequest(ctx context.Context, doc *DiscoveryDocument, resolved *ResolvedMethod,
	pathParams, queryParams map[string]string, body io.Reader, headers map[string]string) (*http.Response, error) {

	requestURL, err := c.BuildRequestURL(doc, resolved, pathParams)
	if err != nil {
		return nil, fmt.Errorf("failed to build request URL: %w", err)
	}

	// Add query parameters
	if len(queryParams) > 0 {
		u, err := url.Parse(requestURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}
		q := u.Query()
		for key, value := range queryParams {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
		requestURL = u.String()
	}

	req, err := http.NewRequestWithContext(ctx, resolved.HTTPMethod, requestURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	c.logDebug("Executing %s %s", resolved.HTTPMethod, requestURL)

	return c.doWithRetry(req)
}

// JoinPath is a helper to safely join URL paths
func JoinPath(base, p string) string {
	return path.Join(base, p)
}
