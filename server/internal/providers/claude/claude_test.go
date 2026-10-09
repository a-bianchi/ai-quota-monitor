package claude

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"
)

// Simulated endpoints; request and response shapes follow Claude Code's client, not a real account.
func TestLoginRefreshAndUsage(t *testing.T) {
	var tokens []map[string]string
	revoked := false
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			tokens = append(tokens, body)
			if revoked || body["code"] == "bad" {
				w.WriteHeader(400)
				return
			}
			n := len(tokens)
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "access-" + string(rune('0'+n)), "refresh_token": "refresh-" + string(rune('0'+n)), "expires_in": 3600})
		case "/usage":
			if r.Header.Get("anthropic-beta") != "oauth-2025-04-20" {
				t.Error("missing beta header")
			}
			// The token from the login is no longer accepted; only refreshed ones are.
			if r.Header.Get("Authorization") == "Bearer access-2" {
				w.WriteHeader(401)
				return
			}
			_, _ = w.Write([]byte(`{"five_hour":{"utilization":23.5,"resets_at":"2026-10-09T18:00:00.000000+00:00"},"seven_day":{"utilization":120,"resets_at":null},"seven_day_opus":null,"extra_usage":{"is_enabled":false}}`))
		}
	}))
	defer s.Close()
	m := New(t.TempDir())
	m.Client, m.TokenURL, m.UsageURL = s.Client(), s.URL+"/token", s.URL+"/usage"
	ctx := context.Background()

	if _, e := m.Read(ctx, "acct"); !errors.Is(e, ErrLogin) || m.Status("acct").Status != "auth_required" {
		t.Fatalf("expected login required, got %v", e)
	}
	if e := m.Complete(ctx, "acct", "code"); !errors.Is(e, ErrNoLogin) {
		t.Fatalf("expected no login in progress, got %v", e)
	}
	state, e := m.Start("acct")
	if e != nil || state.Status != "pending" {
		t.Fatal(e, state)
	}
	u, _ := url.Parse(state.VerificationURL)
	q := u.Query()
	if q.Get("client_id") != clientID || q.Get("scope") != "user:profile" || q.Get("code_challenge_method") != "S256" || q.Get("redirect_uri") != redirectURI || len(q.Get("code_challenge")) != 43 {
		t.Fatalf("unexpected authorize URL %s", state.VerificationURL)
	}
	if e = m.Complete(ctx, "acct", "code#other-state"); !errors.Is(e, ErrRejected) || len(tokens) != 0 {
		t.Fatalf("state mismatch must be rejected locally, got %v", e)
	}
	if e = m.Complete(ctx, "acct", "bad#"+q.Get("state")); !errors.Is(e, ErrRejected) {
		t.Fatalf("expected rejected code, got %v", e)
	}
	if e = m.Complete(ctx, "acct", " good#"+q.Get("state")+"\n"); e != nil {
		t.Fatal(e)
	}
	if got := tokens[len(tokens)-1]; got["code"] != "good" || got["code_verifier"] == "" || got["grant_type"] != "authorization_code" {
		t.Fatalf("unexpected exchange %v", got)
	}
	if id := <-m.Ready; id != "acct" || m.Status("acct").Status != "authenticated" {
		t.Fatal("login must signal readiness")
	}
	if info, _ := os.Stat(m.path("acct")); info == nil || info.Mode().Perm() != 0600 {
		t.Fatal("credentials must be private")
	}

	// The usage endpoint answers 401 once: the session is refreshed and the rotated tokens are stored.
	before := len(tokens)
	m.status["acct"] = ""
	report, e := m.Read(ctx, "acct")
	if e != nil {
		t.Fatal(e)
	}
	if got := tokens[len(tokens)-1]; len(tokens) != before+1 || got["grant_type"] != "refresh_token" || got["refresh_token"] == "" {
		t.Fatalf("unexpected refresh %v", got)
	}
	five, week := report.Quotas[0], report.Quotas[1]
	if len(report.Quotas) != 2 || five.ID != "five-hour" || *five.Used != 23.5 || !five.ResetsAt.Equal(time.Date(2026, 10, 9, 18, 0, 0, 0, time.UTC)) || week.ID != "seven-day" || *week.Used != 100 || week.ResetsAt != nil {
		t.Fatalf("unexpected quotas %+v", report.Quotas)
	}
	before = len(tokens)
	if _, e = m.Read(ctx, "acct"); e != nil || len(tokens) != before {
		t.Fatalf("a valid session must not refresh again: %v", e)
	}

	// An expired access token with a revoked refresh token needs a new login.
	b, _ := os.ReadFile(m.path("acct"))
	var c credentials
	_ = json.Unmarshal(b, &c)
	c.ExpiresAt = time.Now().Add(-time.Hour)
	_ = m.save("acct", c)
	revoked = true
	if _, e = m.Read(ctx, "acct"); !errors.Is(e, ErrLogin) || m.Status("acct").Status != "auth_required" {
		t.Fatalf("expected login required, got %v", e)
	}
}
