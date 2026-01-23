package auth

import (
	"testing"
	"time"

	"github.com/dl-alexandre/gdrive/internal/types"
)

func TestManager_NeedsRefresh(t *testing.T) {
	mgr := NewManager("/tmp/test")

	tests := []struct {
		name     string
		expiry   time.Time
		expected bool
	}{
		{
			"Expired credentials",
			time.Now().Add(-1 * time.Hour),
			true,
		},
		{
			"Expiring soon (within 5 min)",
			time.Now().Add(3 * time.Minute),
			true,
		},
		{
			"Valid credentials",
			time.Now().Add(1 * time.Hour),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := &types.Credentials{
				ExpiryDate: tt.expiry,
			}
			got := mgr.NeedsRefresh(creds)
			if got != tt.expected {
				t.Errorf("NeedsRefresh() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestManager_ValidateScopes(t *testing.T) {
	mgr := NewManager("/tmp/test")

	creds := &types.Credentials{
		Scopes: []string{
			"https://www.googleapis.com/auth/drive.file",
			"https://www.googleapis.com/auth/drive.readonly",
		},
	}

	// Valid scope check
	err := mgr.ValidateScopes(creds, []string{"https://www.googleapis.com/auth/drive.file"})
	if err != nil {
		t.Errorf("ValidateScopes should pass for existing scope: %v", err)
	}

	// Missing scope check
	err = mgr.ValidateScopes(creds, []string{"https://www.googleapis.com/auth/drive"})
	if err == nil {
		t.Error("ValidateScopes should fail for missing scope")
	}
}
