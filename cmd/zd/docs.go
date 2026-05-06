package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newDocsCmd() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:    "docs",
		Short:  "Regenerate command reference markdown into docs/",
		Hidden: true, // build-time helper, not a user-facing command
		RunE: func(cmd *cobra.Command, _ []string) error {
			if dir == "" {
				dir = "docs/commands"
			}
			if err := os.MkdirAll(dir, 0o750); err != nil {
				return err
			}
			root := cmd.Root()
			if err := doc.GenMarkdownTree(root, dir); err != nil {
				return err
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "wrote command reference to %s\n", filepath.Clean(dir))
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "out", "docs/commands", "output directory")
	return cmd
}
