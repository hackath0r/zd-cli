package zenduty

import (
	"errors"
	"net/http"
	"time"
)

// DefaultHost is the canonical Zenduty API host. Override via Config.Host
// once Xurrent migrates the public API to a xurrent.com domain.
const DefaultHost = "https://www.zenduty.com"

// Config is the runtime configuration for a zenduty.Client. The CLI builds
// this from its resolved profile (flag > env > config-file > default).
type Config struct {
	// Host is the API base URL, with scheme. Defaults to DefaultHost.
	Host string

	// Token is the API key. Sent as "Authorization: Token <Token>".
	// Required.
	Token string

	// AccountID is the 5-character account identifier used by the modern
	// Events ingestion endpoint /integration/{account_id}/generic/{key}/.
	// Optional for read APIs; required for `zd event fire`.
	AccountID string

	// UserAgentVersion is stamped into the User-Agent header. Defaults to
	// "dev" when empty.
	UserAgentVersion string

	// HTTPTimeout is the per-request timeout. Defaults to 30s.
	HTTPTimeout time.Duration

	// Retry overrides retry behaviour. Zero-value uses sane defaults.
	Retry RetryConfig

	// Debug, when set, attaches a request/response logger that writes wire
	// summaries to stderr. Wired up in the cmd/zd layer.
	Debug bool

	// HTTPDoer overrides the underlying transport entirely. Tests use this
	// to plug in httptest.Server backed clients or VCR cassettes.
	HTTPDoer HttpRequestDoer
}

// New constructs a *ClientWithResponses backed by an HTTP client with retry
// and the Token-prefixed Authorization header injected. It does not perform
// any network I/O.
func New(cfg Config) (*ClientWithResponses, error) {
	if cfg.Token == "" {
		return nil, ErrMissingToken
	}
	if cfg.Host == "" {
		cfg.Host = DefaultHost
	}
	if cfg.HTTPTimeout == 0 {
		cfg.HTTPTimeout = 30 * time.Second
	}
	if cfg.UserAgentVersion == "" {
		cfg.UserAgentVersion = "dev"
	}

	doer := cfg.HTTPDoer
	if doer == nil {
		doer = &http.Client{
			Timeout:   cfg.HTTPTimeout,
			Transport: NewRetryTransport(http.DefaultTransport, cfg.Retry),
		}
	}

	return NewClientWithResponses(
		cfg.Host,
		WithHTTPClient(doer),
		WithRequestEditorFn(AuthEditor(cfg.Token)),
		WithRequestEditorFn(UserAgentEditor(cfg.UserAgentVersion)),
	)
}

// AccountIDOrError returns cfg.AccountID or an error explaining why the
// caller needs to set it. Used by event ingestion commands that require
// the 5-character account identifier in the URL path.
func (cfg Config) AccountIDOrError() (string, error) {
	if cfg.AccountID == "" {
		return "", errors.New("zenduty: account_id is required (run `zd config set account_id=XXXXX` or pass --account-id)")
	}
	return cfg.AccountID, nil
}
