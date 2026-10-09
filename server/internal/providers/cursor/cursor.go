// Package cursor reads the signed-in user's own usage from the endpoint behind
// cursor.com's dashboard. The endpoint is undocumented and may change.
package cursor

import (
	"ai-quota-monitor/server/internal/usage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

var ErrAuth = errors.New("provider authentication required")

const cookieName = "WorkosCursorSessionToken"

type Cursor struct {
	Client   *http.Client
	Endpoint string
}

type meter struct {
	Enabled *bool    `json:"enabled"`
	Used    *float64 `json:"used"`
	Limit   *float64 `json:"limit"`
	Auto    *float64 `json:"autoPercentUsed"`
	API     *float64 `json:"apiPercentUsed"`
	Total   *float64 `json:"totalPercentUsed"`
}

type summary struct {
	BillingCycleEnd *time.Time `json:"billingCycleEnd"`
	Individual      struct {
		Plan     *meter `json:"plan"`
		Overall  *meter `json:"overall"`
		OnDemand *meter `json:"onDemand"`
	} `json:"individualUsage"`
	Team struct {
		OnDemand *meter `json:"onDemand"`
	} `json:"teamUsage"`
}

// Collect takes the value of the WorkosCursorSessionToken cookie from cursor.com.
func (p Cursor) Collect(ctx context.Context, session string) (usage.Report, error) {
	session = strings.TrimPrefix(strings.TrimSpace(session), cookieName+"=")
	if session == "" || strings.ContainsAny(session, "; \r\n") {
		return usage.Report{}, ErrAuth
	}
	endpoint := p.Endpoint
	if endpoint == "" {
		endpoint = "https://cursor.com/api/usage-summary"
	}
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if e != nil {
		return usage.Report{}, e
	}
	req.Header.Set("Cookie", cookieName+"="+session)
	req.Header.Set("Accept", "application/json")
	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	res, e := client.Do(req)
	if e != nil {
		return usage.Report{}, fmt.Errorf("Cursor request failed")
	}
	defer res.Body.Close()
	// An expired session may also be answered with a redirect to the login page.
	if res.StatusCode == 401 || res.StatusCode == 403 || (res.StatusCode >= 300 && res.StatusCode < 400) {
		return usage.Report{}, ErrAuth
	}
	if res.StatusCode != 200 {
		return usage.Report{}, fmt.Errorf("Cursor returned HTTP %d", res.StatusCode)
	}
	var body summary
	if e = json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&body); e != nil {
		return usage.Report{}, fmt.Errorf("invalid Cursor response")
	}
	now := time.Now().UTC()
	var quotas []usage.Quota
	add := func(id, name string, percent *float64) {
		if percent == nil || math.IsNaN(*percent) || *percent < 0 {
			return
		}
		used := math.Min(math.Round(*percent*100)/100, 100)
		quotas = append(quotas, usage.Quota{ID: id, Name: name, Unit: "percent", Used: &used, Limit: usage.Number(100), ResetsAt: body.BillingCycleEnd})
	}
	// Included usage: prefer the percentages the dashboard shows; otherwise
	// derive one from used/limit, which holds whatever unit Cursor reports.
	included := body.Individual.Plan
	if included == nil || (included.Total == nil && ratio(included) == nil) {
		included = body.Individual.Overall
	}
	if included != nil {
		if included.Total != nil {
			add("included", "Uso incluido", included.Total)
		} else {
			add("included", "Uso incluido", ratio(included))
		}
		add("auto", "Auto", included.Auto)
		add("api", "Modelos API", included.API)
	}
	add("on-demand", "On-demand", ratio(body.Individual.OnDemand))
	add("team-on-demand", "On-demand equipo", ratio(body.Team.OnDemand))
	r := usage.Report{UpdatedAt: now, Quotas: quotas}
	if len(quotas) == 0 {
		return usage.Report{}, fmt.Errorf("Cursor response has no usage with a known limit")
	}
	if e = usage.ValidateReport(r, now); e != nil {
		return usage.Report{}, e
	}
	return r, nil
}

// ratio returns used/limit as a percentage, or nil when the meter is disabled or has no limit.
func ratio(m *meter) *float64 {
	if m == nil || (m.Enabled != nil && !*m.Enabled) || m.Used == nil || m.Limit == nil || *m.Limit <= 0 {
		return nil
	}
	return usage.Number(*m.Used / *m.Limit * 100)
}
