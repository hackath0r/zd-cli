package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hackath0r/zd-cli/internal/config"
	zerrors "github.com/hackath0r/zd-cli/internal/errors"
	"github.com/hackath0r/zd-cli/internal/output"
	"github.com/hackath0r/zd-cli/internal/version"
	"github.com/hackath0r/zd-cli/internal/zenduty"
)

// Global persistent flag names. Centralised so every subcommand reads
// from the same viper key and we never typo a flag binding.
const (
	flagOutput    = "output"
	flagProfile   = "profile"
	flagHost      = "host"
	flagToken     = "token"
	flagAccountID = "account-id"
	flagDebug     = "debug"
	flagDebugCurl = "debug-curl"
	flagTemplate  = "template"
	flagConfig    = "config"
	flagNoColor   = "no-color"
)

// newRootCmd builds the root cobra command. Done as a function (not a
// package-level var) so tests can construct fresh trees without relying
// on global state.
func newRootCmd() *cobra.Command {
	v := viper.New()
	cobra.OnInitialize(func() { initConfig(v) })

	cmd := &cobra.Command{
		Use:           binaryName(),
		Short:         "Cross-platform CLI for the Zenduty / Xurrent IMR REST API",
		Long:          longDescription(),
		Version:       version.Get().Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	flags := cmd.PersistentFlags()
	flags.String(flagOutput, "", "output format: json | yaml | table | template (default: table on TTY, json when piped)")
	flags.String(flagProfile, "", "config profile name (default: profiles.default_profile in the config file)")
	flags.String(flagHost, "", "API base URL (default: https://www.zenduty.com)")
	flags.String(flagToken, "", "API token (overrides env and profile)")
	flags.String(flagAccountID, "", "5-character account identifier; required for the 'zd event fire' command")
	flags.Bool(flagDebug, false, "log wire-level requests/responses to stderr")
	flags.Bool(flagDebugCurl, false, "print a copy-paste curl reproduction of every request to stderr")
	flags.String(flagTemplate, "", "Go text/template body, used when --output=template")
	flags.String(flagConfig, "", "config file path (default: $XDG_CONFIG_HOME/zd/config.yaml or ~/.config/zd/config.yaml)")
	flags.Bool(flagNoColor, false, "disable color output")

	for _, name := range []string{flagOutput, flagProfile, flagHost, flagToken, flagAccountID, flagDebug, flagDebugCurl, flagTemplate, flagConfig, flagNoColor} {
		_ = v.BindPFlag(name, flags.Lookup(name))
	}

	v.SetEnvPrefix("ZD")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	v.AutomaticEnv()
	// Aliases for the most-used vars so the legacy ZENDUTY_ prefix works.
	_ = v.BindEnv(flagToken, "ZENDUTY_API_TOKEN", "ZD_TOKEN")
	_ = v.BindEnv(flagHost, "ZENDUTY_HOST", "ZD_HOST")
	_ = v.BindEnv(flagAccountID, "ZENDUTY_ACCOUNT_ID", "ZD_ACCOUNT_ID")

	cmd.SetContext(context.WithValue(context.Background(), viperKey{}, v))

	cmd.AddCommand(
		newVersionCmd(),
		newCompletionCmd(),
		newDocsCmd(),
		newConfigCmd(),
		newIncidentCmd(),
		newEventCmd(),
		newOncallCmd(),
		newAccountCmd(),
		newUserCmd(),
		newTeamCmd(),
		newScheduleCmd(),
		newServiceCmd(),
		newEscalationPolicyCmd(),
		newPriorityCmd(),
		newTagCmd(),
		newSLACmd(),
		newPostmortemCmd(),
		newMaintenanceCmd(),
		newRouterCmd(),
		newAlertRuleCmd(),
		newIntegrationCmd(),
		newAnalyticsCmd(),
	)

	return cmd
}

func longDescription() string {
	return `zd is a fast, scriptable command-line interface for the Zenduty / Xurrent
Incident Management & Response (IMR) REST API.

The same binary also runs as 'ximr' through a symlink installed by every
distribution channel.

Pick a profile in ~/.config/zd/config.yaml or pass --token / --host on
each invocation. Output defaults to a human-readable table on a TTY and
to JSON when piped, so scripts and AI-driven skills get clean structured
data without extra flags.`
}

// binaryName returns "zd" or "ximr" depending on how the binary was
// invoked. Used as cobra's Use to keep help and completion accurate.
func binaryName() string {
	exe := os.Args[0]
	switch base := strings.ToLower(filepath.Base(exe)); base {
	case "ximr", "ximr.exe":
		return "ximr"
	default:
		return "zd"
	}
}

// viperKey is a private context key under which the cobra root stores the
// viper instance. Subcommands fetch it via viperFromCmd.
type viperKey struct{}

func viperFromCmd(cmd *cobra.Command) *viper.Viper {
	if v, ok := cmd.Context().Value(viperKey{}).(*viper.Viper); ok && v != nil {
		return v
	}
	// Walk up to the root if a child command lost the context. Fallback
	// only; should never trigger in practice because cobra propagates ctx.
	for c := cmd; c != nil; c = c.Parent() {
		if v, ok := c.Context().Value(viperKey{}).(*viper.Viper); ok && v != nil {
			return v
		}
	}
	return viper.New()
}

func initConfig(v *viper.Viper) {
	cfgPath := v.GetString(flagConfig)
	if cfgPath == "" {
		def, err := config.DefaultPath()
		if err != nil {
			fmt.Fprintln(os.Stderr, "warning: cannot resolve home directory:", err)
			return
		}
		cfgPath = def
	}
	v.Set(flagConfig, cfgPath)
}

// resolveConfig returns the effective zenduty.Config for the current command
// invocation, applying flag > env > profile > default precedence.
func resolveConfig(cmd *cobra.Command) (zenduty.Config, error) {
	v := viperFromCmd(cmd)

	cfgPath := v.GetString(flagConfig)
	file, err := config.Load(cfgPath)
	if err != nil {
		return zenduty.Config{}, zerrors.Configf("read config: %v", err)
	}

	profileName := v.GetString(flagProfile)
	prof, _, err := file.Resolve(profileName)
	if err != nil {
		return zenduty.Config{}, zerrors.Configf("%v", err)
	}

	host := firstNonEmpty(v.GetString(flagHost), prof.Host, zenduty.DefaultHost)
	accountID := firstNonEmpty(v.GetString(flagAccountID), prof.AccountID)

	token := v.GetString(flagToken)
	if token == "" && prof.TokenEnv != "" {
		token = os.Getenv(prof.TokenEnv)
	}
	if token == "" {
		token = prof.Token
	}
	if token == "" {
		return zenduty.Config{}, zerrors.Configf("no API token configured (set --token, ZENDUTY_API_TOKEN, or run `zd config init`)")
	}

	return zenduty.Config{
		Host:             host,
		Token:            token,
		AccountID:        accountID,
		UserAgentVersion: version.Get().Version,
		Debug:            v.GetBool(flagDebug),
	}, nil
}

// clientFromCmd builds a typed Zenduty client for the current command.
// Use this in every RunE that needs to talk to the API.
func clientFromCmd(cmd *cobra.Command) (*zenduty.ClientWithResponses, zenduty.Config, error) {
	cfg, err := resolveConfig(cmd)
	if err != nil {
		return nil, zenduty.Config{}, err
	}
	cli, err := zenduty.New(cfg)
	if err != nil {
		return nil, zenduty.Config{}, zerrors.Configf("init client: %v", err)
	}
	return cli, cfg, nil
}

// outputFormat resolves the output format for the current command, picking
// up the --output flag and falling back to TTY detection on stdout.
func outputFormat(cmd *cobra.Command) (output.Format, error) {
	v := viperFromCmd(cmd)
	return output.Detect(v.GetString(flagOutput), cmd.OutOrStdout())
}

// outputTemplate returns the --template body or "".
func outputTemplate(cmd *cobra.Command) string {
	return viperFromCmd(cmd).GetString(flagTemplate)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
