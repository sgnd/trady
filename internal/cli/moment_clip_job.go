package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/sgnd/trady/internal/inventory"
	"github.com/sgnd/trady/internal/media"
	"github.com/sgnd/trady/internal/out"
)

func cmdMoment(args []string) int {
	if len(args) == 0 || hasHelp(args) {
		out.P.Print(usage("moment"))
		return 0
	}
	switch args[0] {
	case "list":
		return momentList(args[1:])
	case "set":
		return momentSet(args[1:])
	default:
		return out.P.Fail(out.Usage("unknown moment subcommand %q", args[0]))
	}
}

func momentList(args []string) int {
	pos, _, ok := parse("moment list", args, nil, "PROJECT")
	if !ok {
		return out.ExitUsage
	}
	ctx := context.Background()
	p, err := inventory.GetProject(ctx, pos[0])
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

const maxPickTitle = 120

// momentSet records the moments a person or an agent picked, in place of the
// scored ones: the next `project run` cuts exactly these.
func momentSet(args []string) int {
	pos, vals, ok := parse("moment set", args, map[string]bool{"--start": true, "--end": true, "--title": true, "--clear": false}, "PROJECT")
	if !ok {
		return out.ExitUsage
	}
	ctx := context.Background()
	if vals["--clear"] == "true" {
		if vals["--start"] != "" || vals["--end"] != "" || vals["--title"] != "" {
			return out.P.Fail(out.Usage("moment set: --clear takes no --start, --end or --title"))
		}
		p, err := inventory.GetProject(ctx, pos[0])
		if err != nil {
			return out.P.Fail(mapErr(err))
		}
		if err := inventory.ClearPicks(ctx, p.ID); err != nil {
			return out.P.Fail(mapErr(err))
		}
		out.P.Print(map[string]any{"project": p.Name, "cleared": true})
		return 0
	}
	if vals["--start"] == "" || vals["--end"] == "" {
		return out.P.Fail(out.Usage("moment set: give --start SEC and --end SEC (or --clear)"))
	}
	start, err1 := strconv.ParseFloat(vals["--start"], 64)
	end, err2 := strconv.ParseFloat(vals["--end"], 64)
	if err1 != nil || err2 != nil {
		return out.P.Fail(out.Usage("moment set: --start and --end are seconds, such as 30 or 62.5"))
	}
	title := strings.TrimSpace(vals["--title"])
	if title == "" {
		title = "Picked moment"
	}
	if len([]rune(title)) > maxPickTitle {
		return out.P.Fail(out.Usage("moment set: --title is at most %d characters", maxPickTitle))
	}
	p, err := inventory.GetProject(ctx, pos[0])
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	src, err := inventory.GetSource(ctx, p.SourceID)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	if err := media.ValidatePick(start, end, src.DurationSec); err != nil {
		return out.P.Fail(out.Usage("moment set: %v", err))
	}
	m, err := inventory.AddMoment(ctx, inventory.Moment{ProjectID: p.ID, Title: title, Category: "picked", Score: 100, StartSec: start, EndSec: end, Picked: true})
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	out.P.Print(map[string]any{"project": p.Name, "moment": m, "next": "trady project run " + p.Name})
	return 0
}

type momentListData struct {
	Count   int                `json:"count"`
	Moments []inventory.Moment `json:"moments"`
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
		out.P.Print(usage("clip"))
		return 0
	}
	ctx := context.Background()
	switch args[0] {
	case "list":
		pos, _, ok := parse("clip "+args[0], args[1:], nil, "PROJECT")
		if !ok {
			return out.ExitUsage
		}
		p, err := inventory.GetProject(ctx, pos[0])
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
		out.P.Print(usage("job"))
		return 0
	}
	ctx := context.Background()
	switch args[0] {
	case "list":
		if _, _, ok := parse("job list", args[1:], nil); !ok {
			return out.ExitUsage
		}
		list, err := inventory.ListJobs(ctx)
		if err != nil {
			return out.P.Fail(mapErr(err))
		}
		out.P.Print(jobListData{Count: len(list), Jobs: list})
		return 0
	case "get":
		pos, _, ok := parse("job get", args[1:], nil, "ID")
		if !ok {
			return out.ExitUsage
		}
		j, err := inventory.GetJob(ctx, pos[0])
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
	if _, _, ok := parse("list", args, nil); !ok {
		return out.ExitUsage
	}
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
