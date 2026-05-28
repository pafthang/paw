package channels

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Manager struct {
	mu       sync.Mutex
	channels map[string]Channel
	status   map[string]ChannelStatus
}

func NewManager() *Manager {
	return &Manager{
		channels: map[string]Channel{},
		status:   map[string]ChannelStatus{},
	}
}

func (m *Manager) Register(ch Channel) {
	if ch == nil || ch.Name() == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels[ch.Name()] = ch
	if _, ok := m.status[ch.Name()]; !ok {
		m.status[ch.Name()] = ChannelStatus{Name: ch.Name()}
	}
}

func (m *Manager) List() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	names := make([]string, 0, len(m.channels))
	for name := range m.channels {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (m *Manager) StatusAll() []ChannelStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]ChannelStatus, 0, len(m.status))
	for name, st := range m.status {
		if ch, ok := m.channels[name]; ok {
			st = ch.Status()
		}
		out = append(out, st)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (m *Manager) Start(ctx context.Context, name string) (ChannelStatus, error) {
	m.mu.Lock()
	ch, ok := m.channels[name]
	m.mu.Unlock()
	if !ok {
		return ChannelStatus{}, fmt.Errorf("unknown channel %q", name)
	}
	err := ch.Start(ctx)
	st := ch.Status()
	if err != nil {
		st.LastError = err.Error()
		m.setStatus(st)
		return st, err
	}
	st.Running = true
	if st.StartedAt.IsZero() {
		st.StartedAt = time.Now().UTC()
	}
	m.setStatus(st)
	return st, nil
}

func (m *Manager) Stop(ctx context.Context, name string) (ChannelStatus, error) {
	m.mu.Lock()
	ch, ok := m.channels[name]
	m.mu.Unlock()
	if !ok {
		return ChannelStatus{}, fmt.Errorf("unknown channel %q", name)
	}
	err := ch.Stop(ctx)
	st := ch.Status()
	if err != nil {
		st.LastError = err.Error()
		m.setStatus(st)
		return st, err
	}
	st.Running = false
	m.setStatus(st)
	return st, nil
}

func (m *Manager) setStatus(st ChannelStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.status[st.Name] = st
}
