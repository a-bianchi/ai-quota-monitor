package httpapi

import (
	"ai-quota-monitor/server/internal/config"
	"ai-quota-monitor/server/internal/panelauth"
	"ai-quota-monitor/server/internal/secrets"
	"ai-quota-monitor/server/internal/storage"
	"ai-quota-monitor/server/internal/web"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// PBKDF2 hash of "test-password-123" with few rounds, generated outside this code base, to keep tests fast.
const testHash = "pbkdf2-sha256:1000:MDEyMzQ1Njc4OWFiY2RlZg:aVRVJRKXS5I15VLc5oQIrZrtje7zLP2Y-jhdb79Cwkk"

type panelTest struct {
	t   *testing.T
	api *API
	h   http.Handler
	dir string
}

func newPanel(t *testing.T) *panelTest {
	dir := t.TempDir()
	store, e := storage.New(dir)
	if e != nil {
		t.Fatal(e)
	}
	sessions, e := panelauth.Open(dir)
	if e != nil {
		t.Fatal(e)
	}
	vault, e := secrets.Open(dir, strings.Repeat("m", 32))
	if e != nil {
		t.Fatal(e)
	}
	api := &API{
		Accounts: []config.Account{
			{ID: "openrouter-personal", Name: "OpenRouter", Provider: "openrouter", Scope: "personal", Source: "direct", CredentialEnv: "TEST_OPENROUTER_KEY"},
			{ID: "codex-work", Name: "Codex", Provider: "openai", Scope: "work", Source: "direct"},
			{ID: "collected", Name: "Collected", Provider: "kiro", Scope: "work", Source: "collector"},
		},
		Store: store, DeviceToken: "device", CollectorToken: "collector",
		Username: "admin", PasswordHash: testHash, Sessions: sessions, Secrets: vault, Web: web.Handler(),
	}
	return &panelTest{t: t, api: api, h: api.Handler(), dir: dir}
}
func (p *panelTest) call(method, path, bearer, body string) (*httptest.ResponseRecorder, map[string]any) {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if bearer != "" {
		r.Header.Set("Authorization", "Bearer "+bearer)
	}
	w := httptest.NewRecorder()
	p.h.ServeHTTP(w, r)
	out := map[string]any{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w, out
}
func (p *panelTest) want(code int, method, path, bearer, body string) map[string]any {
	p.t.Helper()
	w, out := p.call(method, path, bearer, body)
	if w.Code != code {
		p.t.Fatalf("%s %s got %d, want %d: %s", method, path, w.Code, code, w.Body.String())
	}
	return out
}
func (p *panelTest) login() (access, refresh string) {
	p.t.Helper()
	out := p.want(200, "POST", "/auth/login", "", `{"username":"admin","password":"test-password-123"}`)
	access, _ = out["accessToken"].(string)
	refresh, _ = out["refreshToken"].(string)
	if access == "" || refresh == "" || access == refresh || out["tokenType"] != "Bearer" || out["expiresIn"] != float64(900) {
		p.t.Fatalf("unexpected login response %v", out)
	}
	return access, refresh
}

func TestPanelLoginAndRouteProtection(t *testing.T) {
	p := newPanel(t)
	p.want(401, "POST", "/auth/login", "", `{"username":"admin","password":"wrong"}`)
	p.want(401, "POST", "/auth/login", "", `{"username":"other","password":"test-password-123"}`)
	p.want(400, "POST", "/auth/login", "", `not json`)
	access, refresh := p.login()

	// The panel session reaches the panel routes and the provider logins.
	p.want(200, "GET", "/usage", access, "")
	p.want(200, "GET", "/accounts", access, "")
	p.want(404, "GET", "/accounts/codex-work/codex/login", access, "") // authorized; no Codex manager in this test
	// It is not a collector: ingestion stays with COLLECTOR_TOKEN.
	p.want(401, "PUT", "/accounts/collected/usage", access, `{}`)
	// Panel-only routes reject everything else, including the static tokens.
	for _, token := range []string{"", "device", "collector", refresh, access + "x"} {
		p.want(401, "GET", "/accounts", token, "")
		p.want(401, "PUT", "/accounts/openrouter-personal/credential", token, `{"value":"k"}`)
	}
	p.want(401, "GET", "/usage", refresh, "")
	// Static tokens keep their previous permissions.
	p.want(200, "GET", "/usage", "device", "")
	p.want(200, "GET", "/usage", "collector", "")
	p.want(401, "GET", "/accounts/codex-work/codex/login", "device", "")
	p.want(404, "GET", "/accounts/codex-work/codex/login", "collector", "")
}

func TestPanelRefreshRotationAndLogout(t *testing.T) {
	p := newPanel(t)
	access, refresh := p.login()
	p.want(401, "POST", "/auth/refresh", "", `{"refreshToken":"`+access+`"}`)
	p.want(400, "POST", "/auth/refresh", "", `nope`)
	out := p.want(200, "POST", "/auth/refresh", "", `{"refreshToken":"`+refresh+`"}`)
	access2, _ := out["accessToken"].(string)
	refresh2, _ := out["refreshToken"].(string)
	if access2 == "" || access2 == access || refresh2 == refresh {
		t.Fatalf("tokens must rotate: %v", out)
	}
	p.want(401, "GET", "/usage", access, "")
	p.want(200, "GET", "/usage", access2, "")

	// Sessions survive a restart: a new server over the same data directory accepts the token.
	sessions, e := panelauth.Open(p.dir)
	if e != nil {
		t.Fatal(e)
	}
	restarted := &panelTest{t: t, api: &API{Username: "admin", PasswordHash: testHash, Sessions: sessions}, dir: p.dir}
	restarted.h = restarted.api.Handler()
	restarted.want(200, "GET", "/accounts", access2, "")

	// Logout revokes the session on the server.
	if w, _ := p.call("POST", "/auth/logout", access2, ""); w.Code != 204 {
		t.Fatalf("logout got %d", w.Code)
	}
	p.want(401, "GET", "/usage", access2, "")
	p.want(401, "POST", "/auth/refresh", "", `{"refreshToken":"`+refresh2+`"}`)

	// Logout also works with only the refresh token, once the access token expired.
	access3, refresh3 := p.login()
	if w, _ := p.call("POST", "/auth/logout", "", `{"refreshToken":"`+refresh3+`"}`); w.Code != 204 {
		t.Fatalf("logout got %d", w.Code)
	}
	p.want(401, "GET", "/usage", access3, "")

	// Reusing a rotated refresh token revokes the session it belonged to.
	access4, refresh4 := p.login()
	out = p.want(200, "POST", "/auth/refresh", "", `{"refreshToken":"`+refresh4+`"}`)
	access5, _ := out["accessToken"].(string)
	p.want(401, "POST", "/auth/refresh", "", `{"refreshToken":"`+refresh4+`"}`)
	p.want(401, "GET", "/usage", access4, "")
	p.want(401, "GET", "/usage", access5, "")
}

func TestPanelLoginRateLimit(t *testing.T) {
	p := newPanel(t)
	for i := 0; i < 5; i++ {
		p.want(401, "POST", "/auth/login", "", `{"username":"admin","password":"wrong"}`)
	}
	w, _ := p.call("POST", "/auth/login", "", `{"username":"admin","password":"test-password-123"}`)
	if w.Code != 429 || w.Header().Get("Retry-After") == "" {
		t.Fatalf("expected 429 with Retry-After, got %d", w.Code)
	}
}

func TestPanelDisabledKeepsStaticTokens(t *testing.T) {
	p := newPanel(t)
	p.api.Username, p.api.PasswordHash = "", ""
	p.want(503, "POST", "/auth/login", "", `{"username":"","password":""}`)
	p.want(503, "POST", "/auth/refresh", "", `{"refreshToken":"x"}`)
	p.want(503, "GET", "/accounts", "collector", "")
	p.want(200, "GET", "/usage", "device", "")
	p.want(404, "GET", "/accounts/codex-work/codex/login", "collector", "")
}

func TestPanelCredentials(t *testing.T) {
	p := newPanel(t)
	access, _ := p.login()
	changed := 0
	p.api.OnCredential = func() { changed++ }
	origin := func() any {
		t.Helper()
		w, out := p.call("GET", "/accounts", access, "")
		if strings.Contains(w.Body.String(), "sk-") {
			t.Fatalf("credential leaked: %s", w.Body.String())
		}
		if out["credentialStorage"] != true {
			t.Fatalf("storage must be enabled: %v", out)
		}
		first := out["accounts"].([]any)[0].(map[string]any)
		codex := out["accounts"].([]any)[1].(map[string]any)
		if codex["credential"] != nil || out["accounts"].([]any)[2].(map[string]any)["credential"] != nil {
			t.Fatal("only credential accounts expose a credential state")
		}
		c := first["credential"].(map[string]any)
		if c["configured"] != (c["origin"] != "") {
			t.Fatalf("inconsistent state %v", c)
		}
		return c["origin"]
	}
	if origin() != "" || p.api.Credential("openrouter-personal", "TEST_OPENROUTER_KEY") != "" {
		t.Fatal("expected no credential")
	}
	// Environment variable as fallback.
	t.Setenv("TEST_OPENROUTER_KEY", "sk-from-env")
	if origin() != "env" || p.api.Credential("openrouter-personal", "TEST_OPENROUTER_KEY") != "sk-from-env" {
		t.Fatal("expected the environment credential")
	}
	for _, body := range []string{`{}`, `{"value":"   "}`, `{"value":"two\nlines"}`, `nope`} {
		p.want(400, "PUT", "/accounts/openrouter-personal/credential", access, body)
	}
	p.want(404, "PUT", "/accounts/codex-work/credential", access, `{"value":"sk-x"}`)
	p.want(404, "PUT", "/accounts/missing/credential", access, `{"value":"sk-x"}`)
	w, _ := p.call("PUT", "/accounts/openrouter-personal/credential", access, `{"value":" sk-from-panel "}`)
	if w.Code != 200 || strings.Contains(w.Body.String(), "sk-") || changed != 1 {
		t.Fatalf("store got %d %s changed=%d", w.Code, w.Body.String(), changed)
	}
	// The stored credential wins over the environment variable and survives a restart.
	if origin() != "stored" || p.api.Credential("openrouter-personal", "TEST_OPENROUTER_KEY") != "sk-from-panel" {
		t.Fatal("stored credential must have priority")
	}
	vault, _ := secrets.Open(p.dir, strings.Repeat("m", 32))
	if v, _ := vault.Get("openrouter-personal"); v != "sk-from-panel" {
		t.Fatal("credential not persisted")
	}
	// Without MASTER_KEY nothing can be stored and the fallback keeps working.
	p.api.Secrets = nil
	p.want(503, "PUT", "/accounts/openrouter-personal/credential", access, `{"value":"sk-x"}`)
	if _, out := p.call("GET", "/accounts", access, ""); out["credentialStorage"] != false {
		t.Fatal("storage must be reported as disabled")
	}
	if p.api.Credential("openrouter-personal", "TEST_OPENROUTER_KEY") != "sk-from-env" {
		t.Fatal("fallback must keep working")
	}
}

func TestPanelStaticFiles(t *testing.T) {
	p := newPanel(t)
	for path, contentType := range map[string]string{"/": "text/html", "/static/app.js": "javascript", "/static/app.css": "text/css", "/static/i18n/es.json": "application/json", "/static/i18n/en.json": "application/json"} {
		w, _ := p.call("GET", path, "", "")
		csp := w.Header().Get("Content-Security-Policy")
		if w.Code != 200 || !strings.Contains(w.Header().Get("Content-Type"), contentType) || !strings.Contains(csp, "script-src 'self'") || strings.Contains(csp, "unsafe") || w.Header().Get("X-Content-Type-Options") != "nosniff" {
			t.Fatalf("%s got %d %q csp=%q", path, w.Code, w.Header().Get("Content-Type"), csp)
		}
	}
	w, _ := p.call("GET", "/", "", "")
	page := w.Body.String()
	// No inline code and nothing loaded from other hosts.
	for _, banned := range []string{"<script>", "<style", " style=", "onclick=", "http://", "https://", "x-html"} {
		if strings.Contains(page, banned) {
			t.Fatalf("index.html must not contain %q", banned)
		}
	}
	p.want(404, "GET", "/static/missing.js", "", "")
	p.want(404, "GET", "/dashboard", "", "")
}
