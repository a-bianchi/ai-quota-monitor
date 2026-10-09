package main

import (
	"ai-quota-monitor/server/internal/config"
	"ai-quota-monitor/server/internal/providers/codex"
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"
)

func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func run() error {
	if e := config.LoadEnv(".env"); e != nil {
		return e
	}
	account := flag.String("account", "", "configured account ID, e.g. codex-work")
	login := flag.Bool("login", false, "sign in with ChatGPT for this account and exit")
	device := flag.Bool("device-auth", false, "use device authentication with -login")
	once := flag.Bool("once", false, "collect and send once, then exit")
	home := flag.String("codex-home", "", "isolated profile directory (default data/codex/<account>)")
	flag.Parse()
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,79}$`).MatchString(*account) {
		return errors.New("provide -account with a valid configured account ID")
	}
	cfg, e := config.Load(filepath.Join("configs", filepath.Base(value("CONFIG_NAME", "config.json"))))
	if e != nil {
		return e
	}
	found := false
	for _, a := range cfg.Accounts {
		if a.ID == *account && a.Provider == "openai" && a.Source == "collector" {
			found = true
		}
	}
	if !found {
		return errors.New("account must be configured with provider=openai and source=collector")
	}
	if *home == "" {
		*home = filepath.Join("data", "codex", *account)
	}
	resolved, e := codex.PrepareHome(*home)
	if e != nil {
		return e
	}
	client := codex.Client{Binary: value("CODEX_BIN", "codex"), Home: resolved}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *login {
		fmt.Println("Iniciá sesión con la cuenta correspondiente a", *account, "(perfil aislado). No cambia el login principal de Codex.")
		return client.Login(ctx, *device)
	}
	token := os.Getenv("COLLECTOR_TOKEN")
	if len(token) < 32 || token == os.Getenv("DEVICE_TOKEN") {
		return errors.New("set COLLECTOR_TOKEN with at least 32 characters, different from DEVICE_TOKEN, in server and collector")
	}
	base := value("QUOTA_API_URL", "http://localhost:8080")
	if e = codex.ValidateURL(base); e != nil {
		return e
	}
	interval, e := time.ParseDuration(value("CODEX_POLL_INTERVAL", "5m"))
	if e != nil || interval < 10*time.Second {
		return errors.New("CODEX_POLL_INTERVAL must be a duration >= 10s")
	}
	collect := func() error {
		attempt, cancel := context.WithTimeout(ctx, 45*time.Second)
		defer cancel()
		limits, e := client.Read(attempt)
		if e != nil {
			return e
		}
		report, e := limits.Report(time.Now())
		if e != nil {
			return e
		}
		if e = codex.Push(attempt, nil, base, *account, token, report); e != nil {
			return e
		}
		slog.Info("Codex quotas sent", "account", *account, "quotas", len(report.Quotas))
		return nil
	}
	if *once {
		return collect()
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if e = collect(); e != nil && ctx.Err() == nil {
			slog.Warn("Codex collection failed", "account", *account, "error", e)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
func main() {
	if e := run(); e != nil {
		slog.Error("collector failed", "error", e)
		os.Exit(1)
	}
}
