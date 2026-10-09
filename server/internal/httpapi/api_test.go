package httpapi

import (
	"ai-quota-monitor/server/internal/config"
	"ai-quota-monitor/server/internal/storage"
	"ai-quota-monitor/server/internal/usage"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAuthIngestionAndStaleness(t *testing.T) {
	store, e := storage.New(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	api := &API{Accounts: []config.Account{{ID: "work", Name: "Codex", Provider: "openai", Scope: "work", Source: "collector"}, {ID: "demo", Name: "Demo", Provider: "openai", Scope: "personal", Source: "mock"}}, Store: store, DeviceToken: "device", CollectorToken: "collector", StaleAfter: 15 * time.Minute}
	h := api.Handler()
	call := func(method, path, token, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w := call("GET", "/usage", "", ""); w.Code != 401 {
		t.Fatalf("unauthorized=%d", w.Code)
	}
	report := usage.Report{UpdatedAt: time.Now().UTC().Add(-time.Hour), Quotas: []usage.Quota{{ID: "weekly", Name: "Semanal", Unit: "percent", Used: usage.Number(6), Limit: usage.Number(100)}}}
	body, _ := json.Marshal(report)
	if w := call("PUT", "/accounts/work/usage", "device", string(body)); w.Code != 401 {
		t.Fatal("device must not ingest")
	}
	if w := call("PUT", "/accounts/work/usage", "collector", string(body)); w.Code != 200 {
		t.Fatalf("ingest %d %s", w.Code, w.Body.String())
	}
	if w := call("PUT", "/accounts/demo/usage", "collector", string(body)); w.Code != 404 {
		t.Fatal("must not overwrite mock accounts")
	}
	if w := call("GET", "/usage", "collector", ""); w.Code != 200 {
		t.Fatalf("collector must read usage, got %d", w.Code)
	}
	w := call("GET", "/usage?scope=work", "device", "")
	var response struct {
		Accounts []usage.Subscription `json:"accounts"`
	}
	if e = json.Unmarshal(w.Body.Bytes(), &response); e != nil {
		t.Fatal(e)
	}
	if len(response.Accounts) != 1 || response.Accounts[0].Status != "stale" || *response.Accounts[0].Quotas[0].Used != 6 {
		t.Fatalf("unexpected response %s", w.Body.String())
	}
	report.UpdatedAt = report.UpdatedAt.Add(-time.Hour)
	body, _ = json.Marshal(report)
	if w := call("PUT", "/accounts/work/usage", "collector", string(body)); w.Code != 409 {
		t.Fatal("must reject older observations")
	}
	if w := call("GET", "/usage?scope=invalid", "device", ""); w.Code != 400 {
		t.Fatal("invalid scope accepted")
	}
	report.UpdatedAt = time.Now().Add(time.Hour)
	body, _ = json.Marshal(report)
	if w := call("PUT", "/accounts/work/usage", "collector", string(body)); w.Code != 400 {
		t.Fatal("future timestamp accepted")
	}
	report.UpdatedAt = time.Now()
	report.Quotas[0].Used = usage.Number(120)
	body, _ = json.Marshal(report)
	if w := call("PUT", "/accounts/work/usage", "collector", string(body)); w.Code != 400 {
		t.Fatal("invalid percentage accepted")
	}
}

func TestEmptyCollectorTokenDoesNotReadUsage(t *testing.T) {
	h := (&API{DeviceToken: "device"}).Handler()
	for _, header := range []string{"Bearer ", "Bearer"} {
		r := httptest.NewRequest("GET", "/usage", nil)
		r.Header.Set("Authorization", header)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != 401 {
			t.Fatalf("header=%q got=%d", header, w.Code)
		}
	}
}

func TestTimezoneAndLanguage(t *testing.T) {
	store, e := storage.New(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	reset := time.Date(2030, 1, 2, 3, 0, 0, 0, time.UTC)
	names := []string{"Semanal", "5 horas", "1 horas", "7 días Opus", "GPT · Semanal", "Saldo prepago", "Nombre propio"}
	report := usage.Report{UpdatedAt: time.Now().UTC()}
	for _, n := range names {
		report.Quotas = append(report.Quotas, usage.Quota{ID: n, Name: n, Unit: "percent", Used: usage.Number(1), Limit: usage.Number(100), ResetsAt: &reset})
	}
	if e = store.Put("direct:a", report); e != nil {
		t.Fatal(e)
	}
	h := (&API{Accounts: []config.Account{{ID: "a", Name: "A", Provider: "openai", Scope: "work", Source: "direct"}}, Store: store, DeviceToken: "device", StaleAfter: time.Hour}).Handler()
	get := func(zone, lang string) (int, string, []usage.Quota) {
		r := httptest.NewRequest("GET", "/usage", nil)
		r.Header.Set("Authorization", "Bearer device")
		if zone != "" {
			r.Header.Set("X-Timezone", zone)
		}
		if lang != "" {
			r.Header.Set("Accept-Language", lang)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		var body struct {
			Accounts []usage.Subscription `json:"accounts"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &body)
		if len(body.Accounts) == 0 {
			return w.Code, w.Body.String(), nil
		}
		return w.Code, w.Body.String(), body.Accounts[0].Quotas
	}
	code, raw, quotas := get("", "")
	if code != 200 || !strings.Contains(raw, `"resetsAt":"2030-01-02T00:00:00-03:00"`) || quotas[0].Name != "Semanal" {
		t.Fatalf("default must be Argentina time and Spanish: %s", raw)
	}
	if _, raw, _ = get("UTC", ""); !strings.Contains(raw, `"resetsAt":"2030-01-02T03:00:00Z"`) {
		t.Fatalf("UTC must be available on request: %s", raw)
	}
	code, raw, quotas = get("Asia/Tokyo", "en-US,en;q=0.9,es;q=0.8")
	if code != 200 || !strings.Contains(raw, `"resetsAt":"2030-01-02T12:00:00+09:00"`) || !strings.Contains(raw, `+09:00","status"`) || !strings.Contains(raw, `+09:00","schemaVersion"`) {
		t.Fatalf("times must use the requested zone: %s", raw)
	}
	for i, want := range []string{"Weekly", "5 hours", "1 hour", "7 days Opus", "GPT · Weekly", "Prepaid balance", "Nombre propio"} {
		if quotas[i].Name != want {
			t.Fatalf("name %q, want %q", quotas[i].Name, want)
		}
	}
	if _, _, quotas = get("", "fr, es;q=0.5, en;q=0.1"); quotas[0].Name != "Semanal" {
		t.Fatal("first supported language must win")
	}
	if code, _, _ = get("Mars/Olympus", ""); code != 400 {
		t.Fatalf("invalid zone got %d", code)
	}
	if !strings.Contains(raw, `"usageLevel":"low"`) || quotas[0].UsageLevel != "low" {
		t.Fatalf("usage level must be returned: %s", raw)
	}
	if stored, _ := store.Get("direct:a"); stored.Quotas[0].Name != "Semanal" || stored.Quotas[0].ResetsAt.Location() != time.UTC || stored.Quotas[0].UsageLevel != "" {
		t.Fatal("stored report must not be modified")
	}
}

func TestLoginPermissions(t *testing.T) {
	a := &API{DeviceToken: "device", CollectorToken: "collector"}
	h := a.Handler()
	for _, route := range []string{"POST codex", "GET codex", "POST claude", "PUT claude", "GET claude"} {
		method, provider, _ := strings.Cut(route, " ")
		for _, token := range []string{"", "device", "collector"} {
			req := httptest.NewRequest(method, "/accounts/some-account/"+provider+"/login", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			want := 401
			if token == "collector" {
				want = 404
			}
			if w.Code != want {
				t.Fatalf("%s token=%s got=%d want=%d", route, token, w.Code, want)
			}
		}
	}
}
