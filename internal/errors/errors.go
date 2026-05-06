// Package errors defines the typed exit-code errors used across cmd/zd.
//
// Every cobra RunE that returns a *ExitError lets main translate it into
// the documented exit-code contract:
//
//	0 - success
//	1 - API error (4xx / 5xx response, retries exhausted)
//	2 - usage error (bad flags, missing required args)
//	3 - config error (no token, no profile, malformed config file)
//	4 - network or retry exhausted (DNS, TLS, connection refused, deadline)
//
// Errors print a concise human line on stderr plus a structured JSON
// payload (when --debug or stderr is not a TTY) so skills can parse them.
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Exit codes used by the CLI.
const (
	CodeOK      = 0
	CodeAPI     = 1
	CodeUsage   = 2
	CodeConfig  = 3
	CodeNetwork = 4
)

// ExitError carries an exit code, a human message, and an optional cause.
// cmd/zd unwraps these in main and exits with the embedded Code.
type ExitError struct {
	Code   int
	Msg    string
	URL    string
	Status int
	Cause  error
}

// Error returns a one-line human message.
func (e *ExitError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Cause)
	}
	return e.Msg
}

// Unwrap supports errors.Is / errors.As.
func (e *ExitError) Unwrap() error { return e.Cause }

// MarshalJSON renders the structured error payload that the CLI writes
// to stderr to aid scripting. Keys are stable; missing fields are omitted.
func (e *ExitError) MarshalJSON() ([]byte, error) {
	type wire struct {
		Error  string `json:"error"`
		Code   int    `json:"code"`
		Status int    `json:"status,omitempty"`
		URL    string `json:"url,omitempty"`
	}
	return json.Marshal(wire{
		Error:  e.Error(),
		Code:   e.Code,
		Status: e.Status,
		URL:    e.URL,
	})
}

// Usagef constructs a usage error (exit 2) with formatted message.
func Usagef(format string, args ...any) *ExitError {
	return &ExitError{Code: CodeUsage, Msg: fmt.Sprintf(format, args...)}
}

// Configf constructs a configuration error (exit 3).
func Configf(format string, args ...any) *ExitError {
	return &ExitError{Code: CodeConfig, Msg: fmt.Sprintf(format, args...)}
}

// Networkf constructs a transport-level error (exit 4).
func Networkf(cause error, format string, args ...any) *ExitError {
	return &ExitError{Code: CodeNetwork, Msg: fmt.Sprintf(format, args...), Cause: cause}
}

// API constructs an API error (exit 1) with the offending status / URL
// embedded so structured output captures useful debugging context.
func API(status int, url, body string) *ExitError {
	msg := fmt.Sprintf("API request failed (HTTP %d)", status)
	if body != "" {
		msg = fmt.Sprintf("%s: %s", msg, body)
	}
	return &ExitError{Code: CodeAPI, Msg: msg, Status: status, URL: url}
}

// FromHTTP inspects an *http.Response and returns nil on success or an
// appropriate *ExitError on a non-2xx response. The body is read up to
// maxBodyBytes for inclusion in the error message.
func FromHTTP(resp *http.Response, maxBodyBytes int64) error {
	if resp == nil {
		return nil
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body := ""
	if resp.Body != nil {
		buf, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
		_ = resp.Body.Close()
		body = string(buf)
	}
	url := ""
	if resp.Request != nil && resp.Request.URL != nil {
		url = resp.Request.URL.String()
	}
	e := API(resp.StatusCode, url, body)
	return e
}

// Code returns the exit code for an arbitrary error: ExitError honours its
// embedded Code; everything else is treated as a generic API error.
func Code(err error) int {
	if err == nil {
		return CodeOK
	}
	var ee *ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return CodeAPI
}
