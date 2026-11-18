package store

import (
	"context"
	"errors"
	"testing"
)

func TestPassphraseMode(t *testing.T) {
	t.Setenv("TRADY_HOME", t.TempDir())
	t.Setenv("TRADY_VAULT_PASS", "a-test-passphrase-long-enough")
	ctx := context.Background()
	db, _, err := Open(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	if m, err := KeyMode(ctx); err != nil || m != "passphrase" {
		t.Fatalf("key mode %q, %v", m, err)
	}
	// reopen with the same passphrase works
	db, _, err = Open(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	// without it the store is locked
	t.Setenv("TRADY_VAULT_PASS", "")
	if _, _, err := Open(ctx, false); !errors.Is(err, ErrLocked) {
		t.Fatalf("want ErrLocked, got %v", err)
	}
}

func TestShortPassphrase(t *testing.T) {
	t.Setenv("TRADY_HOME", t.TempDir())
	t.Setenv("TRADY_VAULT_PASS", "short")
	if _, _, err := Open(context.Background(), true); !errors.Is(err, ErrKey) {
		t.Fatalf("want ErrKey, got %v", err)
	}
}

func TestKeyfileMode(t *testing.T) {
	t.Setenv("TRADY_HOME", t.TempDir())
	t.Setenv("TRADY_VAULT_PASS", "")
	t.Setenv("TRADY_KEYCHAIN", "off")
	ctx := context.Background()
	db, _, err := Open(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	if m, _ := KeyMode(ctx); m != "keyfile" {
		t.Fatalf("key mode %q", m)
	}
	db, _, err = Open(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestNoStore(t *testing.T) {
	t.Setenv("TRADY_HOME", t.TempDir())
	if _, _, err := Open(context.Background(), false); !errors.Is(err, ErrNoStore) {
		t.Fatalf("want ErrNoStore, got %v", err)
	}
	if _, err := KeyMode(context.Background()); !errors.Is(err, ErrNoStore) {
		t.Fatalf("KeyMode: want ErrNoStore, got %v", err)
	}
}

func TestUnknownKeyMode(t *testing.T) {
	if _, err := ResolveMaster(map[string]string{"key_mode": "weird"}); !errors.Is(err, ErrKey) {
		t.Fatalf("want ErrKey, got %v", err)
	}
}
