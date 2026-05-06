package main

import (
	"github.com/spf13/cobra"

	"github.com/hackath0r/zd-cli/internal/output"
	"github.com/hackath0r/zd-cli/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version, commit and build date",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			format, err := outputFormat(cmd)
			if err != nil {
				return err
			}
			info := version.Get()

			spec := &output.TableSpec{
				Headers: []string{"version", "commit", "date"},
				Row: func(item any) []string {
					i := item.(version.Info)
					return []string{i.Version, i.Commit, i.Date}
				},
			}
			return output.Render(cmd.OutOrStdout(), []version.Info{info}, format, output.Options{
				Table:    spec,
				Template: outputTemplate(cmd),
			})
		},
	}
}
