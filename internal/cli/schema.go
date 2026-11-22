package cli

import (
	"strings"

	"github.com/sgnd/trady/internal/out"
)

// schemas is the one list of usage lines. Help text and `trady schema`
// both print from it, and a test checks the README against it.
var schemas = map[string][]string{
	"version": {"trady version"},
	"doctor":  {"trady doctor"},
	"guide":   {"trady guide", "trady guide install [--project] [--dry-run]"},
	"schema":  {"trady schema [COMMAND]"},
	"list":    {"trady list"},
	"secret": {
		"trady secret set NAME [--from-env VAR]",
		"trady secret list",
		"trady secret show NAME [--reveal]",
		"trady secret rm NAME",
	},
	"source": {
		"trady source add PATH [--name NAME]",
		"trady source list",
		"trady source transcribe NAME|ID",
		"trady source transcript NAME|ID",
		"trady source rm NAME|ID",
	},
	"project": {
		"trady project create SOURCE [--name NAME]",
		"trady project list",
		"trady project get NAME|ID",
		"trady project status NAME|ID",
		"trady project run NAME|ID [--out DIR]",
		"trady project rm NAME|ID",
	},
	"moment": {"trady moment list PROJECT", "trady moment set PROJECT --start SEC --end SEC [--title TITLE]", "trady moment set PROJECT --clear"},
	"clip":   {"trady clip list PROJECT"},
	"job":    {"trady job list", "trady job get ID"},
}

// usage is the help text of a command, from schemas.
func usage(cmd string) string {
	return "usage: " + strings.Join(schemas[cmd], "\n       ")
}

func cmdSchema(args []string) int {
	if len(args) == 0 {
		out.P.Print(map[string]any{"commands": schemas})
		return 0
	}
	if len(args) > 1 {
		return out.P.Fail(out.Usage("schema: unexpected argument %q", args[1]))
	}
	s, ok := schemas[args[0]]
	if !ok {
		var keys []string
		for k := range schemas {
			keys = append(keys, k)
		}
		return out.P.Fail(out.Usage("unknown command %q (have: %s)", args[0], strings.Join(keys, ", ")))
	}
	out.P.Print(map[string]any{"command": args[0], "usage": s})
	return 0
}
