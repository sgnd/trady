package cli

import (
	"os"
	"strings"
	"testing"
)

// TestReadmeMatchesSchema checks that the README command table names every
// command and subcommand in schemas, and nothing that is gone.
func TestReadmeMatchesSchema(t *testing.T) {
	b, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	readme := string(b)
	start := strings.Index(readme, "## Commands")
	end := strings.Index(readme[start+1:], "\n## ")
	if start < 0 || end < 0 {
		t.Fatal("README has no Commands section")
	}
	table := readme[start : start+1+end]

	for cmd, lines := range schemas {
		subs := map[string]bool{}
		for _, l := range lines {
			f := strings.Fields(l)
			if len(f) > 2 && !strings.HasPrefix(f[2], "[") && strings.ToLower(f[2]) == f[2] {
				subs[f[2]] = true
			}
		}
		if len(subs) == 0 {
			if !strings.Contains(table, "`"+cmd) {
				t.Errorf("README command table is missing %q", cmd)
			}
			continue
		}
		for sub := range subs {
			if !strings.Contains(table, "`"+cmd+" ") && !strings.Contains(table, "`"+cmd+"`") {
				t.Errorf("README command table is missing %q", cmd)
			}
			if !strings.Contains(table, sub) {
				t.Errorf("README command table is missing %s %s", cmd, sub)
			}
		}
	}
	if strings.Contains(table, "export") {
		t.Error("README names clip export, which is gone")
	}
}

func TestUsageComesFromSchema(t *testing.T) {
	for cmd, lines := range schemas {
		u := usage(cmd)
		for _, l := range lines {
			if !strings.Contains(u, l) {
				t.Errorf("usage(%q) is missing %q", cmd, l)
			}
		}
	}
}
