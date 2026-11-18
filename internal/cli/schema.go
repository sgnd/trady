package cli

import (
	"strings"

	"github.com/sgnd/trady/internal/out"
)

var schemas = map[string]string{
	"version": "trady version",
	"doctor":  "trady doctor",
	"guide":   "trady guide [install]",
	"schema":  "trady schema [COMMAND]",
	"secret":  "trady secret set|list|show|rm …",
	"source":  "trady source add|list|transcribe|transcript|rm …",
	"project": "trady project create|list|get|status|run|rm …",
	"moment":  "trady moment list PROJECT",
	"clip":    "trady clip list|export PROJECT",
	"job":     "trady job list|get …",
	"list":    "trady list",
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
	out.P.Print(map[string]string{"command": args[0], "usage": s})
	return 0
}
