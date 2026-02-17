package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

// TestPKCESupport validates PKCE implementation
func TestPKCESupport(t *testing.T) {
	config := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Scopes:       []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d/callback", addr.Port)

	flow, err := NewOAuthFlow(config, listener, redirectURL)
	if err != nil {
		t.Fatalf("NewOAuthFlow failed: %v", err)
	}
	defer flow.Close()

	// Test 1: Code verifier is generated
	if flow.codeVerifier == "" {
		t.Error("Code verifier not generated")
	}

	// Test 2: Code verifier has correct length (43-128 chars for base64url)
	if len(flow.codeVerifier) < 43 || len(flow.codeVerifier) > 128 {
		t.Errorf("Code verifier length %d outside valid range 43-128", len(flow.codeVerifier))
	}

	// Test 3: Auth URL contains PKCE parameters
	authURL := flow.GetAuthURL()
	parsedURL, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("Failed to parse auth URL: %v", err)
	}

	query := parsedURL.Query()

	codeChallenge := query.Get("code_challenge")
	if codeChallenge == "" {
		t.Error("Auth URL missing code_challenge parameter")
	}

	challengeMethod := query.Get("code_challenge_method")
	if challengeMethod != "S256" {
		t.Errorf("Expected code_challenge_method=S256, got %s", challengeMethod)
	}

	// Test 4: Code challenge is properly computed S256 hash
	expectedChallenge := codeChallengeS256(flow.codeVerifier)
	if codeChallenge != expectedChallenge {
		t.Errorf("Code challenge mismatch: expected %s, got %s", expectedChallenge, codeChallenge)
	}

	// Test 5: State parameter is included
	state := query.Get("state")
	if state == "" {
		t.Error("Auth URL missing state parameter")
	}
	if state != flow.state {
		t.Errorf("State mismatch: expected %s, got %s", flow.state, state)
	}

	// Test 6: Offline access is requested
	accessType := query.Get("access_type")
	if accessType != "offline" {
		t.Errorf("Expected access_type=offline, got %s", accessType)
	}
}

// TestEphemeralPortSelection validates dynamic port assignment
func TestEphemeralPortSelection(t *testing.T) {
	config := &oauth2.Config{
		ClientID: "test-client-id",
		Scopes:   []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	// Test 1: Create multiple flows to ensure no port conflicts
	flows := make([]*OAuthFlow, 5)
	for i := 0; i < 5; i++ {
		flow, err := newLoopbackFlow(config)
		if err != nil {
			t.Fatalf("Failed to create flow %d: %v", i, err)
		}
		flows[i] = flow
		defer flow.Close()
	}

	// Test 2: Verify each flow has unique port
	ports := make(map[int]bool)
	for i, flow := range flows {
		if flow.listener == nil {
			t.Errorf("Flow %d has nil listener", i)
			continue
		}

		addr := flow.listener.Addr().(*net.TCPAddr)
		port := addr.Port

		// Port should be non-zero
		if port == 0 {
			t.Errorf("Flow %d has port 0", i)
		}

		// Port should be unique
		if ports[port] {
			t.Errorf("Flow %d has duplicate port %d", i, port)
		}
		ports[port] = true

		// Redirect URL should match actual port
		expectedRedirect := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
		if flow.redirectURL != expectedRedirect {
			t.Errorf("Flow %d redirect URL mismatch: expected %s, got %s",
				i, expectedRedirect, flow.redirectURL)
		}
	}

	// Test 3: Verify listener binds to loopback (127.0.0.1)
	for i, flow := range flows {
		addr := flow.listener.Addr().(*net.TCPAddr)
		if !addr.IP.IsLoopback() {
			t.Errorf("Flow %d not bound to loopback: %s", i, addr.IP)
		}
	}
}

// TestStateValidation validates CSRF state parameter checking
func TestStateValidation(t *testing.T) {
	config := &oauth2.Config{
		ClientID: "test-client-id",
		Scopes:   []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	flow, err := newLoopbackFlow(config)
	if err != nil {
		t.Fatalf("Failed to create flow: %v", err)
	}
	defer flow.Close()

	tests := []struct {
		name          string
		state         string
		code          string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid state",
			state:       flow.state,
			code:        "test-code",
			expectError: false,
		},
		{
			name:          "invalid state",
			state:         "wrong-state",
			code:          "test-code",
			expectError:   true,
			errorContains: "invalid state",
		},
		{
			name:          "missing code",
			state:         flow.state,
			code:          "",
			expectError:   true,
			errorContains: "auth error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset channels
			flow.codeChan = make(chan string, 1)
			flow.errChan = make(chan error, 1)

			// Create test request
			reqURL := fmt.Sprintf("/callback?state=%s&code=%s",
				url.QueryEscape(tt.state),
				url.QueryEscape(tt.code))
			req := httptest.NewRequest("GET", reqURL, nil)
			w := httptest.NewRecorder()

			// Handle callback
			flow.handleCallback(w, req)

			// Check result
			select {
			case receivedCode := <-flow.codeChan:
				if tt.expectError {
					t.Errorf("Expected error but got code: %s", receivedCode)
				}
				if receivedCode != tt.code {
					t.Errorf("Code mismatch: expected %s, got %s", tt.code, receivedCode)
				}
			case err := <-flow.errChan:
				if !tt.expectError {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Error should contain '%s', got: %v", tt.errorContains, err)
				}
			case <-time.After(100 * time.Millisecond):
				if !tt.expectError {
					t.Error("Timeout waiting for result")
				}
			}
		})
	}
}

