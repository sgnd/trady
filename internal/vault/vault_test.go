package vault

import (
	"context"
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
