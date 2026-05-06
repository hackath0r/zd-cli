package config

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.yaml")
	f, err := Load(path)
	if err != nil {
		t.Fatalf("Load(missing) should not error: %v", err)
	}
	if f == nil || len(f.Profiles) != 0 {
		t.Fatalf("expected empty File, got %+v", f)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	in := &File{
		DefaultProfile: "prod",
		Profiles: map[string]Profile{
			"prod": {
				Host:        "https://www.zenduty.com",
				TokenEnv:    "ZENDUTY_API_TOKEN",
				AccountID:   "acct-123",
				DefaultTeam: "team-abc",
			},
			"dev": {
				Host:  "https://staging.zenduty.com",
				Token: "raw-token-discouraged",
			},
		},
	}
	if err := Save(path, in); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.DefaultProfile != in.DefaultProfile {
		t.Fatalf("default_profile: got %q, want %q", got.DefaultProfile, in.DefaultProfile)
	}
	if len(got.Profiles) != 2 {
		t.Fatalf("profiles count: got %d, want 2", len(got.Profiles))
	}
	if got.Profiles["prod"].AccountID != "acct-123" {
		t.Fatalf("prod profile not preserved: %+v", got.Profiles["prod"])
	}
}

func TestResolveImplicitWhenNoDefault(t *testing.T) {
	f := &File{Profiles: map[string]Profile{}}
	p, name, err := f.Resolve("")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if name != "" || (p != Profile{}) {
		t.Fatalf("expected zero result, got name=%q profile=%+v", name, p)
	}
}

func TestResolveDefault(t *testing.T) {
	f := &File{
		DefaultProfile: "prod",
		Profiles:       map[string]Profile{"prod": {Host: "https://x"}},
	}
	p, name, err := f.Resolve("")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if name != "prod" || p.Host != "https://x" {
		t.Fatalf("got name=%q profile=%+v", name, p)
	}
}

func TestResolveMissing(t *testing.T) {
	f := &File{
		Profiles: map[string]Profile{"prod": {Host: "https://x"}},
	}
	_, _, err := f.Resolve("staging")
	if !errors.Is(err, ErrNoProfile) {
		t.Fatalf("expected ErrNoProfile, got %v", err)
	}
}

func TestDefaultPath(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if filepath.Base(filepath.Dir(got)) != "zd" {
		t.Fatalf("expected zd dir in %q", got)
	}
}