// TestCallbackServerTimeout validates timeout handling
func TestCallbackServerTimeout(t *testing.T) {
	config := &oauth2.Config{
		ClientID: "test-client-id",
		Scopes:   []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	flow, err := newLoopbackFlow(config)
	if err != nil {
		t.Fatalf("Failed to create flow: %v", err)
	}
	defer flow.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	flow.StartCallbackServer(ctx)

	// Wait for code with short timeout (should timeout)
	code, err := flow.WaitForCode(100 * time.Millisecond)
	if err == nil {
		t.Errorf("Expected timeout error, got code: %s", code)
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

// TestIsHeadlessEnv validates headless environment detection
func TestIsHeadlessEnv(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		unsetVars   []string
		expected    bool
		skipOnMacOS bool
	}{
		{
			name:      "normal environment",
			envVars:   map[string]string{"DISPLAY": ":0"}, // Set DISPLAY to simulate GUI environment
			unsetVars: []string{"CI", "GITHUB_ACTIONS", "SSH_CONNECTION", "SSH_TTY", "GDRV_NO_BROWSER"},
			expected:  false,
		},
		{
			name:      "GDRV_NO_BROWSER set",
			envVars:   map[string]string{"GDRV_NO_BROWSER": "1", "DISPLAY": ":0"},
			unsetVars: []string{"CI", "GITHUB_ACTIONS", "SSH_CONNECTION", "SSH_TTY"},
			expected:  true,
		},
		{
			name:      "CI environment",
			envVars:   map[string]string{"CI": "true", "DISPLAY": ":0"},
			unsetVars: []string{"GITHUB_ACTIONS", "SSH_CONNECTION", "SSH_TTY", "GDRV_NO_BROWSER"},
			expected:  true,
		},
		{
			name:      "GitHub Actions",
			envVars:   map[string]string{"GITHUB_ACTIONS": "true", "DISPLAY": ":0"},
			unsetVars: []string{"CI", "SSH_CONNECTION", "SSH_TTY", "GDRV_NO_BROWSER"},
			expected:  true,
		},
		{
			name:      "SSH connection",
			envVars:   map[string]string{"SSH_CONNECTION": "192.168.1.1 22 192.168.1.2 54321", "DISPLAY": ":0"},
			unsetVars: []string{"CI", "GITHUB_ACTIONS", "SSH_TTY", "GDRV_NO_BROWSER"},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Unset environment variables first
			for _, k := range tt.unsetVars {
				os.Unsetenv(k)
			}

			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			result := isHeadlessEnv()
			if result != tt.expected {
				t.Errorf("Expected isHeadlessEnv=%v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCodeVerifierGeneration validates code verifier randomness
func TestCodeVerifierGeneration(t *testing.T) {
	verifiers := make(map[string]bool)

	// Generate multiple verifiers and ensure they're unique
	for i := 0; i < 100; i++ {
		verifier, err := generateCodeVerifier()
		if err != nil {
			t.Fatalf("Failed to generate verifier %d: %v", i, err)
		}

		// Check length
		if len(verifier) < 43 || len(verifier) > 128 {
			t.Errorf("Verifier %d length %d outside valid range", i, len(verifier))
		}

		// Check uniqueness
		if verifiers[verifier] {
			t.Errorf("Duplicate verifier generated: %s", verifier)
		}
		verifiers[verifier] = true

		// Verify it's valid base64url
		if strings.ContainsAny(verifier, "+/=") {
			t.Errorf("Verifier %d contains invalid base64url characters: %s", i, verifier)
		}
	}
}

// TestCodeChallengeComputation validates S256 challenge computation
func TestCodeChallengeComputation(t *testing.T) {
	tests := []struct {
		verifier string
		// Expected challenges computed externally for validation
	}{
		{
			verifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			// This is the standard test vector from RFC 7636
		},
	}

	for _, tt := range tests {
		challenge := codeChallengeS256(tt.verifier)

		// Challenge should be base64url encoded (no +/= characters)
		if strings.ContainsAny(challenge, "+/=") {
			t.Errorf("Challenge contains invalid base64url characters: %s", challenge)
		}

		// Challenge should be exactly 43 characters (256 bits base64url encoded)
		if len(challenge) != 43 {
			t.Errorf("Challenge length should be 43, got %d", len(challenge))
		}

		// Challenge should be deterministic
		challenge2 := codeChallengeS256(tt.verifier)
		if challenge != challenge2 {
			t.Error("Code challenge computation is not deterministic")
		}
	}
}

// TestOAuthFlowWithMockServer validates full OAuth flow with mock server
func TestOAuthFlowWithMockServer(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}

			codeVerifier := r.FormValue("code_verifier")
			if codeVerifier == "" {
				t.Error("Token exchange missing code_verifier parameter")
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"access_token": "mock_access_token",
				"refresh_token": "mock_refresh_token",
				"expires_in": 3600,
				"token_type": "Bearer"
			}`)
		}
	}))
	defer mockServer.Close()

	config := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Scopes:       []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
	}

	flow, err := newLoopbackFlow(config)
	if err != nil {
		t.Fatalf("Failed to create flow: %v", err)
	}
	defer flow.Close()

	ctx := context.Background()
	creds, err := flow.ExchangeCode(ctx, "mock_auth_code")
	if err != nil {
		t.Fatalf("ExchangeCode failed: %v", err)
	}

	if creds.AccessToken != "mock_access_token" {
		t.Errorf("Expected access_token=mock_access_token, got %s", creds.AccessToken)
	}
	if creds.RefreshToken != "mock_refresh_token" {
		t.Errorf("Expected refresh_token=mock_refresh_token, got %s", creds.RefreshToken)
	}
}

func TestNewOAuthFlow_NilConfig(t *testing.T) {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	defer listener.Close()

	_, err := NewOAuthFlow(nil, listener, "http://127.0.0.1:8080/callback")
	if err == nil {
		t.Error("NewOAuthFlow should fail with nil config")
	}
}

func TestNewOAuthFlow_NoRedirectURL(t *testing.T) {
	config := &oauth2.Config{
		ClientID: "test-client-id",
		Scopes:   []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	defer listener.Close()

	_, err := NewOAuthFlow(config, listener, "")
	if err == nil {
		t.Error("NewOAuthFlow should fail with no redirect URL")
	}
}

func TestOAuthFlow_GetAuthURL_HasRequiredParams(t *testing.T) {
	config := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Scopes:       []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	flow, err := newLoopbackFlow(config)
	if err != nil {
		t.Fatalf("Failed to create flow: %v", err)
	}
	defer flow.Close()

	authURL := flow.GetAuthURL()
	parsedURL, _ := url.Parse(authURL)
	query := parsedURL.Query()

	if query.Get("client_id") != "test-client-id" {
		t.Error("Auth URL missing client_id")
	}
	if query.Get("response_type") != "code" {
		t.Error("Auth URL missing response_type=code")
	}
	if query.Get("scope") == "" {
		t.Error("Auth URL missing scope")
	}
}

func TestOAuthFlow_StartCallbackServer_ContextCancel(t *testing.T) {
	config := &oauth2.Config{
		ClientID: "test-client-id",
		Scopes:   []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	flow, err := newLoopbackFlow(config)
	if err != nil {
		t.Fatalf("Failed to create flow: %v", err)
	}
	defer flow.Close()

	ctx, cancel := context.WithCancel(context.Background())
	flow.StartCallbackServer(ctx)

	cancel()
	time.Sleep(100 * time.Millisecond)
}

// Regression: --no-browser must not call openBrowser.
func TestAuthenticate_NoBrowser_DoesNotOpenBrowser(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"tok","refresh_token":"rtok","expires_in":3600,"token_type":"Bearer"}`)
	}))
	defer mockServer.Close()

	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)
	mgr.oauthConfig = &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Scopes:       []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
	}

	browserCalled := false
	var capturedURL string
	openBrowser := func(u string) error {
		browserCalled = true
		capturedURL = u
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	type authResult struct {
		err error
	}
	ch := make(chan authResult, 1)
	go func() {
		_, err := mgr.Authenticate(ctx, "test-no-browser", openBrowser, OAuthAuthOptions{NoBrowser: true})
		ch <- authResult{err: err}
	}()

	time.Sleep(200 * time.Millisecond)

	if browserCalled {
		t.Error("openBrowser should not be called when NoBrowser=true")
	}
	_ = capturedURL

	cancel()
	select {
	case <-ch:
	case <-time.After(6 * time.Second):
		t.Fatal("Authenticate did not return after context cancel")
	}
}

// Regression: --no-browser must bind a real listener that accepts connections.
// The old code used newManualFlow which passed nil listener, causing a 404.
func TestAuthenticate_NoBrowser_ListenerAcceptsConnections(t *testing.T) {
	flow, err := newLoopbackFlow(&oauth2.Config{
		ClientID: "test-client-id",
		Scopes:   []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	})
	if err != nil {
		t.Fatalf("newLoopbackFlow failed: %v", err)
	}
	defer flow.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	flow.StartCallbackServer(ctx)

	if flow.listener == nil {
		t.Fatal("Loopback flow must have a non-nil listener")
	}

	addr := flow.listener.Addr().(*net.TCPAddr)
	callbackURL := fmt.Sprintf("http://127.0.0.1:%d/callback?state=%s&code=test",
		addr.Port, url.QueryEscape(flow.state))

	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("Connection to callback port refused (original bug): %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Callback returned %d, want 200", resp.StatusCode)
	}
}

