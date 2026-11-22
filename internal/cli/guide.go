package cli

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sgnd/trady/internal/out"
	"github.com/sgnd/trady/skills"
)

//go:embed guide.md
var guideMD string

// isAgent says whether prompts and installs must wait for a person. Tests replace it.
var isAgent = out.IsAgent

func cmdGuide(args []string) int {
	if len(args) > 0 && args[0] == "install" {
		pos, vals, ok := parse("guide install", args[1:], map[string]bool{"--project": false, "--dry-run": false})
		_ = pos
		if !ok {
			return out.ExitUsage
		}
		return guideInstall(vals["--project"] == "true", vals["--dry-run"] == "true")
	}
	if _, _, ok := parse("guide", args, nil); !ok {
		return out.ExitUsage
	}
	out.P.Print(strings.TrimSpace(guideMD) + "\n")
	return 0
}

// skillDirs are the folders AI agents read skills from: Claude Code reads
// .claude/skills; Codex, Cursor, Gemini CLI and others read .agents/skills.
var skillDirs = []string{".claude/skills", ".agents/skills"}

type skillFile struct {
	Path   string `json:"path"`
	Action string `json:"action"` // created | updated | unchanged
	Backup string `json:"backup,omitempty"`
}

type skillResult struct {
	Files  []skillFile `json:"files"`
	dryRun bool
}

func (r skillResult) Text() string {
	var b strings.Builder
	changed := false
	for _, f := range r.Files {
		action := f.Action
		if r.dryRun && action != "unchanged" {
			action = "would " + map[string]string{"created": "create", "updated": "update"}[action]
		}
		b.WriteString(action + "  " + out.SafeText(f.Path) + "\n")
		if f.Backup != "" {
			b.WriteString("  old version kept as " + out.SafeText(f.Backup) + "\n")
		}
		changed = changed || f.Action != "unchanged"
	}
	if changed && !r.dryRun {
		b.WriteString("Restart your AI agent so it sees the skill.")
	}
	return strings.TrimRight(b.String(), "\n")
}

// guideInstall teaches Claude Code, Codex, Cursor and Gemini CLI to use trady
// by writing the skill that is built into this binary. An agent may look
// (--dry-run) and may install into the project it works in (--project); the
// home-wide install changes every project on the machine, so a person runs it.
func guideInstall(project, dry bool) int {
	if isAgent() && !dry && !project {
		return out.P.Fail(out.NeedsPerson("guide install"))
	}
	base, err := os.UserHomeDir()
	if project {
		base, err = os.Getwd()
	}
	if err != nil {
		return out.P.Fail(out.Errf(out.ExitError, "NO_FOLDER", "set HOME, or run it from the project folder", "cannot find the folder to install into: %v", err))
	}
	var paths []string
	for _, d := range skillDirs {
		paths = append(paths, filepath.Join(base, d, "trady", "SKILL.md"))
	}
	if project {
		if err := noLinksUnder(base, paths); err != nil {
			return out.P.Fail(out.Errf(out.ExitError, "WRITE_FAILED", "remove the link, or install for your home folder without --project", "%v", err))
		}
	}
	res := skillResult{Files: []skillFile{}, dryRun: dry}
	for _, p := range paths {
		f, err := installSkill(p, dry)
		if err != nil {
			return out.P.Fail(out.Errf(out.ExitError, "WRITE_FAILED", "check that you may write to this folder", "%v", err))
		}
		res.Files = append(res.Files, f)
	}
	out.P.Print(res)
	return 0
}

// installSkill writes the skill to path. A different old copy is kept as a
// backup first. Every write goes through a temporary file of a random name
// in the same folder, so no name trady writes to is one a link could hold.
func installSkill(path string, dryRun bool) (skillFile, error) {
	f := skillFile{Path: path, Action: "created"}
	old, err := os.ReadFile(path)
	switch {
	case err == nil && string(old) == skills.Trady:
		f.Action = "unchanged"
		return f, nil
	case err == nil:
		f.Action, f.Backup = "updated", backupName(path, time.Now())
	case !errors.Is(err, fs.ErrNotExist):
		return f, err
	}
	if dryRun {
		return f, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return f, err
	}
	if f.Backup != "" {
		if err := writeAtomic(f.Backup, old, 0o600, false); err != nil {
			return f, err
		}
	}
	return f, writeAtomic(path, []byte(skills.Trady), 0o644, true)
}

// backupName is PATH.bak, or PATH.bak.<UTC time> when that is taken. A file
// already there (the person's own .bak, or a link) is never written over.
func backupName(path string, now time.Time) string {
	name := path + ".bak"
	if _, err := os.Lstat(name); err == nil {
		name = path + ".bak." + now.UTC().Format("20060102T150405Z")
	}
	return name
}

// writeAtomic writes b next to path and renames it into place. With
// replace false it refuses a name that exists by then.
func writeAtomic(path string, b []byte, mode os.FileMode, replace bool) error {
	if !replace {
		if _, err := os.Lstat(path); err == nil {
			return fmt.Errorf("%s appeared while trady wrote it; nothing was changed, run it again", path)
		}
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// noLinksUnder refuses a link on the way from base to any of paths. A
// project folder is often a repository someone else wrote, and a link
// committed there could point the write at a file elsewhere.
func noLinksUnder(base string, paths []string) error {
	for _, p := range paths {
		rel, err := filepath.Rel(base, p)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("%s is not in %s", p, base)
		}
		cur := base
		for _, part := range strings.Split(rel, string(filepath.Separator)) {
			cur = filepath.Join(cur, part)
			fi, err := os.Lstat(cur)
			if errors.Is(err, fs.ErrNotExist) {
				break
			}
			if err != nil {
				return err
			}
			if fi.Mode()&(fs.ModeSymlink|fs.ModeIrregular) != 0 { // Irregular: a Windows junction
				return fmt.Errorf("%s is a link, and trady does not follow links in a project folder; nothing was changed", cur)
			}
		}
	}
	return nil
}
