package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/dl-alexandre/gdrv/internal/utils"
)

func TestOAuthClientSecretHintBySource(t *testing.T) {
	tests := []struct {
		name     string
		source   oauthClientSource
		secret   string
		contains string
	}{
		{
			name:     "flags without secret",
			source:   oauthClientSourceFlags,
			secret:   "",
			contains: "--client-secret",
		},
		{
			name:     "env without secret",
			source:   oauthClientSourceEnv,
			secret:   "",
			contains: "GDRV_CLIENT_SECRET",
		},
		{
			name:     "config without secret",
			source:   oauthClientSourceConfig,
			secret:   "",
			contains: "oauthClientSecret",
		},
		{
			name:     "bundled without secret",
			source:   oauthClientSourceBundled,
			secret:   "",
			contains: "bundled OAuth credentials",
		},
		{
			name:     "unknown source",
			source:   oauthClientSource(""),
			secret:   "",
			contains: "",
		},
		{
			name:     "secret configured",
			source:   oauthClientSourceBundled,
			secret:   "configured",
			contains: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hint := oauthClientSecretHint(tc.source, tc.secret)
			if tc.contains == "" {
				if hint != "" {
					t.Fatalf("expected empty hint, got %q", hint)
				}
				return
			}
			if !strings.Contains(hint, tc.contains) {
				t.Fatalf("expected hint to contain %q, got %q", tc.contains, hint)
			}
		})
	}
}

func TestBuildAuthFlowErrorAddsMissingSecretHint(t *testing.T) {
	err := errors.New("oauth2: \"invalid_request\" \"client_secret is missing.\"")
	builder := buildAuthFlowError(err, oauthClientSourceBundled, "client-id", "")
	cliErr := builder.Build()

	if cliErr.Code != utils.ErrCodeAuthRequired {
		t.Fatalf("expected code %q, got %q", utils.ErrCodeAuthRequired, cliErr.Code)
	}
	if !strings.Contains(cliErr.Message, "client_secret is missing") {
		t.Fatalf("expected original error in message, got %q", cliErr.Message)
	}
	if !strings.Contains(cliErr.Message, "bundled OAuth credentials") {
		t.Fatalf("expected bundled hint in message, got %q", cliErr.Message)
	}
	if got, ok := cliErr.Context["oauthClientSource"]; !ok || got != string(oauthClientSourceBundled) {
		t.Fatalf("expected oauthClientSource context %q, got %#v", oauthClientSourceBundled, got)
	}
	if got, ok := cliErr.Context["clientIdConfigured"].(bool); !ok || !got {
		t.Fatalf("expected clientIdConfigured=true, got %#v", cliErr.Context["clientIdConfigured"])
	}
	if got, ok := cliErr.Context["clientSecretConfigured"].(bool); !ok || got {
		t.Fatalf("expected clientSecretConfigured=false, got %#v", cliErr.Context["clientSecretConfigured"])
	}
}

func TestBuildAuthFlowErrorDoesNotAddHintForUnrelatedError(t *testing.T) {
	err := errors.New("network timeout")
	builder := buildAuthFlowError(err, oauthClientSourceBundled, "client-id", "")
	cliErr := builder.Build()

	if cliErr.Message != "network timeout" {
		t.Fatalf("expected unchanged message, got %q", cliErr.Message)
	}
}
