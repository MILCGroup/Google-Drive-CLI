package iamadmin

import (
	"testing"
	"time"

	"cloud.google.com/go/iam/admin/apiv1/adminpb"
	"github.com/milcgroup/gdrv/internal/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConvertServiceAccount(t *testing.T) {
	tests := []struct {
		name     string
		input    *adminpb.ServiceAccount
		expected types.ServiceAccount
	}{
		{
			name: "full service account",
			input: &adminpb.ServiceAccount{
				Name:           "projects/my-project/serviceAccounts/my-sa@my-project.iam.gserviceaccount.com",
				ProjectId:      "my-project",
				Email:          "my-sa@my-project.iam.gserviceaccount.com",
				DisplayName:    "My Service Account",
				Description:    "Service account for testing",
				Disabled:       false,
				Oauth2ClientId: "123456789",
			},
			expected: types.ServiceAccount{
				Name:           "projects/my-project/serviceAccounts/my-sa@my-project.iam.gserviceaccount.com",
				ProjectId:      "my-project",
				Email:          "my-sa@my-project.iam.gserviceaccount.com",
				DisplayName:    "My Service Account",
				Description:    "Service account for testing",
				Disabled:       false,
				Oauth2ClientId: "123456789",
			},
		},
		{
			name: "disabled service account",
			input: &adminpb.ServiceAccount{
				Name:      "projects/my-project/serviceAccounts/disabled@my-project.iam.gserviceaccount.com",
				ProjectId: "my-project",
				Email:     "disabled@my-project.iam.gserviceaccount.com",
				Disabled:  true,
			},
			expected: types.ServiceAccount{
				Name:      "projects/my-project/serviceAccounts/disabled@my-project.iam.gserviceaccount.com",
				ProjectId: "my-project",
				Email:     "disabled@my-project.iam.gserviceaccount.com",
				Disabled:  true,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertServiceAccount(tc.input)
			if result.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tc.expected.Name)
			}
			if result.ProjectId != tc.expected.ProjectId {
				t.Errorf("ProjectId: got %q, want %q", result.ProjectId, tc.expected.ProjectId)
			}
			if result.Email != tc.expected.Email {
				t.Errorf("Email: got %q, want %q", result.Email, tc.expected.Email)
			}
			if result.DisplayName != tc.expected.DisplayName {
				t.Errorf("DisplayName: got %q, want %q", result.DisplayName, tc.expected.DisplayName)
			}
			if result.Description != tc.expected.Description {
				t.Errorf("Description: got %q, want %q", result.Description, tc.expected.Description)
			}
			if result.Disabled != tc.expected.Disabled {
				t.Errorf("Disabled: got %v, want %v", result.Disabled, tc.expected.Disabled)
			}
			if result.Oauth2ClientId != tc.expected.Oauth2ClientId {
				t.Errorf("Oauth2ClientId: got %q, want %q", result.Oauth2ClientId, tc.expected.Oauth2ClientId)
			}
		})
	}
}

func TestConvertServiceAccountKey(t *testing.T) {
	validAfter := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	validBefore := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    *adminpb.ServiceAccountKey
		expected types.ServiceAccountKey
	}{
		{
			name: "full key",
			input: &adminpb.ServiceAccountKey{
				Name:            "projects/my-project/serviceAccounts/my-sa@my-project.iam.gserviceaccount.com/keys/key-id",
				KeyAlgorithm:    adminpb.ServiceAccountKeyAlgorithm(2), // KEY_ALG_RSA_2048
				KeyOrigin:       adminpb.ServiceAccountKeyOrigin(2),    // GOOGLE_PROVIDED
				ValidAfterTime:  timestamppb.New(validAfter),
				ValidBeforeTime: timestamppb.New(validBefore),
			},
			expected: types.ServiceAccountKey{
				Name:         "projects/my-project/serviceAccounts/my-sa@my-project.iam.gserviceaccount.com/keys/key-id",
				KeyAlgorithm: "KEY_ALG_RSA_2048",
				KeyOrigin:    "GOOGLE_PROVIDED",
				ValidAfter:   "2024-01-01 00:00:00",
				ValidBefore:  "2025-01-01 00:00:00",
			},
		},
		{
			name: "key without validity times",
			input: &adminpb.ServiceAccountKey{
				Name:         "projects/my-project/serviceAccounts/my-sa@my-project.iam.gserviceaccount.com/keys/key-id-2",
				KeyAlgorithm: adminpb.ServiceAccountKeyAlgorithm(0), // KEY_ALG_UNSPECIFIED
				KeyOrigin:    adminpb.ServiceAccountKeyOrigin(0),    // ORIGIN_UNSPECIFIED
			},
			expected: types.ServiceAccountKey{
				Name:         "projects/my-project/serviceAccounts/my-sa@my-project.iam.gserviceaccount.com/keys/key-id-2",
				KeyAlgorithm: "KEY_ALG_UNSPECIFIED",
				KeyOrigin:    "ORIGIN_UNSPECIFIED",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertServiceAccountKey(tc.input)
			if result.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tc.expected.Name)
			}
			if result.KeyAlgorithm != tc.expected.KeyAlgorithm {
				t.Errorf("KeyAlgorithm: got %q, want %q", result.KeyAlgorithm, tc.expected.KeyAlgorithm)
			}
			if result.KeyOrigin != tc.expected.KeyOrigin {
				t.Errorf("KeyOrigin: got %q, want %q", result.KeyOrigin, tc.expected.KeyOrigin)
			}
			if result.ValidAfter != tc.expected.ValidAfter {
				t.Errorf("ValidAfter: got %q, want %q", result.ValidAfter, tc.expected.ValidAfter)
			}
			if result.ValidBefore != tc.expected.ValidBefore {
				t.Errorf("ValidBefore: got %q, want %q", result.ValidBefore, tc.expected.ValidBefore)
			}
		})
	}
}

func TestConvertRole(t *testing.T) {
	tests := []struct {
		name     string
		input    *adminpb.Role
		expected types.IAMRole
	}{
		{
			name: "full role",
			input: &adminpb.Role{
				Name:        "projects/my-project/roles/customRole",
				Title:       "Custom Role",
				Description: "A custom role for testing",
				Stage:       adminpb.Role_GA,
			},
			expected: types.IAMRole{
				Name:        "projects/my-project/roles/customRole",
				Title:       "Custom Role",
				Description: "A custom role for testing",
				Stage:       "GA",
			},
		},
		{
			name: "minimal role",
			input: &adminpb.Role{
				Name:  "organizations/my-org/roles/basicRole",
				Title: "Basic Role",
				Stage: adminpb.Role_ALPHA,
			},
			expected: types.IAMRole{
				Name:  "organizations/my-org/roles/basicRole",
				Title: "Basic Role",
				Stage: "ALPHA",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := convertRole(tc.input)
			if result.Name != tc.expected.Name {
				t.Errorf("Name: got %q, want %q", result.Name, tc.expected.Name)
			}
			if result.Title != tc.expected.Title {
				t.Errorf("Title: got %q, want %q", result.Title, tc.expected.Title)
			}
			if result.Description != tc.expected.Description {
				t.Errorf("Description: got %q, want %q", result.Description, tc.expected.Description)
			}
			if result.Stage != tc.expected.Stage {
				t.Errorf("Stage: got %q, want %q", result.Stage, tc.expected.Stage)
			}
		})
	}
}

func TestManagerClose(t *testing.T) {
	t.Run("close with nil client", func(t *testing.T) {
		m := &Manager{}
		err := m.Close()
		if err != nil {
			t.Errorf("Close() with nil client should not error, got: %v", err)
		}
	})
}
