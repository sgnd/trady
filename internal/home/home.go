// Package home resolves TRADY_HOME (default ~/.trady) and keeps it private.
package home

import (
	"fmt"
	"os"
	"path/filepath"
)

// Dir is the folder for the sealed store and runtime files.
// Override with TRADY_HOME.
func Dir() (string, error) {
	if h := os.Getenv("TRADY_HOME"); h != "" {
		return filepath.Clean(h), nil
	}
	d, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home: user home: %w", err)
	}
	return filepath.Join(d, ".trady"), nil
}

// Ensure creates Dir with mode 0700 if missing.
func Ensure() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return "", fmt.Errorf("home: make %s: %w", d, err)
	}
	return d, nil
}
