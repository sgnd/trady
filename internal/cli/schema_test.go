package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
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

// commandSet is every command in schemas: "noun" for a command without
// subcommands (and for the bare form of one with), "noun verb" otherwise.
func commandSet() []string {
	seen := map[string]bool{}
	for noun, lines := range schemas {
		for _, l := range lines {
			f := strings.Fields(l)
			if len(f) > 2 && !strings.HasPrefix(f[2], "[") && !strings.HasPrefix(f[2], "-") && strings.ToLower(f[2]) == f[2] {
				seen[noun+" "+f[2]] = true
			} else {
				seen[noun] = true // `trady guide`, `trady doctor`, `trady schema [COMMAND]`
			}
		}
	}
	var got []string
	for c := range seen {
		got = append(got, c)
	}
	sort.Strings(got)
	return got
}

var commandLine = regexp.MustCompile("(?m)^- `trady ([a-z-]+(?: [a-z-]+)?)`")

func listedCommands(t *testing.T, file string) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", file))
	if err != nil {
		t.Fatal(err)
	}
	body := string(b)
	i := strings.Index(body, "\n## Commands\n")
	if i < 0 {
		t.Fatalf("%s has no \"## Commands\" section", file)
	}
	rest := body[i+len("\n## Commands\n"):]
	if j := strings.Index(rest, "\n## "); j >= 0 {
		rest = rest[:j]
	}
	var got []string
	for _, m := range commandLine.FindAllStringSubmatch(rest, -1) {
		got = append(got, m[1])
	}
	if len(got) == 0 {
		t.Fatalf("%s: no command lines in its Commands section", file)
	}
	return got
}

func checkSameCommands(t *testing.T, file string, got []string) {
	t.Helper()
	want := commandSet()
	sort.Strings(got)
	got = slices.Compact(got)
	for _, w := range want {
		if !slices.Contains(got, w) {
			t.Errorf("%s does not list `trady %s`, but the binary has it", file, w)
		}
	}
	for _, g := range got {
		if !slices.Contains(want, g) {
			t.Errorf("%s lists `trady %s`, but the binary has no such command", file, g)
		}
	}
}

func TestSkillListsEveryCommand(t *testing.T) {
	checkSameCommands(t, "skills/trady/SKILL.md", listedCommands(t, "skills/trady/SKILL.md"))
}

func TestAgentsListsEveryCommand(t *testing.T) {
	checkSameCommands(t, "AGENTS.md", listedCommands(t, "AGENTS.md"))
}
