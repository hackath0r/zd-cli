package zenduty

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newRespBody(s string) io.ReadCloser { return io.NopCloser(strings.NewReader(s)) }

func TestRetryTransport_RetriesOn5xxThenSucceeds(t *testing.T) {
	var calls int32
	base := rtFunc(func(_ *http.Request) (*http.Response, error) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			return &http.Response{StatusCode: 502, Body: newRespBody("bad gateway"), Header: make(http.Header)}, nil
		}
		return &http.Response{StatusCode: 200, Body: newRespBody("ok"), Header: make(http.Header)}, nil
	})
	cfg := RetryConfig{MaxAttempts: 4, InitialBackoff: 1 * time.Millisecond, MaxBackoff: 5 * time.Millisecond}
	rt := NewRetryTransport(base, cfg)

	req, _ := http.NewRequest(http.MethodGet, "https://x", http.NoBody)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status: got %d, want 200", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("call count: got %d, want 3", got)
	}
}

func TestRetryTransport_GivesUpAfterMaxAttempts(t *testing.T) {
	var calls int32
	base := rtFunc(func(_ *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return &http.Response{StatusCode: 503, Body: newRespBody("nope"), Header: make(http.Header)}, nil
	})
	cfg := RetryConfig{MaxAttempts: 3, InitialBackoff: 1 * time.Millisecond, MaxBackoff: 2 * time.Millisecond}
	rt := NewRetryTransport(base, cfg)

	req, _ := http.NewRequest(http.MethodGet, "https://x", http.NoBody)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 503 {
		t.Fatalf("expected to surface last 503, got %d", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Fatalf("call count: got %d, want 3", got)
	}
}

func TestRetryTransport_PermanentErrorBailsOut(t *testing.T) {
	var calls int32
	target := errors.New("dns: no such host")
	base := rtFunc(func(_ *http.Request) (*http.Response, error) {
		atomic.AddInt32(&calls, 1)
		return nil, target
	})
	cfg := RetryConfig{MaxAttempts: 3, InitialBackoff: 1 * time.Millisecond, MaxBackoff: 2 * time.Millisecond}
	rt := NewRetryTransport(base, cfg)

	req, _ := http.NewRequest(http.MethodGet, "https://x", http.NoBody)
	_, err := rt.RoundTrip(req)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if got := atomic.LoadInt32(&calls); got > 3 {
		t.Fatalf("call count: got %d", got)
	}
}

func TestRetryTransport_ReplaysBody(t *testing.T) {
	var seen []string
	base := rtFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		seen = append(seen, string(body))
		if len(seen) < 2 {
			return &http.Response{StatusCode: 503, Body: newRespBody(""), Header: make(http.Header)}, nil
		}
		return &http.Response{StatusCode: 200, Body: newRespBody("ok"), Header: make(http.Header)}, nil
	})
	cfg := RetryConfig{MaxAttempts: 3, InitialBackoff: 1 * time.Millisecond, MaxBackoff: 2 * time.Millisecond}
	rt := NewRetryTransport(base, cfg)

	req, _ := http.NewRequest(http.MethodPost, "https://x", bytes.NewBufferString("payload"))
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	defer resp.Body.Close()
	if len(seen) != 2 || seen[0] != "payload" || seen[1] != "payload" {
		t.Fatalf("body not replayed correctly: %#v", seen)
	}
}

func TestShouldRetryStatus(t *testing.T) {
	cases := map[int]bool{200: false, 400: false, 404: false, 429: true, 500: true, 503: true, 599: true, 600: false}
	for code, want := range cases {
		if got := shouldRetryStatus(code); got != want {
			t.Fatalf("shouldRetryStatus(%d): got %v, want %v", code, got, want)
		}
	}
}
