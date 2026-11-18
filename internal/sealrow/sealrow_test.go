package sealrow

import (
	"bytes"
	"testing"
)

func TestSealOpen(t *testing.T) {
	master := make([]byte, MasterLen)
	for i := range master {
		master[i] = byte(i + 1)
	}
	k, err := NewKey(master)
	if err != nil {
		t.Fatal(err)
	}
	r := Row{ID: "abc", Type: "source", Version: 1}
	sealed, err := k.Seal(r, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	plain, err := k.Open(r, sealed)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plain, []byte("hello")) {
		t.Fatalf("got %q", plain)
	}
	r.Version = 2
	if _, err := k.Open(r, sealed); err != ErrOpen {
		t.Fatalf("want ErrOpen, got %v", err)
	}
}
