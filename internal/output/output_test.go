package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestDetect_Explicit(t *testing.T) {
	cases := map[string]Format{
		"json":     FormatJSON,
		"YAML":     FormatYAML,
		"table":    FormatTable,
		"template": FormatTemplate,
	}
	for in, want := range cases {
		got, err := Detect(in, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("Detect(%q) returned err=%v", in, err)
		}
		if got != want {
			t.Fatalf("Detect(%q): got %q, want %q", in, got, want)
		}
	}
}

func TestDetect_Invalid(t *testing.T) {
	if _, err := Detect("xml", &bytes.Buffer{}); err == nil {
		t.Fatalf("Detect(xml): expected error, got nil")
	}
}

func TestDetect_DefaultsToJSONForBuffer(t *testing.T) {
	got, err := Detect("", &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}
	if got != FormatJSON {
		t.Fatalf("Detect default for buffer: got %q, want json", got)
	}
}

func TestRender_JSON(t *testing.T) {
	var buf bytes.Buffer
	type item struct {
		Name string `json:"name"`
	}
	if err := Render(&buf, []item{{Name: "alpha"}, {Name: "beta"}}, FormatJSON, Options{}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	var got []item
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v\noutput=%s", err, buf.String())
	}
	if len(got) != 2 || got[0].Name != "alpha" || got[1].Name != "beta" {
		t.Fatalf("unexpected JSON output: %+v", got)
	}
}

func TestRender_YAML(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, map[string]int{"k": 1}, FormatYAML, Options{}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(buf.String(), "k: 1") {
		t.Fatalf("yaml output missing key: %q", buf.String())
	}
}

func TestRender_Table(t *testing.T) {
	type item struct {
		Name  string
		Count int
	}
	spec := &TableSpec{
		Headers: []string{"name", "count"},
		Row: func(v any) []string {
			it := v.(item)
			return []string{it.Name, "x" + it.Name}
		},
	}
	var buf bytes.Buffer
	err := Render(&buf, []item{{Name: "alpha"}, {Name: "beta"}}, FormatTable, Options{Table: spec})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "COUNT") {
		t.Fatalf("expected uppercased headers, got %q", out)
	}
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "xbeta") {
		t.Fatalf("expected row cells, got %q", out)
	}
}

func TestRender_TableMissingSpecFallsBackToJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, []int{1, 2, 3}, FormatTable, Options{}); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(buf.String(), "[\n") {
		t.Fatalf("expected JSON fallback, got %q", buf.String())
	}
}

func TestRender_TableWrongCellCount(t *testing.T) {
	type item struct{ Name string }
	spec := &TableSpec{
		Headers: []string{"name", "count"},
		Row:     func(v any) []string { return []string{v.(item).Name} },
	}
	var buf bytes.Buffer
	err := Render(&buf, []item{{Name: "alpha"}}, FormatTable, Options{Table: spec})
	if err == nil {
		t.Fatalf("expected error for cell-count mismatch")
	}
}

func TestRender_Template(t *testing.T) {
	var buf bytes.Buffer
	err := Render(&buf, map[string]string{"name": "alpha"}, FormatTemplate, Options{Template: "{{.name}}"})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "alpha" {
		t.Fatalf("template output: got %q, want alpha", buf.String())
	}
}

func TestRender_TemplateRequiresBody(t *testing.T) {
	var buf bytes.Buffer
	if err := Render(&buf, nil, FormatTemplate, Options{}); err == nil {
		t.Fatalf("expected error when template body is empty")
	}
}

func TestRender_UnsupportedFormat(t *testing.T) {
	if err := Render(&bytes.Buffer{}, nil, Format("xml"), Options{}); err == nil {
		t.Fatalf("expected error for unsupported format")
	}
}
