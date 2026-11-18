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
		return sourceList(ctx)
	case "rm", "remove":
		if len(args) < 2 {
			return out.P.Fail(out.Usage("source rm needs NAME or ID"))
		}
		if err := inventory.RemoveSource(ctx, args[1]); err != nil {
			return out.P.Fail(mapErr(err))
		}
		out.P.Print(map[string]any{"removed": args[1]})
		return 0
	default:
		return out.P.Fail(out.Usage("unknown source subcommand %q", args[0]))
	}
}

func sourceAdd(ctx context.Context, args []string) int {
	var path, name string
	for i := 0; i < len(args); i++ {
		if args[i] == "--name" && i+1 < len(args) {
			name = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(args[i], "-") {
			return out.P.Fail(out.Usage("unknown flag %s", args[i]))
		}
		path = args[i]
	}
	if path == "" {
		return out.P.Fail(out.Usage("source add needs a PATH"))
	}
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