func TestAuthenticate_BrowserOpens(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"tok","refresh_token":"rtok","expires_in":3600,"token_type":"Bearer"}`)
	}))
	defer mockServer.Close()

	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)
	mgr.oauthConfig = &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Scopes:       []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
	}

	browserCalled := false
	openBrowser := func(u string) error {
		browserCalled = true
		parsed, err := url.Parse(u)
		if err != nil {
			return err
		}
		state := parsed.Query().Get("state")
		redirectURI := parsed.Query().Get("redirect_uri")
		callbackURL := fmt.Sprintf("%s?state=%s&code=browser-code", redirectURI, url.QueryEscape(state))
		resp, err := http.Get(callbackURL)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	}

	t.Setenv("DISPLAY", ":0")
	for _, k := range []string{"CI", "GITHUB_ACTIONS", "SSH_CONNECTION", "SSH_TTY", "GDRV_NO_BROWSER"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	creds, err := mgr.Authenticate(ctx, "test-browser", openBrowser, OAuthAuthOptions{NoBrowser: false})
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	if !browserCalled {
		t.Error("openBrowser should be called when NoBrowser=false")
	}
	if creds.AccessToken != "tok" {
		t.Errorf("Expected access_token 'tok', got %q", creds.AccessToken)
	}
}

func TestAuthenticate_BrowserFailsStillWaitsForCallback(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"tok2","refresh_token":"rtok2","expires_in":3600,"token_type":"Bearer"}`)
	}))
	defer mockServer.Close()

	tmpDir := t.TempDir()
	mgr := NewManager(tmpDir)
	mgr.oauthConfig = &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Scopes:       []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
	}

	urlCh := make(chan string, 1)
	openBrowser := func(u string) error {
		urlCh <- u
		return fmt.Errorf("browser not available")
	}

	t.Setenv("DISPLAY", ":0")
	for _, k := range []string{"CI", "GITHUB_ACTIONS", "SSH_CONNECTION", "SSH_TTY", "GDRV_NO_BROWSER"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	type authResult struct {
		accessToken string
		err         error
	}
	ch := make(chan authResult, 1)
	go func() {
		creds, err := mgr.Authenticate(ctx, "test-browser-fail", openBrowser, OAuthAuthOptions{NoBrowser: false})
		if err != nil {
			ch <- authResult{err: err}
			return
		}
		ch <- authResult{accessToken: creds.AccessToken}
	}()

	var capturedURL string
	select {
	case capturedURL = <-urlCh:
	case <-time.After(3 * time.Second):
		t.Fatal("openBrowser was not called")
	}

	parsed, err := url.Parse(capturedURL)
	if err != nil {
		t.Fatalf("Failed to parse auth URL: %v", err)
	}
	state := parsed.Query().Get("state")
	redirectURI := parsed.Query().Get("redirect_uri")
	callbackURL := fmt.Sprintf("%s?state=%s&code=manual-code", redirectURI, url.QueryEscape(state))

	resp, err := http.Get(callbackURL)
	if err != nil {
		t.Fatalf("Callback request failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Callback returned status %d, want 200", resp.StatusCode)
	}

	res := <-ch
	if res.err != nil {
		t.Fatalf("Authenticate failed: %v", res.err)
	}
	if res.accessToken != "tok2" {
		t.Errorf("Expected access_token 'tok2', got %q", res.accessToken)
	}
}

func TestGenerateState_Uniqueness(t *testing.T) {
	states := make(map[string]bool)

	for i := 0; i < 50; i++ {
		state, err := generateState()
		if err != nil {
			t.Fatalf("Failed to generate state: %v", err)
		}

		if states[state] {
			t.Errorf("Duplicate state generated: %s", state)
		}
		states[state] = true

		if len(state) == 0 {
			t.Error("Generated state is empty")
		}
	}
}
