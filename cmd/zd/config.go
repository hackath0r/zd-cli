package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/hackath0r/zd-cli/internal/config"
	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage zd configuration profiles",
		Long: `Manage the per-user configuration file at $XDG_CONFIG_HOME/zd/config.yaml
(or ~/.config/zd/config.yaml). Configuration is layered: command-line flags
override environment variables override file values override defaults.`,
	}
	cmd.AddCommand(
		newConfigInitCmd(),
		newConfigShowCmd(),
		newConfigUseCmd(),
		newConfigSetCmd(),
		newConfigDoctorCmd(),
		newConfigPathCmd(),
	)
	return cmd
}

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Print the resolved config file path",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := config.DefaultPath()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), path)
			return nil
		},
	}
}

func newConfigInitCmd() *cobra.Command {
	var nonInteractive bool
	var name string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create or update a profile in the config file",
		Long: `Interactively create a new profile (or update an existing one) and write it
to the config file. Pass --no-interactive to write a profile from flags only.

Generate an API token at: https://www.zenduty.com (Account -> API Keys).`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := config.DefaultPath()
			if err != nil {
				return zerrors.Configf("resolve config path: %v", err)
			}
			file, err := config.Load(path)
			if err != nil {
				return zerrors.Configf("load %s: %v", path, err)
			}

			r := bufio.NewReader(os.Stdin)
			out := cmd.OutOrStdout()

			profileName := strings.TrimSpace(name)
			if profileName == "" && !nonInteractive {
				profileName = ask(r, out, "profile name", firstNonEmpty(file.DefaultProfile, "default"))
			}
			if profileName == "" {
				profileName = "default"
			}

			existing := file.Profiles[profileName]

			host := existing.Host
			tokenEnv := existing.TokenEnv
			accountID := existing.AccountID
			defaultTeam := existing.DefaultTeam

			if !nonInteractive {
				host = ask(r, out, "API host", firstNonEmpty(host, zenduty.DefaultHost))
				tokenEnv = ask(r, out, "env var holding the API token (recommended over storing the token directly)", firstNonEmpty(tokenEnv, "ZENDUTY_API_TOKEN"))
				accountID = ask(r, out, "5-character account_id (required for 'zd event fire'; press Enter to skip)", accountID)
				defaultTeam = ask(r, out, "default team unique_id (optional)", defaultTeam)
			}

			file.Profiles[profileName] = config.Profile{
				Host:        host,
				TokenEnv:    tokenEnv,
				AccountID:   accountID,
				DefaultTeam: defaultTeam,
			}
			if file.DefaultProfile == "" {
				file.DefaultProfile = profileName
			}

			if err := config.Save(path, file); err != nil {
				return zerrors.Configf("write %s: %v", path, err)
			}
			fmt.Fprintf(out, "wrote profile %q to %s\n", profileName, path)
			fmt.Fprintf(out, "next: export %s=<your-api-token> && zd config doctor\n", tokenEnv)
			return nil
		},
	}
	cmd.Flags().BoolVar(&nonInteractive, "no-interactive", false, "skip prompts; write a profile from flags only")
	cmd.Flags().StringVar(&name, "name", "", "profile name (default: prompt or 'default')")
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print the resolved config file contents",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := config.DefaultPath()
			if err != nil {
				return err
			}
			file, err := config.Load(path)
			if err != nil {
				return zerrors.Configf("load %s: %v", path, err)
			}
			// Redact any inline tokens before printing.
			for name, prof := range file.Profiles {
				if prof.Token != "" {
					prof.Token = "***redacted***"
					file.Profiles[name] = prof
				}
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "# %s\n", path)
			fmt.Fprintf(out, "default_profile: %s\n", file.DefaultProfile)
			fmt.Fprintln(out, "profiles:")
			for name, p := range file.Profiles {
				fmt.Fprintf(out, "  %s:\n", name)
				if p.Host != "" {
					fmt.Fprintf(out, "    host: %s\n", p.Host)
				}
				if p.TokenEnv != "" {
					fmt.Fprintf(out, "    token_env: %s\n", p.TokenEnv)
				}
				if p.AccountID != "" {
					fmt.Fprintf(out, "    account_id: %s\n", p.AccountID)
				}
				if p.DefaultTeam != "" {
					fmt.Fprintf(out, "    default_team: %s\n", p.DefaultTeam)
				}
			}
			return nil
		},
	}
}

func newConfigUseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile>",
		Short: "Set the default profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.DefaultPath()
			if err != nil {
				return err
			}
			file, err := config.Load(path)
			if err != nil {
				return zerrors.Configf("load %s: %v", path, err)
			}
			if _, ok := file.Profiles[args[0]]; !ok {
				return zerrors.Configf("profile %q does not exist; run `zd config init --name %s` to create it", args[0], args[0])
			}
			file.DefaultProfile = args[0]
			if err := config.Save(path, file); err != nil {
				return zerrors.Configf("write %s: %v", path, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "default_profile -> %s\n", args[0])
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	var profile string
	cmd := &cobra.Command{
		Use:   "set key=value [key=value ...]",
		Short: "Update profile fields",
		Long: `Update one or more fields on a profile. Recognised keys:

  host          API base URL
  token_env     environment variable holding the token (recommended)
  account_id    5-character account identifier
  default_team  team unique_id used as a fallback for team-scoped commands

Tokens themselves are not accepted on the command line for security; use
token_env=NAME and export the token in your shell.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.DefaultPath()
			if err != nil {
				return err
			}
			file, err := config.Load(path)
			if err != nil {
				return zerrors.Configf("load %s: %v", path, err)
			}
			if profile == "" {
				profile = file.DefaultProfile
			}
			if profile == "" {
				profile = "default"
			}
			p := file.Profiles[profile]
			for _, kv := range args {
				k, v, ok := strings.Cut(kv, "=")
				if !ok {
					return zerrors.Usagef("expected key=value, got %q", kv)
				}
				switch k {
				case "host":
					p.Host = v
				case "token_env":
					p.TokenEnv = v
				case "account_id":
					p.AccountID = v
				case "default_team":
					p.DefaultTeam = v
				default:
					return zerrors.Usagef("unknown key %q (allowed: host, token_env, account_id, default_team)", k)
				}
			}
			if file.Profiles == nil {
				file.Profiles = map[string]config.Profile{}
			}
			file.Profiles[profile] = p
			if err := config.Save(path, file); err != nil {
				return zerrors.Configf("write %s: %v", path, err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "updated profile %q in %s\n", profile, path)
			return nil
		},
	}
	cmd.Flags().StringVar(&profile, "name", "", "profile name (default: current default_profile)")
	return cmd
}

func newConfigDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Verify configuration by calling the API",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cli, cfg, err := clientFromCmd(cmd)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "host:       %s\n", cfg.Host)
			fmt.Fprintf(out, "account_id: %s\n", redacted(cfg.AccountID))
			fmt.Fprintf(out, "token:      %s\n", redacted(cfg.Token))
			fmt.Fprintln(out)

			ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
			defer cancel()

			start := time.Now()
			resp, err := cli.GetApiAccountMembersWithResponse(ctx)
			took := time.Since(start)
			if err != nil {
				return zerrors.Networkf(err, "calling GET /api/account/members/")
			}
			fmt.Fprintf(out, "GET /api/account/members/ -> %d in %s\n", resp.StatusCode(), took.Round(time.Millisecond))

			switch {
			case resp.StatusCode() == http.StatusOK:
				fmt.Fprintln(out, "ok: configuration looks healthy")
				return nil
			case resp.StatusCode() == http.StatusUnauthorized:
				return zerrors.Configf("token rejected (401); generate a fresh API key at https://www.zenduty.com -> Account -> API Keys")
			case resp.StatusCode() == http.StatusForbidden:
				return zerrors.Configf("token lacks permissions (403); ensure the key was issued for this account")
			default:
				return zerrors.API(resp.StatusCode(), "GET /api/account/members/", strings.TrimSpace(string(resp.Body)))
			}
		},
	}
}

func ask(r *bufio.Reader, w io.Writer, prompt, def string) string {
	if def != "" {
		fmt.Fprintf(w, "%s [%s]: ", prompt, def)
	} else {
		fmt.Fprintf(w, "%s: ", prompt)
	}
	line, _ := r.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

func redacted(s string) string {
	if s == "" {
		return "(empty)"
	}
	if len(s) <= 6 {
		return strings.Repeat("*", len(s))
	}
	return s[:3] + strings.Repeat("*", len(s)-6) + s[len(s)-3:]
}
