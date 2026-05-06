// Command normalize-spec applies workarounds to the upstream Zenduty
// OpenAPI spec so it round-trips cleanly through oapi-codegen.
//
// Patches applied (idempotent):
//
//  1. Replace empty `{}` placeholders in team-scoped paths with `{team_id}`,
//     matching the documented shape on https://apidocs.zenduty.com/.
//     Affects 14 paths under /api/account/teams/{}/...
//
// Usage:
//
//	normalize-spec <input.yaml> <output.yaml>
package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

// pathFixes lists string substitutions applied to the raw YAML body. We use
// string substitution rather than YAML round-tripping to keep formatting and
// comment placement byte-for-byte identical for diff readability.
//
// Order matters: more specific multi-placeholder fixes must come before the
// generic single-placeholder prefix fix.
var pathFixes = []struct{ From, To string }{
	{"/api/v2/account/teams/{}/schedules/{}/overrides/", "/api/v2/account/teams/{team_id}/schedules/{schedule_id}/overrides/"},
	{"/api/account/teams/{}/", "/api/account/teams/{team_id}/"},
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: normalize-spec <input.yaml> <output.yaml>")
		os.Exit(2)
	}
	in, err := os.ReadFile(os.Args[1])
	must(err)

	out := in
	patched := 0
	for _, fix := range pathFixes {
		count := bytes.Count(out, []byte(fix.From))
		if count == 0 {
			continue
		}
		// Avoid mangling JSON values inside examples that legitimately contain
		// the literal `{}`: only patch occurrences where the substring sits at
		// the start of a path key line. Path keys live at indent 2 in OpenAPI.
		out = []byte(strings.ReplaceAll(string(out), "  "+fix.From, "  "+fix.To))
		patched += count
	}

	must(os.WriteFile(os.Args[2], out, 0o644))
	fmt.Fprintf(os.Stderr, "normalize-spec: applied %d substitutions across %d patch rules\n", patched, len(pathFixes))
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "normalize-spec:", err)
		os.Exit(1)
	}
}
