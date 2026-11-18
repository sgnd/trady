// Package sealrow seals one record of trady's local store. A sealed row is
//
//	format (1 byte) | suite (1 byte) | nonce (12 bytes) | AES-256-GCM ciphertext
//
// bound to the row's ID, type and version.
package sealrow

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

const Format byte = 1

const SuiteAESGCM byte = 1

const (
	headerLen = 2
	nonceLen  = 12
	MasterLen = 32
	hkdfInfo  = "trady local store v1"
)

var (
	ErrNewerFormat = errors.New("sealrow: row written by a newer trady; update trady")
	ErrOpen        = errors.New("sealrow: row does not open (wrong key, damaged, or moved)")
)

type Key struct {
	aead cipher.AEAD
}

func NewKey(master []byte) (*Key, error) {
	if len(master) != MasterLen {
		return nil, fmt.Errorf("sealrow: master key is %d bytes, want %d", len(master), MasterLen)
	}
	k, err := hkdf.Key(sha256.New, master, nil, hkdfInfo, 32)
	if err != nil {
		return nil, err
	}
	blk, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}
	a, err := cipher.NewGCM(blk)
	if err != nil {
		return nil, err
	}
	return &Key{aead: a}, nil
}

type Row struct {
	ID      string
	Type    string
	Version uint64
}

func (r Row) aad(format, suite byte) []byte {
	b := make([]byte, 0, 2+8+len(r.ID)+8+len(r.Type)+8)
	b = append(b, format, suite)
	b = binary.BigEndian.AppendUint64(b, uint64(len(r.ID)))
	b = append(b, r.ID...)
	b = binary.BigEndian.AppendUint64(b, uint64(len(r.Type)))
	b = append(b, r.Type...)
	return binary.BigEndian.AppendUint64(b, r.Version)
}

func (k *Key) Seal(r Row, plain []byte) ([]byte, error) {
	if r.ID == "" || r.Type == "" {
		return nil, errors.New("sealrow: a row needs an ID and a type")
	}
	out := make([]byte, headerLen+nonceLen, headerLen+nonceLen+len(plain)+k.aead.Overhead())
	out[0], out[1] = Format, SuiteAESGCM
	if _, err := rand.Read(out[headerLen:]); err != nil {
		return nil, err
	}
	return k.aead.Seal(out, out[headerLen:], plain, r.aad(Format, SuiteAESGCM)), nil
}

func (k *Key) Open(r Row, sealed []byte) ([]byte, error) {
	if len(sealed) < headerLen {
		return nil, ErrOpen
	}
	if sealed[0] > Format {
		return nil, ErrNewerFormat
	}
	if sealed[0] != Format || sealed[1] != SuiteAESGCM {
		if sealed[1] > SuiteAESGCM {
			return nil, ErrNewerFormat
		}
		return nil, ErrOpen
	}
	if len(sealed) < headerLen+nonceLen+k.aead.Overhead() {
		return nil, ErrOpen
	}
	plain, err := k.aead.Open(nil, sealed[headerLen:headerLen+nonceLen], sealed[headerLen+nonceLen:], r.aad(sealed[0], sealed[1]))
	if err != nil {
		return nil, ErrOpen
	}
	return plain, nil
}

func NewID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
