//go:build darwin

package store

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestKeychainKeyNotInArguments(t *testing.T) {
	t.Setenv("TRADY_HOME", t.TempDir())
	var gotArgs []string
	var gotStdin string
	old := runSecurity
	runSecurity = func(stdin string, args ...string) ([]byte, error) {
		gotArgs, gotStdin = args, stdin
		return nil, nil
	}
	t.Cleanup(func() { runSecurity = old })

	key, ok := keychainNewKey()
	if !ok {
		t.Fatal("keychainNewKey failed")
	}
	hexKey := hex.EncodeToString(key)
	if strings.Contains(strings.Join(gotArgs, " "), hexKey) {
		t.Fatalf("key is in the arguments: %v", gotArgs)
	}
	if len(gotArgs) != 1 || gotArgs[0] != "-i" {
		t.Fatalf("want security -i, got %v", gotArgs)
	}
	if !strings.Contains(gotStdin, hexKey) || !strings.HasPrefix(gotStdin, "add-generic-password -U") {
		t.Fatalf("stdin does not add the key: %q", gotStdin)
	}
}
