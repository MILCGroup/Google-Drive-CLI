package types

import "time"

// Credentials represents OAuth2 or service account credentials
type Credentials struct {
	AccessToken         string    `json:"access_token"`
	RefreshToken        string    `json:"refresh_token,omitempty"`
	ExpiryDate          time.Time `json:"expiry_date"`
	Scopes              []string  `json:"scopes"`
	Type                AuthType  `json:"type"`
	ServiceAccountEmail string    `json:"service_account_email,omitempty"`
	ImpersonatedUser    string    `json:"impersonated_user,omitempty"`
}

type AuthType string

const (
	AuthTypeOAuth          AuthType = "oauth"
	AuthTypeServiceAccount AuthType = "service_account"
	AuthTypeImpersonated   AuthType = "impersonated"
)

// StoredCredentials represents credentials as stored in secure storage
type StoredCredentials struct {
	Profile             string   `json:"profile"`
	AccessToken         string   `json:"access_token"`
	RefreshToken        string   `json:"refresh_token,omitempty"`
	ExpiryDate          string   `json:"expiry_date"`
	Scopes              []string `json:"scopes"`
	Type                AuthType `json:"type"`
	ServiceAccountEmail string   `json:"service_account_email,omitempty"`
	ImpersonatedUser    string   `json:"impersonated_user,omitempty"`
}
