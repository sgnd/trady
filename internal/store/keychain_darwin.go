//go:build darwin

package store

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

const securityTool = "/usr/bin/security"

func keychainService() string {
	h := sha256.Sum256([]byte(Path("")))
	return "trady vault " + hex.EncodeToString(h[:6])
}

// runSecurity runs the security tool with stdin; tests replace it.
var runSecurity = func(stdin string, args ...string) ([]byte, error) {
	cmd := exec.Command(securityTool, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	cmd.Stderr = io.Discard
	return cmd.Output()
}

func keychainGet() ([]byte, error) {
	out, err := runSecurity("", "find-generic-password", "-s", keychainService(), "-a", "trady", "-w")
	if err != nil {
		return nil, err
	}
	s := strings.TrimSpace(string(out))
	b, err := hex.DecodeString(s)
	if err != nil {
		return []byte(s), nil
	}
	return b, nil
}

func keychainNewKey() ([]byte, bool) {
	key := make([]byte, keyLen)
	if _, err := rand.Read(key); err != nil {
		return nil, false
	}
	// `security -i` reads the command from stdin, so the key stays out of
	// the process list.
	line := fmt.Sprintf("add-generic-password -U -a trady -s %q -w %s\n", keychainService(), hex.EncodeToString(key))
	if _, err := runSecurity(line, "-i"); err != nil {
		return nil, false
	}
	return key, true
}
