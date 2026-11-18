// Package sealdb is trady's local store: one file of sealed rows
// (AES-256-GCM via sealrow). Pure Go — no cgo, no SQLite dependency.
package sealdb

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/sgnd/trady/internal/sealrow"
)

const Format = 1

var (
	ErrNewerFormat = errors.New("sealdb: store made by a newer trady; update trady")
	ErrNotFound    = errors.New("sealdb: record not found")
	ErrReadOnly    = errors.New("sealdb: change inside a read step")
	ErrTampered    = errors.New("sealdb: store was changed outside trady")
	ErrWrongKey    = errors.New("sealdb: the key does not open this store")
)

const keycheckID, keycheckType = "keycheck", "keycheck"

type fileStore struct {
	Format  int               `json:"format"`
	Meta    map[string]string `json:"meta"`
	Records []diskRec         `json:"records"`
}

type diskRec struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Version uint64 `json:"version"`
	MTime   int64  `json:"mtime"`
	Deleted bool   `json:"deleted"`
	Data    []byte `json:"data"`
}

type DB struct {
	path string
	key  *sealrow.Key
	mu   sync.Mutex
	fs   *fileStore
}

func Open(ctx context.Context, path string, key *sealrow.Key) (*DB, error) {
	_ = ctx
	if key == nil {
		return nil, errors.New("sealdb: no key")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	d := &DB{path: path, key: key}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		d.fs = &fileStore{Format: Format, Meta: map[string]string{"format": fmt.Sprintf("%d", Format)}}
		if err := d.Update(ctx, func(tx *Tx) error {
			_, err := tx.Put(keycheckType, keycheckID, []byte("trady"))
			return err
		}); err != nil {
			return nil, err
		}
		_ = os.Chmod(path, 0o600)
		return d, nil
	}
	if err != nil {
		return nil, err
	}
	var fs fileStore
	if err := json.Unmarshal(b, &fs); err != nil {
		return nil, fmt.Errorf("sealdb: corrupt store: %w", err)
	}
	if fs.Format > Format {
		return nil, ErrNewerFormat
	}
	if fs.Meta == nil {
		fs.Meta = map[string]string{}
	}
	d.fs = &fs
	err = d.View(ctx, func(tx *Tx) error {
		_, err := tx.Get(keycheckID)
		if errors.Is(err, sealrow.ErrOpen) {
			return ErrWrongKey
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	_ = os.Chmod(path, 0o600)
	return d, nil
}

func (d *DB) Close() error { return nil }

func Peek(ctx context.Context, path string) (map[string]string, error) {
	_ = ctx
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var fs fileStore
	if err := json.Unmarshal(b, &fs); err != nil {
		return nil, err
	}
	if fs.Meta == nil {
		fs.Meta = map[string]string{}
	}
	out := map[string]string{}
	for k, v := range fs.Meta {
		out[k] = v
	}
	out["format"] = fmt.Sprintf("%d", fs.Format)
	return out, nil
}

func (d *DB) View(ctx context.Context, fn func(*Tx) error) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return fn(&Tx{ctx: ctx, db: d, write: false})
}

func (d *DB) Update(ctx context.Context, fn func(*Tx) error) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	snap := cloneStore(d.fs)
	tx := &Tx{ctx: ctx, db: d, write: true}
	if err := fn(tx); err != nil {
		d.fs = snap
		return err
	}
	return d.persist()
}

func cloneStore(fs *fileStore) *fileStore {
	out := &fileStore{Format: fs.Format, Meta: map[string]string{}, Records: append([]diskRec{}, fs.Records...)}
	for k, v := range fs.Meta {
		out.Meta[k] = v
	}
	return out
}

func (d *DB) persist() error {
	b, err := json.MarshalIndent(d.fs, "", "  ")
	if err != nil {
		return err
	}
	tmp := d.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, d.path)
}

type Record struct {
	ID      string
	Type    string
	Version uint64
	MTime   time.Time
	Deleted bool
	Data    []byte
}

