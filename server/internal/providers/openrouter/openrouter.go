package openrouter

import (
	"ai-quota-monitor/server/internal/usage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var ErrAuth = errors.New("provider authentication required")

type OpenRouter struct {
	Client   *http.Client
	Endpoint string
}

func (p OpenRouter) Collect(ctx context.Context, key string) (usage.Report, error) {
	if key == "" {
		return usage.Report{}, ErrAuth
	}
	endpoint := p.Endpoint
	if endpoint == "" {
		endpoint = "https://openrouter.ai/api/v1/credits"
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if e != nil {
		return usage.Report{}, e
	}
	req.Header.Set("Authorization", "Bearer "+key)
	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	res, e := client.Do(req)
	if e != nil {
		return usage.Report{}, fmt.Errorf("OpenRouter request failed")
	}
	defer res.Body.Close()
	if res.StatusCode == 401 || res.StatusCode == 403 {
		return usage.Report{}, ErrAuth
	}
	if res.StatusCode != 200 {
		return usage.Report{}, fmt.Errorf("OpenRouter returned HTTP %d", res.StatusCode)
	}
	var body struct {
		Data *struct {
			Credits *float64 `json:"total_credits"`
			Usage   *float64 `json:"total_usage"`
		} `json:"data"`
	}
	if e = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&body); e != nil {
		return usage.Report{}, fmt.Errorf("invalid OpenRouter response")
	}
	if body.Data == nil || body.Data.Credits == nil || body.Data.Usage == nil {
		return usage.Report{}, fmt.Errorf("incomplete OpenRouter response")
	}
	r := usage.Report{UpdatedAt: time.Now().UTC(), Quotas: []usage.Quota{{ID: "balance", Name: "Saldo prepago", Unit: "usd", Used: body.Data.Usage, Limit: body.Data.Credits}}}
	if e = usage.ValidateReport(r, time.Now()); e != nil {
		return usage.Report{}, e
	}
	return r, nil
}
