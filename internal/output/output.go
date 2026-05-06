// Package output renders any Go value to one of four output formats:
// json, yaml, table, or template. The choice is normally driven by the
// --output global flag, with a sensible default selected from the TTY
// state of the destination writer (table when interactive, json when piped).
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// Format identifies an output format.
type Format string

// Supported formats.
const (
	FormatJSON     Format = "json"
	FormatYAML     Format = "yaml"
	FormatTable    Format = "table"
	FormatTemplate Format = "template"
)

// String makes Format flag-friendly.
func (f Format) String() string { return string(f) }

// Detect chooses an output format. If explicit is non-empty it wins
// (validated). Otherwise we pick table when w is a TTY, json when piped.
func Detect(explicit string, w io.Writer) (Format, error) {
	if explicit != "" {
		switch Format(strings.ToLower(explicit)) {
		case FormatJSON, FormatYAML, FormatTable, FormatTemplate:
			return Format(strings.ToLower(explicit)), nil
		default:
			return "", fmt.Errorf("unknown --output value %q (want one of json, yaml, table, template)", explicit)
		}
	}
	if isTerminal(w) {
		return FormatTable, nil
	}
	return FormatJSON, nil
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

// TableSpec describes how to render a slice of items as a table. Headers
// is the column titles in display order; Row maps a single item to a
// slice of cell strings (length must equal len(Headers)).
type TableSpec struct {
	Headers []string
	Row     func(item any) []string
}

// Options carries optional rendering hints. TableSpec is required when the
// resolved Format is FormatTable; Template is required for FormatTemplate.
type Options struct {
	Table    *TableSpec
	Template string // text/template body used by FormatTemplate
}

// Render dispatches to the format-specific encoder. Returning an error
// surfaces directly to the caller so commands can decide whether to
// degrade gracefully.
func Render(w io.Writer, v any, format Format, opts Options) error {
	switch format {
	case FormatJSON:
		return renderJSON(w, v)
	case FormatYAML:
		return renderYAML(w, v)
	case FormatTable:
		if opts.Table == nil {
			return renderJSON(w, v)
		}
		return renderTable(w, v, opts.Table)
	case FormatTemplate:
		if opts.Template == "" {
			return fmt.Errorf("--template is required when --output=template")
		}
		return renderTemplate(w, v, opts.Template)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func renderJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func renderYAML(w io.Writer, v any) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return err
	}
	return enc.Close()
}

func renderTable(w io.Writer, v any, spec *TableSpec) error {
	tw := tabwriter.NewWriter(w, 0, 2, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, strings.Join(upper(spec.Headers), "\t")); err != nil {
		return err
	}

	items, ok := flatten(v)
	if !ok {
		return fmt.Errorf("table renderer: value is not a slice; pass --output json instead")
	}
	for _, item := range items {
		cells := spec.Row(item)
		if len(cells) != len(spec.Headers) {
			return fmt.Errorf("table renderer: row produced %d cells, expected %d", len(cells), len(spec.Headers))
		}
		if _, err := fmt.Fprintln(tw, strings.Join(cells, "\t")); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func renderTemplate(w io.Writer, v any, body string) error {
	tmpl, err := template.New("zd").Option("missingkey=zero").Parse(body)
	if err != nil {
		return fmt.Errorf("parse --template: %w", err)
	}
	if err := tmpl.Execute(w, v); err != nil {
		return err
	}
	if !strings.HasSuffix(body, "\n") {
		_, _ = fmt.Fprintln(w)
	}
	return nil
}

// flatten converts a slice value (typed or []any) to []any. Non-slice
// values yield a single-element slice so single-object renders also work
// in table mode.
func flatten(v any) ([]any, bool) {
	switch s := v.(type) {
	case nil:
		return nil, true
	case []any:
		return s, true
	}
	// Use reflection for typed slices.
	rv := derefSlice(v)
	if rv == nil {
		return []any{v}, true
	}
	out := make([]any, 0, rv.Len)
	for i := 0; i < rv.Len; i++ {
		out = append(out, rv.At(i))
	}
	return out, true
}

func upper(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToUpper(s)
	}
	return out
}
