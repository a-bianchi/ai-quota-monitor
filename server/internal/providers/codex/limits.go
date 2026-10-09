package codex

import (
	"ai-quota-monitor/server/internal/usage"
	"errors"
	"fmt"
	"sort"
	"time"
)

type Window struct {
	UsedPercent        *float64 `json:"usedPercent"`
	WindowDurationMins *int     `json:"windowDurationMins"`
	ResetsAt           *int64   `json:"resetsAt"`
}
type Bucket struct {
	LimitID   string  `json:"limitId"`
	LimitName string  `json:"limitName"`
	Primary   *Window `json:"primary"`
	Secondary *Window `json:"secondary"`
}
type Limits struct {
	RateLimits *Bucket           `json:"rateLimits"`
	ByID       map[string]Bucket `json:"rateLimitsByLimitId"`
}

func (l Limits) Report(now time.Time) (usage.Report, error) {
	buckets := l.ByID
	if len(buckets) == 0 && l.RateLimits != nil {
		id := l.RateLimits.LimitID
		if id == "" {
			id = "codex"
		}
		buckets = map[string]Bucket{id: *l.RateLimits}
	}
	ids := make([]string, 0, len(buckets))
	for id := range buckets {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	report := usage.Report{UpdatedAt: now.UTC(), Quotas: []usage.Quota{}}
	for _, id := range ids {
		b := buckets[id]
		for _, slot := range []struct {
			name   string
			window *Window
		}{{"primary", b.Primary}, {"secondary", b.Secondary}} {
			w := slot.window
			if w == nil || w.UsedPercent == nil {
				continue
			}
			// An expired snapshot must not be recorded as fresh after a reset.
			if w.ResetsAt != nil && *w.ResetsAt > 0 && *w.ResetsAt <= now.Unix() {
				continue
			}
			label := slot.name
			if w.WindowDurationMins != nil {
				mins := *w.WindowDurationMins
				switch {
				case mins == 10080:
					label = "Semanal"
				case mins > 0 && mins%1440 == 0:
					label = fmt.Sprintf("%d días", mins/1440)
				case mins > 0 && mins%60 == 0:
					label = fmt.Sprintf("%d horas", mins/60)
				case mins > 0:
					label = fmt.Sprintf("%d minutos", mins)
				}
			}
			if id != "codex" {
				name := b.LimitName
				if name == "" {
					name = id
				}
				label = name + " · " + label
			}
			var reset *time.Time
			if w.ResetsAt != nil && *w.ResetsAt > 0 {
				t := time.Unix(*w.ResetsAt, 0).UTC()
				reset = &t
			}
			report.Quotas = append(report.Quotas, usage.Quota{ID: id + ":" + slot.name, Name: label, Unit: "percent", Used: w.UsedPercent, Limit: usage.Number(100), ResetsAt: reset})
		}
	}
	if len(report.Quotas) == 0 {
		return report, errors.New("Codex returned no current quota windows; no observation was sent")
	}
	if e := usage.ValidateReport(report, now); e != nil {
		return report, e
	}
	return report, nil
}
