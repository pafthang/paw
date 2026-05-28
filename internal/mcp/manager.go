package mcp

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"sync"
	"time"
)

type Status struct {
	Name      string    `json:"name"`
	Running   bool      `json:"running"`
	PID       int       `json:"pid,omitempty"`
	StartedAt time.Time `json:"started_at,omitempty"`
	LastError string    `json:"last_error,omitempty"`
}

type Manager struct {
	mu       sync.Mutex
	procs    map[string]*exec.Cmd
	statuses map[string]Status
}

func NewManager() *Manager {
	return &Manager{
		procs:    map[string]*exec.Cmd{},
		statuses: map[string]Status{},
	}
}

func (m *Manager) Status() []Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Status, 0, len(m.statuses))
	for _, st := range m.statuses {
		out = append(out, st)
	}
	return out
}

func (m *Manager) Start(ctx context.Context, name string, cfg ServerConfig) (Status, error) {
	if err := ValidateName(name); err != nil {
		return Status{}, err
	}
	if cfg.Command == "" {
		return Status{}, errors.New("command is required")
	}

	m.mu.Lock()
	if _, ok := m.procs[name]; ok {
		st := m.statuses[name]
		m.mu.Unlock()
		return st, errors.New("already running")
	}
	m.mu.Unlock()

	cmd := exec.CommandContext(ctx, cfg.Command, cfg.Args...)
	cmd.Env = os.Environ()
	for k, v := range cfg.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	if err := cmd.Start(); err != nil {
		return Status{}, err
	}
	st := Status{Name: name, Running: true, PID: cmd.Process.Pid, StartedAt: time.Now().UTC()}

	m.mu.Lock()
	m.procs[name] = cmd
	m.statuses[name] = st
	m.mu.Unlock()

	go func() {
		err := cmd.Wait()
		m.mu.Lock()
		defer m.mu.Unlock()
		delete(m.procs, name)
		st := m.statuses[name]
		st.Running = false
		st.PID = 0
		if err != nil {
			st.LastError = err.Error()
		}
		m.statuses[name] = st
	}()

	return st, nil
}

func (m *Manager) Stop(ctx context.Context, name string) (Status, error) {
	if err := ValidateName(name); err != nil {
		return Status{}, err
	}
	m.mu.Lock()
	cmd := m.procs[name]
	st := m.statuses[name]
	m.mu.Unlock()
	if cmd == nil || cmd.Process == nil {
		return st, errors.New("not running")
	}
	_ = ctx
	_ = cmd.Process.Kill()
	return st, nil
}
