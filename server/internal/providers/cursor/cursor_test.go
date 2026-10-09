package cursor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Payloads follow shapes reported by third parties; they are not captures from a real account.
func TestCursor(t *testing.T) {
	for _, tc := range []struct {
		name      string
		code      int
		body      string
		wantError bool
		auth      bool
		want      map[string]float64
	}{
		{"individual percentages", 200, `{"billingCycleEnd":"2026-11-01T00:00:00.000Z","individualUsage":{"plan":{"enabled":true,"used":900,"limit":2000,"autoPercentUsed":12.5,"apiPercentUsed":40,"totalPercentUsed":45},"onDemand":{"enabled":false,"used":0,"limit":null}}}`, false, false, map[string]float64{"included": 45, "auto": 12.5, "api": 40}},
		{"used and limit only", 200, `{"billingCycleEnd":"2026-11-01T00:00:00Z","individualUsage":{"plan":{"enabled":true,"used":125,"limit":500}}}`, false, false, map[string]float64{"included": 25}},
		{"over the limit", 200, `{"individualUsage":{"overall":{"enabled":true,"used":300,"limit":200}}}`, false, false, map[string]float64{"included": 100}},
		{"team on-demand", 200, `{"individualUsage":{"overall":{"enabled":false,"used":0,"limit":null}},"teamUsage":{"onDemand":{"enabled":true,"used":50,"limit":200}}}`, false, false, map[string]float64{"team-on-demand": 25}},
		{"no known limit", 200, `{"individualUsage":{"overall":{"enabled":false,"used":0,"limit":null}}}`, true, false, nil},
		{"not json", 200, `<html>`, true, false, nil},
		{"expired session", 401, `{"error":"not_authenticated"}`, true, true, nil},
		{"login redirect", 307, ``, true, true, nil},
		{"server error", 500, `{}`, true, false, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Cookie") != "WorkosCursorSessionToken=test-session" {
					t.Error("missing session cookie")
				}
				w.WriteHeader(tc.code)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer s.Close()
			report, e := (Cursor{Endpoint: s.URL, Client: s.Client()}).Collect(context.Background(), "WorkosCursorSessionToken=test-session\n")
			if (e != nil) != tc.wantError {
				t.Fatalf("error=%v", e)
			}
			if tc.auth != errors.Is(e, ErrAuth) {
				t.Fatalf("auth error mismatch: %v", e)
			}
			if len(report.Quotas) != len(tc.want) {
				t.Fatalf("unexpected quotas: %+v", report.Quotas)
			}
			for _, q := range report.Quotas {
				if want, ok := tc.want[q.ID]; !ok || *q.Used != want || q.Unit != "percent" {
					t.Fatalf("unexpected quota: %+v", q)
				}
			}
			if tc.name == "individual percentages" && report.Quotas[0].ResetsAt == nil {
				t.Fatal("missing reset time")
			}
		})
	}
	if _, e := (Cursor{}).Collect(context.Background(), ""); !errors.Is(e, ErrAuth) {
		t.Fatal("empty session must require auth")
	}
}
