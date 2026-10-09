package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var ErrLogin = errors.New("Codex requires ChatGPT login; run the collector with -login for this account")

type Client struct {
	Binary string
	Home   string
}

// SessionEnv isolates Codex credentials from the main app and removes API-key
// overrides and monitor secrets from the child process environment.
func SessionEnv(home string) []string {
	out := []string{}
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		switch key {
		case "CODEX_HOME", "OPENAI_API_KEY", "CODEX_API_KEY", "CODEX_ACCESS_TOKEN", "DEVICE_TOKEN", "COLLECTOR_TOKEN", "OPENROUTER_PERSONAL_KEY", "CURSOR_WORK_SESSION", "AUTH_USERNAME", "AUTH_PASSWORD", "AUTH_PASSWORD_HASH", "MASTER_KEY", "ADMIN_TOKEN":
			continue
		}
		out = append(out, entry)
	}
	return append(out, "CODEX_HOME="+home)
}
func PrepareHome(home string) (string, error) {
	abs, e := filepath.Abs(home)
	if e != nil {
		return "", e
	}
	user, e := os.UserHomeDir()
	if e != nil {
		return "", e
	}
	if abs == filepath.Join(user, ".codex") {
		return "", errors.New("use a separate Codex home; the main app session is not a collector profile")
	}
	if e = os.MkdirAll(abs, 0700); e != nil {
		return "", e
	}
	resolved, e := filepath.EvalSymlinks(abs)
	if e != nil {
		return "", e
	}
	mainHome := filepath.Join(user, ".codex")
	if canonical, e := filepath.EvalSymlinks(mainHome); e == nil {
		mainHome = canonical
	}
	if resolved == mainHome {
		return "", errors.New("collector home cannot alias the main app session")
	}
	return resolved, nil
}
func (c Client) Login(ctx context.Context, device bool) error {
	args := []string{"login", "-c", `cli_auth_credentials_store="file"`}
	if device {
		args = append(args, "--device-auth")
	}
	cmd := exec.CommandContext(ctx, c.Binary, args...)
	cmd.Env = SessionEnv(c.Home)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type rpc struct {
	enc     *json.Encoder
	scanner *bufio.Scanner
	next    int
}

func (r *rpc) call(method string, params any, result any) error {
	r.next++
	id := r.next
	request := map[string]any{"id": id, "method": method}
	if params != nil {
		request["params"] = params
	}
	if e := r.enc.Encode(request); e != nil {
		return errors.New("cannot write to Codex App Server")
	}
	for r.scanner.Scan() {
		var msg struct {
			ID     *int            `json:"id"`
			Method string          `json:"method"`
			Result json.RawMessage `json:"result"`
			Error  *struct {
				Code int `json:"code"`
			} `json:"error"`
		}
		if e := json.Unmarshal(r.scanner.Bytes(), &msg); e != nil {
			return errors.New("invalid Codex RPC response")
		}
		if msg.Method != "" {
			if msg.ID != nil {
				return errors.New("unexpected Codex server request; no tool actions are approved by this collector")
			}
			continue
		}
		if msg.ID == nil || *msg.ID != id {
			continue
		}
		if msg.Error != nil {
			return fmt.Errorf("Codex %s failed (RPC code %d)", method, msg.Error.Code)
		}
		if len(msg.Result) == 0 {
			return errors.New("missing Codex RPC result")
		}
		if e := json.Unmarshal(msg.Result, result); e != nil {
			return errors.New("invalid Codex RPC result")
		}
		return nil
	}
	if e := r.scanner.Err(); e != nil {
		return errors.New("Codex RPC stream failed")
	}
	return errors.New("Codex App Server closed before responding")
}

type session struct {
	cmd       *exec.Cmd
	input     io.WriteCloser
	transport *rpc
}

func (s *session) close() { _ = s.input.Close(); _ = s.cmd.Process.Kill(); _ = s.cmd.Wait() }
func (c Client) open(ctx context.Context) (*session, error) {
	// Plugins are not needed to read limits; left enabled, Codex clones ~80 MB of them into every home.
	cmd := exec.CommandContext(ctx, c.Binary, "app-server", "-c", `cli_auth_credentials_store="file"`, "--disable", "plugins", "--disable", "remote_plugin")
	cmd.Env = SessionEnv(c.Home)
	cmd.Stderr = io.Discard
	input, e := cmd.StdinPipe()
	if e != nil {
		return nil, e
	}
	output, e := cmd.StdoutPipe()
	if e != nil {
		return nil, e
	}
	if e = cmd.Start(); e != nil {
		return nil, errors.New("cannot start Codex; check CODEX_BIN or container installation")
	}
	scanner := bufio.NewScanner(output)
	scanner.Buffer(make([]byte, 4096), 2<<20)
	s := &session{cmd: cmd, input: input, transport: &rpc{enc: json.NewEncoder(input), scanner: scanner}}
	var init any
	if e = s.transport.call("initialize", map[string]any{"clientInfo": map[string]string{"name": "ai_quota_monitor", "title": "AI Quota Monitor", "version": "0.2.0"}}, &init); e != nil {
		s.close()
		return nil, e
	}
	if e = s.transport.enc.Encode(map[string]any{"method": "initialized", "params": map[string]any{}}); e != nil {
		s.close()
		return nil, e
	}
	return s, nil
}
func (c Client) Read(ctx context.Context) (Limits, error) {
	var limits Limits
	s, e := c.open(ctx)
	if e != nil {
		return limits, e
	}
	defer s.close()
	var account struct {
		Account *struct {
			Type string `json:"type"`
		} `json:"account"`
	}
	if e = s.transport.call("account/read", map[string]bool{"refreshToken": false}, &account); e != nil {
		return limits, e
	}
	if account.Account == nil || account.Account.Type != "chatgpt" {
		return limits, ErrLogin
	}
	if e = s.transport.call("account/rateLimits/read", nil, &limits); e != nil {
		return limits, e
	}
	return limits, nil
}

type DevicePrompt struct {
	LoginID         string `json:"loginId"`
	VerificationURL string `json:"verificationUrl"`
	UserCode        string `json:"userCode"`
}

// StartDevice keeps the App Server alive until wait returns, so it can poll the
// authorization server and persist/refresh the managed credentials itself.
func (c Client) StartDevice(ctx context.Context) (DevicePrompt, func() error, func(), error) {
	var prompt DevicePrompt
	s, e := c.open(ctx)
	if e != nil {
		return prompt, nil, nil, e
	}
	if e = s.transport.call("account/login/start", map[string]string{"type": "chatgptDeviceCode"}, &prompt); e != nil {
		s.close()
		return prompt, nil, nil, e
	}
	if prompt.LoginID == "" || prompt.UserCode == "" || prompt.VerificationURL != "https://auth.openai.com/codex/device" {
		s.close()
		return prompt, nil, nil, errors.New("invalid device authorization response")
	}
	wait := func() error {
		for s.transport.scanner.Scan() {
			var msg struct {
				ID     json.RawMessage `json:"id"`
				Method string          `json:"method"`
				Params struct {
					LoginID string `json:"loginId"`
					Success bool   `json:"success"`
				} `json:"params"`
			}
			if e := json.Unmarshal(s.transport.scanner.Bytes(), &msg); e != nil {
				return errors.New("invalid login event")
			}
			if len(msg.ID) > 0 && string(msg.ID) != "null" {
				return errors.New("unexpected request during device login")
			}
			if msg.Method == "account/login/completed" && msg.Params.LoginID == prompt.LoginID {
				if !msg.Params.Success {
					return errors.New("device authorization failed or expired")
				}
				return nil
			}
		}
		return errors.New("device authorization ended before completion")
	}
	return prompt, wait, s.close, nil
}
