package sealdb

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sgnd/trady/internal/sealrow"
)

func TestPutGetList(t *testing.T) {
	dir := t.TempDir()
	master := make([]byte, sealrow.MasterLen)
	for i := range master {
		master[i] = byte(i + 3)
	}
	key, err := sealrow.NewKey(master)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(dir, "store.db"), key)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var id string
	err = db.Update(ctx, func(tx *Tx) error {
		r, err := tx.Put("source", "", []byte(`{"name":"a"}`))
		if err != nil {
			return err
		}
		id = r.ID
		return tx.SetMeta("key_mode", "keyfile")
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.View(ctx, func(tx *Tx) error {
		r, err := tx.Get(id)
		if err != nil {
			return err
		}
		if string(r.Data) != `{"name":"a"}` {
			t.Fatalf("data %s", r.Data)
		}
		list, err := tx.List("source")
		if err != nil {
			return err
		}
		if len(list) != 1 {
			t.Fatalf("list %d", len(list))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
