package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type AuthMetadata struct {
	Profile        string   `json:"profile"`
	ClientIDHash   string   `json:"clientIdHash"`
	ClientIDLast4  string   `json:"clientIdLast4,omitempty"`
	Scopes         []string `json:"scopes,omitempty"`
	ExpiryDate     string   `json:"expiryDate,omitempty"`
	RefreshToken   bool     `json:"refreshToken"`
	CredentialType string   `json:"credentialType,omitempty"`
	StorageBackend string   `json:"storageBackend,omitempty"`
	UpdatedAt      string   `json:"updatedAt"`
}

func metadataFilePath(configDir, key string) string {
	return filepath.Join(configDir, "credentials", key+metadataSuffix)
}

func writeMetadata(configDir, key string, metadata *AuthMetadata) error {
	if err := os.MkdirAll(filepath.Join(configDir, "credentials"), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(metadataFilePath(configDir, key), data, 0600)
}

func readMetadata(path string) (*AuthMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var metadata AuthMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

func metadataTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
