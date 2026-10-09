package panelauth

import (
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPBKDF2Vector(t *testing.T) {
	// Published PBKDF2-HMAC-SHA256 test vector (password "password", salt "salt", 4096 rounds).
	got := hex.EncodeToString(pbkdf2([]byte("password"), []byte("salt"), 4096, 32))
	if got != "c5e478d59288c841aa530db6845c4c8d962893a001ce4e11a4963873aa98134a" {
		t.Fatalf("unexpected key %s", got)
	}
	long := hex.EncodeToString(pbkdf2([]byte("passwordPASSWORDpassword"), []byte("saltSALTsaltSALTsaltSALTsaltSALTsalt"), 4096, 40))
	if long != "348c89dbcbd32b2f32d814b8116e84cf2b17347ebc1800181c4e2a1fb8dd53e1c635518c7dac47e9" {
		t.Fatalf("unexpected multi-block key %s", long)
	}
}
func TestPasswordHash(t *testing.T) {
	if _, e := HashPassword("short"); e == nil {
		t.Fatal("short password accepted")
	}
	h, e := hashWith("correct horse battery", 1000)
	if e != nil || strings.ContainsAny(h, "$ \"'") || strings.Contains(h, "correct") {
		t.Fatalf("unexpected hash %q %v", h, e)
	}
	other, _ := hashWith("correct horse battery", 1000)
	if h == other {
		t.Fatal("hashes must be salted")
	}
	if CheckHash(h) != nil || !VerifyPassword(h, "correct horse battery") || VerifyPassword(h, "wrong") || VerifyPassword("plain-text", "plain-text") || CheckHash("plain-text") == nil {
		t.Fatal("verification mismatch")
	}
}
func TestLimiter(t *testing.T) {
	l := NewLimiter(2, time.Hour)
	l.Fail("a")
	if l.Blocked("a") != 0 {
		t.Fatal("blocked too early")
	}
	l.Fail("a")
	if l.Blocked("a") <= 0 || l.Blocked("b") != 0 {
		t.Fatal("expected only a to be blocked")
	}
	l.Reset("a")
	if l.Blocked("a") != 0 {
		t.Fatal("reset must unblock")
	}
	l = NewLimiter(1, time.Millisecond)
	l.Fail("a")
	time.Sleep(5 * time.Millisecond)
	if l.Blocked("a") != 0 {
		t.Fatal("window must expire")
	}
}
func TestSessions(t *testing.T) {
	dir := t.TempDir()
	s, e := Open(dir)
	if e != nil {
		t.Fatal(e)
	}
	first, e := s.Create()
	if e != nil || !s.Valid(first.AccessToken) || s.Valid(first.RefreshToken) || s.Valid("") || first.ExpiresIn != 900 {
		t.Fatal("login must create a valid access token", e)
	}
	// Only hashes reach the disk, in a private file.
	raw, _ := os.ReadFile(filepath.Join(dir, "panel-sessions.json"))
	info, _ := os.Stat(filepath.Join(dir, "panel-sessions.json"))
	if strings.Contains(string(raw), first.AccessToken) || strings.Contains(string(raw), first.RefreshToken) || info.Mode().Perm() != 0600 {
		t.Fatal("tokens must not be stored in clear")
	}
	// Sessions survive a restart.
	if s, e = Open(dir); e != nil || !s.Valid(first.AccessToken) {
		t.Fatal("session lost after reopening", e)
	}
	second, e := s.Refresh(first.RefreshToken)
	if e != nil || !s.Valid(second.AccessToken) || s.Valid(first.AccessToken) || second.RefreshToken == first.RefreshToken {
		t.Fatal("refresh must rotate both tokens", e)
	}
	if _, e = s.Refresh(first.AccessToken); !errors.Is(e, ErrInvalid) {
		t.Fatal("an access token must not refresh")
	}
	// Reusing the rotated refresh token revokes the whole session.
	if _, e = s.Refresh(first.RefreshToken); !errors.Is(e, ErrInvalid) || s.Valid(second.AccessToken) {
		t.Fatal("reuse must revoke the session")
	}
	if _, e = s.Refresh(second.RefreshToken); !errors.Is(e, ErrInvalid) {
		t.Fatal("revoked session must not refresh")
	}
	// Logout by access or refresh token; other sessions are untouched.
	a, _ := s.Create()
	b, _ := s.Create()
	if e = s.Revoke(a.AccessToken); e != nil || s.Valid(a.AccessToken) || !s.Valid(b.AccessToken) {
		t.Fatal("revoke by access token failed")
	}
	if _, e = s.Refresh(a.RefreshToken); !errors.Is(e, ErrInvalid) {
		t.Fatal("refresh token must die with its session")
	}
	if e = s.Revoke(b.RefreshToken); e != nil || s.Valid(b.AccessToken) {
		t.Fatal("revoke by refresh token failed")
	}
	// Expiry.
	c, _ := s.Create()
	s.now = func() time.Time { return time.Now().Add(AccessLifetime + time.Second) }
	if s.Valid(c.AccessToken) {
		t.Fatal("access token must expire")
	}
	if c, e = s.Refresh(c.RefreshToken); e != nil || !s.Valid(c.AccessToken) {
		t.Fatal("refresh must work after the access token expired", e)
	}
	s.now = func() time.Time { return time.Now().Add(2 * RefreshLifetime) }
	if _, e = s.Refresh(c.RefreshToken); !errors.Is(e, ErrInvalid) {
		t.Fatal("refresh token must expire")
	}
	// The number of sessions is bounded.
	s.now = time.Now
	for i := 0; i < maxSessions+5; i++ {
		if _, e = s.Create(); e != nil {
			t.Fatal(e)
		}
	}
	if len(s.sessions) != maxSessions {
		t.Fatalf("sessions=%d", len(s.sessions))
	}
}
