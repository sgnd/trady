package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// allowedNetImporters may import "net" without making trady dial.
// golang.org/x/sys/windows (pulled in by golang.org/x/term for hidden input)
// imports net for socket types only; it never opens a connection itself.
var allowedNetImporters = map[string]bool{
	"golang.org/x/sys/windows": true,
}

// TestNoNetworkCode fails when network code can get into the binary.
// trady is private by default: it makes no network call. A feature that
// needs the network must be opt-in and must change this test on purpose.
// It checks every target OS, so a Mac run also catches a Windows-only import.
func TestNoNetworkCode(t *testing.T) {
	for _, goos := range []string{"darwin", "linux", "windows"} {
		cmd := exec.Command("go", "list", "-deps", "-f", "{{.ImportPath}} {{join .Imports \" \"}}", ".")
		cmd.Env = append(os.Environ(), "GOOS="+goos, "CGO_ENABLED=0")
		b, err := cmd.Output()
		if err != nil {
			t.Fatalf("%s: %v", goos, err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
			f := strings.Fields(line)
			pkg, imports := f[0], f[1:]
			// packages that dial or speak a network protocol are never allowed
			if strings.HasPrefix(pkg, "net/") && pkg != "net/netip" || pkg == "crypto/tls" {
				t.Errorf("%s: the binary contains %s; trady must not make network calls", goos, pkg)
			}
			for _, imp := range imports {
				if imp == "net" && pkg != "net" && !allowedNetImporters[pkg] {
					t.Errorf("%s: %s imports net; trady must not make network calls", goos, pkg)
				}
			}
		}
	}
}
