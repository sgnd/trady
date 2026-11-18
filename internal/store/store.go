// Package store opens the sealed trady home under TRADY_HOME / ~/.trady.
package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/crypto/scrypt"

	"github.com/sgnd/trady/internal/home"
	"github.com/sgnd/trady/internal/sealdb"
	"github.com/sgnd/trady/internal/sealrow"
)

const (
	keyLen      = 32
	PassphraseN = 1 << 15
	DBFile      = "store.db"
	KeyFile     = "store.key"
)

var (
	ErrNoStore = errors.New("store: no store yet")
	ErrLocked  = errors.New("store: locked (set TRADY_VAULT_PASS or unlock the keychain)")
	ErrKey     = errors.New("store: key problem")
)

// Path joins name under the home directory.
func Path(name string) string {
	d, err := home.Dir()
	if err != nil {
		return name
	}
	if name == "" {
		return d
	}
	return filepath.Join(d, name)
}

func EnsureHome() error {
	_, err := home.Ensure()
	return err
}

// Open opens store.db, creating it when create is true.
func Open(ctx context.Context, create bool) (*sealdb.DB, string, error) {
	if err := EnsureHome(); err != nil {
		return nil, "", err
	}
	path := Path(DBFile)
	meta, err := sealdb.Peek(ctx, path)
	if errors.Is(err, os.ErrNotExist) {
		if !create {
			return nil, "", ErrNoStore
		}
		master, newMeta, err := newMaster()
		if err != nil {
			return nil, "", err
		}
		key, err := sealrow.NewKey(master)
		if err != nil {
			return nil, "", err
		}
		db, err := sealdb.Open(ctx, path, key)
		if err != nil {
			return nil, "", err
		}
		err = db.Update(ctx, func(tx *sealdb.Tx) error {
			for k, v := range newMeta {
				if err := tx.SetMeta(k, v); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			db.Close()
			return nil, "", err
		}
		return db, newMeta["key_mode"], nil
	}
	if err != nil {
		return nil, "", err
	}
	master, err := ResolveMaster(meta)
	if err != nil {
		return nil, "", err
	}
	key, err := sealrow.NewKey(master)
	if err != nil {
		return nil, "", err
	}
	db, err := sealdb.Open(ctx, path, key)
	if err != nil {
		return nil, meta["key_mode"], err
	}
	return db, meta["key_mode"], nil
}

// ResolveMaster returns the 32-byte store master from meta + env/keychain/keyfile.
func ResolveMaster(meta map[string]string) ([]byte, error) {
	return loadMaster(meta)
}

func KeyMode(ctx context.Context) (string, error) {
	path := Path(DBFile)
	meta, err := sealdb.Peek(ctx, path)
	if errors.Is(err, os.ErrNotExist) {
		return "", ErrNoStore
	}
	if err != nil {
		return "", err
	}
	return meta["key_mode"], nil
}

func loadMaster(meta map[string]string) ([]byte, error) {
	switch meta["key_mode"] {
	case "passphrase":
		p := os.Getenv("TRADY_VAULT_PASS")
		if p == "" {
			return nil, ErrLocked
		}
		salt, err := hex.DecodeString(meta["salt"])
		n, nerr := strconv.Atoi(meta["scrypt_n"])
		if err != nil || len(salt) < 16 || nerr != nil || n < 2 {
			return nil, fmt.Errorf("%w: passphrase settings damaged", ErrKey)
		}
		return scrypt.Key([]byte(p), salt, n, 8, 1, keyLen)
	case "keychain":
		k, err := keychainGet()
		if err != nil {
			return nil, err
		}
		if len(k) < keyLen {
			return nil, fmt.Errorf("%w: keychain item too short", ErrKey)
		}
		return k[:keyLen], nil
	case "keyfile":
		b, err := os.ReadFile(Path(KeyFile))
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s missing", ErrKey, Path(KeyFile))
		}
		if err != nil {
			return nil, err
		}
		if len(b) < keyLen {
			return nil, fmt.Errorf("%w: %s too short", ErrKey, Path(KeyFile))
		}
		return b[:keyLen], nil
	default:
		return nil, fmt.Errorf("%w: unknown key_mode %q", ErrKey, meta["key_mode"])
	}
}

func newMaster() ([]byte, map[string]string, error) {
	if p := os.Getenv("TRADY_VAULT_PASS"); p != "" {
		if len(p) < 15 {
			return nil, nil, fmt.Errorf("%w: TRADY_VAULT_PASS must be at least 15 characters", ErrKey)
		}
		salt := make([]byte, 16)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
		k, err := scrypt.Key([]byte(p), salt, PassphraseN, 8, 1, keyLen)
		if err != nil {
			return nil, nil, err
		}
		return k, map[string]string{
			"key_mode": "passphrase",
			"salt":     hex.EncodeToString(salt),
			"scrypt_n": strconv.Itoa(PassphraseN),
		}, nil
	}
	if os.Getenv("TRADY_KEYCHAIN") != "off" {
		if k, ok := keychainNewKey(); ok && len(k) >= keyLen {
			return k[:keyLen], map[string]string{"key_mode": "keychain"}, nil
		}
	}
	k, err := writeKeyFile(Path(KeyFile))
	if err != nil {
		return nil, nil, err
	}
	return k, map[string]string{"key_mode": "keyfile"}, nil
}

func writeKeyFile(kp string) ([]byte, error) {
	key := make([]byte, keyLen)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(kp), ".tmp-key-*")
	if err != nil {
		return nil, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(key); err != nil {
		tmp.Close()
		return nil, err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}
	if err := os.Link(tmpName, kp); err == nil {
		return key, nil
	} else if !errors.Is(err, os.ErrExist) {
		f, err2 := os.OpenFile(kp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err2, os.ErrExist) {
			return os.ReadFile(kp)
		}
		if err2 != nil {
			return nil, err2
		}
		if _, err := f.Write(key); err != nil {
			f.Close()
			os.Remove(kp)
			return nil, err
		}
		_ = f.Sync()
		f.Close()
		return key, nil
	}
	return os.ReadFile(kp)
}
