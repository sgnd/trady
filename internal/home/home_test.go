package home

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureTRADY_HOME(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "trady-home")
	t.Setenv("TRADY_HOME", home)
	got, err := Ensure()
	if err != nil {
		t.Fatal(err)
	}
	if got != home {
		t.Fatalf("Dir = %q, want %q", got, home)
	}
	st, err := os.Stat(got)
	if err != nil {
		t.Fatal(err)
	}
	if !st.IsDir() {
		t.Fatalf("%s is not a directory", got)
	}
}
