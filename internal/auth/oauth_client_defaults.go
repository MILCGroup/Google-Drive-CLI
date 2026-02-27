package auth

import "fmt"

var (
	BundledOAuthClientID     string
	BundledOAuthClientSecret string
	BundledBuildSource       string
)

func GetBundledOAuthClient() (string, string, bool) {
	clientID := BundledOAuthClientID
	clientSecret := BundledOAuthClientSecret

	if clientID == "" {
		return "", "", false
	}

	return clientID, clientSecret, true
}

func IsOfficialBuild() bool {
	return BundledBuildSource == "official"
}

func GetBuildInfo() string {
	if BundledBuildSource == "official" {
		return "official"
	}
	return "source"
}

func RequireOfficialBuild() error {
	if BundledBuildSource != "official" {
		return fmt.Errorf("bundled OAuth credentials require official signed build. Build from source with GDRV_CLIENT_ID/GDRV_CLIENT_SECRET env vars, or download official release from %s",
			"https://github.com/MILCGroup/Google-Drive-CLI/releases")
	}
	return nil
}