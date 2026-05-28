package channels

import (
	"context"
	"testing"
)

type fakeChannel struct {
	name    string
	running bool
}

func (f *fakeChannel) Name() string { return f.name }
func (f *fakeChannel) Start(ctx context.Context) error {
	f.running = true
	return nil
}
func (f *fakeChannel) Stop(ctx context.Context) error {
	f.running = false
	return nil
}
func (f *fakeChannel) Status() ChannelStatus {
	return ChannelStatus{Name: f.name, Running: f.running}
}

func TestManager_StartStop(t *testing.T) {
	m := NewManager()
	ch := &fakeChannel{name: "demo"}
	m.Register(ch)
	if len(m.List()) != 1 {
		t.Fatalf("list=%v", m.List())
	}
	if _, err := m.Start(context.Background(), "demo"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if !ch.running {
		t.Fatalf("expected running")
	}
	if _, err := m.Stop(context.Background(), "demo"); err != nil {
		t.Fatalf("stop: %v", err)
	}
	if ch.running {
		t.Fatalf("expected stopped")
	}
}
