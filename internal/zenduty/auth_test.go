package zenduty

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestAuthEditor_UsesTokenScheme is the foot-gun guard: the upstream
// OpenAPI spec misleads with `bearerFormat: Token <api_key>` but the API
// actually requires "Authorization: Token <key>", not "Bearer". This test
// asserts the prefix explicitly and would fail loudly if anyone "fixed"
// the editor to use Bearer.
func TestAuthEditor_UsesTokenScheme(t *testing.T) {
	t.Parallel()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.invalid/", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	const tok = "secret-key-abc"
	if err := AuthEditor(tok)(context.Background(), req); err != nil {
		t.Fatalf("AuthEditor: %v", err)
	}

	got := req.Header.Get("Authorization")
	const want = "Token secret-key-abc"
	if got != want {
		t.Fatalf("Authorization header = %q, want %q", got, want)
	}
	if strings.HasPrefix(got, "Bearer ") {
		t.Fatal("Authorization header uses Bearer scheme; Zenduty requires the Token scheme")
	}

	if accept := req.Header.Get("Accept"); accept != "application/json" {
		t.Fatalf("Accept header = %q, want application/json", accept)
	}
}

func TestAuthEditor_EmptyTokenErrors(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.invalid/", nil)
	err := AuthEditor("")(context.Background(), req)
	if !errors.Is(err, ErrMissingToken) {
		t.Fatalf("AuthEditor(empty): err = %v, want ErrMissingToken", err)
	}
}

func TestAuthEditor_PreservesUserSetAccept(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.invalid/", nil)
	req.Header.Set("Accept", "application/xml")
	if err := AuthEditor("k")(context.Background(), req); err != nil {
		t.Fatalf("AuthEditor: %v", err)
	}
	if got := req.Header.Get("Accept"); got != "application/xml" {
		t.Fatalf("Accept header = %q, expected user-set value to be preserved", got)
	}
}

func TestUserAgentEditor_StampsCLIVersion(t *testing.T) {
	t.Parallel()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.invalid/", nil)
	if err := UserAgentEditor("v1.2.3")(context.Background(), req); err != nil {
		t.Fatalf("UserAgentEditor: %v", err)
	}
	got := req.Header.Get("User-Agent")
	if !strings.Contains(got, "zd-cli/v1.2.3") {
		t.Fatalf("User-Agent = %q, want substring zd-cli/v1.2.3", got)
	}
}

// TestNew_WiresAuthAndUA verifies that New() actually plumbs the editors
// through to outgoing requests by routing them at an httptest server.
func TestNew_WiresAuthAndUA(t *testing.T) {
	t.Parallel()

	var gotAuth, gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	cli, err := New(Config{
		Host:             srv.URL,
		Token:            "deadbeef",
		UserAgentVersion: "test",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// The simplest endpoint is "list account members". We don't care about
	// the response shape, only that the request carried the right headers.
	resp, err := cli.GetApiAccountMembersWithResponse(context.Background())
	if err != nil {
		t.Fatalf("GetApiAccountMembersWithResponse: %v", err)
	}
	if resp.StatusCode() != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode())
	}
	if gotAuth != "Token deadbeef" {
		t.Fatalf("server saw Authorization=%q, want %q", gotAuth, "Token deadbeef")
	}
	if !strings.HasPrefix(gotUA, "zd-cli/test") {
		t.Fatalf("server saw User-Agent=%q, want prefix zd-cli/test", gotUA)
	}
}
