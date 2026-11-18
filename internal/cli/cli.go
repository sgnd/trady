// Package cli wires trady commands (person + AI agent).
package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sgnd/trady/internal/out"
)

// Version is set at build time via -ldflags -X.
var Version = "dev"

// Execute runs trady with os.Args[1:] and returns the process exit code.
func Execute() int {
	args := os.Args[1:]
	args, format := out.ParseOutputFlag(args)
	out.P.Version = Version
	out.P.Format = format

	if len(args) == 0 {
		printRootHelp(os.Stdout)
		return 0
	}
	switch args[0] {
	case "version", "-v", "--version":
		out.P.Reset("version")
		return cmdVersion(args[1:])
	case "doctor":
		out.P.Reset("doctor")
		return cmdDoctor(args[1:])
	case "secret":
		out.P.Reset("secret")
		return cmdSecret(args[1:])
	case "guide":
		out.P.Reset("guide")
		return cmdGuide(args[1:])
	case "schema":
		out.P.Reset("schema")
		return cmdSchema(args[1:])
	case "source":
		out.P.Reset("source")
		return cmdSource(args[1:])
	case "project":
		out.P.Reset("project")
		return cmdProject(args[1:])
	case "moment":
		out.P.Reset("moment")
		return cmdMoment(args[1:])
	case "clip":
		out.P.Reset("clip")
		return cmdClip(args[1:])
	case "job":
		out.P.Reset("job")
		return cmdJob(args[1:])
	case "list":
		out.P.Reset("list")
		return cmdList(args[1:])
	case "help", "-h", "--help":
		printRootHelp(os.Stdout)
		return 0
	default:
		return out.P.Fail(out.Usage("unknown command %q", args[0]).WithHint("run `trady` for the command list"))
	}
}

func printRootHelp(w io.Writer) {
	fmt.Fprint(w, strings.TrimSpace(`
trady: local-first CLI for short captioned clips from long recordings

State lives under ~/.trady (TRADY_HOME). Sealed store + vault (e2ee).
Runtime is this CLI — not trady.efolusi.com.

START        version · doctor · guide · schema
SECRETS      secret set|list|show|rm
MEDIA        source add|list|rm
PROJECTS     project create|list|get|status|run|rm
CLIPS        moment list · clip list|export
JOBS         job list|get
INVENTORY    list

Sister architecture to komando. Agent-safe: set TRADY_AGENT=1; use -o json.
`)+"\n")
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	s := err.Error()
	switch {
	case strings.Contains(s, "not found") || strings.Contains(s, "ErrNotFound") || s == "sealdb: record not found":
		return out.Errf(out.ExitNotFound, "NOT_FOUND", "check the name; `trady list` shows what trady knows", "%v", err)
	case strings.Contains(s, "already exists"):
		return out.Errf(out.ExitConflict, "CONFLICT", "pick another --name, or remove the old one first", "%v", err)
	case strings.Contains(s, "key problem"):
		return out.Errf(out.ExitPolicy, "KEY", "fix the vault key; `trady doctor` checks this computer", "%v", err)
	case strings.Contains(s, "locked"):
		return out.Errf(out.ExitPolicy, "LOCKED", "set TRADY_VAULT_PASS or unlock the keychain", "%v", err)
	default:
		return err
	}
}
