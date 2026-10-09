// Package claude reads the limits of a Claude subscription with an OAuth
// session kept on the server. It uses the public OAuth client of Claude Code
// and the usage endpoint behind its /usage screen; neither is a documented API.
package claude

import (
	"ai-quota-monitor/server/internal/usage"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	ErrLogin    = errors.New("Claude requires login")
	ErrNoLogin  = errors.New("no Claude login in progress")
	ErrRejected = errors.New("Claude rejected the authorization code")
)

const (
	clientID    = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	redirectURI = "https://platform.claude.com/oauth/code/callback"
	// Enough to read usage; the session cannot be used for inference.
	scope = "user:profile"
)

type State struct {
	Status          string     `json:"status"`
	VerificationURL string     `json:"verificationUrl,omitempty"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`
}
type pending struct {
	verifier, state, url string
	expires              time.Time
}
type credentials struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresAt    time.Time `json:"expiresAt"`
}
type Manager struct {
	Client                           *http.Client
	AuthorizeURL, TokenURL, UsageURL string
	Ready                            chan string
	root                             string
	mu                               sync.Mutex
	logins                           map[string]pending
	status                           map[string]string
	// session serializes token use: refresh tokens rotate on every refresh.
	session sync.Mutex
}

func New(root string) *Manager {
	return &Manager{
		Client:       &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
		AuthorizeURL: "https://claude.com/cai/oauth/authorize",
		TokenURL:     "https://platform.claude.com/v1/oauth/token",
		UsageURL:     "https://api.anthropic.com/api/oauth/usage",
		Ready:        make(chan string, 64),
		root:         root,
		logins:       map[string]pending{},
		status:       map[string]string{},
	}
}
func random() (string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func (m *Manager) Status(id string) State {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.logins[id]; ok && time.Now().Before(p.expires) {
		return State{Status: "pending", VerificationURL: p.url, ExpiresAt: &p.expires}
	}
	if s := m.status[id]; s != "" {
		return State{Status: s}
	}
	return State{Status: "unknown"}
}

// Start prepares a PKCE authorization. The user opens the URL, signs in and
// gets a code to hand back through Complete.
func (m *Manager) Start(id string) (State, error) {
	verifier, e := random()
	if e != nil {
		return State{}, e
	}
	state, e := random()
	if e != nil {
		return State{}, e
	}
	challenge := sha256.Sum256([]byte(verifier))
	q := url.Values{
		"code":                  {"true"},
		"client_id":             {clientID},
		"response_type":         {"code"},
		"redirect_uri":          {redirectURI},
		"scope":                 {scope},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(challenge[:])},
		"code_challenge_method": {"S256"},
		"state":                 {state},
	}
	p := pending{verifier: verifier, state: state, url: m.AuthorizeURL + "?" + q.Encode(), expires: time.Now().UTC().Add(15 * time.Minute)}
	m.mu.Lock()
	m.logins[id] = p
	m.mu.Unlock()
	return m.Status(id), nil
}

// Complete exchanges the code shown after authorizing, formatted "code#state".
func (m *Manager) Complete(ctx context.Context, id, code string) error {
	m.mu.Lock()
	p, ok := m.logins[id]
	m.mu.Unlock()
	if !ok || !time.Now().Before(p.expires) {
		return ErrNoLogin
	}
	code, state, _ := strings.Cut(strings.TrimSpace(code), "#")
	if code == "" || (state != "" && state != p.state) {
		return ErrRejected
	}
	m.session.Lock()
	defer m.session.Unlock()
	c, e := m.token(ctx, map[string]string{"grant_type": "authorization_code", "code": code, "redirect_uri": redirectURI, "client_id": clientID, "code_verifier": p.verifier, "state": p.state}, credentials{})
	if errors.Is(e, ErrLogin) {
		return ErrRejected
	}
	if e != nil {
		return e
	}
	if e = m.save(id, c); e != nil {
		return e
	}
	m.mu.Lock()
	delete(m.logins, id)
	m.status[id] = "authenticated"
	m.mu.Unlock()
	select {
	case m.Ready <- id:
	default:
	}
	return nil
}
func (m *Manager) token(ctx context.Context, body map[string]string, old credentials) (credentials, error) {
	payload, _ := json.Marshal(body)
	req, e := http.NewRequestWithContext(ctx, http.MethodPost, m.TokenURL, bytes.NewReader(payload))
	if e != nil {
		return old, e
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "ai-quota-monitor")
	res, e := m.Client.Do(req)
	if e != nil {
		return old, errors.New("Claude token request failed")
	}
	defer res.Body.Close()
	if res.StatusCode == 400 || res.StatusCode == 401 || res.StatusCode == 403 {
		return old, ErrLogin
	}
	if res.StatusCode != 200 {
		return old, fmt.Errorf("Claude token endpoint returned HTTP %d", res.StatusCode)
	}
	var out struct {
		AccessToken  string  `json:"access_token"`
		RefreshToken string  `json:"refresh_token"`
		ExpiresIn    float64 `json:"expires_in"`
	}
	if e = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&out); e != nil || out.AccessToken == "" {
		return old, errors.New("invalid Claude token response")
	}
	if out.RefreshToken == "" {
		out.RefreshToken = old.RefreshToken
	}
	return credentials{AccessToken: out.AccessToken, RefreshToken: out.RefreshToken, ExpiresAt: time.Now().UTC().Add(time.Duration(out.ExpiresIn) * time.Second)}, nil
}
func (m *Manager) path(id string) string { return filepath.Join(m.root, id, "credentials.json") }
func (m *Manager) save(id string, c credentials) error {
	if e := os.MkdirAll(filepath.Dir(m.path(id)), 0700); e != nil {
		return e
	}
	b, _ := json.Marshal(c)
	tmp := m.path(id) + ".tmp"
	if e := os.WriteFile(tmp, b, 0600); e != nil {
		return e
	}
	return os.Rename(tmp, m.path(id))
}
func (m *Manager) refresh(ctx context.Context, id string, c credentials) (credentials, error) {
	if c.RefreshToken == "" {
		return c, ErrLogin
	}
	c, e := m.token(ctx, map[string]string{"grant_type": "refresh_token", "refresh_token": c.RefreshToken, "client_id": clientID, "scope": scope}, c)
	if e != nil {
		return c, e
	}
	return c, m.save(id, c)
}

