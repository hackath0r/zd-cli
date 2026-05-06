package main

import (
	"reflect"
	"testing"
	"time"
)

func TestStatusName(t *testing.T) {
	cases := []struct {
		in   *int
		want string
	}{
		{nil, ""},
		{intPtr(1), "triggered"},
		{intPtr(2), "acknowledged"},
		{intPtr(3), "resolved"},
		{intPtr(99), "99"},
	}
	for _, c := range cases {
		if got := statusName(c.in); got != c.want {
			t.Fatalf("statusName(%v): got %q, want %q", c.in, got, c.want)
		}
	}
}

func TestUrgencyName(t *testing.T) {
	cases := []struct {
		in   *int
		want string
	}{
		{nil, ""},
		{intPtr(0), "low"},
		{intPtr(1), "high"},
		{intPtr(7), "7"},
	}
	for _, c := range cases {
		if got := urgencyName(c.in); got != c.want {
			t.Fatalf("urgencyName(%v): got %q, want %q", c.in, got, c.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		s    string
		n    int
		want string
	}{
		{"alpha", 10, "alpha"},
		{"alpha", 5, "alpha"},
		{"alphabet", 5, "alph…"},
		{"x", 0, ""},
	}
	for _, c := range cases {
		if got := truncate(c.s, c.n); got != c.want {
			t.Fatalf("truncate(%q, %d): got %q, want %q", c.s, c.n, got, c.want)
		}
	}
}

func TestLooksLikeUUID(t *testing.T) {
	cases := map[string]bool{
		"123456":                               false, // numeric
		"abc":                                  false, // too short
		"d290f1ee6c544b0190e6d701748f0851":     true,  // long alphanumeric
		"d290f1ee-6c54-4b01-90e6-d701748f0851": true,  // canonical UUID
		"":                                     false,
	}
	for in, want := range cases {
		if got := looksLikeUUID(in); got != want {
			t.Fatalf("looksLikeUUID(%q): got %v, want %v", in, got, want)
		}
	}
}

func TestDerefHelpers(t *testing.T) {
	if derefStr(nil) != "" {
		t.Fatalf("derefStr(nil) should be empty")
	}
	s := "hello"
	if derefStr(&s) != "hello" {
		t.Fatalf("derefStr(*hello): wrong value")
	}
	if derefInt(nil) != 0 {
		t.Fatalf("derefInt(nil) should be 0")
	}
	if derefBool(nil) {
		t.Fatalf("derefBool(nil) should be false")
	}
	now := time.Now()
	if derefTimeRFC3339(nil) != "" {
		t.Fatalf("derefTimeRFC3339(nil) should be empty")
	}
	if got := derefTimeRFC3339(&now); got == "" {
		t.Fatalf("derefTimeRFC3339(now) returned empty")
	}
}

func TestStrPtrOrNil(t *testing.T) {
	if strPtrOrNil("") != nil {
		t.Fatalf("strPtrOrNil(empty) should be nil")
	}
	v := strPtrOrNil("x")
	if v == nil || *v != "x" {
		t.Fatalf("strPtrOrNil(x) wrong")
	}
}

func TestAsString(t *testing.T) {
	if got := asString(nil, "fallback"); got != "fallback" {
		t.Fatalf("asString(nil) wrong: %q", got)
	}
	empty := ""
	if got := asString(&empty, "fallback"); got != "fallback" {
		t.Fatalf("asString(&empty) wrong: %q", got)
	}
	v := "value"
	if got := asString(&v, "fallback"); got != "value" {
		t.Fatalf("asString(&v) wrong: %q", got)
	}
}

func TestDecodeList_AsArray(t *testing.T) {
	body := []byte(`[{"id":"a"},{"id":"b"}]`)
	out, err := decodeList[map[string]string](body)
	if err != nil {
		t.Fatalf("decodeList: %v", err)
	}
	if len(out) != 2 || out[0]["id"] != "a" || out[1]["id"] != "b" {
		t.Fatalf("unexpected list: %+v", out)
	}
}

func TestDecodeList_AsSingleObject(t *testing.T) {
	body := []byte(`{"id":"a"}`)
	out, err := decodeList[map[string]string](body)
	if err != nil {
		t.Fatalf("decodeList: %v", err)
	}
	if len(out) != 1 || out[0]["id"] != "a" {
		t.Fatalf("expected single-element list, got %+v", out)
	}
}

func TestDecodeList_NullEmpty(t *testing.T) {
	out, err := decodeList[map[string]any]([]byte(`null`))
	if err != nil {
		t.Fatalf("decodeList(null): %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("expected nil/empty list, got %+v", out)
	}
}

func TestDecodeOne_AsObject(t *testing.T) {
	out, err := decodeOne[map[string]string]([]byte(`{"id":"a"}`))
	if err != nil {
		t.Fatalf("decodeOne: %v", err)
	}
	if out == nil || (*out)["id"] != "a" {
		t.Fatalf("expected object, got %+v", out)
	}
}

func TestDecodeOne_FromArray(t *testing.T) {
	out, err := decodeOne[map[string]string]([]byte(`[{"id":"a"}]`))
	if err != nil {
		t.Fatalf("decodeOne: %v", err)
	}
	if out == nil || (*out)["id"] != "a" {
		t.Fatalf("expected first element, got %+v", out)
	}
}

func TestDecodeOne_NullReturnsNil(t *testing.T) {
	out, err := decodeOne[map[string]string]([]byte(`null`))
	if err != nil {
		t.Fatalf("decodeOne: %v", err)
	}
	if out != nil {
		t.Fatalf("expected nil for null body, got %+v", out)
	}
}

func TestPtrInt(t *testing.T) {
	p := ptrInt(42)
	if p == nil || *p != 42 {
		t.Fatalf("ptrInt failed")
	}
	if !reflect.DeepEqual(*p, 42) {
		t.Fatalf("ptrInt deep-equal mismatch")
	}
}

func intPtr(i int) *int { return &i }
