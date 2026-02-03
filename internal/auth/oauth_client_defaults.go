package auth

// DefaultPublicOAuthClientID is a public OAuth client ID embedded in the binary.
// It is safe to embed because public/installed clients rely on PKCE, not a secret.
const DefaultPublicOAuthClientID = "243995828363-egpe1e3ked6bac0ed9qd8boa2o1jmfe1.apps.googleusercontent.com"

// BundledOAuthClientID and BundledOAuthClientSecret can be set at build time
// via -ldflags. The secret is optional for public clients.
var (
	BundledOAuthClientID     string
	BundledOAuthClientSecret string
)

// GetBundledOAuthClient returns the bundled/default OAuth client credentials.
// The secret may be empty for public clients.
func GetBundledOAuthClient() (string, string, bool) {
	clientID := BundledOAuthClientID
	if clientID == "" {
		clientID = DefaultPublicOAuthClientID
	}
	if clientID == "" {
		return "", "", false
	}
	return clientID, BundledOAuthClientSecret, true
}
