// Command zd is the CLI for the Zenduty / Xurrent IMR REST API.
//
// The same binary also runs as `ximr` when invoked via the symlink installed
// by every distribution channel (Homebrew, Scoop, install.sh, install.ps1).
package main

import (
	"fmt"
	"os"

	"github.com/hackath0r/zd-cli/internal/version"
)

func main() {
	// Placeholder entrypoint. Cobra root command lands in cmd/zd/root.go in
	// the next implementation step.
	info := version.Get()
	fmt.Fprintf(os.Stdout, "zd %s (commit %s, built %s)\n", info.Version, info.Commit, info.Date)
}
