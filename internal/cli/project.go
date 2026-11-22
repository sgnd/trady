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
		out.P.Print(usage("project"))
		return 0
	}
	ctx := context.Background()
	switch args[0] {
	case "create":
		return projectCreate(ctx, args[1:])
	case "list":
		if _, _, ok := parse("project list", args[1:], nil); !ok {
			return out.ExitUsage
		}
		return projectList(ctx)
	case "get", "status":
		pos, _, ok := parse("project "+args[0], args[1:], nil, "NAME|ID")
		if !ok {
			return out.ExitUsage
		}
		p, err := inventory.GetProject(ctx, pos[0])
		if err != nil {
			return out.P.Fail(mapErr(err))
		}
		out.P.Print(p)
		return 0
	case "run":
		return projectRun(ctx, args[1:])
	case "rm", "remove":
		pos, _, ok := parse("project rm", args[1:], nil, "NAME|ID")
		if !ok {
			return out.ExitUsage
		}
		if err := inventory.RemoveProject(ctx, pos[0]); err != nil {
			return out.P.Fail(mapErr(err))
		}
		out.P.Print(map[string]any{"removed": pos[0]})
		return 0
	default:
		return out.P.Fail(out.Usage("unknown project subcommand %q", args[0]))
	}
}

func projectCreate(ctx context.Context, args []string) int {
	pos, vals, ok := parse("project create", args, map[string]bool{"--name": true}, "SOURCE")
	if !ok {
		return out.ExitUsage
	}
	source, name := pos[0], vals["--name"]
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
	pos, vals, ok := parse("project run", args, map[string]bool{"--out": true}, "NAME|ID")
	if !ok {
		return out.ExitUsage
	}
	idOrName, outDir := pos[0], vals["--out"]
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

	// a new run replaces the clips and the scored moments of the last one;
	// the moments picked with `moment set` stay
	if err := inventory.ClearSuggested(ctx, p.ID); err != nil {
		return failJob(ctx, job, p, err)
	}

	segs, captionNote := captionsFor(ctx, src, &job)

	p.Status = "selecting"
	_ = inventory.SaveProject(ctx, p)
	job.Message = "selecting moments"
	_ = inventory.SaveJob(ctx, job)

	moments, selection, err := chooseMoments(ctx, p, src, segs)
	if err != nil {
		return failJob(ctx, job, p, err)
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
			err := media.Cut(pctx, src.Path, m.StartSec, m.EndSec, clipPath, media.Captions{
				ASS: media.ASS(segs, m.StartSec, m.EndSec, src.Width, src.Height),
				SRT: media.SRT(segs, m.StartSec, m.EndSec),
			})
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
		"project":   p,
		"job":       job,
		"moments":   len(moments),
		"rendered":  rendered,
		"out_dir":   outDir,
		"captions":  captionNote,
		"selection": selection,
	})
	return 0
}

// captionsFor returns the source's transcript, making it first when
// whisper.cpp and a model are on this machine. The note says what happened.
func captionsFor(ctx context.Context, src inventory.Source, job *inventory.Job) ([]inventory.Segment, string) {
	if t, err := inventory.GetTranscript(ctx, src.ID); err == nil {
		return t.Segments, captionNote("from the saved transcript", t.Segments)
	}
	if media.WhisperPath() == "" || media.ModelPath() == "" {
		return nil, "off: install whisper-cli and a model to burn in captions (see `trady doctor`)"
	}
	job.Message = "transcribing"
	_ = inventory.SaveJob(ctx, *job)
	tctx, cancel := media.WithTimeout(ctx)
	defer cancel()
	segs, model, err := media.Transcribe(tctx, src.Path)
	if err != nil {
		return nil, "off: transcription failed: " + err.Error()
	}
	if _, err := inventory.SaveTranscript(ctx, src.ID, model, segs); err != nil {
		return nil, "off: could not save the transcript: " + err.Error()
	}
	return segs, captionNote("from a new transcript", segs)
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

// chooseMoments returns the moments to cut: the ones picked with
// `moment set` when there are any, else the best windows of the transcript,
// else evenly spaced placeholders. The note says which.
func chooseMoments(ctx context.Context, p inventory.Project, src inventory.Source, segs []inventory.Segment) ([]inventory.Moment, string, error) {
	existing, err := inventory.ListMoments(ctx, p.ID)
	if err != nil {
		return nil, "", err
	}
	var picks []inventory.Moment
	for _, m := range existing {
		if m.Picked {
			picks = append(picks, m)
		}
	}
	if len(picks) > 0 {
		return picks, "picked with `trady moment set`", nil
	}
	suggested := media.ScoreMoments(p.ID, segs, media.MomentCount(src.DurationSec))
	selection := "scored from the transcript"
	if len(suggested) == 0 {
		suggested = media.SuggestMoments(p.ID, src.DurationSec)
		selection = "evenly spaced placeholders (no transcript with speech)"
	}
	var moments []inventory.Moment
	for _, m := range suggested {
		saved, err := inventory.AddMoment(ctx, m)
		if err != nil {
			return nil, "", err
		}
		moments = append(moments, saved)
	}
	return moments, selection, nil
}

// captionNote says how the captions are made: word by word when the
// transcript has word times, else one line per segment (an older transcript).
func captionNote(from string, segs []inventory.Segment) string {
	for _, s := range segs {
		if len(s.Words) > 0 {
			return fmt.Sprintf("burned in word by word, %s (%d segments)", from, len(segs))
		}
	}
	return fmt.Sprintf("burned in one line per segment, %s (%d segments); run `trady source transcribe` again for word by word", from, len(segs))
}
