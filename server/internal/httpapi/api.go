package httpapi

import (
	"ai-quota-monitor/server/internal/config"
	"ai-quota-monitor/server/internal/providers/claude"
	"ai-quota-monitor/server/internal/providers/codex"
	"ai-quota-monitor/server/internal/storage"
	"ai-quota-monitor/server/internal/usage"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

type API struct {
	CodexAuth      *codex.Manager
	ClaudeAuth     *claude.Manager
	Accounts       []config.Account
	Store          *storage.Store
	DeviceToken    string
	CollectorToken string
	StaleAfter     time.Duration
	// Username and Password enable POST /auth/login; both empty disables it.
	Username, Password string
	LoginFailureDelay  time.Duration
	mu                 sync.RWMutex
	failures           map[string]string
}

func (a *API) Failure(id, status string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.failures == nil {
		a.failures = map[string]string{}
	}
	if status == "" {
		delete(a.failures, id)
	} else {
		a.failures[id] = status
	}
}
func (a *API) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	m.Handle("POST /auth/login", a.session(a.login))
	m.Handle("POST /auth/refresh", a.session(a.refresh))
	m.Handle("GET /usage", a.auth(http.HandlerFunc(a.list), a.DeviceToken, a.CollectorToken))
	m.Handle("PUT /accounts/{id}/usage", a.auth(http.HandlerFunc(a.ingest), a.CollectorToken))
	m.Handle("POST /accounts/{id}/codex/login", a.auth(http.HandlerFunc(a.codexLogin), a.CollectorToken))
	m.Handle("GET /accounts/{id}/codex/login", a.auth(http.HandlerFunc(a.codexStatus), a.CollectorToken))
	m.Handle("POST /accounts/{id}/claude/login", a.auth(http.HandlerFunc(a.claudeLogin), a.CollectorToken))
	m.Handle("PUT /accounts/{id}/claude/login", a.auth(http.HandlerFunc(a.claudeCode), a.CollectorToken))
	m.Handle("GET /accounts/{id}/claude/login", a.auth(http.HandlerFunc(a.claudeStatus), a.CollectorToken))
	return m
}

