package zenduty

import (
	"context"
	"errors"
	"net/http"
)

// ErrMissingToken is returned by AuthEditor when invoked with an empty token.
var ErrMissingToken = errors.New("zenduty: missing API token")

// AuthEditor returns a RequestEditorFn that injects the Zenduty / Xurrent IMR
// API authorization header on every outgoing request.
//
// Important: the API expects "Authorization: Token <key>", NOT "Bearer".
// The upstream OpenAPI spec declares bearerAuth with a misleading
// `bearerFormat: Token <api_key>` hint, which would lead a naive client to
// send "Authorization: Bearer <key>" and fail with 401. We always use the
// "Token" prefix here.
func AuthEditor(token string) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		if token == "" {
			return ErrMissingToken
		}
		req.Header.Set("Authorization", "Token "+token)
		// Some Zenduty endpoints are picky about Accept; force application/json.
		if req.Header.Get("Accept") == "" {
			req.Header.Set("Accept", "application/json")
		}
		return nil
	}
}

// UserAgentEditor stamps a recognizable User-Agent on every request so
// Zenduty / Xurrent can attribute traffic to the CLI.
func UserAgentEditor(version string) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", "zd-cli/"+version+" (+https://github.com/hackath0r/zd-cli)")
		return nil
	}
}
