package panelauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

const (
	AccessLifetime  = 15 * time.Minute
	RefreshLifetime = 30 * 24 * time.Hour
	maxSessions     = 32
)

var ErrInvalid = errors.New("invalid or expired token")

type Tokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

// session keeps only hashes: a copy of the file does not reveal usable tokens.
type session struct {
	AccessHash     string    `json:"accessHash"`
	AccessExpires  time.Time `json:"accessExpires"`
	RefreshHash    string    `json:"refreshHash"`
	PreviousHash   string    `json:"previousRefreshHash,omitempty"`
	RefreshExpires time.Time `json:"refreshExpires"`
	CreatedAt      time.Time `json:"createdAt"`
}

// Sessions stores panel sessions on disk so they survive restarts.
type Sessions struct {
	mu       sync.Mutex
	path     string
	sessions map[string]*session
	now      func() time.Time
}

func Open(dir string) (*Sessions, error) {
	if e := os.MkdirAll(dir, 0700); e != nil {
		return nil, e
	}
	s := &Sessions{path: filepath.Join(dir, "panel-sessions.json"), sessions: map[string]*session{}, now: time.Now}
	b, e := os.ReadFile(s.path)
	if errors.Is(e, os.ErrNotExist) {
		return s, nil
	}
	if e != nil {
		return nil, e
	}
	if e = json.Unmarshal(b, &s.sessions); e != nil {
		return nil, e
	}
	if s.sessions == nil {
		s.sessions = map[string]*session{}
	}
	return s, nil
}
func hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func random(prefix string) (string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return prefix + base64.RawURLEncoding.EncodeToString(b), nil
}

// save drops expired sessions, keeps the newest ones and writes the file atomically.
func (s *Sessions) save() error {
	now := s.now()
	ids := make([]string, 0, len(s.sessions))
	for id, v := range s.sessions {
		if !now.Before(v.RefreshExpires) {
			delete(s.sessions, id)
			continue
		}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return s.sessions[ids[i]].CreatedAt.After(s.sessions[ids[j]].CreatedAt) })
	for _, id := range ids[min(len(ids), maxSessions):] {
		delete(s.sessions, id)
	}
	b, e := json.Marshal(s.sessions)
	if e != nil {
		return e
	}
	tmp := s.path + ".tmp"
	if e = os.WriteFile(tmp, b, 0600); e != nil {
		return e
	}
	return os.Rename(tmp, s.path)
}
func (s *Sessions) issue(v *session) (Tokens, error) {
	access, e := random("aqm_a_")
	if e != nil {
		return Tokens{}, e
	}
	refresh, e := random("aqm_r_")
	if e != nil {
		return Tokens{}, e
	}
	now := s.now()
	v.AccessHash, v.AccessExpires = hash(access), now.Add(AccessLifetime)
	v.RefreshHash, v.RefreshExpires = hash(refresh), now.Add(RefreshLifetime)
	return Tokens{AccessToken: access, RefreshToken: refresh, ExpiresIn: int(AccessLifetime.Seconds())}, nil
}

// Create starts a session after a successful login.
func (s *Sessions) Create() (Tokens, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, e := random("")
	if e != nil {
		return Tokens{}, e
	}
	v := &session{CreatedAt: s.now()}
	tokens, e := s.issue(v)
	if e != nil {
		return Tokens{}, e
	}
	s.sessions[id] = v
	return tokens, s.save()
}

// Refresh rotates both tokens; the refresh token used stops working. Presenting
// an already rotated refresh token revokes the session, since it means the
// token was copied.
func (s *Sessions) Refresh(refreshToken string) (Tokens, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	h := hash(refreshToken)
	for id, v := range s.sessions {
		if v.PreviousHash == h {
			delete(s.sessions, id)
			_ = s.save()
			return Tokens{}, ErrInvalid
		}
		if v.RefreshHash != h {
			continue
		}
		if !s.now().Before(v.RefreshExpires) {
			delete(s.sessions, id)
			_ = s.save()
			return Tokens{}, ErrInvalid
		}
		v.PreviousHash = v.RefreshHash
		tokens, e := s.issue(v)
		if e != nil {
			return Tokens{}, e
		}
		return tokens, s.save()
	}
	return Tokens{}, ErrInvalid
}

// Valid reports whether accessToken belongs to a live session.
func (s *Sessions) Valid(accessToken string) bool {
	if s == nil || accessToken == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	h := hash(accessToken)
	for _, v := range s.sessions {
		if v.AccessHash == h {
			return s.now().Before(v.AccessExpires)
		}
	}
	return false
}

// Revoke ends the session that owns the given access or refresh token.
func (s *Sessions) Revoke(token string) error {
	if token == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	h := hash(token)
	for id, v := range s.sessions {
		if v.AccessHash == h || v.RefreshHash == h {
			delete(s.sessions, id)
			return s.save()
		}
	}
	return nil
}
