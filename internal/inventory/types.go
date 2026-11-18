// Package inventory holds sealed media domain records (source, project, …).
package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/sgnd/trady/internal/sealdb"
	"github.com/sgnd/trady/internal/store"
)

const (
	TypeSource  = "source"
	TypeProject = "project"
	TypeMoment  = "moment"
	TypeClip    = "clip"
	TypeJob     = "job"
)

type Source struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	DurationSec float64   `json:"duration_sec,omitempty"`
	Width       int       `json:"width,omitempty"`
	Height      int       `json:"height,omitempty"`
	Codec       string    `json:"codec,omitempty"`
	ProbeNote   string    `json:"probe_note,omitempty"`
	AddedAt     time.Time `json:"added_at"`
}

type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SourceID  string    `json:"source_id"`
	Status    string    `json:"status"` // pending|probing|ready|selecting|rendering|done|failed
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Moment struct {
	ID        string  `json:"id"`
	ProjectID string  `json:"project_id"`
	Title     string  `json:"title"`
	Category  string  `json:"category"`
	Score     int     `json:"score"`
	StartSec  float64 `json:"start_sec"`
	EndSec    float64 `json:"end_sec"`
}

type Clip struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	MomentID  string `json:"moment_id"`
	Path      string `json:"path,omitempty"`
	Status    string `json:"status"` // pending|rendered|failed|skipped
	Note      string `json:"note,omitempty"`
}

type Job struct {
	ID         string     `json:"id"`
	Kind       string     `json:"kind"`
	ProjectID  string     `json:"project_id,omitempty"`
	Status     string     `json:"status"` // queued|running|done|failed
	Message    string     `json:"message,omitempty"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

func withDB(ctx context.Context, create bool, fn func(*sealdb.DB) error) error {
	db, _, err := store.Open(ctx, create)
	if err != nil {
		return err
	}
	defer db.Close()
	return fn(db)
}

func putJSON(ctx context.Context, typ, id string, v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	var outID string
	err = withDB(ctx, true, func(db *sealdb.DB) error {
		return db.Update(ctx, func(tx *sealdb.Tx) error {
			r, err := tx.Put(typ, id, b)
			if err != nil {
				return err
			}
			outID = r.ID
			return nil
		})
	})
	return outID, err
}

func listJSON[T any](ctx context.Context, typ string, decode func([]byte) (T, error)) ([]T, error) {
	var out []T
	err := withDB(ctx, false, func(db *sealdb.DB) error {
		return db.View(ctx, func(tx *sealdb.Tx) error {
			rows, err := tx.List(typ)
			if err != nil {
				return err
			}
			for _, r := range rows {
				v, err := decode(r.Data)
				if err != nil {
					return err
				}
				out = append(out, v)
			}
			return nil
		})
	})
	if errors.Is(err, store.ErrNoStore) {
		return nil, nil
	}
	return out, err
}

func getByID[T any](ctx context.Context, id string, decode func([]byte) (T, error)) (T, error) {
	var zero T
	var found T
	err := withDB(ctx, false, func(db *sealdb.DB) error {
		return db.View(ctx, func(tx *sealdb.Tx) error {
			r, err := tx.Get(id)
			if err != nil {
				return err
			}
			if r.Deleted {
				return sealdb.ErrNotFound
			}
			v, err := decode(r.Data)
			if err != nil {
				return err
			}
			found = v
			return nil
		})
	})
	if err != nil {
		return zero, err
	}
	return found, nil
}

func deleteID(ctx context.Context, id string) error {
	return withDB(ctx, false, func(db *sealdb.DB) error {
		return db.Update(ctx, func(tx *sealdb.Tx) error {
			return tx.Delete(id)
		})
	})
}
