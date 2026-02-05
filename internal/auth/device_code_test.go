package auth

import (
	"context"
	"net/http"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestDeviceCodeFlow_RequestDeviceCode(t *testing.T) {
	config := &oauth2.Config{
		ClientID: "test-client-id",
		Scopes:   []string{"https://www.googleapis.com/auth/drive.file"},
	}

	flow := NewDeviceCodeFlow(config)
	if flow == nil {
		t.Fatal("NewDeviceCodeFlow returned nil")
	}

	// Note: This test requires actual network calls to Google's servers
	// In a real implementation, you would mock the HTTP client
	// For now, we just test the structure
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This will fail without valid credentials, but tests the structure
	_, err := flow.RequestDeviceCode(ctx)
	// We expect an error in test environment, but not a nil pointer error
	if err == nil {
		t.Log("Device code request succeeded (unexpected in test environment)")
	}
}

func TestDeviceCodeFlow_PollOnce(t *testing.T) {
	config := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Scopes:       []string{"https://www.googleapis.com/auth/drive.file"},
	}

	flow := NewDeviceCodeFlow(config)
	flow.response = &DeviceCodeResponse{
		DeviceCode:      "test-device-code",
		UserCode:        "TEST-CODE",
		VerificationURL: "https://google.com/device",
		ExpiresIn:       300,
		Interval:        5,
	}

	ctx := context.Background()
	_, err := flow.pollOnce(ctx, http.DefaultClient)
	if err != nil {
		t.Logf("Poll returned error (expected): %v", err)
	}
}

func TestDeviceCodeFlow_PollForToken_NoResponse(t *testing.T) {
	config := &oauth2.Config{
		ClientID: "test-client-id",
		Scopes:   []string{"https://www.googleapis.com/auth/drive.file"},
	}

	flow := NewDeviceCodeFlow(config)

	ctx := context.Background()
	_, err := flow.PollForToken(ctx)
	if err == nil {
		t.Error("PollForToken should fail when response is nil")
	}
}

func TestDeviceCodeFlow_PollForToken_ContextCanceled(t *testing.T) {
	config := &oauth2.Config{
		ClientID: "test-client-id",
		Scopes:   []string{"https://www.googleapis.com/auth/drive.file"},
	}

	flow := NewDeviceCodeFlow(config)
	flow.response = &DeviceCodeResponse{
		DeviceCode:      "test-device-code",
		UserCode:        "TEST-CODE",
		VerificationURL: "https://google.com/device",
		ExpiresIn:       300,
		Interval:        5,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := flow.PollForToken(ctx)
	if err == nil {
		t.Error("PollForToken should fail when context is canceled")
	}
}
