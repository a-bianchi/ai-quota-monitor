package storage

import (
	"ai-quota-monitor/server/internal/usage"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	mu      sync.RWMutex
	path    string
	reports map[string]usage.Report
}

func New(dir string) (*Store, error) {
	if e := os.MkdirAll(dir, 0700); e != nil {
		return nil, e
	}
	s := &Store{path: filepath.Join(dir, "usage.json"), reports: map[string]usage.Report{}}
	b, e := os.ReadFile(s.path)
	if errors.Is(e, os.ErrNotExist) {
		return s, nil
	}
	if e != nil {
		return nil, e
	}
	if e = json.Unmarshal(b, &s.reports); e != nil {
		return nil, e
	}
	if s.reports == nil {
		s.reports = map[string]usage.Report{}
	}
	return s, nil
}
func (s *Store) Get(id string) (usage.Report, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.reports[id]
	return r, ok
}
func (s *Store) Put(id string, r usage.Report) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if old, ok := s.reports[id]; ok && r.UpdatedAt.Before(old.UpdatedAt) {
		return errors.New("observation is older than stored data")
	}
	next := make(map[string]usage.Report, len(s.reports)+1)
	for k, v := range s.reports {
		next[k] = v
	}
	next[id] = r
	b, e := json.Marshal(next)
	if e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(s.path), "usage-*.tmp")
	if e != nil {
		return e
	}
	name := f.Name()
	defer os.Remove(name)
	if _, e = f.Write(b); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	if e = os.Rename(name, s.path); e != nil {
		return e
	}
	s.reports = next
	return nil
}
