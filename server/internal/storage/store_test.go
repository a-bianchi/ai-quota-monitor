package storage

import (
	"ai-quota-monitor/server/internal/usage"
	"testing"
	"time"
)

func TestPersistence(t *testing.T) {
	dir := t.TempDir()
	s, e := New(dir)
	if e != nil {
		t.Fatal(e)
	}
	r := usage.Report{UpdatedAt: time.Now().UTC(), Quotas: []usage.Quota{{ID: "credits", Name: "Credits", Unit: "credits", Used: usage.Number(1.25)}}}
	if e = s.Put("collector:account", r); e != nil {
		t.Fatal(e)
	}
	reopened, e := New(dir)
	if e != nil {
		t.Fatal(e)
	}
	got, ok := reopened.Get("collector:account")
	if !ok || *got.Quotas[0].Used != 1.25 || got.Quotas[0].Limit != nil {
		t.Fatal("persistence lost values or null")
	}
	r.UpdatedAt = r.UpdatedAt.Add(-time.Minute)
	if e = s.Put("collector:account", r); e == nil {
		t.Fatal("older report accepted")
	}
	if _, ok = reopened.Get("direct:account"); ok {
		t.Fatal("source namespaces must be independent")
	}
}