type Tx struct {
	ctx   context.Context
	db    *DB
	write bool
}

const headLen = 9

func head(del bool, mtime int64, data []byte) []byte {
	b := make([]byte, headLen, headLen+len(data))
	if del {
		b[0] = 1
	}
	binary.BigEndian.PutUint64(b[1:], uint64(mtime))
	return append(b, data...)
}

func (t *Tx) open(r diskRec) (Record, error) {
	plain, err := t.db.key.Open(sealrow.Row{ID: r.ID, Type: r.Type, Version: r.Version}, r.Data)
	if err != nil {
		return Record{}, fmt.Errorf("record %s: %w", r.ID, err)
	}
	if len(plain) < headLen || (plain[0] != 0) != r.Deleted || int64(binary.BigEndian.Uint64(plain[1:])) != r.MTime {
		return Record{}, fmt.Errorf("record %s: %w", r.ID, ErrTampered)
	}
	return Record{ID: r.ID, Type: r.Type, Version: r.Version, MTime: time.Unix(0, r.MTime), Deleted: r.Deleted, Data: plain[headLen:]}, nil
}

func (t *Tx) Get(id string) (Record, error) {
	for _, r := range t.db.fs.Records {
		if r.ID == id {
			return t.open(r)
		}
	}
	return Record{}, ErrNotFound
}

func (t *Tx) List(typ string) ([]Record, error) {
	var out []Record
	for _, r := range t.db.fs.Records {
		if r.Type != typ || r.Deleted {
			continue
		}
		rec, err := t.open(r)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].MTime.Equal(out[j].MTime) {
			return out[i].ID < out[j].ID
		}
		return out[i].MTime.Before(out[j].MTime)
	})
	return out, nil
}

func (t *Tx) Put(typ, id string, data []byte) (Record, error) {
	if !t.write {
		return Record{}, ErrReadOnly
	}
	var ver uint64
	idx := -1
	for i, r := range t.db.fs.Records {
		if r.ID == id && id != "" {
			if r.Type != typ {
				return Record{}, fmt.Errorf("sealdb: record %s is a %s, not a %s", id, r.Type, typ)
			}
			ver = r.Version
			idx = i
			break
		}
	}
	if id == "" {
		var err error
		id, err = sealrow.NewID()
		if err != nil {
			return Record{}, err
		}
	}
	now := time.Now().UnixNano()
	sealed, err := t.db.key.Seal(sealrow.Row{ID: id, Type: typ, Version: ver + 1}, head(false, now, data))
	if err != nil {
		return Record{}, err
	}
	rec := diskRec{ID: id, Type: typ, Version: ver + 1, MTime: now, Deleted: false, Data: sealed}
	if idx >= 0 {
		t.db.fs.Records[idx] = rec
	} else {
		t.db.fs.Records = append(t.db.fs.Records, rec)
	}
	return Record{ID: id, Type: typ, Version: ver + 1, MTime: time.Unix(0, now), Data: data}, nil
}

func (t *Tx) Delete(id string) error {
	if !t.write {
		return ErrReadOnly
	}
	for i, r := range t.db.fs.Records {
		if r.ID != id {
			continue
		}
		now := time.Now().UnixNano()
		sealed, err := t.db.key.Seal(sealrow.Row{ID: id, Type: r.Type, Version: r.Version + 1}, head(true, now, nil))
		if err != nil {
			return err
		}
		t.db.fs.Records[i] = diskRec{ID: id, Type: r.Type, Version: r.Version + 1, MTime: now, Deleted: true, Data: sealed}
		return nil
	}
	return ErrNotFound
}

func (t *Tx) Meta(k string) (string, error) {
	return t.db.fs.Meta[k], nil
}

func (t *Tx) SetMeta(k, v string) error {
	if !t.write {
		return ErrReadOnly
	}
	if k == "format" {
		return errors.New("sealdb: the format is set by the store itself")
	}
	t.db.fs.Meta[k] = v
	return nil
}
