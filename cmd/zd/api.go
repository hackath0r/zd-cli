package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	zerrors "github.com/hackath0r/zd-cli/internal/errors"
)

// readJSONBody resolves the JSON body for a write command in a single
// place. Precedence (first hit wins):
//
//  1. --data <inline JSON>
//  2. --body @path/to/file.json
//  3. --body -        (stdin, when followed by no other flags)
//  4. stdin (only if not a TTY and the user explicitly opts in via --stdin)
//
// Any literal merges from convenience flags happen on top of the
// decoded body in the caller. Returning a json.RawMessage keeps the
// caller free to either pass the raw bytes through to the typed client
// (via the *WithBodyWithResponse variant) or to decode it into a
// typed payload struct first.
func readJSONBody(cmd *cobra.Command, fromStdinFlag string) (json.RawMessage, error) {
	flags := cmd.Flags()

	if flags.Changed("data") {
		raw, _ := flags.GetString("data")
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, zerrors.Usagef("--data is empty")
		}
		if err := validateJSON(raw); err != nil {
			return nil, zerrors.Usagef("--data: %v", err)
		}
		return json.RawMessage(raw), nil
	}

	if flags.Changed("body") {
		path, _ := flags.GetString("body")
		path = strings.TrimSpace(path)
		switch {
		case path == "":
			return nil, zerrors.Usagef("--body is empty (use @path/to/file.json or - for stdin)")
		case path == "-" || strings.HasPrefix(path, "-"):
			return readStdin(cmd)
		case strings.HasPrefix(path, "@"):
			return readFile(strings.TrimPrefix(path, "@"))
		default:
			return readFile(path)
		}
	}

	if fromStdinFlag != "" && flags.Changed(fromStdinFlag) {
		return readStdin(cmd)
	}

	return nil, nil
}

func readFile(path string) (json.RawMessage, error) {
	if path == "" {
		return nil, zerrors.Usagef("empty file path")
	}
	data, err := os.ReadFile(path) //nolint:gosec // user-supplied JSON body is the point
	if err != nil {
		return nil, zerrors.Configf("read %s: %v", path, err)
	}
	if err := validateJSON(string(data)); err != nil {
		return nil, zerrors.Usagef("%s: %v", path, err)
	}
	return json.RawMessage(data), nil
}

func readStdin(cmd *cobra.Command) (json.RawMessage, error) {
	data, err := io.ReadAll(cmd.InOrStdin())
	if err != nil {
		return nil, zerrors.Networkf(err, "read stdin")
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil, zerrors.Usagef("stdin body is empty")
	}
	if err := validateJSON(string(data)); err != nil {
		return nil, zerrors.Usagef("stdin: %v", err)
	}
	return json.RawMessage(data), nil
}

func validateJSON(s string) error {
	var probe any
	if err := json.Unmarshal([]byte(s), &probe); err != nil {
		return fmt.Errorf("not valid JSON: %w", err)
	}
	return nil
}

// addBodyFlags wires the common --body/--data flags onto a cobra command
// for write operations. Pass the same command back to the caller so it
// can chain additional convenience flags.
func addBodyFlags(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().String("body", "", "path to a JSON request body (prefix with @ or pass - for stdin)")
	cmd.Flags().String("data", "", "inline JSON request body")
	return cmd
}

// mergeBody merges literal convenience-flag values into a JSON body.
// 'extra' is a map of fields the caller already populated from flags;
// fields already present in 'body' win when both sources set the same
// key (file/stdin is the source of truth, flags are augmentations).
//
// Returns a JSON object payload suitable for *WithBodyWithResponse.
func mergeBody(body json.RawMessage, extra map[string]any) (json.RawMessage, error) {
	merged := map[string]any{}
	if len(body) > 0 && string(body) != "null" {
		if err := json.Unmarshal(body, &merged); err != nil {
			return nil, zerrors.Usagef("body must be a JSON object to merge with flags: %v", err)
		}
	}
	for k, v := range extra {
		if _, ok := merged[k]; ok {
			continue
		}
		if v == nil {
			continue
		}
		merged[k] = v
	}
	if len(merged) == 0 {
		return nil, zerrors.Usagef("no body provided; pass --body @file.json or convenience flags like --name")
	}
	out, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("marshal merged body: %w", err)
	}
	return out, nil
}

// nonEmpty returns nil for empty strings so flag-only callers don't
// inject explicit "" into the merged JSON object.
func nonEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// bodyReader returns an io.Reader for a json.RawMessage; used with
// *WithBodyWithResponse client methods.
func bodyReader(body json.RawMessage) io.Reader {
	return strings.NewReader(string(body))
}

// jsonContentType is the canonical content-type for write requests.
const jsonContentType = "application/json"
