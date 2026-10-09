package httpapi

import (
	"ai-quota-monitor/server/internal/panelauth"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// panelEnabled reports whether the web panel login is configured.
func (a *API) panelEnabled() bool {
	return a.Username != "" && a.PasswordHash != "" && a.Sessions != nil
}
func bearer(r *http.Request) string {
	token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	return token
}

// panelOnly wraps the session endpoints: never cached, small bodies, off without a user.
func (a *API) panelOnly(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if !a.panelEnabled() {
			write(w, 503, map[string]string{"error": "panel login disabled; set AUTH_USERNAME and AUTH_PASSWORD_HASH"})
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
		next(w, r)
	})
}
func writeTokens(w http.ResponseWriter, t panelauth.Tokens) {
	write(w, 200, map[string]any{"tokenType": "Bearer", "accessToken": t.AccessToken, "expiresIn": t.ExpiresIn, "refreshToken": t.RefreshToken})
}
func (a *API) login(w http.ResponseWriter, r *http.Request) {
	// Failures are counted per client address and globally, so guessing is
	// slowed down even when requests come from many addresses or through a proxy.
	client, _, e := net.SplitHostPort(r.RemoteAddr)
	if e != nil {
		client = r.RemoteAddr
	}
	keys := []string{"ip:" + client, "all"}
	limiters := []*panelauth.Limiter{a.loginLimiter, a.globalLimiter}
	for i, key := range keys {
		if wait := limiters[i].Blocked(key); wait > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
			write(w, 429, map[string]string{"error": "too many failed logins; try again later"})
			return
		}
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if e := json.NewDecoder(r.Body).Decode(&body); e != nil {
		write(w, 400, map[string]string{"error": "expected JSON with username and password"})
		return
	}
	// The hash is always computed so the response time does not reveal the username.
	password := panelauth.VerifyPassword(a.PasswordHash, body.Password)
	user := subtle.ConstantTimeCompare([]byte(body.Username), []byte(a.Username)) == 1
	if !user || !password {
		for i, key := range keys {
			limiters[i].Fail(key)
		}
		slog.Warn("panel login failed", "client", client)
		write(w, 401, map[string]string{"error": "invalid username or password"})
		return
	}
	a.loginLimiter.Reset(keys[0])
	tokens, e := a.Sessions.Create()
	if e != nil {
		slog.Error("could not persist panel session")
		write(w, 500, map[string]string{"error": "could not create session"})
		return
	}
	writeTokens(w, tokens)
}
func (a *API) refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if e := json.NewDecoder(r.Body).Decode(&body); e != nil {
		write(w, 400, map[string]string{"error": "expected JSON with refreshToken"})
		return
	}
	tokens, e := a.Sessions.Refresh(body.RefreshToken)
	if errors.Is(e, panelauth.ErrInvalid) {
		write(w, 401, map[string]string{"error": "invalid or expired refresh token; log in again"})
		return
	}
	if e != nil {
		slog.Error("could not persist panel session")
		write(w, 500, map[string]string{"error": "could not refresh session"})
		return
	}
	writeTokens(w, tokens)
}

// logout revokes the session of the access token in Authorization or of the
// refresh token in the body, so it also works once the access token expired.
func (a *API) logout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	for _, token := range []string{bearer(r), body.RefreshToken} {
		if e := a.Sessions.Revoke(token); e != nil {
			slog.Error("could not persist panel logout")
			write(w, 500, map[string]string{"error": "could not close session"})
			return
		}
	}
	w.WriteHeader(204)
}

// Credential returns the provider credential of an account: the one stored
// from the panel wins over the environment variable named by credentialEnv.
func (a *API) Credential(id, envName string) string {
	if v, ok := a.Secrets.Get(id); ok {
		return v
	}
	if envName == "" {
		return ""
	}
	return os.Getenv(envName)
}

// accounts describes what the panel can configure for each account. It never
// returns credential values, only whether one is set and where it comes from.
func (a *API) accounts(w http.ResponseWriter, r *http.Request) {
	type credential struct {
		Configured bool   `json:"configured"`
		Origin     string `json:"origin"`
	}
	type account struct {
		ID         string      `json:"id"`
		Name       string      `json:"name"`
		Provider   string      `json:"provider"`
		Scope      string      `json:"scope"`
		Source     string      `json:"source"`
		Login      string      `json:"login"`
		Credential *credential `json:"credential"`
	}
	out := []account{}
	for _, c := range a.Accounts {
		item := account{ID: c.ID, Name: c.Name, Provider: c.Provider, Scope: c.Scope, Source: c.Source}
		if c.Source == "direct" {
			switch {
			case c.Provider == "openai" && a.CodexAuth != nil:
				item.Login = "codex"
			case c.Provider == "anthropic" && a.ClaudeAuth != nil:
				item.Login = "claude"
			case c.CredentialEnv != "":
				item.Credential = &credential{}
				if _, ok := a.Secrets.Get(c.ID); ok {
					*item.Credential = credential{true, "stored"}
				} else if os.Getenv(c.CredentialEnv) != "" {
					*item.Credential = credential{true, "env"}
				}
			}
		}
		out = append(out, item)
	}
	write(w, 200, map[string]any{"credentialStorage": a.Secrets.Enabled(), "accounts": out})
}

// credential stores or replaces the provider credential of an account.
func (a *API) credential(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	found := false
	for _, c := range a.Accounts {
		if c.ID == id && c.Source == "direct" && c.CredentialEnv != "" {
			found = true
		}
	}
	if !found {
		write(w, 404, map[string]string{"error": "account does not use a credential"})
		return
	}
	if !a.Secrets.Enabled() {
		write(w, 503, map[string]string{"error": "credential storage disabled; set MASTER_KEY"})
		return
	}
	var body struct {
		Value string `json:"value"`
	}
	if e := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&body); e != nil {
		write(w, 400, map[string]string{"error": "expected JSON with value"})
		return
	}
	value := strings.TrimSpace(body.Value)
	if value == "" || len(value) > 8192 || strings.ContainsFunc(value, func(c rune) bool { return c < 0x20 || c == 0x7f }) {
		write(w, 400, map[string]string{"error": "value must be a single line of 1 to 8192 characters"})
		return
	}
	if e := a.Secrets.Set(id, value); e != nil {
		slog.Error("could not store credential", "account", id)
		write(w, 500, map[string]string{"error": "could not store credential"})
		return
	}
	slog.Info("credential replaced from panel", "account", id)
	if a.OnCredential != nil {
		a.OnCredential()
	}
	write(w, 200, map[string]any{"configured": true, "origin": "stored"})
}