// Read returns the subscription windows. It sends no prompts.
func (m *Manager) Read(ctx context.Context, id string) (usage.Report, error) {
	m.session.Lock()
	report, e := m.read(ctx, id)
	m.session.Unlock()
	m.mu.Lock()
	if e == nil {
		m.status[id] = "authenticated"
	} else if errors.Is(e, ErrLogin) {
		m.status[id] = "auth_required"
	}
	m.mu.Unlock()
	return report, e
}
func (m *Manager) read(ctx context.Context, id string) (usage.Report, error) {
	var c credentials
	b, e := os.ReadFile(m.path(id))
	if e != nil || json.Unmarshal(b, &c) != nil {
		return usage.Report{}, ErrLogin
	}
	if c.AccessToken == "" || time.Until(c.ExpiresAt) < time.Minute {
		if c, e = m.refresh(ctx, id, c); e != nil {
			return usage.Report{}, e
		}
	}
	res, e := m.get(ctx, c)
	if e == nil && res.StatusCode == 401 {
		res.Body.Close()
		if c, e = m.refresh(ctx, id, c); e != nil {
			return usage.Report{}, e
		}
		res, e = m.get(ctx, c)
	}
	if e != nil {
		return usage.Report{}, errors.New("Claude usage request failed")
	}
	defer res.Body.Close()
	if res.StatusCode == 401 || res.StatusCode == 403 {
		return usage.Report{}, ErrLogin
	}
	if res.StatusCode != 200 {
		return usage.Report{}, fmt.Errorf("Claude usage returned HTTP %d", res.StatusCode)
	}
	type window struct {
		Utilization *float64 `json:"utilization"`
		ResetsAt    *string  `json:"resets_at"`
	}
	var body struct {
		FiveHour       *window `json:"five_hour"`
		SevenDay       *window `json:"seven_day"`
		SevenDayOpus   *window `json:"seven_day_opus"`
		SevenDaySonnet *window `json:"seven_day_sonnet"`
	}
	if e = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&body); e != nil {
		return usage.Report{}, errors.New("invalid Claude usage response")
	}
	now := time.Now().UTC()
	report := usage.Report{UpdatedAt: now}
	for _, w := range []struct {
		id, name string
		w        *window
	}{{"five-hour", "5 horas", body.FiveHour}, {"seven-day", "7 días", body.SevenDay}, {"seven-day-opus", "7 días Opus", body.SevenDayOpus}, {"seven-day-sonnet", "7 días Sonnet", body.SevenDaySonnet}} {
		if w.w == nil || w.w.Utilization == nil || math.IsNaN(*w.w.Utilization) {
			continue
		}
		q := usage.Quota{ID: w.id, Name: w.name, Unit: "percent", Used: usage.Number(math.Max(0, math.Min(*w.w.Utilization, 100))), Limit: usage.Number(100)}
		if w.w.ResetsAt != nil {
			if t, e := time.Parse(time.RFC3339, *w.w.ResetsAt); e == nil {
				q.ResetsAt = &t
			}
		}
		report.Quotas = append(report.Quotas, q)
	}
	if e = usage.ValidateReport(report, now); e != nil {
		return usage.Report{}, errors.New("Claude usage response has no known windows")
	}
	return report, nil
}
func (m *Manager) get(ctx context.Context, c credentials) (*http.Response, error) {
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, m.UsageURL, nil)
	if e != nil {
		return nil, e
	}
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	req.Header.Set("anthropic-beta", "oauth-2025-04-20")
	req.Header.Set("User-Agent", "ai-quota-monitor")
	return m.Client.Do(req)
}
