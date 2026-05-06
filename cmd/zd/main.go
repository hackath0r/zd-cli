// Command zd is the CLI for the Zenduty / Xurrent IMR REST API.
//
// The same binary also runs as `ximr` when invoked through the symlink
// installed by every distribution channel (Homebrew formula, Scoop manifest,
// install.sh, install.ps1, deb/rpm post-install). Behaviour is identical;
// only the displayed command name changes in help output.
package main

import (
	"errors"
	stderrors "errors"
	"fmt"
	"os"

	zerrors "github.com/hackath0r/zd-cli/internal/errors"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		writeError(err)
		os.Exit(zerrors.Code(err))
	}
}

func writeError(err error) {
	if err == nil {
		return
	}
	var ee *zerrors.ExitError
	if stderrors.As(err, &ee) {
		fmt.Fprintln(os.Stderr, "error:", ee.Error())
		return
	}
	if errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(os.Stderr, "error:", err)
		return
	}
	fmt.Fprintln(os.Stderr, "error:", err)
}
