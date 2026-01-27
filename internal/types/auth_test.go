package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestCredentials_JSONMarshaling(t *testing.T) {
	now := time.Now()
	creds := Credentials{
		AccessToken:         "access-token-123",
		RefreshToken:        "refresh-token-456",
		ExpiryDate:          now,
		Scopes:              []string{"scope1", "scope2"},
		Type:                AuthTypeOAuth,
		ClientID:            "client-123",
		ServiceAccountEmail: "",
		ImpersonatedUser:    "",
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("Failed to marshal Credentials: %v", err)
	}

	var decoded Credentials
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal Credentials: %v", err)
	}

	if decoded.AccessToken != creds.AccessToken {
		t.Errorf("AccessToken = %s, want %s", decoded.AccessToken, creds.AccessToken)
	}

	if decoded.RefreshToken != creds.RefreshToken {
		t.Errorf("RefreshToken = %s, want %s", decoded.RefreshToken, creds.RefreshToken)
	}

	if !decoded.ExpiryDate.Equal(creds.ExpiryDate) {
		t.Errorf("ExpiryDate = %v, want %v", decoded.ExpiryDate, creds.ExpiryDate)
	}

	if len(decoded.Scopes) != len(creds.Scopes) {
		t.Errorf("Scopes length = %d, want %d", len(decoded.Scopes), len(creds.Scopes))
	}

	if decoded.Type != creds.Type {
		t.Errorf("Type = %s, want %s", decoded.Type, creds.Type)
	}

	if decoded.ClientID != creds.ClientID {
		t.Errorf("ClientID = %s, want %s", decoded.ClientID, creds.ClientID)
	}
}

func TestAuthType_Constants(t *testing.T) {
	tests := []struct {
		authType AuthType
		want     string
	}{
		{AuthTypeOAuth, "oauth"},
		{AuthTypeServiceAccount, "service_account"},
		{AuthTypeImpersonated, "impersonated"},
	}

	for _, tt := range tests {
		t.Run(string(tt.authType), func(t *testing.T) {
			if string(tt.authType) != tt.want {
				t.Errorf("AuthType = %s, want %s", tt.authType, tt.want)
			}
		})
	}
}

func TestCredentials_OAuthType(t *testing.T) {
	creds := Credentials{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiryDate:   time.Now().Add(time.Hour),
		Scopes:       []string{"drive"},
		Type:         AuthTypeOAuth,
		ClientID:     "client-123",
	}

	if creds.Type != AuthTypeOAuth {
		t.Errorf("Type = %s, want %s", creds.Type, AuthTypeOAuth)
	}

	if creds.ClientID == "" {
		t.Error("ClientID should not be empty for OAuth")
	}

	if creds.ServiceAccountEmail != "" {
		t.Error("ServiceAccountEmail should be empty for OAuth")
	}
}

func TestCredentials_ServiceAccountType(t *testing.T) {
	creds := Credentials{
		AccessToken:         "access-token",
		ExpiryDate:          time.Now().Add(time.Hour),
		Scopes:              []string{"drive"},
		Type:                AuthTypeServiceAccount,
		ServiceAccountEmail: "service@example.iam.gserviceaccount.com",
	}

	if creds.Type != AuthTypeServiceAccount {
		t.Errorf("Type = %s, want %s", creds.Type, AuthTypeServiceAccount)
	}

	if creds.ServiceAccountEmail == "" {
		t.Error("ServiceAccountEmail should not be empty for service account")
	}

	// Service accounts typically don't have refresh tokens
	if creds.RefreshToken != "" {
		t.Log("Note: Service accounts typically don't use refresh tokens")
	}
}

func TestCredentials_ImpersonatedType(t *testing.T) {
	creds := Credentials{
		AccessToken:         "access-token",
		ExpiryDate:          time.Now().Add(time.Hour),
		Scopes:              []string{"drive"},
		Type:                AuthTypeImpersonated,
		ServiceAccountEmail: "service@example.iam.gserviceaccount.com",
		ImpersonatedUser:    "user@example.com",
	}

	if creds.Type != AuthTypeImpersonated {
		t.Errorf("Type = %s, want %s", creds.Type, AuthTypeImpersonated)
	}

	if creds.ImpersonatedUser == "" {
		t.Error("ImpersonatedUser should not be empty for impersonated type")
	}

	if creds.ServiceAccountEmail == "" {
		t.Error("ServiceAccountEmail should not be empty for impersonated type")
	}
}

