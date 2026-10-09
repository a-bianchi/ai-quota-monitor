package codex

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"time"
)

var ErrBusy = errors.New("Codex account login or collection is in progress")

type State struct {
	Status          string     `json:"status"`
	VerificationURL string     `json:"verificationUrl,omitempty"`
	UserCode        string     `json:"userCode,omitempty"`
	ExpiresAt       *time.Time `json:"expiresAt,omitempty"`
}
type entry struct {
	state  State
	cancel context.CancelFunc
	busy   bool
}
type Manager struct {
	mu           sync.Mutex
	ctx          context.Context
	binary, root string
	entries      map[string]*entry
	Ready        chan string
}

func New(ctx context.Context, binary, root string) *Manager {
	return &Manager{ctx: ctx, binary: binary, root: root, entries: map[string]*entry{}, Ready: make(chan string, 64)}
}
func (m *Manager) client(id string) (Client, error) {
	home, e := PrepareHome(filepath.Join(m.root, id))
	return Client{Binary: m.binary, Home: home}, e
}
func (m *Manager) Status(id string) State {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e := m.entries[id]; e != nil {
		return e.state
	}
	return State{Status: "unknown"}
}
func (m *Manager) Read(ctx context.Context, id string) (Limits, error) {
	m.mu.Lock()
	e := m.entries[id]
	if e == nil {
		e = &entry{state: State{Status: "unknown"}}
		m.entries[id] = e
	}
	if e.busy {
		m.mu.Unlock()
		return Limits{}, ErrBusy
	}
	e.busy = true
	m.mu.Unlock()
	defer func() { m.mu.Lock(); e.busy = false; m.mu.Unlock() }()
	c, err := m.client(id)
	if err != nil {
		return Limits{}, err
	}
	limits, err := c.Read(ctx)
	m.mu.Lock()
	if err == nil {
		e.state = State{Status: "authenticated"}
	} else if errors.Is(err, ErrLogin) {
		e.state = State{Status: "auth_required"}
	}
	m.mu.Unlock()
	return limits, err
}
func (m *Manager) Start(request context.Context, id string) (State, error) {
	m.mu.Lock()
	old := m.entries[id]
	if old != nil && old.busy {
		state := old.state
		m.mu.Unlock()
		if state.Status == "pending" || state.Status == "starting" {
			return state, nil
		}
		return State{}, ErrBusy
	}
	ctx, cancel := context.WithTimeout(m.ctx, 15*time.Minute)
	e := &entry{state: State{Status: "starting"}, cancel: cancel, busy: true}
	m.entries[id] = e
	m.mu.Unlock()
	ready := make(chan struct{})
	go func() {
		defer cancel()
		fail := func() { m.mu.Lock(); e.state = State{Status: "failed"}; e.busy = false; m.mu.Unlock() }
		c, err := m.client(id)
		if err != nil {
			fail()
			close(ready)
			return
		}
		startup := time.AfterFunc(15*time.Second, cancel)
		prompt, wait, closeSession, err := c.StartDevice(ctx)
		startup.Stop()
		if err != nil {
			fail()
			close(ready)
			return
		}

		deadline, _ := ctx.Deadline()
		m.mu.Lock()
		e.state = State{Status: "pending", VerificationURL: prompt.VerificationURL, UserCode: prompt.UserCode, ExpiresAt: &deadline}
		m.mu.Unlock()
		close(ready)
		err = wait()
		closeSession()
		m.mu.Lock()
		e.busy = false
		if err == nil {
			e.state = State{Status: "authenticated"}
		} else {
			e.state = State{Status: "failed"}
		}
		m.mu.Unlock()
		if err == nil {
			select {
			case m.Ready <- id:
			default:
			}
		}
	}()
	select {
	case <-request.Done():
		return State{}, request.Err()
	case <-ready:
		return m.Status(id), nil
	}
}
