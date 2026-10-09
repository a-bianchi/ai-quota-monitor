package codex

import (
	"ai-quota-monitor/server/internal/usage"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMain doubles as a fake app-server subprocess, exercising the real stdio
// transport, handshake, account check, notifications and process cleanup.
func TestMain(m *testing.M) {
	if os.Getenv("QUOTA_FAKE_CODEX") == "1" {
		scanner := bufio.NewScanner(os.Stdin)
		encoder := json.NewEncoder(os.Stdout)
		for scanner.Scan() {
			var request struct {
				ID     *int   `json:"id"`
				Method string `json:"method"`
			}
			if json.Unmarshal(scanner.Bytes(), &request) != nil {
				os.Exit(2)
			}
			if request.Method == "initialized" {
				continue
			}
			var result any
			switch request.Method {
			case "initialize":
				result = map[string]string{"userAgent": "fake"}
			case "account/read":
				if os.Getenv("QUOTA_FAKE_LOGGED_OUT") == "1" {
					result = map[string]any{"account": nil}
				} else {
					result = map[string]any{"account": map[string]string{"type": "chatgpt"}}
				}
			case "account/login/start":
				if os.Getenv("QUOTA_FAKE_DELAY_LOGIN") == "1" {
					time.Sleep(time.Second)
				}
				result = map[string]string{"type": "chatgptDeviceCode", "loginId": "test-login", "verificationUrl": "https://auth.openai.com/codex/device", "userCode": "ABCD-1234"}
			case "account/rateLimits/read":
				result = map[string]any{"rateLimits": map[string]any{"limitId": "codex", "primary": map[string]any{"usedPercent": 11, "windowDurationMins": 300, "resetsAt": time.Now().Add(time.Hour).Unix()}}}
			default:
				os.Exit(3)
			}
			_ = encoder.Encode(map[string]any{"method": "account/updated", "params": map[string]string{"authMode": "chatgpt"}})
			_ = encoder.Encode(map[string]any{"id": request.ID, "result": result})
			if request.Method == "account/login/start" {
				event := map[string]any{"loginId": "test-login", "success": os.Getenv("QUOTA_FAKE_FAIL_LOGIN") != "1"}
				_ = encoder.Encode(map[string]any{"method": "account/login/completed", "params": event})
			}
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}
func TestClientProtocol(t *testing.T) {
	t.Setenv("QUOTA_FAKE_CODEX", "1")
	t.Setenv("OPENAI_API_KEY", "must-not-reach-child")
	c := Client{Binary: os.Args[0], Home: t.TempDir()}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	limits, e := c.Read(ctx)
	if e != nil {
		t.Fatal(e)
	}
	report, e := limits.Report(time.Now())
	if e != nil {
		t.Fatal(e)
	}
	if len(report.Quotas) != 1 || *report.Quotas[0].Used != 11 || report.Quotas[0].Name != "5 horas" {
		t.Fatalf("unexpected quotas %+v", report)
	}
	t.Setenv("QUOTA_FAKE_LOGGED_OUT", "1")
	if _, e = c.Read(ctx); e != ErrLogin {
		t.Fatalf("expected login error, got %v", e)
	}
	for _, entry := range SessionEnv(c.Home) {
		if strings.HasPrefix(entry, "OPENAI_API_KEY=") {
			t.Fatal("API key leaked into child")
		}
	}
}
func TestLimitsMapping(t *testing.T) {
	now := time.Now().UTC()
	reset := now.Add(time.Hour).Unix()
	raw := fmt.Sprintf(`{"rateLimits":{"limitId":"legacy","primary":{"usedPercent":99}},"rateLimitsByLimitId":{"codex":{"primary":null,"secondary":{"usedPercent":6,"windowDurationMins":10080,"resetsAt":%d}},"extra":{"limitName":"Extra","primary":{"usedPercent":42,"windowDurationMins":15,"resetsAt":%d}}}}`, reset, reset)
	var l Limits
	if e := json.Unmarshal([]byte(raw), &l); e != nil {
		t.Fatal(e)
	}
	report, e := l.Report(now)
	if e != nil {
		t.Fatal(e)
	}
	if len(report.Quotas) != 2 || report.Quotas[0].ID != "codex:secondary" || report.Quotas[0].Name != "Semanal" || report.Quotas[1].Name != "Extra · 15 minutos" {
		t.Fatalf("unexpected %+v", report)
	}
	if !report.Quotas[0].ResetsAt.Equal(time.Unix(reset, 0)) {
		t.Fatal("reset timestamp changed")
	}
	// Never manufacture a 5h window when the API only returns a weekly one.
	l = Limits{RateLimits: &Bucket{Secondary: &Window{UsedPercent: usage.Number(1)}}}
	report, e = l.Report(now)
	if e != nil || len(report.Quotas) != 1 {
		t.Fatal("legacy fallback failed")
	}
	past := now.Add(-time.Minute).Unix()
	l.RateLimits.Secondary.ResetsAt = &past
	if _, e = l.Report(now); e == nil {
		t.Fatal("expired window treated as fresh")
	}
	if _, e = (Limits{}).Report(now); e == nil {
		t.Fatal("empty limits treated as data")
	}
}
func TestPush(t *testing.T) {
	report := usage.Report{UpdatedAt: time.Now().UTC(), Quotas: []usage.Quota{{ID: "5h", Name: "5 horas", Unit: "percent", Used: usage.Number(11), Limit: usage.Number(100)}}}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != "/accounts/codex-work/usage" || r.Header.Get("Authorization") != "Bearer collector" {
			t.Error("incorrect ingestion request")
		}
		var got usage.Report
		if e := json.NewDecoder(r.Body).Decode(&got); e != nil {
			t.Error(e)
		}
		if len(got.Quotas) != 1 || *got.Quotas[0].Used != 11 {
			t.Error("wrong payload")
		}
		w.WriteHeader(200)
	}))
	defer s.Close()
	if e := Push(context.Background(), s.Client(), s.URL, "codex-work", "collector", report); e != nil {
		t.Fatal(e)
	}
	if e := ValidateURL("https://token@host"); e == nil {
		t.Fatal("credentials in URL accepted")
	}
}
func TestSessionIsolation(t *testing.T) {
	dir := t.TempDir()
	home, e := PrepareHome(filepath.Join(dir, "profile"))
	if e != nil {
		t.Fatal(e)
	}
	if !filepath.IsAbs(home) {
		t.Fatal("expected absolute profile")
	}
	user, e := os.UserHomeDir()
	if e != nil {
		t.Fatal(e)
	}
	if _, e = PrepareHome(filepath.Join(user, ".codex")); e == nil {
		t.Fatal("main session allowed")
	}
}

func TestDeviceLogin(t *testing.T) {
	t.Setenv("QUOTA_FAKE_CODEX", "1")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c := Client{Binary: os.Args[0], Home: t.TempDir()}
	prompt, wait, closeSession, e := c.StartDevice(ctx)
	if e != nil {
		t.Fatal(e)
	}
	if prompt.UserCode != "ABCD-1234" || prompt.VerificationURL != "https://auth.openai.com/codex/device" {
		t.Fatal("wrong device prompt")
	}
	if e = wait(); e != nil {
		t.Fatal(e)
	}
	closeSession()
	t.Setenv("QUOTA_FAKE_FAIL_LOGIN", "1")
	_, wait, closeSession, e = c.StartDevice(ctx)
	if e != nil {
		t.Fatal(e)
	}
	defer closeSession()
	if e = wait(); e == nil {
		t.Fatal("failed authorization reported success")
	}
}
