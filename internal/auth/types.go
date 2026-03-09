package auth

import (
	"time"

	"github.com/dl-alexandre/gdrv/internal/types"
)

// AuthSource represents the source of authentication credentials
type AuthSource string

const (
	// AuthSourceToken - GDRV_TOKEN env var (pre-obtained access token)
	AuthSourceToken AuthSource = "token"
	// AuthSourceCredentialsFile - GDRV_CREDENTIALS_FILE env var
	AuthSourceCredentialsFile AuthSource = "credentials_file"
	// AuthSourceProfile - Stored profile credentials (keyring/encrypted file)
	AuthSourceProfile AuthSource = "profile"
	// AuthSourcePlaintext - Plaintext file fallback (legacy/development)
	AuthSourcePlaintext AuthSource = "plaintext"
)

// AuthResolution contains metadata about how credentials were resolved
// This is always returned, even on failure, to enable debugging and transparency
type AuthResolution struct {
	Source      AuthSource `json:"source"`
	Reason      string     `json:"reason"`
	Subject     string     `json:"subject"` // profile name or impersonated user
	Scopes      []string   `json:"scopes"`
	Path        string     `json:"path,omitempty"`       // File path (if applicable)
	ExpiresAt   *time.Time `json:"expires_at,omitempty"` // Token expiry (if known)
	Refreshable bool       `json:"refreshable"`
	Timestamp   time.Time  `json:"timestamp"`
}

// ResolveOptions configures the auth resolution process
type ResolveOptions struct {
	Profile         string
	RequiredScopes  []string
	ImpersonateUser string // CLI flag value (takes precedence over GDRV_IMPERSONATE_USER)
}

// ResolvedCredentials bundles the resolution metadata with the actual credentials
type ResolvedCredentials struct {
	Resolution *AuthResolution
	Creds      *types.Credentials
}