func TestStoredCredentials_JSONMarshaling(t *testing.T) {
	stored := StoredCredentials{
		Profile:             "default",
		AccessToken:         "access-token-123",
		RefreshToken:        "refresh-token-456",
		ExpiryDate:          "2024-12-31T23:59:59Z",
		Scopes:              []string{"scope1", "scope2"},
		Type:                AuthTypeOAuth,
		ClientID:            "client-123",
		ServiceAccountEmail: "",
		ImpersonatedUser:    "",
	}

	data, err := json.Marshal(stored)
	if err != nil {
		t.Fatalf("Failed to marshal StoredCredentials: %v", err)
	}

	var decoded StoredCredentials
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal StoredCredentials: %v", err)
	}

	if decoded.Profile != stored.Profile {
		t.Errorf("Profile = %s, want %s", decoded.Profile, stored.Profile)
	}

	if decoded.AccessToken != stored.AccessToken {
		t.Errorf("AccessToken = %s, want %s", decoded.AccessToken, stored.AccessToken)
	}

	if decoded.RefreshToken != stored.RefreshToken {
		t.Errorf("RefreshToken = %s, want %s", decoded.RefreshToken, stored.RefreshToken)
	}

	if decoded.ExpiryDate != stored.ExpiryDate {
		t.Errorf("ExpiryDate = %s, want %s", decoded.ExpiryDate, stored.ExpiryDate)
	}

	if len(decoded.Scopes) != len(stored.Scopes) {
		t.Errorf("Scopes length = %d, want %d", len(decoded.Scopes), len(stored.Scopes))
	}

	if decoded.Type != stored.Type {
		t.Errorf("Type = %s, want %s", decoded.Type, stored.Type)
	}

	if decoded.ClientID != stored.ClientID {
		t.Errorf("ClientID = %s, want %s", decoded.ClientID, stored.ClientID)
	}
}

func TestStoredCredentials_DifferentProfiles(t *testing.T) {
	profiles := []string{"default", "work", "personal"}

	for _, profile := range profiles {
		t.Run(profile, func(t *testing.T) {
			stored := StoredCredentials{
				Profile:      profile,
				AccessToken:  "token",
				ExpiryDate:   "2024-12-31T23:59:59Z",
				Scopes:       []string{"drive"},
				Type:         AuthTypeOAuth,
			}

			if stored.Profile != profile {
				t.Errorf("Profile = %s, want %s", stored.Profile, profile)
			}
		})
	}
}

func TestCredentials_RefreshTokenOmitEmpty(t *testing.T) {
	// Service account without refresh token
	creds := Credentials{
		AccessToken:         "access-token",
		ExpiryDate:          time.Now(),
		Scopes:              []string{"drive"},
		Type:                AuthTypeServiceAccount,
		ServiceAccountEmail: "service@example.com",
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(data)
	// refresh_token should not be present when empty
	if containsMiddle(jsonStr, `"refresh_token":""`) {
		t.Error("Empty refresh_token should be omitted")
	}
}

func TestCredentials_OptionalFieldsOmitted(t *testing.T) {
	creds := Credentials{
		AccessToken: "access-token",
		ExpiryDate:  time.Now(),
		Scopes:      []string{"drive"},
		Type:        AuthTypeOAuth,
	}

	data, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(data)

	// Optional fields should be omitted when empty
	if containsMiddle(jsonStr, `"client_id":""`) {
		t.Error("Empty client_id should be omitted")
	}

	if containsMiddle(jsonStr, `"service_account_email":""`) {
		t.Error("Empty service_account_email should be omitted")
	}

	if containsMiddle(jsonStr, `"impersonated_user":""`) {
		t.Error("Empty impersonated_user should be omitted")
	}
}

func TestStoredCredentials_OptionalFieldsOmitted(t *testing.T) {
	stored := StoredCredentials{
		Profile:     "default",
		AccessToken: "access-token",
		ExpiryDate:  "2024-12-31T23:59:59Z",
		Scopes:      []string{"drive"},
		Type:        AuthTypeOAuth,
	}

	data, err := json.Marshal(stored)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	jsonStr := string(data)

	// Optional fields should be omitted when empty
	if containsMiddle(jsonStr, `"refresh_token":""`) {
		t.Error("Empty refresh_token should be omitted")
	}

	if containsMiddle(jsonStr, `"client_id":""`) {
		t.Error("Empty client_id should be omitted")
	}
}

func TestCredentials_MultipleScopes(t *testing.T) {
	scopes := []string{
		"https://www.googleapis.com/auth/drive",
		"https://www.googleapis.com/auth/drive.file",
		"https://www.googleapis.com/auth/spreadsheets",
	}

	creds := Credentials{
		AccessToken: "access-token",
		ExpiryDate:  time.Now(),
		Scopes:      scopes,
		Type:        AuthTypeOAuth,
	}

	if len(creds.Scopes) != len(scopes) {
		t.Errorf("Scopes length = %d, want %d", len(creds.Scopes), len(scopes))
	}

	for i, scope := range creds.Scopes {
		if scope != scopes[i] {
			t.Errorf("Scope[%d] = %s, want %s", i, scope, scopes[i])
		}
	}
}

func TestCredentials_ExpiryDateHandling(t *testing.T) {
	now := time.Now()
	future := now.Add(1 * time.Hour)
	past := now.Add(-1 * time.Hour)

	tests := []struct {
		name       string
		expiryDate time.Time
		isExpired  bool
	}{
		{"future expiry", future, false},
		{"past expiry", past, true},
		{"current time", now, false}, // Consider current time as not expired
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds := Credentials{
				AccessToken: "access-token",
				ExpiryDate:  tt.expiryDate,
				Scopes:      []string{"drive"},
				Type:        AuthTypeOAuth,
			}

			// Note: The struct doesn't have an IsExpired method,
			// but we can verify the expiry date is stored correctly
			if !creds.ExpiryDate.Equal(tt.expiryDate) {
				t.Errorf("ExpiryDate = %v, want %v", creds.ExpiryDate, tt.expiryDate)
			}
		})
	}
}
