package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/sgnd/trady/internal/inventory"
	"github.com/sgnd/trady/internal/out"
)

func cmdMoment(args []string) int {
	if len(args) == 0 || hasHelp(args) || args[0] != "list" {
		out.P.Print("usage: trady moment list PROJECT")
		if len(args) > 0 && args[0] != "list" {
			return out.P.Fail(out.Usage("unknown moment subcommand"))
		}
		return 0
	}
	if len(args) < 2 {
		return out.P.Fail(out.Usage("moment list needs a PROJECT"))
	}
	ctx := context.Background()
	p, err := inventory.GetProject(ctx, args[1])
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	list, err := inventory.ListMoments(ctx, p.ID)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	out.P.Print(momentListData{Count: len(list), Moments: list})
	return 0
}

type momentListData struct {
	Count   int                 `json:"count"`
	Moments []inventory.Moment  `json:"moments"`
}

func (d momentListData) Text() string {
	if d.Count == 0 {
		return "no moments: trady project run PROJECT"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%-8s %-12s %5s %7s %7s %s\n", "SCORE", "CATEGORY", "START", "END", "ID", "TITLE")
	for _, m := range d.Moments {
		id := m.ID
		if len(id) > 8 {
			id = id[:8]
		}
		fmt.Fprintf(&b, "%-8d %-12s %5.0f %7.0f %7s %s\n", m.Score, m.Category, m.StartSec, m.EndSec, id, m.Title)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func cmdClip(args []string) int {
	if len(args) == 0 || hasHelp(args) {
		out.P.Print("usage: trady clip list PROJECT\n       trady clip export PROJECT  (alias: see paths from last project run)")
		return 0
	}
	ctx := context.Background()
	switch args[0] {
	case "list", "export":
		if len(args) < 2 {
			return out.P.Fail(out.Usage("clip %s needs a PROJECT", args[0]))
		}
		p, err := inventory.GetProject(ctx, args[1])
		if err != nil {
			return out.P.Fail(mapErr(err))
		}
		list, err := inventory.ListClips(ctx, p.ID)
		if err != nil {
			return out.P.Fail(mapErr(err))
		}
		out.P.Print(clipListData{Count: len(list), Clips: list})
		return 0
	default:
		return out.P.Fail(out.Usage("unknown clip subcommand %q", args[0]))
	}
}

type clipListData struct {
	Count int              `json:"count"`
	Clips []inventory.Clip `json:"clips"`
}

func (d clipListData) Text() string {
	if d.Count == 0 {
		return "no clips: trady project run PROJECT"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%-10s %-10s %s\n", "STATUS", "ID", "PATH/NOTE")
	for _, c := range d.Clips {
		id := c.ID
		if len(id) > 8 {
			id = id[:8]
		}
		extra := c.Path
		if extra == "" {
			extra = c.Note
		}
		fmt.Fprintf(&b, "%-10s %-10s %s\n", c.Status, id, extra)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func cmdJob(args []string) int {
	if len(args) == 0 || hasHelp(args) {
		out.P.Print("usage: trady job list\n       trady job get ID")
		return 0
	}
	ctx := context.Background()
	switch args[0] {
	case "list":
		list, err := inventory.ListJobs(ctx)
		if err != nil {
			return out.P.Fail(mapErr(err))
		}
		out.P.Print(jobListData{Count: len(list), Jobs: list})
		return 0
	case "get":
		if len(args) < 2 {
			return out.P.Fail(out.Usage("job get needs ID"))
		}
		j, err := inventory.GetJob(ctx, args[1])
		if err != nil {
			return out.P.Fail(mapErr(err))
		}
		out.P.Print(j)
		return 0
	default:
		return out.P.Fail(out.Usage("unknown job subcommand %q", args[0]))
	}
}

type jobListData struct {
	Count int             `json:"count"`
	Jobs  []inventory.Job `json:"jobs"`
}

func (d jobListData) Text() string {
	if d.Count == 0 {
		return "no jobs yet"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%-16s %-12s %-10s %s\n", "ID", "KIND", "STATUS", "MESSAGE")
	for _, j := range d.Jobs {
		id := j.ID
		if len(id) > 12 {
			id = id[:12]
		}
		fmt.Fprintf(&b, "%-16s %-12s %-10s %s\n", id, j.Kind, j.Status, j.Message)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func cmdList(args []string) int {
	ctx := context.Background()
	sources, err := inventory.ListSources(ctx)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	projects, err := inventory.ListProjects(ctx)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	out.P.Print(map[string]any{
		"sources":  sources,
		"projects": projects,
	})
	return 0
}
