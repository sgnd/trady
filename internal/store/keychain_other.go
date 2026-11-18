//go:build !darwin

package store

import "fmt"

func keychainGet() ([]byte, error) {
	return nil, fmt.Errorf("%w: no keychain on this OS (use TRADY_VAULT_PASS or store.key)", ErrKey)
}

func keychainNewKey() ([]byte, bool) { return nil, false }
