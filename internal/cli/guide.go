package cli

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sgnd/trady/internal/out"
)

//go:embed guide.md
var guideMD string

func cmdGuide(args []string) int {
	if len(args) > 0 && args[0] == "install" {
		if _, _, ok := parse("guide install", args[1:], nil); !ok {
			return out.ExitUsage
		}
		return guideInstall()
	}
	if _, _, ok := parse("guide", args, nil); !ok {
		return out.ExitUsage
	}
	out.P.Print(strings.TrimSpace(guideMD) + "\n")
	return 0
}

func guideInstall() int {
	if out.IsAgent() {
		return out.P.Fail(out.NeedsPerson("guide install"))
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return out.P.Fail(err)
	}
	dst := filepath.Join(home, ".cursor", "skills", "trady", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return out.P.Fail(err)
	}
	if err := os.WriteFile(dst, skillBody(), 0o644); err != nil {
		return out.P.Fail(err)
	}
	out.P.Print(map[string]any{"installed": dst})
	fmt.Fprintf(os.Stderr, "Installed the Cursor skill to %s. Restart the agent to load it.\n", dst)
	return 0
}

// skillBody is the Cursor skill that guide install writes. It comes only from
// the binary: a file in the current folder could plant agent instructions.
func skillBody() []byte {
	return []byte("---\nname: trady\ndescription: One local command line that turns long recordings into short captioned clips.\n---\n\n" + guideMD)
}
