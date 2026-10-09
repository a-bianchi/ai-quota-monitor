package panelauth

import (
	"sync"
	"time"
)

// Limiter blocks a key after too many failures inside a sliding window.
type Limiter struct {
	Max    int
	Window time.Duration
	mu     sync.Mutex
	fails  map[string][]time.Time
}

func NewLimiter(max int, window time.Duration) *Limiter {
	return &Limiter{Max: max, Window: window, fails: map[string][]time.Time{}}
}
func (l *Limiter) recent(key string, now time.Time) []time.Time {
	kept := l.fails[key][:0]
	for _, t := range l.fails[key] {
		if now.Sub(t) < l.Window {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		delete(l.fails, key)
	} else {
		l.fails[key] = kept
	}
	return kept
}

// Blocked returns how long the key must wait, or zero when it may try.
func (l *Limiter) Blocked(key string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if recent := l.recent(key, now); len(recent) >= l.Max {
		return recent[len(recent)-l.Max].Add(l.Window).Sub(now)
	}
	return 0
}
func (l *Limiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.fails[key] = append(l.recent(key, now), now)
}
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.fails, key)
}
