package usage

import (
	"fmt"
	"math"
	"time"
)

type Quota struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Unit     string     `json:"unit"`
	Used     *float64   `json:"used"`
	Limit    *float64   `json:"limit"`
	ResetsAt *time.Time `json:"resetsAt"`
}
type Subscription struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Provider  string     `json:"provider"`
	Scope     string     `json:"scope"`
	Source    string     `json:"source"`
	UpdatedAt *time.Time `json:"updatedAt"`
	Status    string     `json:"status"`
	Quotas    []Quota    `json:"quotas"`
}
type Report struct {
	UpdatedAt time.Time `json:"updatedAt"`
	Quotas    []Quota   `json:"quotas"`
}

func Number(n float64) *float64 { return &n }
func ValidateReport(r Report, now time.Time) error {
	if r.UpdatedAt.IsZero() || r.UpdatedAt.After(now.Add(time.Minute)) {
		return fmt.Errorf("updatedAt must be a valid observation time, not in the future")
	}
	if len(r.Quotas) == 0 || len(r.Quotas) > 16 {
		return fmt.Errorf("expected between 1 and 16 quotas")
	}
	seen := map[string]bool{}
	for _, q := range r.Quotas {
		if q.ID == "" || len(q.ID) > 80 || q.Name == "" || len(q.Name) > 120 || seen[q.ID] {
			return fmt.Errorf("quota IDs must be unique; IDs and names are required and bounded")
		}
		seen[q.ID] = true
		switch q.Unit {
		case "percent", "credits", "usd", "tokens":
		default:
			return fmt.Errorf("invalid quota unit")
		}
		for _, n := range []*float64{q.Used, q.Limit} {
			if n != nil && (math.IsNaN(*n) || math.IsInf(*n, 0) || *n < 0) {
				return fmt.Errorf("quota values must be finite and nonnegative")
			}
		}
		if q.Unit == "percent" && ((q.Limit != nil && *q.Limit != 100) || (q.Used != nil && *q.Used > 100)) {
			return fmt.Errorf("percent limit must equal 100 and used cannot exceed 100")
		}
	}
	return nil
}
