package secrets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStore(t *testing.T) {
	dir := t.TempDir()
	master := strings.Repeat("k", 32)
	if s, e := Open(dir, ""); s != nil || e != nil || s.Enabled() {
		t.Fatal("empty master key must disable storage")
	}
	var disabled *Store
	if _, ok := disabled.Get("a"); ok || disabled.Set("a", "v") == nil {
		t.Fatal("disabled store must not hold values")
	}
	if _, e := Open(dir, "short"); e == nil {
		t.Fatal("short master key accepted")
	}
	s, e := Open(dir, master)
	if e != nil || !s.Enabled() {
		t.Fatal(e)
	}
	if e = s.Set("openrouter-personal", "sk-secret-value"); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(dir, "credentials.json")
	raw, _ := os.ReadFile(path)
	info, _ := os.Stat(path)
	if strings.Contains(string(raw), "sk-secret-value") || info.Mode().Perm() != 0600 {
		t.Fatal("credential must be encrypted in a private file")
	}
	// Persisted and readable after reopening; replacing overwrites.
	if s, e = Open(dir, master); e != nil {
		t.Fatal(e)
	}
	if v, ok := s.Get("openrouter-personal"); !ok || v != "sk-secret-value" {
		t.Fatal("credential lost after reopening")
	}
	if e = s.Set("openrouter-personal", "sk-new"); e != nil {
		t.Fatal(e)
	}
	if v, _ := s.Get("openrouter-personal"); v != "sk-new" {
		t.Fatal("credential not replaced")
	}
	if _, ok := s.Get("missing"); ok {
		t.Fatal("unexpected credential")
	}
	// An entry cannot be moved to another account, and a different master key reads nothing.
	s.values["other"] = s.values["openrouter-personal"]
	if _, ok := s.Get("other"); ok {
		t.Fatal("entry must be bound to its account")
	}
	if other, _ := Open(dir, strings.Repeat("x", 32)); other == nil {
		t.Fatal("expected a store")
	} else if _, ok := other.Get("openrouter-personal"); ok {
		t.Fatal("wrong master key must not decrypt")
	}
}
