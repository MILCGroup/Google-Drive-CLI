package errors

import (
	"net/http"
	"testing"
	"time"

	"github.com/milcgroup/gdrv/internal/utils"
	"golang.org/x/oauth2"
)

func TestClassifyAuthRefreshErrorInvalidGrant(t *testing.T) {
	retrieveErr := &oauth2.RetrieveError{
		Response: &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     http.Header{"Date": []string{time.Now().UTC().Format(http.TimeFormat)}},
		},
		Body: []byte(`{"error":"invalid_grant","error_description":"expired"}`),
	}

	err := ClassifyAuthRefreshError(retrieveErr)
	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.CLIError.Code != utils.ErrCodeAuthExpired {
		t.Fatalf("expected auth expired, got %s", appErr.CLIError.Code)
	}
	if appErr.CLIError.Context == nil || appErr.CLIError.Context["suggestedAction"] == "" {
		t.Fatalf("expected suggestedAction context to be set")
	}
}

func TestClassifyAuthRefreshErrorInvalidClient(t *testing.T) {
	retrieveErr := &oauth2.RetrieveError{
		Response: &http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     http.Header{"Date": []string{time.Now().UTC().Format(http.TimeFormat)}},
		},
		Body: []byte(`{"error":"invalid_client","error_description":"invalid"}`),
	}

	err := ClassifyAuthRefreshError(retrieveErr)
	appErr, ok := err.(*utils.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if appErr.CLIError.Code != utils.ErrCodeAuthRequired {
		t.Fatalf("expected auth required, got %s", appErr.CLIError.Code)
	}
}
