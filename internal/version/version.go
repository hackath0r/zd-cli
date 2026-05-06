// Package version exposes build-time version metadata, populated via -ldflags
// at build time by goreleaser and the Makefile. The defaults below are used
// when running with `go run` or `go install` without ldflags.
package version

// Build metadata populated via -ldflags at build time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// Info aggregates the build metadata for printing and structured output.
type Info struct {
	Version string `json:"version" yaml:"version"`
	Commit  string `json:"commit"  yaml:"commit"`
	Date    string `json:"date"    yaml:"date"`
}

// Get returns the current build info as a struct.
func Get() Info {
	return Info{Version: Version, Commit: Commit, Date: Date}
}
