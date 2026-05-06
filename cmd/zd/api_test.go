package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newDummyCmdWithBodyFlags() *cobra.Command {
	cmd := &cobra.Command{Use: "x"}
	addBodyFlags(cmd)
	return cmd
}

func TestReadJSONBody_Data(t *testing.T) {
	cmd := newDummyCmdWithBodyFlags()
	if err := cmd.ParseFlags([]string{"--data", `{"name":"foo"}`}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	got, err := readJSONBody(cmd, "")
	if err != nil {
		t.Fatalf("readJSONBody: %v", err)
	}
	var probe map[string]any
	if err := json.Unmarshal(got, &probe); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if probe["name"] != "foo" {
		t.Fatalf("expected name=foo, got %v", probe)
	}
}

func TestReadJSONBody_DataInvalidJSON(t *testing.T) {
	cmd := newDummyCmdWithBodyFlags()
	if err := cmd.ParseFlags([]string{"--data", `{not-json`}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, err := readJSONBody(cmd, ""); err == nil {
		t.Fatalf("expected error for invalid JSON")
	}
}

func TestReadJSONBody_FileWithAtPrefix(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.json")
	if err := os.WriteFile(path, []byte(`{"name":"team-from-file"}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	cmd := newDummyCmdWithBodyFlags()
	if err := cmd.ParseFlags([]string{"--body", "@" + path}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	got, err := readJSONBody(cmd, "")
	if err != nil {
		t.Fatalf("readJSONBody: %v", err)
	}
	if !strings.Contains(string(got), "team-from-file") {
		t.Fatalf("expected file body, got %s", got)
	}
}

func TestReadJSONBody_FilePathPlain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "body.json")
	if err := os.WriteFile(path, []byte(`{"x":1}`), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	cmd := newDummyCmdWithBodyFlags()
	if err := cmd.ParseFlags([]string{"--body", path}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, err := readJSONBody(cmd, ""); err != nil {
		t.Fatalf("readJSONBody: %v", err)
	}
}

func TestReadJSONBody_StdinDash(t *testing.T) {
	cmd := newDummyCmdWithBodyFlags()
	cmd.SetIn(bytes.NewBufferString(`{"k":"v"}`))
	if err := cmd.ParseFlags([]string{"--body", "-"}); err != nil {
		t.Fatalf("parse: %v", err)
	}
	got, err := readJSONBody(cmd, "")
	if err != nil {
		t.Fatalf("readJSONBody: %v", err)
	}
	if string(got) != `{"k":"v"}` {
		t.Fatalf("got %s", got)
	}
}

func TestReadJSONBody_NoFlagsReturnsNil(t *testing.T) {
	cmd := newDummyCmdWithBodyFlags()
	got, err := readJSONBody(cmd, "")
	if err != nil {
		t.Fatalf("readJSONBody: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil body, got %s", got)
	}
}

func TestMergeBody_FlagsOnly(t *testing.T) {
	out, err := mergeBody(nil, map[string]any{
		"name":  "foo",
		"empty": nil,
		"role":  3,
	})
	if err != nil {
		t.Fatalf("mergeBody: %v", err)
	}
	var probe map[string]any
	if err := json.Unmarshal(out, &probe); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if probe["name"] != "foo" {
		t.Fatalf("expected name=foo, got %v", probe)
	}
	if _, has := probe["empty"]; has {
		t.Fatalf("nil flags must be omitted, got %v", probe)
	}
	if probe["role"].(float64) != 3 {
		t.Fatalf("expected role=3, got %v", probe["role"])
	}
}

func TestMergeBody_BodyWinsOverFlag(t *testing.T) {
	body := json.RawMessage(`{"name":"from-body"}`)
	out, err := mergeBody(body, map[string]any{"name": "from-flag"})
	if err != nil {
		t.Fatalf("mergeBody: %v", err)
	}
	if !strings.Contains(string(out), "from-body") {
		t.Fatalf("body must take precedence, got %s", out)
	}
}

func TestMergeBody_EmptyEverywhereFails(t *testing.T) {
	if _, err := mergeBody(nil, map[string]any{"name": nil}); err == nil {
		t.Fatalf("expected error when nothing was provided")
	}
}

func TestMergeBody_RejectsNonObjectBody(t *testing.T) {
	if _, err := mergeBody(json.RawMessage(`["not","an","object"]`), map[string]any{"name": "x"}); err == nil {
		t.Fatalf("expected error for non-object body")
	}
}

func TestNonEmpty(t *testing.T) {
	if nonEmpty("") != nil {
		t.Fatalf("empty string must return nil")
	}
	if nonEmpty("abc") != "abc" {
		t.Fatalf("non-empty string must return value")
	}
}

func TestBodyReader(t *testing.T) {
	r := bodyReader(json.RawMessage(`{"hello":"world"}`))
	buf := make([]byte, 64)
	n, _ := r.Read(buf)
	if !strings.Contains(string(buf[:n]), "hello") {
		t.Fatalf("body reader did not yield bytes: %q", string(buf[:n]))
	}
}
