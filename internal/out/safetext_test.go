package out

import "testing"

func TestSafeText(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain text", "plain text"},
		{"tab\tand\nnewline", "tab\tand\nnewline"},
		{"clear\x1b[2Jscreen", `clear\e[2Jscreen`},
		{"color\x1b[31mred", `color\e[31mred`},
		{"bell\x07", `bell\x07`},
		{"c1\u009b2J", `c1\x9b2J`},
		{"unicode ünï", "unicode ünï"},
	}
	for _, c := range cases {
		if got := SafeText(c.in); got != c.want {
			t.Errorf("SafeText(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
