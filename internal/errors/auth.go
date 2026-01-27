package errors

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dl-alexandre/gdrv/internal/utils"
	"golang.org/x/oauth2"
)

type oauthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorURI         string `json:"error_uri"`
}

func ClassifyAuthRefreshError(err error) error {
	if err == nil {
		return nil
	}

	builder := utils.NewCLIError(utils.ErrCodeAuthExpired, "Token refresh failed.")

	var retrieveErr *oauth2.RetrieveError
	if errors.As(err, &retrieveErr) {
		builder.WithHTTPStatus(retrieveErr.Response.StatusCode)
		resp := oauthErrorResponse{}
		if retrieveErr.Body != nil {
			_ = json.Unmarshal(retrieveErr.Body, &resp)
		}
		if resp.Error != "" {
			builder.WithContext("oauthError", resp.Error)
		}
		if resp.ErrorDescription != "" {
			builder.WithContext("oauthErrorDescription", resp.ErrorDescription)
		}
		switch resp.Error {
		case "invalid_grant":
			builder = utils.NewCLIError(utils.ErrCodeAuthExpired, "Refresh token expired or revoked.").WithContext("suggestedAction", "run 'gdrv auth login' to re-authenticate")
		case "invalid_client", "unauthorized_client":
			builder = utils.NewCLIError(utils.ErrCodeAuthRequired, "OAuth client is not authorized.").WithContext("suggestedAction", "verify client ID/secret and run 'gdrv auth login'")
		case "access_denied":
			builder = utils.NewCLIError(utils.ErrCodeAuthRequired, "Access denied during token refresh.").WithContext("suggestedAction", "run 'gdrv auth login' to re-authenticate")
		default:
			builder.WithContext("suggestedAction", "run 'gdrv auth login' to re-authenticate")
		}
		applyClockSkew(builder, retrieveErr.Response.Header)
		return utils.NewAppError(builder.Build())
	}

	builder.WithContext("suggestedAction", "run 'gdrv auth login' to re-authenticate")
	return utils.NewAppError(builder.Build())
}

func applyAuthRemediation(builder *utils.CLIErrorBuilder, status int, header http.Header) {
	switch status {
	case http.StatusUnauthorized:
		builder.WithContext("suggestedAction", "run 'gdrv auth login' to re-authenticate")
	case http.StatusForbidden:
		builder.WithContext("suggestedAction", "verify OAuth scopes and shared drive access, then re-authenticate if needed")
	}
	applyClockSkew(builder, header)
}

func applyClockSkew(builder *utils.CLIErrorBuilder, header http.Header) {
	if header == nil {
		return
	}
	dateHeader := header.Get("Date")
	if dateHeader == "" {
		return
	}
	remoteTime, err := http.ParseTime(dateHeader)
	if err != nil {
		return
	}
	skew := time.Since(remoteTime)
	if skew < 0 {
		skew = -skew
	}
	if skew > 5*time.Minute {
		builder.WithContext("clockSkewSeconds", int64(skew.Seconds()))
	}
}
