package inventory

import (
	"context"
	"errors"
	"testing"

	"github.com/sgnd/trady/internal/sealdb"
)

func setup(t *testing.T) context.Context {
	t.Helper()
	t.Setenv("TRADY_HOME", t.TempDir())
	t.Setenv("TRADY_VAULT_PASS", "")
	t.Setenv("TRADY_KEYCHAIN", "off")
	return context.Background()
}

func TestEmptyListsAreNotNil(t *testing.T) {
	ctx := setup(t)
	s, err := ListSources(ctx)
	if err != nil || s == nil {
		t.Fatalf("sources %v, %v", s, err)
	}
	m, err := ListMoments(ctx, "")
	if err != nil || m == nil {
		t.Fatalf("moments %v, %v", m, err)
	}
}

func TestSourceAndProject(t *testing.T) {
	ctx := setup(t)
	s, err := AddSource(ctx, "", "/media/ep1.mp4", Source{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "ep1" {
		t.Fatalf("default name %q", s.Name)
	}
	if _, err := AddSource(ctx, "ep1", "/other.mp4", Source{}); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate source: want ErrConflict, got %v", err)
	}
	p, err := CreateProject(ctx, "", "ep1")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "ep1-project" || p.SourceID != s.ID || p.Status != "pending" {
		t.Fatalf("bad project %+v", p)
	}
	if _, err := CreateProject(ctx, "ep1-project", s.ID); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate project: want ErrConflict, got %v", err)
	}
	if _, err := CreateProject(ctx, "x", "nope"); !errors.Is(err, sealdb.ErrNotFound) {
		t.Fatalf("missing source: want ErrNotFound, got %v", err)
	}
	if got, err := GetProject(ctx, p.ID); err != nil || got.Name != p.Name {
		t.Fatalf("get by id: %+v, %v", got, err)
	}
}

func TestRemoveProjectRemovesChildren(t *testing.T) {
	ctx := setup(t)
	if _, err := AddSource(ctx, "src", "/a.mp4", Source{}); err != nil {
		t.Fatal(err)
	}
	keep, _ := CreateProject(ctx, "keep", "src")
	gone, _ := CreateProject(ctx, "gone", "src")
	for _, p := range []Project{keep, gone} {
		m, err := AddMoment(ctx, Moment{ProjectID: p.ID, StartSec: 0, EndSec: 10})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := AddClip(ctx, Clip{ProjectID: p.ID, MomentID: m.ID}); err != nil {
			t.Fatal(err)
		}
	}
	if err := RemoveProject(ctx, "gone"); err != nil {
		t.Fatal(err)
	}
	ms, _ := ListMoments(ctx, "")
	cs, _ := ListClips(ctx, "")
	if len(ms) != 1 || ms[0].ProjectID != keep.ID || len(cs) != 1 || cs[0].ProjectID != keep.ID {
		t.Fatalf("left moments %+v clips %+v", ms, cs)
	}
	if err := RemoveProject(ctx, "gone"); !errors.Is(err, sealdb.ErrNotFound) {
		t.Fatalf("second remove: want ErrNotFound, got %v", err)
	}
}

func TestJobs(t *testing.T) {
	ctx := setup(t)
	j, err := AddJob(ctx, Job{Kind: "project.run", Status: "running"})
	if err != nil {
		t.Fatal(err)
	}
	if j.StartedAt.IsZero() {
		t.Fatal("StartedAt not set")
	}
	j.Status = "done"
	if err := SaveJob(ctx, j); err != nil {
		t.Fatal(err)
	}
	got, err := GetJob(ctx, j.ID)
	if err != nil || got.Status != "done" {
		t.Fatalf("got %+v, %v", got, err)
	}
}
