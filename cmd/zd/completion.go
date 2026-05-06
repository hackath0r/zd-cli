package main

import (
	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "completion {bash|zsh|fish|powershell}",
		Short:                 "Generate shell completion scripts",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Long: `Generate shell completion scripts for zd.

Bash:
  source <(zd completion bash)
  # to load on every shell:
  zd completion bash > /etc/bash_completion.d/zd

Zsh (one-time):
  source <(zd completion zsh)
  # to load on every shell:
  zd completion zsh > "${fpath[1]}/_zd"

Fish:
  zd completion fish | source
  zd completion fish > ~/.config/fish/completions/zd.fish

PowerShell:
  zd completion powershell | Out-String | Invoke-Expression
  zd completion powershell > $PROFILE`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletionV2(out, true)
			case "zsh":
				return cmd.Root().GenZshCompletion(out)
			case "fish":
				return cmd.Root().GenFishCompletion(out, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(out)
			}
			return nil
		},
	}
	return cmd
}
