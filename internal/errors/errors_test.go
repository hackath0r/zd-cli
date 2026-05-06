package errors

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestExitError_ErrorWithCause(t *testing.T) {
	cause := errors.New("dial tcp: timeout")
	e := Networkf(cause, "API call failed")
	if !strings.Contains(e.Error(), "API call failed") {
		t.Fatalf("error string missing message: %q", e.Error())
	}
	if !strings.Contains(e.Error(), "dial tcp: timeout") {
		t.Fatalf("error string missing cause: %q", e.Error())
	}
}

func TestExitError_NilSafe(t *testing.T) {
	var nilExit *ExitError
	if nilExit.Error() != "" {
		t.Fatalf("nil ExitError should stringify to empty")
	}
}

func TestExitError_Unwrap(t *testing.T) {
	cause := errors.New("boom")
	e := Networkf(cause, "x")
	if !errors.Is(e, cause) {
		t.Fatalf("errors.Is should find wrapped cause")
	}
}

func TestExitError_MarshalJSON(t *testing.T) {
	e := API(429, "https://x/api", "rate limited")
	b, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}
	var wire struct {
		Error  string `json:"error"`
		Code   int    `json:"code"`
		Status int    `json:"status"`
		URL    string `json:"url"`
	}
	if err := json.Unmarshal(b, &wire); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if wire.Code != CodeAPI {
		t.Fatalf("code: got %d, want %d", wire.Code, CodeAPI)
	}
	if wire.Status != 429 {
		t.Fatalf("status: got %d, want 429", wire.Status)
	}
	if wire.URL != "https://x/api" {
		t.Fatalf("url: got %q", wire.URL)
	}
}

func TestCode_TypedAndUntyped(t *testing.T) {
	if got := Code(nil); got != CodeOK {
		t.Fatalf("Code(nil): got %d, want %d", got, CodeOK)
	}
	if got := Code(Usagef("bad")); got != CodeUsage {
		t.Fatalf("Code(Usagef): got %d, want %d", got, CodeUsage)
	}
	if got := Code(Configf("bad")); got != CodeConfig {
		t.Fatalf("Code(Configf): got %d, want %d", got, CodeConfig)
	}
	if got := Code(errors.New("plain")); got != CodeAPI {
		t.Fatalf("Code(plain): got %d, want %d", got, CodeAPI)
	}
}

func TestFromHTTP(t *testing.T) {
	if err := FromHTTP(nil, 1024); err != nil {
		t.Fatalf("nil response should be no-op")
	}
	ok := &http.Response{StatusCode: 204, Body: io.NopCloser(strings.NewReader(""))}
	if err := FromHTTP(ok, 1024); err != nil {
		t.Fatalf("204 should be ok, got %v", err)
	}
	bad := &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader("server boom")),
		Request:    mustReq(t, "https://x/api/incidents/"),
	}
	err := FromHTTP(bad, 1024)
	var ee *ExitError
	if !errors.As(err, &ee) {
		t.Fatalf("expected *ExitError, got %T", err)
	}
	if ee.Status != 500 || ee.Code != CodeAPI {
		t.Fatalf("unexpected ExitError: %+v", ee)
	}
	if !strings.Contains(ee.Msg, "server boom") {
		t.Fatalf("body excerpt missing: %q", ee.Msg)
	}
}

func mustReq(t *testing.T, url string) *http.Request {
	t.Helper()
	r, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	return r
}
