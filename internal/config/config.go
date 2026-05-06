// Package config loads and persists the zd CLI's profile-based
// configuration from ~/.config/zd/config.yaml.
//
// Resolution order for any value (token, host, account_id, default_team)
// is enforced by the cmd/zd layer:
//
//  1. command-line flag (--token, --host, --account-id)
//  2. environment variable (ZENDUTY_API_TOKEN or per-profile token_env)
//  3. profile file value
//  4. compiled-in default
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// File is the on-disk schema for ~/.config/zd/config.yaml.
type File struct {
	DefaultProfile string             `yaml:"default_profile"`
	Profiles       map[string]Profile `yaml:"profiles"`
}

// Profile is a single named environment.
type Profile struct {
	Host        string `yaml:"host,omitempty"`
	TokenEnv    string `yaml:"token_env,omitempty"`
	Token       string `yaml:"token,omitempty"` // discouraged; use TokenEnv
	AccountID   string `yaml:"account_id,omitempty"`
	DefaultTeam string `yaml:"default_team,omitempty"`
}

// ErrNoProfile is returned by Resolve when the named profile is missing.
var ErrNoProfile = errors.New("config: profile not found")

// DefaultPath returns the canonical config file path:
// $XDG_CONFIG_HOME/zd/config.yaml or ~/.config/zd/config.yaml.
func DefaultPath() (string, error) {
	if env := os.Getenv("XDG_CONFIG_HOME"); env != "" {
		return filepath.Join(env, "zd", "config.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "zd", "config.yaml"), nil
}

// Load reads the config file at path. A missing file is not an error; it
// returns an empty File so first-run users can layer flag/env values.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &File{Profiles: map[string]Profile{}}, nil
		}
		return nil, err
	}
	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if f.Profiles == nil {
		f.Profiles = map[string]Profile{}
	}
	return &f, nil
}

// Save writes the config file, creating parent directories as needed
// and using restrictive 0600 perms because it may contain a token.
func Save(path string, f *File) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// Resolve returns the profile to use given an explicit name (may be ""
// which falls back to DefaultProfile).
func (f *File) Resolve(name string) (Profile, string, error) {
	if name == "" {
		name = f.DefaultProfile
	}
	if name == "" {
		return Profile{}, "", nil // fully implicit; caller layers flags/env
	}
	p, ok := f.Profiles[name]
	if !ok {
		return Profile{}, name, fmt.Errorf("%w: %s", ErrNoProfile, name)
	}
	return p, name, nil
}
