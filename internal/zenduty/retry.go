package zenduty

import (
	"bytes"
	"context"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"
)

// RetryConfig tunes the retry behaviour of the HTTP transport. Zero values
// fall back to the defaults below.
type RetryConfig struct {
	// MaxAttempts is the total number of attempts including the first call.
	// Zero means "use DefaultMaxAttempts".
	MaxAttempts int

	// InitialBackoff is the delay before the first retry. Subsequent retries
	// use exponential backoff up to MaxBackoff.
	InitialBackoff time.Duration

	// MaxBackoff caps the per-retry delay.
	MaxBackoff time.Duration
}

// Retry defaults. Tuned conservatively for an interactive CLI: we want
// resilience against transient blips without making the user wait minutes.
const (
	DefaultMaxAttempts    = 4
	DefaultInitialBackoff = 200 * time.Millisecond
	DefaultMaxBackoff     = 4 * time.Second
)

// retryTransport wraps an http.RoundTripper and retries on 429 + 5xx with
// exponential backoff. The request body is buffered up-front so it can be
// replayed on each attempt.
type retryTransport struct {
	base http.RoundTripper
	cfg  RetryConfig
}

// NewRetryTransport wraps base with retry semantics. If base is nil,
// http.DefaultTransport is used.
func NewRetryTransport(base http.RoundTripper, cfg RetryConfig) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = DefaultMaxAttempts
	}
	if cfg.InitialBackoff <= 0 {
		cfg.InitialBackoff = DefaultInitialBackoff
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = DefaultMaxBackoff
	}
	return &retryTransport{base: base, cfg: cfg}
}

func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil && req.Body != http.NoBody {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		_ = req.Body.Close()
	}

	var lastResp *http.Response
	var lastErr error

	for attempt := 0; attempt < rt.cfg.MaxAttempts; attempt++ {
		// Restore body for each attempt.
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(bodyBytes)), nil
			}
		}

		resp, err := rt.base.RoundTrip(req)
		lastResp, lastErr = resp, err

		switch {
		case err != nil:
			if !isTransientErr(err) {
				return nil, err
			}
		case !shouldRetryStatus(resp.StatusCode):
			return resp, nil
		default:
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}

		if attempt == rt.cfg.MaxAttempts-1 {
			break
		}

		delay := rt.backoff(attempt, resp)
		select {
		case <-time.After(delay):
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}

	return lastResp, lastErr
}

func (rt *retryTransport) backoff(attempt int, resp *http.Response) time.Duration {
	// Honour Retry-After when the server provides one.
	if resp != nil {
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(ra); err == nil && secs >= 0 {
				d := time.Duration(secs) * time.Second
				if d > rt.cfg.MaxBackoff {
					d = rt.cfg.MaxBackoff
				}
				return d
			}
			if t, err := http.ParseTime(ra); err == nil {
				if d := time.Until(t); d > 0 {
					if d > rt.cfg.MaxBackoff {
						d = rt.cfg.MaxBackoff
					}
					return d
				}
			}
		}
	}
	// Exponential backoff: initial * 2^attempt, capped.
	d := time.Duration(float64(rt.cfg.InitialBackoff) * math.Pow(2, float64(attempt)))
	if d > rt.cfg.MaxBackoff {
		d = rt.cfg.MaxBackoff
	}
	return d
}

func shouldRetryStatus(code int) bool {
	if code == http.StatusTooManyRequests {
		return true
	}
	return code >= 500 && code <= 599
}

func isTransientErr(err error) bool {
	// context cancellation is not transient.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	return true
}
