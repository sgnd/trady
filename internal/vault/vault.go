// Package vault keeps named secrets inside the sealed store.
// Values are sealed again under "trady vault entries v1".
package vault

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/sgnd/trady/internal/sealdb"
	"github.com/sgnd/trady/internal/sealrow"
	"github.com/sgnd/trady/internal/store"
)

const (
	secretType     = "secret"
	vaultHKDFLabel = "trady vault entries v1"
)

type Item struct {
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

type sealedEntry struct {
	Name      string    `json:"name"`
	Nonce     []byte    `json:"nonce"`
	Data      []byte    `json:"data"`
	UpdatedAt time.Time `json:"updated_at"`
}

func openDB(ctx context.Context, create bool) (*sealdb.DB, cipher.AEAD, error) {
	db, _, err := store.Open(ctx, create)
	if err != nil {
		return nil, nil, err
	}
	meta, err := sealdb.Peek(ctx, store.Path(store.DBFile))
	if err != nil {
		db.Close()
		return nil, nil, err
	}
	master, err := store.ResolveMaster(meta)
	if err != nil {
		db.Close()
		return nil, nil, err
	}
	aead, err := vaultAEAD(master)
	if err != nil {
		db.Close()
		return nil, nil, err
	}
	return db, aead, nil
}

func vaultAEAD(master []byte) (cipher.AEAD, error) {
	k, err := hkdf.Key(sha256.New, master, nil, vaultHKDFLabel, 32)
	if err != nil {
		return nil, err
	}
	blk, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(blk)
}

func sealValue(aead cipher.AEAD, name, value string) (nonce, data []byte, err error) {
	nonce = make([]byte, aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	aad := []byte(name)
	data = aead.Seal(nil, nonce, []byte(value), aad)
	return nonce, data, nil
}

func openValue(aead cipher.AEAD, name string, nonce, data []byte) (string, error) {
	plain, err := aead.Open(nil, nonce, data, []byte(name))
	if err != nil {
		return "", fmt.Errorf("vault: secret %q does not open", name)
	}
	return string(plain), nil
}

func Set(ctx context.Context, name, value string) error {
	if name == "" || value == "" {
		return errors.New("vault: name and value required")
	}
	db, aead, err := openDB(ctx, true)
	if err != nil {
		return err
	}
	defer db.Close()
	nonce, data, err := sealValue(aead, name, value)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	payload, err := json.Marshal(sealedEntry{Name: name, Nonce: nonce, Data: data, UpdatedAt: now})
	if err != nil {
		return err
	}
	return db.Update(ctx, func(tx *sealdb.Tx) error {
		rows, err := tx.List(secretType)
		if err != nil {
			return err
		}
		id := ""
		for _, r := range rows {
			var e sealedEntry
			if err := json.Unmarshal(r.Data, &e); err != nil {
				return err
			}
			if e.Name == name {
				id = r.ID
				break
			}
		}
		_, err = tx.Put(secretType, id, payload)
		return err
	})
}

func List(ctx context.Context) ([]Item, error) {
	db, _, err := openDB(ctx, false)
	if errors.Is(err, store.ErrNoStore) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer db.Close()
	var items []Item
	err = db.View(ctx, func(tx *sealdb.Tx) error {
		rows, err := tx.List(secretType)
		if err != nil {
			return err
		}
		for _, r := range rows {
			var e sealedEntry
			if err := json.Unmarshal(r.Data, &e); err != nil {
				return err
			}
			items = append(items, Item{Name: e.Name, UpdatedAt: e.UpdatedAt})
		}
		return nil
	})
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	return items, err
}

func Get(ctx context.Context, name string) (string, error) {
	db, aead, err := openDB(ctx, false)
	if err != nil {
		return "", err
	}
	defer db.Close()
	var found *sealedEntry
	err = db.View(ctx, func(tx *sealdb.Tx) error {
		rows, err := tx.List(secretType)
		if err != nil {
			return err
		}
		for _, r := range rows {
			var e sealedEntry
			if err := json.Unmarshal(r.Data, &e); err != nil {
				return err
			}
			if e.Name == name {
				found = &e
				return nil
			}
		}
		return sealdb.ErrNotFound
	})
	if err != nil {
		return "", err
	}
	return openValue(aead, name, found.Nonce, found.Data)
}

func Delete(ctx context.Context, name string) error {
	db, _, err := openDB(ctx, false)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.Update(ctx, func(tx *sealdb.Tx) error {
		rows, err := tx.List(secretType)
		if err != nil {
			return err
		}
		for _, r := range rows {
			var e sealedEntry
			if err := json.Unmarshal(r.Data, &e); err != nil {
				return err
			}
			if e.Name == name {
				return tx.Delete(r.ID)
			}
		}
		return sealdb.ErrNotFound
	})
}

// Exists reports whether a secret name is present (no value revealed).
func Exists(ctx context.Context, name string) (bool, error) {
	items, err := List(ctx)
	if err != nil {
		return false, err
	}
	for _, it := range items {
		if it.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// Ensure sealrow MasterLen is referenced for clarity in keyfile path.
var _ = sealrow.MasterLen
