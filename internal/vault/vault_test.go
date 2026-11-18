package vault

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetListGetRm(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TRADY_HOME", dir)
	t.Setenv("TRADY_KEYCHAIN", "off")
	t.Setenv("TRADY_VAULT_PASS", "test-passphrase-ok")

	ctx := context.Background()
	if err := Set(ctx, "api", "secret-value"); err != nil {
		t.Fatal(err)
	}
	items, err := List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Name != "api" {
		t.Fatalf("%+v", items)
	}
	v, err := Get(ctx, "api")
	if err != nil {
		t.Fatal(err)
	}
	if v != "secret-value" {
		t.Fatalf("got %q", v)
	}
	if err := Delete(ctx, "api"); err != nil {
		t.Fatal(err)
	}
}

func setup(t *testing.T) context.Context {
	t.Helper()
	t.Setenv("TRADY_HOME", t.TempDir())
	t.Setenv("TRADY_KEYCHAIN", "off")
	t.Setenv("TRADY_VAULT_PASS", "test-passphrase-ok")
	return context.Background()
}

func TestOverwriteKeepsOneEntry(t *testing.T) {
	ctx := setup(t)
	for _, v := range []string{"one", "two"} {
		if err := Set(ctx, "api", v); err != nil {
			t.Fatal(err)
		}
	}
	items, _ := List(ctx)
	if len(items) != 1 {
		t.Fatalf("want 1 entry, got %d", len(items))
	}
	if v, _ := Get(ctx, "api"); v != "two" {
		t.Fatalf("got %q", v)
	}
}

func TestMissingAndInvalid(t *testing.T) {
	ctx := setup(t)
	if ok, err := Exists(ctx, "nope"); ok || err != nil {
		t.Fatalf("exists: %v %v", ok, err)
	}
	if _, err := Get(ctx, "nope"); err == nil {
		t.Fatal("get missing: want error")
	}
	if err := Delete(ctx, "nope"); err == nil {
		t.Fatal("delete missing: want error")
	}
	for _, n := range []string{"", "has space", "a/b", strings.Repeat("x", 65)} {
		if err := Set(ctx, n, "v"); err == nil {
			t.Fatalf("name %q: want error", n)
		}
	}
	if err := Set(ctx, "ok", ""); err == nil {
		t.Fatal("empty value: want error")
	}
}

func TestValueIsNotStoredInPlainText(t *testing.T) {
	ctx := setup(t)
	const value = "very-recognisable-secret-value-123"
	if err := Set(ctx, "api", value); err != nil {
		t.Fatal(err)
	}
	err := filepath.WalkDir(os.Getenv("TRADY_HOME"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if bytes.Contains(b, []byte(value)) {
			t.Errorf("%s holds the secret value in plain text", p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWrongPassphrase(t *testing.T) {
	ctx := setup(t)
	if err := Set(ctx, "api", "v"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("TRADY_VAULT_PASS", "another-passphrase-x")
	if _, err := Get(ctx, "api"); err == nil {
		t.Fatal("wrong passphrase: want error")
	}
}
