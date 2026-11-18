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
		return guideInstall()
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
	skillPath := filepath.Join("skills", "trady", "SKILL.md")
	// Prefer repo skill when developing; else embed guide text as skill body.
	body, err := os.ReadFile(skillPath)
	if err != nil {
		body = []byte("---\nname: trady\ndescription: Local-first CLI for short captioned clips from long recordings.\n---\n\n" + guideMD)
	}
	if err := os.WriteFile(dst, body, 0o644); err != nil {
		return out.P.Fail(err)
	}
	out.P.Print(map[string]any{"installed": dst})
	fmt.Fprintf(os.Stderr, "Installed Cursor skill to %s — restart the agent.\n", dst)
	return 0
}
