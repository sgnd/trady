package out

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestParseOutputFlag(t *testing.T) {
	cases := []struct {
		in   []string
		rest []string
		fmt  string
	}{
		{[]string{"list"}, []string{"list"}, "auto"},
		{[]string{"list", "-o", "json"}, []string{"list"}, "json"},
		{[]string{"-o=text", "list"}, []string{"list"}, "text"},
		{[]string{"list", "--output=json"}, []string{"list"}, "json"},
		{[]string{"list", "--output", "text", "x"}, []string{"list", "x"}, "text"},
	}
	for _, c := range cases {
		rest, f := ParseOutputFlag(c.in)
		if !reflect.DeepEqual(rest, c.rest) || f != c.fmt {
			t.Errorf("%v: got %v %q, want %v %q", c.in, rest, f, c.rest, c.fmt)
		}
	}
}

func TestPrintJSON(t *testing.T) {
	var b bytes.Buffer
	p := &Printer{Out: &b, Format: "json", Version: "v1"}
	p.Reset("demo")
	p.Print(map[string]int{"n": 1})
	var e Envelope
	if err := json.Unmarshal(b.Bytes(), &e); err != nil {
		t.Fatal(err)
	}
	if !e.OK || e.Command != "demo" || e.Meta.Contract != Contract || e.Meta.Version != "v1" {
		t.Fatalf("bad envelope: %+v", e)
	}
}

func TestFail(t *testing.T) {
	var b bytes.Buffer
	p := &Printer{Out: &b, Format: "json"}
	if code := p.Fail(errors.New("boom")); code != ExitError {
		t.Fatalf("plain error: exit %d", code)
	}
	if !strings.Contains(b.String(), `"code":"ERROR"`) {
		t.Fatalf("plain error: %s", b.String())
	}

	var errb bytes.Buffer
	p = &Printer{Out: &b, Err: &errb, Format: "text"}
	if code := p.Fail(Usage("bad %s", "arg")); code != ExitUsage {
		t.Fatalf("usage: exit %d", code)
	}
	if !strings.Contains(errb.String(), "USAGE: bad arg") || !strings.Contains(errb.String(), "hint:") {
		t.Fatalf("usage text: %q", errb.String())
	}
}
