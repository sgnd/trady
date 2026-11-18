package cli

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	flags := map[string]bool{"--name": true, "--reveal": false}
	cases := []struct {
		in   []string
		pos  []string
		vals map[string]string
		bad  bool
	}{
		{[]string{"a"}, []string{"a"}, map[string]string{}, false},
		{[]string{"a", "--name", "x"}, []string{"a"}, map[string]string{"--name": "x"}, false},
		{[]string{"--name=x", "a"}, []string{"a"}, map[string]string{"--name": "x"}, false},
		{[]string{"--reveal", "a"}, []string{"a"}, map[string]string{"--reveal": "true"}, false},
		{[]string{"--", "-odd"}, []string{"-odd"}, map[string]string{}, false},
		{[]string{"--nope"}, nil, nil, true},
		{[]string{"a", "--name"}, nil, nil, true},
		{[]string{"--name="}, nil, nil, true},
		{[]string{"--reveal=yes"}, nil, nil, true},
	}
	for _, c := range cases {
		pos, vals, err := parseArgs("t", c.in, flags)
		if c.bad {
			if err == nil {
				t.Errorf("%v: want error", c.in)
			}
			continue
		}
		if err != nil || !reflect.DeepEqual(pos, c.pos) || !reflect.DeepEqual(vals, c.vals) {
			t.Errorf("%v: got %v %v %v", c.in, pos, vals, err)
		}
	}
}

func TestWant(t *testing.T) {
	if err := want("t", []string{"a"}, "NAME"); err != nil {
		t.Fatal(err)
	}
	if err := want("t", nil, "NAME"); err == nil {
		t.Fatal("missing: want error")
	}
	if err := want("t", []string{"a", "b"}, "NAME"); err == nil {
		t.Fatal("extra: want error")
	}
}
