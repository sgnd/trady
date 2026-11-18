package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sgnd/trady/internal/inventory"
	"github.com/sgnd/trady/internal/media"
	"github.com/sgnd/trady/internal/out"
)

func cmdSource(args []string) int {
	if len(args) == 0 || hasHelp(args) {
		out.P.Print("usage: trady source add PATH [--name NAME]\n       trady source list\n       trady source rm NAME|ID")
		return 0
	}
	ctx := context.Background()
	switch args[0] {
	case "add":
		return sourceAdd(ctx, args[1:])
	case "list":
		if _, _, ok := parse("source list", args[1:], nil); !ok {
			return out.ExitUsage
		}
		return sourceList(ctx)
	case "rm", "remove":
		pos, _, ok := parse("source rm", args[1:], nil, "NAME|ID")
		if !ok {
			return out.ExitUsage
		}
		if err := inventory.RemoveSource(ctx, pos[0]); err != nil {
			return out.P.Fail(mapErr(err))
		}
		out.P.Print(map[string]any{"removed": pos[0]})
		return 0
	default:
		return out.P.Fail(out.Usage("unknown source subcommand %q", args[0]))
	}
}

func sourceAdd(ctx context.Context, args []string) int {
	pos, vals, ok := parse("source add", args, map[string]bool{"--name": true}, "PATH")
	if !ok {
		return out.ExitUsage
	}
	path, name := pos[0], vals["--name"]
	abs, err := filepath.Abs(path)
	if err != nil {
		return out.P.Fail(err)
	}
	if _, err := os.Stat(abs); err != nil {
		return out.P.Fail(out.Errf(out.ExitNotFound, "FILE_NOT_FOUND", "give a readable media file path", "%v", err))
	}
	pctx, cancel := media.WithTimeout(ctx)
	defer cancel()
	probe, err := media.Probe(pctx, abs)
	if err != nil {
		return out.P.Fail(err)
	}
	s, err := inventory.AddSource(ctx, name, abs, probe)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	out.P.Print(s)
	return 0
}

type sourceListData struct {
	Count   int                `json:"count"`
	Sources []inventory.Source `json:"sources"`
}

func (d sourceListData) Text() string {
	if d.Count == 0 {
		return "no sources yet: trady source add PATH"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%-16s %-20s %8s %s\n", "ID", "NAME", "SECONDS", "PATH")
	for _, s := range d.Sources {
		id := s.ID
		if len(id) > 12 {
			id = id[:12]
		}
		fmt.Fprintf(&b, "%-16s %-20s %8.1f %s\n", id, s.Name, s.DurationSec, s.Path)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func sourceList(ctx context.Context) int {
	list, err := inventory.ListSources(ctx)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	out.P.Print(sourceListData{Count: len(list), Sources: list})
	return 0
}