// auth accepts any of the given tokens, or an access token from /auth/login.
// The first token is required for the endpoint to be enabled.
func (a *API) auth(next http.Handler, tokens ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if tokens[0] == "" {
			write(w, 503, map[string]string{"error": "endpoint disabled"})
			return
		}
		got := []byte(r.Header.Get("Authorization"))
		ok := 0
		if bearer, found := strings.CutPrefix(string(got), "Bearer "); found && a.validToken(bearer, "access") {
			ok = 1
		}
		for _, token := range tokens {
			if token != "" {
				ok |= subtle.ConstantTimeCompare(got, []byte("Bearer "+token))
			}
		}
		if ok != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			write(w, 401, map[string]string{"error": "unauthorized"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func (a *API) list(w http.ResponseWriter, r *http.Request) {
	scope := r.URL.Query().Get("scope")
	if scope != "" && scope != "personal" && scope != "work" {
		write(w, 400, map[string]string{"error": "invalid scope"})
		return
	}
	zone := r.Header.Get("X-Timezone")
	if zone == "" {
		zone = defaultZone
	}
	loc, e := time.LoadLocation(zone)
	if e != nil {
		write(w, 400, map[string]string{"error": "invalid X-Timezone; use an IANA name such as America/Argentina/Buenos_Aires or UTC"})
		return
	}
	lang := language(r.Header.Get("Accept-Language"))
	now := time.Now().UTC()
	out := []usage.Subscription{}
	for _, c := range a.Accounts {
		if scope != "" && scope != c.Scope {
			continue
		}
		s := usage.Subscription{ID: c.ID, Name: c.Name, Provider: c.Provider, Scope: c.Scope, Source: c.Source, Status: "unavailable", Quotas: []usage.Quota{}}
		if report, ok := a.Store.Get(c.Source + ":" + c.ID); ok {
			updated := report.UpdatedAt.In(loc)
			s.UpdatedAt = &updated
			s.Quotas = localize(report.Quotas, lang, loc)
			s.Status = "ok"
			if now.Sub(report.UpdatedAt) > a.StaleAfter {
				s.Status = "stale"
			}
			for _, q := range report.Quotas {
				if q.ResetsAt != nil && !q.ResetsAt.After(now) {
					s.Status = "stale"
				}
			}
		}
		a.mu.RLock()
		failure := a.failures[c.ID]
		a.mu.RUnlock()
		if failure != "" {
			s.Status = failure
		}
		out = append(out, s)
	}
	write(w, 200, map[string]any{"schemaVersion": 1, "generatedAt": now.In(loc), "accounts": out})
}
func (a *API) ingest(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	found := false
	for _, c := range a.Accounts {
		if c.ID == id && c.Source == "collector" {
			found = true
			break
		}
	}
	if !found {
		write(w, 404, map[string]string{"error": "collector account not found"})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	var report usage.Report
	if e := d.Decode(&report); e != nil {
		write(w, 400, map[string]string{"error": "invalid JSON report"})
		return
	}
	if e := d.Decode(new(any)); e != io.EOF {
		write(w, 400, map[string]string{"error": "expected one JSON object"})
		return
	}
	if e := usage.ValidateReport(report, time.Now()); e != nil {
		write(w, 400, map[string]string{"error": e.Error()})
		return
	}
	if old, ok := a.Store.Get("collector:" + id); ok && report.UpdatedAt.Before(old.UpdatedAt) {
		write(w, 409, map[string]string{"error": "observation is older than stored data"})
		return
	}
	if e := a.Store.Put("collector:"+id, report); e != nil {
		slog.Error("persist report failed", "account", id)
		write(w, 500, map[string]string{"error": "could not persist report"})
		return
	}
	a.Failure(id, "")
	write(w, 200, map[string]string{"status": "ok"})
}

func (a *API) codexAccount(id string) bool {
	for _, c := range a.Accounts {
		if c.ID == id && c.Provider == "openai" && c.Source == "direct" {
			return true
		}
	}
	return false
}
func (a *API) codexLogin(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if a.CodexAuth == nil || !a.codexAccount(id) {
		write(w, 404, map[string]string{"error": "direct Codex account not found"})
		return
	}
	state, err := a.CodexAuth.Start(r.Context(), id)
	if err != nil {
		write(w, 409, map[string]string{"error": "account busy or login interrupted"})
		return
	}
	if state.Status == "failed" {
		write(w, 502, map[string]string{"error": "could not start device login; check Codex CLI and device authorization settings"})
		return
	}
	write(w, 202, state)
}
func (a *API) codexStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if a.CodexAuth == nil || !a.codexAccount(id) {
		write(w, 404, map[string]string{"error": "direct Codex account not found"})
		return
	}
	write(w, 200, a.CodexAuth.Status(id))
}

func (a *API) claudeAccount(id string) bool {
	if a.ClaudeAuth == nil {
		return false
	}
	for _, c := range a.Accounts {
		if c.ID == id && c.Provider == "anthropic" && c.Source == "direct" {
			return true
		}
	}
	return false
}
func (a *API) claudeLogin(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !a.claudeAccount(id) {
		write(w, 404, map[string]string{"error": "direct Claude account not found"})
		return
	}
	state, err := a.ClaudeAuth.Start(id)
	if err != nil {
		write(w, 500, map[string]string{"error": "could not start login"})
		return
	}
	write(w, 202, state)
}
func (a *API) claudeCode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !a.claudeAccount(id) {
		write(w, 404, map[string]string{"error": "direct Claude account not found"})
		return
	}
	var body struct {
		Code string `json:"code"`
	}
	if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&body); e != nil || body.Code == "" {
		write(w, 400, map[string]string{"error": "expected JSON with the authorization code"})
		return
	}
	switch err := a.ClaudeAuth.Complete(r.Context(), id, body.Code); {
	case err == nil:
		write(w, 200, a.ClaudeAuth.Status(id))
	case errors.Is(err, claude.ErrNoLogin):
		write(w, 409, map[string]string{"error": "no login in progress or it expired; start again with POST"})
	case errors.Is(err, claude.ErrRejected):
		write(w, 400, map[string]string{"error": "authorization code rejected; start again with POST"})
	default:
		slog.Warn("Claude login failed", "account", id)
		write(w, 502, map[string]string{"error": "could not complete login with Claude"})
	}
}
func (a *API) claudeStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if !a.claudeAccount(id) {
		write(w, 404, map[string]string{"error": "direct Claude account not found"})
		return
	}
	write(w, 200, a.ClaudeAuth.Status(id))
}
