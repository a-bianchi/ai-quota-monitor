package openrouter

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenRouter(t *testing.T) {
	for _, tc := range []struct {
		name      string
		code      int
		body      string
		wantError bool
		auth      bool
	}{
		{"credits", 200, `{"data":{"total_credits":25,"total_usage":6.6}}`, false, false},
		{"missing field", 200, `{"data":{"total_credits":25}}`, true, false},
		{"bad credentials", 403, `{}`, true, true},
		{"server error", 500, `{}`, true, false},
		{"negative value", 200, `{"data":{"total_credits":-1,"total_usage":0}}`, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer test-key" {
					t.Error("missing bearer token")
				}
				w.WriteHeader(tc.code)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer s.Close()
			report, e := (OpenRouter{Endpoint: s.URL, Client: s.Client()}).Collect(context.Background(), "test-key")
			if (e != nil) != tc.wantError {
				t.Fatalf("error=%v", e)
			}
			if tc.auth && !errors.Is(e, ErrAuth) {
				t.Fatal("expected auth error")
			}
			if !tc.wantError {
				q := report.Quotas[0]
				if *q.Used != 6.6 || *q.Limit != 25 || q.Unit != "usd" {
					t.Fatalf("unexpected credits: %+v", q)
				}
			}
		})
	}
}
