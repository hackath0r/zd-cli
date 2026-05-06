package main

import (
	"context"
	"encoding/json"
	"errors"
	stderrors "errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/output"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

// renderResult is the small bridge between command RunEs and the output
// renderer. Every command produces some JSON-shaped value plus an
// optional table projection; the caller passes both and we honour the
// resolved --output flag.
func renderResult(cmd *cobra.Command, value any, table *output.TableSpec) error {
	format, err := outputFormat(cmd)
	if err != nil {
		return err
	}
	return output.Render(cmd.OutOrStdout(), value, format, output.Options{
		Table:    table,
		Template: outputTemplate(cmd),
	})
}

// httpResponse is the small subset of *<Op>Response we need to build
// errors uniformly. Every generated *WithResponse type embeds *http.Response
// as HTTPResponse and exposes Body []byte and StatusCode().
type httpResponse interface {
	StatusCode() int
}

// requireOK ensures the API call succeeded; otherwise it builds a typed
// ExitError carrying the status code, URL and a body excerpt.
func requireOK(resp httpResponse, body []byte, url string) error {
	code := resp.StatusCode()
	if code >= 200 && code < 300 {
		return nil
	}
	return zerrors.API(code, url, strings.TrimSpace(string(body)))
}

// callContext returns a context that respects the parent context but caps
// the per-call deadline to give CLI users predictable behaviour against
// slow networks. 30s mirrors the default HTTP timeout in zenduty.Config.
func callContext(parent context.Context) (context.Context, context.CancelFunc) {
	if _, ok := parent.Deadline(); ok {
		return parent, func() {}
	}
	return context.WithTimeout(parent, 30*time.Second)
}

// ptrStr / ptrInt build pointers to literals; lots of generated request
// bodies use *string / *int because the OpenAPI schema marks fields as
// optional.
func ptrStr(s string) *string { return &s }
func ptrInt(i int) *int       { return &i }

// strPtrOrNil returns a pointer to s, or nil if s is empty. Useful when
// wiring optional flags into request bodies.
func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// derefStr dereferences a *string, returning "" on nil.
func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// derefInt dereferences a *int, returning 0 on nil.
func derefInt(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}

// derefBool dereferences a *bool, returning false on nil.
func derefBool(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

// derefTimeRFC3339 returns a compact RFC3339 timestamp or "" when nil.
func derefTimeRFC3339(p *time.Time) string {
	if p == nil || p.IsZero() {
		return ""
	}
	return p.UTC().Format(time.RFC3339)
}

// statusName converts an integer incident status into its human label,
// matching Zenduty's documented mapping (1=triggered, 2=acknowledged,
// 3=resolved). Unknown values are stringified verbatim.
func statusName(s *int) string {
	if s == nil {
		return ""
	}
	switch *s {
	case 1:
		return "triggered"
	case 2:
		return "acknowledged"
	case 3:
		return "resolved"
	default:
		return strconv.Itoa(*s)
	}
}

// urgencyName returns a human label for incident urgency.
func urgencyName(u *int) string {
	if u == nil {
		return ""
	}
	switch *u {
	case 0:
		return "low"
	case 1:
		return "high"
	default:
		return strconv.Itoa(*u)
	}
}

// truncate clips s to n characters, adding an ellipsis when shortened.
// Useful for table cells that hold free-form titles.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}

// looksLikeUUID returns true when s smells like a Zenduty unique_id.
// Zenduty unique_ids are 12-character lowercase alphanumeric strings or
// 36-character RFC4122 UUIDs depending on the resource. We treat anything
// non-numeric and longer than 6 characters as a unique_id candidate, and
// pure digits as an incident_number.
func looksLikeUUID(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return false
	}
	return len(s) > 6
}

// callAPI is a tiny wrapper that resolves the client + cfg, runs the
// supplied callback with a deadline-aware context, and tags transport
// errors as Network so they exit 4 instead of 1.
func callAPI[T any](cmd *cobra.Command, do func(ctx context.Context, cli *zenduty.ClientWithResponses, cfg zenduty.Config) (*T, error)) (*T, zenduty.Config, error) {
	cli, cfg, err := clientFromCmd(cmd)
	if err != nil {
		return nil, cfg, err
	}
	ctx, cancel := callContext(cmd.Context())
	defer cancel()

	out, err := do(ctx, cli, cfg)
	if err != nil {
		var ee *zerrors.ExitError
		if stderrors.As(err, &ee) {
			return nil, cfg, err
		}
		return nil, cfg, zerrors.Networkf(err, "API call failed")
	}
	return out, cfg, nil
}

// errIs is a typed alias to keep RunEs short.
func errIs(err, target error) bool { return errors.Is(err, target) }

// httpResp returns the http.Response from a generated *WithResponse type.
// Generated types embed an *http.Response field named HTTPResponse.
type httpRespHaver interface {
	httpResponse
}

// urlFromResp pulls the request URL off a generated response wrapper for
// inclusion in error messages. Returns "" if the request URL is missing.
func urlFromResp(r *http.Response) string {
	if r == nil || r.Request == nil || r.Request.URL == nil {
		return ""
	}
	return r.Request.URL.String()
}

// asString ensures a printable label even for nil-or-empty pointers.
func asString(p *string, fallback string) string {
	if p == nil || *p == "" {
		return fallback
	}
	return *p
}

// decodeList parses an HTTP response body as a JSON array of T, falling
// back to a single-element list when the upstream OpenAPI spec mistypes
// the endpoint as returning a single object. Many Zenduty list endpoints
// have the latter quirk in their spec.
func decodeList[T any](body []byte) ([]T, error) {
	body = []byte(strings.TrimSpace(string(body)))
	if len(body) == 0 || string(body) == "null" {
		return nil, nil
	}
	if body[0] == '[' {
		var out []T
		if err := json.Unmarshal(body, &out); err != nil {
			return nil, err
		}
		return out, nil
	}
	var single T
	if err := json.Unmarshal(body, &single); err != nil {
		return nil, err
	}
	return []T{single}, nil
}

// decodeOne parses an HTTP response body as a single JSON object of T.
// Useful when the spec mistakenly types a single-object endpoint as a
// list, the inverse of decodeList.
func decodeOne[T any](body []byte) (*T, error) {
	body = []byte(strings.TrimSpace(string(body)))
	if len(body) == 0 || string(body) == "null" {
		return nil, nil
	}
	if body[0] == '[' {
		var arr []T
		if err := json.Unmarshal(body, &arr); err != nil {
			return nil, err
		}
		if len(arr) == 0 {
			return nil, nil
		}
		return &arr[0], nil
	}
	var out T
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// notImplemented returns a RunE that exits 2 (usage) explaining the
// command is part of the spec but not yet wrapped by the CLI. Used as
// scaffolding for the v0.x admin CRUDs roadmap.
func notImplemented(api string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		return &zerrors.ExitError{
			Code: zerrors.CodeUsage,
			Msg:  fmt.Sprintf("not yet implemented in this CLI release; the underlying API (%s) is in the OpenAPI spec and will be wrapped in a future version. Track progress at https://github.com/hackath0r/zd-cli/issues", api),
		}
	}
}
