package main

import (
	"ai-quota-monitor/server/internal/config"
	"ai-quota-monitor/server/internal/httpapi"
	"ai-quota-monitor/server/internal/providers/claude"
	"ai-quota-monitor/server/internal/providers/codex"
	"ai-quota-monitor/server/internal/providers/cursor"
	"ai-quota-monitor/server/internal/providers/openrouter"
	"ai-quota-monitor/server/internal/storage"
	"ai-quota-monitor/server/internal/usage"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
	// Embedded so X-Timezone works in images without a zone database.
	_ "time/tzdata"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func duration(key, fallback string) (time.Duration, error) {
	return time.ParseDuration(env(key, fallback))
}
func run() error {
	if e := config.LoadEnv(".env"); e != nil {
		return e
	}
	device := os.Getenv("DEVICE_TOKEN")
	collector := os.Getenv("COLLECTOR_TOKEN")
	if len(device) < 32 {
		return errors.New("DEVICE_TOKEN must contain at least 32 characters")
	}
	if collector != "" && (len(collector) < 32 || collector == device) {
		return errors.New("COLLECTOR_TOKEN must contain at least 32 characters and differ from DEVICE_TOKEN")
	}
	username, password := os.Getenv("AUTH_USERNAME"), os.Getenv("AUTH_PASSWORD")
	if (username == "") != (password == "") {
		return errors.New("set both AUTH_USERNAME and AUTH_PASSWORD, or neither")
	}
	if username != "" && (len(password) < 8 || collector == "") {
		return errors.New("login requires AUTH_PASSWORD with at least 8 characters and COLLECTOR_TOKEN")
	}
	interval, e := duration("POLL_INTERVAL", "5m")
	if e != nil || interval < time.Second {
		return errors.New("POLL_INTERVAL must be a duration of at least 1s")
	}
	stale, e := duration("STALE_AFTER", "15m")
	if e != nil || stale < interval {
		return errors.New("STALE_AFTER must be a duration >= POLL_INTERVAL")
	}
	port, e := strconv.Atoi(env("HTTP_PORT", "8080"))
	if e != nil || port < 1 || port > 65535 {
		return errors.New("invalid HTTP_PORT")
	}
	configPath := filepath.Join("configs", filepath.Base(env("CONFIG_NAME", "config.json")))
	cfg, e := config.Load(configPath)
	if e != nil {
		return e
	}
	store, e := storage.New(env("DATA_DIR", "./data"))
	if e != nil {
		return e
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	codexManager := codex.New(ctx, env("CODEX_BIN", "codex"), filepath.Join(env("DATA_DIR", "./data"), "codex"))
	claudeManager := claude.New(filepath.Join(env("DATA_DIR", "./data"), "claude"))
	slog.Info("configuration loaded", "path", configPath)
	api := &httpapi.API{CodexAuth: codexManager, ClaudeAuth: claudeManager, Accounts: cfg.Accounts, Store: store, DeviceToken: device, CollectorToken: collector, StaleAfter: stale, Username: username, Password: password}
	var mocks map[string][]usage.Quota
	for _, a := range cfg.Accounts {
		if a.Source == "mock" {
			b, e := os.ReadFile(env("MOCK_FILE", "configs/mock-usage.json"))
			if e != nil {
				return e
			}
			if e = json.Unmarshal(b, &mocks); e != nil {
				return e
			}
			break
		}
	}
	for _, a := range cfg.Accounts {
		if a.Source == "mock" {
			report := usage.Report{UpdatedAt: time.Now().UTC(), Quotas: mocks[a.ID]}
			if e = usage.ValidateReport(report, time.Now()); e != nil {
				return errors.New("invalid or missing mock data for " + a.ID)
			}
		}
	}
	collect := func() {
		for _, a := range cfg.Accounts {
			if ctx.Err() != nil {
				return
			}
			var report usage.Report
			var err error
			switch a.Source {
			case "collector":
				continue
			case "mock":
				report = usage.Report{UpdatedAt: time.Now().UTC(), Quotas: mocks[a.ID]}
			case "direct":
				switch a.Provider {
				case "openai":
					attempt, cancel := context.WithTimeout(ctx, 45*time.Second)
					var limits codex.Limits
					limits, err = codexManager.Read(attempt, a.ID)
					if err == nil {
						report, err = limits.Report(time.Now())
					}
					cancel()
					if errors.Is(err, codex.ErrBusy) {
						continue
					}
				case "anthropic":
					report, err = claudeManager.Read(ctx, a.ID)
				case "cursor":
					report, err = (cursor.Cursor{}).Collect(ctx, os.Getenv(a.CredentialEnv))
				default:
					report, err = (openrouter.OpenRouter{}).Collect(ctx, os.Getenv(a.CredentialEnv))
				}
			}
			if err != nil {
				status := "stale"
				if _, ok := store.Get(a.Source + ":" + a.ID); !ok {
					status = "unavailable"
				}
				if errors.Is(err, openrouter.ErrAuth) || errors.Is(err, cursor.ErrAuth) || errors.Is(err, codex.ErrLogin) || errors.Is(err, claude.ErrLogin) {
					status = "auth_required"
				}
				api.Failure(a.ID, status)
				slog.Warn("collection failed", "account", a.ID, "status", status)
				continue
			}
			if err = store.Put(a.Source+":"+a.ID, report); err != nil {
				slog.Error("persist failed", "account", a.ID)
				api.Failure(a.ID, "unavailable")
				continue
			}
			api.Failure(a.ID, "")
		}
	}
	// Collect asynchronously so provider outages do not block HTTP startup.
	go func() {
		collect()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				collect()
			case <-codexManager.Ready:
				collect()
			case <-claudeManager.Ready:
				collect()
			}
		}
	}()
	server := &http.Server{Addr: env("HTTP_HOST", "127.0.0.1") + ":" + strconv.Itoa(port), Handler: api.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16 << 10}
	done := make(chan error, 1)
	go func() { slog.Info("server listening", "address", server.Addr); done <- server.ListenAndServe() }()
	select {
	case e := <-done:
		if !errors.Is(e, http.ErrServerClosed) {
			return e
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	}
	return nil
}
func main() {
	if e := run(); e != nil {
		slog.Error("startup failed", "error", e)
		os.Exit(1)
	}
}
