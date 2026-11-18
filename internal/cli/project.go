package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sgnd/trady/internal/inventory"
	"github.com/sgnd/trady/internal/media"
	"github.com/sgnd/trady/internal/out"
	"github.com/sgnd/trady/internal/store"
)

func cmdProject(args []string) int {
	if len(args) == 0 || hasHelp(args) {
		out.P.Print(strings.TrimSpace(`
usage: trady project create SOURCE [--name NAME]
       trady project list
       trady project get NAME|ID
       trady project status NAME|ID
       trady project run NAME|ID [--out DIR]
       trady project rm NAME|ID
`) + "\n")
		return 0
	}
	ctx := context.Background()
	switch args[0] {
	case "create":
		return projectCreate(ctx, args[1:])
	case "list":
		return projectList(ctx)
	case "get", "status":
		if len(args) < 2 {
			return out.P.Fail(out.Usage("project %s needs NAME or ID", args[0]))
		}
		p, err := inventory.GetProject(ctx, args[1])
		if err != nil {
			return out.P.Fail(mapErr(err))
		}
		out.P.Print(p)
		return 0
	case "run":
		return projectRun(ctx, args[1:])
	case "rm", "remove":
		if len(args) < 2 {
			return out.P.Fail(out.Usage("project rm needs NAME or ID"))
		}
		if err := inventory.RemoveProject(ctx, args[1]); err != nil {
			return out.P.Fail(mapErr(err))
		}
		out.P.Print(map[string]any{"removed": args[1]})
		return 0
	default:
		return out.P.Fail(out.Usage("unknown project subcommand %q", args[0]))
	}
}

func projectCreate(ctx context.Context, args []string) int {
	var source, name string
	for i := 0; i < len(args); i++ {
		if args[i] == "--name" && i+1 < len(args) {
			name = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(args[i], "-") {
			return out.P.Fail(out.Usage("unknown flag %s", args[i]))
		}
		source = args[i]
	}
	if source == "" {
		return out.P.Fail(out.Usage("project create needs a SOURCE name or id"))
	}
	p, err := inventory.CreateProject(ctx, name, source)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	out.P.Print(p)
	return 0
}

type projectListData struct {
	Count    int                 `json:"count"`
	Projects []inventory.Project `json:"projects"`
}

func (d projectListData) Text() string {
	if d.Count == 0 {
		return "no projects yet: trady project create SOURCE"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%-16s %-24s %-10s %s\n", "ID", "NAME", "STATUS", "SOURCE")
	for _, p := range d.Projects {
		id := p.ID
		if len(id) > 12 {
			id = id[:12]
		}
		sid := p.SourceID
		if len(sid) > 12 {
			sid = sid[:12]
		}
		fmt.Fprintf(&b, "%-16s %-24s %-10s %s\n", id, p.Name, p.Status, sid)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func projectList(ctx context.Context) int {
	list, err := inventory.ListProjects(ctx)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	out.P.Print(projectListData{Count: len(list), Projects: list})
	return 0
}

func projectRun(ctx context.Context, args []string) int {
	var idOrName, outDir string
	for i := 0; i < len(args); i++ {
		if args[i] == "--out" && i+1 < len(args) {
			outDir = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(args[i], "-") {
			return out.P.Fail(out.Usage("unknown flag %s", args[i]))
		}
		idOrName = args[i]
	}
	if idOrName == "" {
		return out.P.Fail(out.Usage("project run needs NAME or ID"))
	}
	p, err := inventory.GetProject(ctx, idOrName)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	src, err := inventory.GetSource(ctx, p.SourceID)
	if err != nil {
		return out.P.Fail(mapErr(err))
	}
	if outDir == "" {
		outDir = filepath.Join(store.Path(""), "exports", p.ID)
	}

	job, err := inventory.AddJob(ctx, inventory.Job{Kind: "project.run", ProjectID: p.ID, Status: "running", Message: "selecting moments"})
	if err != nil {
		return out.P.Fail(mapErr(err))
	}

	p.Status = "selecting"
	_ = inventory.SaveProject(ctx, p)

	suggested := media.SuggestMoments(p.ID, src.DurationSec)
	var moments []inventory.Moment
	for _, m := range suggested {
		saved, err := inventory.AddMoment(ctx, m)
		if err != nil {
			return failJob(ctx, job, p, err)
		}
		moments = append(moments, saved)
	}

	p.Status = "rendering"
	_ = inventory.SaveProject(ctx, p)
	job.Message = "rendering clips"
	_ = inventory.SaveJob(ctx, job)

	ffmpeg, _ := media.LookPath()
	rendered := 0
	for _, m := range moments {
		c := inventory.Clip{ProjectID: p.ID, MomentID: m.ID, Status: "pending"}
		name := m.ID
		if len(name) > 8 {
			name = name[:8]
		}
		clipPath := filepath.Join(outDir, fmt.Sprintf("%s.mp4", name))
		if ffmpeg == "" {
			c.Status = "skipped"
			c.Note = "ffmpeg not installed; moment kept, no file"
			c.Path = ""
		} else {
			pctx, cancel := media.WithTimeout(ctx)
			err := media.Cut(pctx, src.Path, m.StartSec, m.EndSec, clipPath)
			cancel()
			if err != nil {
				c.Status = "failed"
				c.Note = err.Error()
			} else {
				c.Status = "rendered"
				c.Path = clipPath
				rendered++
			}
		}
		if _, err := inventory.AddClip(ctx, c); err != nil {
			return failJob(ctx, job, p, err)
		}
	}

	now := time.Now().UTC()
	job.Status = "done"
	job.FinishedAt = &now
	job.Message = fmt.Sprintf("%d moments, %d clips rendered", len(moments), rendered)
	_ = inventory.SaveJob(ctx, job)
	p.Status = "done"
	p.Error = ""
	_ = inventory.SaveProject(ctx, p)

	out.P.Print(map[string]any{
		"project":  p,
		"job":      job,
		"moments":  len(moments),
		"rendered": rendered,
		"out_dir":  outDir,
		"note":     "moments are evenly spaced placeholders until local AI scoring lands",
	})
	return 0
}

func failJob(ctx context.Context, job inventory.Job, p inventory.Project, err error) int {
	now := time.Now().UTC()
	job.Status = "failed"
	job.FinishedAt = &now
	job.Message = err.Error()
	_ = inventory.SaveJob(ctx, job)
	p.Status = "failed"
	p.Error = err.Error()
	_ = inventory.SaveProject(ctx, p)
	return out.P.Fail(mapErr(err))
}
