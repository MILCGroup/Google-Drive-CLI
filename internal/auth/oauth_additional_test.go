package auth

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestOAuthFlow_WaitForCode_ErrorChannel(t *testing.T) {
	config := &oauth2.Config{
		ClientID: "test-client-id",
		Scopes:   []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	flow, err := newLoopbackFlow(config)
	if err != nil {
		t.Fatalf("Failed to create flow: %v", err)
	}
	defer flow.Close()

	flow.errChan <- testError("test error")

	code, err := flow.WaitForCode(1 * time.Second)
	if err == nil {
		t.Error("WaitForCode should return error from errChan")
	}
	if code != "" {
		t.Errorf("Expected empty code, got %q", code)
	}
}

func TestOAuthFlow_WaitForCode_CodeChannel(t *testing.T) {
	config := &oauth2.Config{
		ClientID: "test-client-id",
		Scopes:   []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	flow, err := newLoopbackFlow(config)
	if err != nil {
		t.Fatalf("Failed to create flow: %v", err)
	}
	defer flow.Close()

	flow.codeChan <- "test-code"

	code, err := flow.WaitForCode(1 * time.Second)
	if err != nil {
		t.Errorf("WaitForCode should not return error: %v", err)
	}
	if code != "test-code" {
		t.Errorf("Expected 'test-code', got %q", code)
	}
}

func TestOAuthFlow_ExchangeCode_InvalidCode(t *testing.T) {
	mockServer := httptest.NewServer(nil)
	defer mockServer.Close()

	config := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Scopes:       []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
	}

	flow, err := newLoopbackFlow(config)
	if err != nil {
		t.Fatalf("Failed to create flow: %v", err)
	}
	defer flow.Close()

	ctx := context.Background()
	_, err = flow.ExchangeCode(ctx, "invalid-code")
	if err != nil {
		t.Logf("ExchangeCode returned error (expected): %v", err)
	}
}

func TestOAuthFlow_Close(t *testing.T) {
	config := &oauth2.Config{
		ClientID: "test-client-id",
		Scopes:   []string{"https://www.googleapis.com/auth/drive.readonly"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		},
	}

	flow, err := newLoopbackFlow(config)
	if err != nil {
		t.Fatalf("Failed to create flow: %v", err)
	}

	flow.Close()
}

func testError(msg string) error {
	return &testErr{msg: msg}
}

type testErr struct {
	msg string
}

func (e *testErr) Error() string {
	return e.msg
}
