package usage

import "testing"

func TestLevel(t *testing.T) {
	for _, tc := range []struct {
		used, limit *float64
		want        string
	}{
		{Number(0), Number(100), "low"},
		{Number(49.99), Number(100), "low"},
		{Number(50), Number(100), "medium"},
		{Number(80), Number(100), "medium"},
		{Number(80.01), Number(100), "high"},
		{Number(150), Number(100), "high"},
		{Number(5), Number(10), "medium"},
		{nil, Number(100), "unknown"},
		{Number(5), nil, "unknown"},
		{Number(5), Number(0), "unknown"},
	} {
		if got := (Quota{Used: tc.used, Limit: tc.limit}).Level(); got != tc.want {
			t.Errorf("used %v limit %v: got %q, want %q", tc.used, tc.limit, got, tc.want)
		}
	}
}
