//go:build darwin

package store

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"os/exec"
	"strings"
	"time"
)

func keychainService() string {
	h := sha256.Sum256([]byte(Path("")))
	return "trady vault " + hex.EncodeToString(h[:6])
}

func keychainGet() ([]byte, error) {
	cmd := exec.Command("/usr/bin/security", "find-generic-password", "-s", keychainService(), "-a", "trady", "-w")
	out, err := cmd.Output()
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
	hexKey := hex.EncodeToString(key)
	ctxDone := make(chan struct{})
	go func() {
		time.Sleep(5 * time.Second)
		close(ctxDone)
	}()
	cmd := exec.Command("/usr/bin/security", "add-generic-password", "-U", "-s", keychainService(), "-a", "trady", "-w", hexKey)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, false
	}
	return key, true
}
