// Package secrets stores provider credentials encrypted at rest with a key
// derived from the server's MASTER_KEY.
package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

const MinMasterKeyLength = 32

type Store struct {
	mu     sync.RWMutex
	path   string
	aead   cipher.AEAD
	values map[string]string
}

// Open returns nil when masterKey is empty: storing credentials is then disabled.
func Open(dir, masterKey string) (*Store, error) {
	if masterKey == "" {
		return nil, nil
	}
	if len(masterKey) < MinMasterKeyLength {
		return nil, errors.New("MASTER_KEY must contain at least 32 characters")
	}
	if e := os.MkdirAll(dir, 0700); e != nil {
		return nil, e
	}
	key := sha256.Sum256([]byte("ai-quota-monitor credentials v1\x00" + masterKey))
	block, e := aes.NewCipher(key[:])
	if e != nil {
		return nil, e
	}
	aead, e := cipher.NewGCM(block)
	if e != nil {
		return nil, e
	}
	s := &Store{path: filepath.Join(dir, "credentials.json"), aead: aead, values: map[string]string{}}
	b, e := os.ReadFile(s.path)
	if errors.Is(e, os.ErrNotExist) {
		return s, nil
	}
	if e != nil {
		return nil, e
	}
	if e = json.Unmarshal(b, &s.values); e != nil {
		return nil, e
	}
	if s.values == nil {
		s.values = map[string]string{}
	}
	return s, nil
}

// Enabled reports whether credentials can be stored.
func (s *Store) Enabled() bool { return s != nil }

// Set encrypts value for the account; the account ID is authenticated so an
// entry cannot be moved to another account.
func (s *Store) Set(id, value string) error {
	if s == nil {
		return errors.New("credential storage is disabled")
	}
	nonce := make([]byte, s.aead.NonceSize())
	if _, e := rand.Read(nonce); e != nil {
		return e
	}
	sealed := base64.StdEncoding.EncodeToString(s.aead.Seal(nonce, nonce, []byte(value), []byte(id)))
	s.mu.Lock()
	defer s.mu.Unlock()
	next := make(map[string]string, len(s.values)+1)
	for k, v := range s.values {
		next[k] = v
	}
	next[id] = sealed
	b, e := json.Marshal(next)
	if e != nil {
		return e
	}
	tmp := s.path + ".tmp"
	if e = os.WriteFile(tmp, b, 0600); e != nil {
		return e
	}
	if e = os.Rename(tmp, s.path); e != nil {
		return e
	}
	s.values = next
	return nil
}

// Get returns the stored credential. Entries that cannot be decrypted, for
// example after MASTER_KEY changed, count as missing.
func (s *Store) Get(id string) (string, bool) {
	if s == nil {
		return "", false
	}
	s.mu.RLock()
	sealed, ok := s.values[id]
	s.mu.RUnlock()
	if !ok {
		return "", false
	}
	raw, e := base64.StdEncoding.DecodeString(sealed)
	if e != nil || len(raw) < s.aead.NonceSize() {
		return "", false
	}
	n := s.aead.NonceSize()
	plain, e := s.aead.Open(nil, raw[:n], raw[n:], []byte(id))
	if e != nil {
		return "", false
	}
	return string(plain), true
}
