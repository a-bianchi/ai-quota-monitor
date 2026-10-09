package codex

import (
	"ai-quota-monitor/server/internal/usage"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func ValidateURL(base string) error {
	u, e := url.Parse(base)
	if e != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Scheme != "http" && u.Scheme != "https") {
		return errors.New("QUOTA_API_URL must be an HTTP(S) base URL without credentials, query or fragment")
	}
	return nil
}
func Push(ctx context.Context, client *http.Client, base, id, token string, report usage.Report) error {
	if e := ValidateURL(base); e != nil {
		return e
	}
	body, e := json.Marshal(report)
	if e != nil {
		return e
	}
	endpoint := strings.TrimRight(base, "/") + "/accounts/" + url.PathEscape(id) + "/usage"
	req, e := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if e != nil {
		return e
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	res, e := client.Do(req)
	if e != nil {
		return errors.New("cannot reach quota API")
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 4096))
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("quota API returned HTTP %d (check account configuration and COLLECTOR_TOKEN)", res.StatusCode)
	}
	return nil
}
