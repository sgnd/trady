package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/sgnd/trady/internal/sealdb"
	"github.com/sgnd/trady/internal/sealrow"
)

func decodeSource(b []byte) (Source, error) {
	var s Source
	return s, json.Unmarshal(b, &s)
}
func decodeProject(b []byte) (Project, error) {
	var p Project
	return p, json.Unmarshal(b, &p)
}
func decodeMoment(b []byte) (Moment, error) {
	var m Moment
	return m, json.Unmarshal(b, &m)
}
func decodeClip(b []byte) (Clip, error) {
	var c Clip
	return c, json.Unmarshal(b, &c)
}
func decodeJob(b []byte) (Job, error) {
	var j Job
	return j, json.Unmarshal(b, &j)
}

// ErrConflict means a source or project with that name already exists.
var ErrConflict = errors.New("already exists")

func AddSource(ctx context.Context, name, path string, probe Source) (Source, error) {
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	if _, err := GetSource(ctx, name); err == nil {
		return Source{}, fmt.Errorf("source %q: %w", name, ErrConflict)
	}
	id, err := sealrow.NewID()
	if err != nil {
		return Source{}, err
	}
	s := Source{
		ID: id, Name: name, Path: path,
		DurationSec: probe.DurationSec, Width: probe.Width, Height: probe.Height,
		Codec: probe.Codec, ProbeNote: probe.ProbeNote, AddedAt: time.Now().UTC(),
	}
	_, err = putJSON(ctx, TypeSource, id, s)
	return s, err
}

func ListSources(ctx context.Context) ([]Source, error) {
	return listJSON(ctx, TypeSource, decodeSource)
}

func GetSource(ctx context.Context, idOrName string) (Source, error) {
	all, err := ListSources(ctx)
	if err != nil {
		return Source{}, err
	}
	for _, s := range all {
		if s.ID == idOrName || s.Name == idOrName {
			return s, nil
		}
	}
	return Source{}, sealdb.ErrNotFound
}

func RemoveSource(ctx context.Context, idOrName string) error {
	s, err := GetSource(ctx, idOrName)
	if err != nil {
		return err
	}
	if t, err := GetTranscript(ctx, s.ID); err == nil {
		if err := deleteID(ctx, t.ID); err != nil {
			return err
		}
	}
	return deleteID(ctx, s.ID)
}

// SaveTranscript stores the transcript of a source, replacing an older one.
func SaveTranscript(ctx context.Context, sourceID, model string, segs []Segment) (Transcript, error) {
	id := ""
	if old, err := GetTranscript(ctx, sourceID); err == nil {
		id = old.ID
	} else {
		var err error
		if id, err = sealrow.NewID(); err != nil {
			return Transcript{}, err
		}
	}
	t := Transcript{ID: id, SourceID: sourceID, Model: model, Segments: segs, CreatedAt: time.Now().UTC()}
	_, err := putJSON(ctx, TypeTranscript, id, t)
	return t, err
}

// GetTranscript returns the transcript of a source.
func GetTranscript(ctx context.Context, sourceID string) (Transcript, error) {
	all, err := listJSON(ctx, TypeTranscript, func(b []byte) (Transcript, error) {
		var t Transcript
		return t, json.Unmarshal(b, &t)
	})
	if err != nil {
		return Transcript{}, err
	}
	for _, t := range all {
		if t.SourceID == sourceID {
			return t, nil
		}
	}
	return Transcript{}, sealdb.ErrNotFound
}

func CreateProject(ctx context.Context, name, sourceIDOrName string) (Project, error) {
	src, err := GetSource(ctx, sourceIDOrName)
	if err != nil {
		return Project{}, fmt.Errorf("source: %w", err)
	}
	if name == "" {
		name = src.Name + "-project"
	}
	if _, err := GetProject(ctx, name); err == nil {
		return Project{}, fmt.Errorf("project %q: %w", name, ErrConflict)
	}
	id, err := sealrow.NewID()
	if err != nil {
		return Project{}, err
	}
	now := time.Now().UTC()
	p := Project{ID: id, Name: name, SourceID: src.ID, Status: "pending", CreatedAt: now, UpdatedAt: now}
	_, err = putJSON(ctx, TypeProject, id, p)
	return p, err
}

func ListProjects(ctx context.Context) ([]Project, error) {
	return listJSON(ctx, TypeProject, decodeProject)
}

func GetProject(ctx context.Context, idOrName string) (Project, error) {
	all, err := ListProjects(ctx)
	if err != nil {
		return Project{}, err
	}
	for _, p := range all {
		if p.ID == idOrName || p.Name == idOrName {
			return p, nil
		}
	}
	return Project{}, sealdb.ErrNotFound
}

func SaveProject(ctx context.Context, p Project) error {
	p.UpdatedAt = time.Now().UTC()
	_, err := putJSON(ctx, TypeProject, p.ID, p)
	return err
}

func RemoveProject(ctx context.Context, idOrName string) error {
	p, err := GetProject(ctx, idOrName)
	if err != nil {
		return err
	}
	// remove the project's moments and clips too, so none are left orphaned
	moments, err := ListMoments(ctx, p.ID)
	if err != nil {
		return err
	}
	for _, m := range moments {
		if err := deleteID(ctx, m.ID); err != nil {
			return err
		}
	}
	clips, err := ListClips(ctx, p.ID)
	if err != nil {
		return err
	}
	for _, c := range clips {
		if err := deleteID(ctx, c.ID); err != nil {
			return err
		}
	}
	return deleteID(ctx, p.ID)
}

func AddMoment(ctx context.Context, m Moment) (Moment, error) {
	id, err := sealrow.NewID()
	if err != nil {
		return Moment{}, err
	}
	m.ID = id
	_, err = putJSON(ctx, TypeMoment, id, m)
	return m, err
}

func ListMoments(ctx context.Context, projectID string) ([]Moment, error) {
	all, err := listJSON(ctx, TypeMoment, decodeMoment)
	if err != nil {
		return nil, err
	}
	out := []Moment{}
	for _, m := range all {
		if projectID == "" || m.ProjectID == projectID {
			out = append(out, m)
		}
	}
	return out, nil
}

func AddClip(ctx context.Context, c Clip) (Clip, error) {
	id, err := sealrow.NewID()
	if err != nil {
		return Clip{}, err
	}
	c.ID = id
	_, err = putJSON(ctx, TypeClip, id, c)
	return c, err
}

func ListClips(ctx context.Context, projectID string) ([]Clip, error) {
	all, err := listJSON(ctx, TypeClip, decodeClip)
	if err != nil {
		return nil, err
	}
	out := []Clip{}
	for _, c := range all {
		if projectID == "" || c.ProjectID == projectID {
			out = append(out, c)
		}
	}
	return out, nil
}

func SaveClip(ctx context.Context, c Clip) error {
	_, err := putJSON(ctx, TypeClip, c.ID, c)
	return err
}

func AddJob(ctx context.Context, j Job) (Job, error) {
	id, err := sealrow.NewID()
	if err != nil {
		return Job{}, err
	}
	j.ID = id
	if j.StartedAt.IsZero() {
		j.StartedAt = time.Now().UTC()
	}
	_, err = putJSON(ctx, TypeJob, id, j)
	return j, err
}

func ListJobs(ctx context.Context) ([]Job, error) {
	return listJSON(ctx, TypeJob, decodeJob)
}

func GetJob(ctx context.Context, id string) (Job, error) {
	return getByID(ctx, id, decodeJob)
}

func SaveJob(ctx context.Context, j Job) error {
	_, err := putJSON(ctx, TypeJob, j.ID, j)
	return err
}
