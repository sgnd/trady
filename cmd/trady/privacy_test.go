package main

import (
	"os/exec"
	"strings"
	"testing"
)

// TestNoNetworkCode fails when any network package gets into the binary.
// trady is private by default: it makes no network call. A feature that
// needs the network must be opt-in and must change this test on purpose.
func TestNoNetworkCode(t *testing.T) {
	b, err := exec.Command("go", "list", "-deps", ".").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, pkg := range strings.Fields(string(b)) {
		if pkg == "net" || strings.HasPrefix(pkg, "net/") || pkg == "crypto/tls" {
			t.Errorf("the binary imports %s; trady must not make network calls", pkg)
		}
	}
}
