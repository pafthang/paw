package bus

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPublishInboundConsumeInbound(t *testing.T) {
	mb := New()
	defer mb.Close()

	msg := InboundMessage{Channel: "telegram", Content: "hello"}
	mb.PublishInbound(msg)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got, ok := mb.ConsumeInbound(ctx)
	if !ok {
		t.Fatal("expected to consume message")
	}
	if got.Content != "hello" {
		t.Fatalf("content mismatch: got %q, want %q", got.Content, "hello")
	}
}

func TestPublishOutboundSubscribeOutbound(t *testing.T) {
	mb := New()
	defer mb.Close()

	msg := OutboundMessage{Channel: "web", Content: "world"}
	mb.PublishOutbound(msg)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got, ok := mb.SubscribeOutbound(ctx)
	if !ok {
		t.Fatal("expected to receive message")
	}
	if got.Content != "world" {
		t.Fatalf("content mismatch: got %q, want %q", got.Content, "world")
	}
}

func TestConsumeInboundContextCancelled(t *testing.T) {
	mb := New()
	defer mb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, ok := mb.ConsumeInbound(ctx)
	if ok {
		t.Fatal("expected false on cancelled context")
	}
}

func TestTryPublishInboundBufferFull(t *testing.T) {
	mb := &MessageBus{
		inbound:     make(chan InboundMessage, 1),
		outbound:    make(chan OutboundMessage, 1),
		handlers:    make(map[string]MessageHandler),
		subscribers: make(map[string]EventHandler),
	}

	if !mb.TryPublishInbound(InboundMessage{Content: "1"}) {
		t.Fatal("first message should succeed")
	}
	if mb.TryPublishInbound(InboundMessage{Content: "2"}) {
		t.Fatal("second message should fail because buffer is full")
	}
}

func TestTryPublishOutboundBufferFull(t *testing.T) {
	mb := &MessageBus{
		inbound:     make(chan InboundMessage, 1),
		outbound:    make(chan OutboundMessage, 1),
		handlers:    make(map[string]MessageHandler),
		subscribers: make(map[string]EventHandler),
	}

	if !mb.TryPublishOutbound(OutboundMessage{Content: "1"}) {
		t.Fatal("first message should succeed")
	}
	if mb.TryPublishOutbound(OutboundMessage{Content: "2"}) {
		t.Fatal("second message should fail because buffer is full")
	}
}

func TestBroadcastDeliveredToAllSubscribers(t *testing.T) {
	mb := New()
	defer mb.Close()

	var count atomic.Int32
	mb.Subscribe("sub1", func(e Event) { count.Add(1) })
	mb.Subscribe("sub2", func(e Event) { count.Add(1) })
	mb.Subscribe("sub3", func(e Event) { count.Add(1) })

	mb.Broadcast(Event{Name: "test"})

	if got := count.Load(); got != 3 {
		t.Fatalf("expected 3 deliveries, got %d", got)
	}
}

func TestBroadcastPanickingHandlerDoesNotCrashBus(t *testing.T) {
	mb := New()
	defer mb.Close()

	var delivered atomic.Int32

	mb.Subscribe("panicker", func(e Event) {
		panic("subscriber exploded")
	})
	mb.Subscribe("normal", func(e Event) {
		delivered.Add(1)
	})

	mb.Broadcast(Event{Name: "test"})
	mb.Broadcast(Event{Name: "test2"})

	if got := delivered.Load(); got == 0 {
		t.Fatal("normal handler should have been called at least once")
	}
}

func TestUnsubscribeStopsDelivery(t *testing.T) {
	mb := New()
	defer mb.Close()

	var count atomic.Int32
	mb.Subscribe("temp", func(e Event) { count.Add(1) })

	mb.Broadcast(Event{Name: "before"})
	if count.Load() != 1 {
		t.Fatal("expected delivery before unsubscribe")
	}

	mb.Unsubscribe("temp")
	mb.Broadcast(Event{Name: "after"})
	if count.Load() != 1 {
		t.Fatal("expected no delivery after unsubscribe")
	}
}

func TestSubscribeOverwritesPrevious(t *testing.T) {
	mb := New()
	defer mb.Close()

	var first, second atomic.Int32
	mb.Subscribe("id1", func(e Event) { first.Add(1) })
	mb.Subscribe("id1", func(e Event) { second.Add(1) })

	mb.Broadcast(Event{Name: "test"})
	if first.Load() != 0 {
		t.Fatal("first handler should have been replaced")
	}
	if second.Load() != 1 {
		t.Fatal("second handler should have been called")
	}
}

func TestRegisterHandlerGetHandler(t *testing.T) {
	mb := New()
	defer mb.Close()

	called := false
	mb.RegisterHandler("telegram", func(msg InboundMessage) error {
		called = true
		return nil
	})

	handler, ok := mb.GetHandler("telegram")
	if !ok {
		t.Fatal("expected handler to be registered")
	}
	_ = handler(InboundMessage{})
	if !called {
		t.Fatal("expected handler to be called")
	}

	_, ok = mb.GetHandler("nonexistent")
	if ok {
		t.Fatal("expected no handler for unregistered channel")
	}
}

func TestBroadcastConcurrentSubscribeUnsubscribe(t *testing.T) {
	mb := New()
	defer mb.Close()

	var wg sync.WaitGroup
	done := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-done:
				return
			default:
				mb.Broadcast(Event{Name: "concurrent"})
			}
		}
	}()

	for range 100 {
		mb.Subscribe("rapid", func(e Event) {})
		mb.Unsubscribe("rapid")
	}

	close(done)
	wg.Wait()
}

func TestPublishInboundConcurrentProducers(t *testing.T) {
	mb := New()
	defer mb.Close()

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for range n {
		go func() {
			defer wg.Done()
			mb.TryPublishInbound(InboundMessage{Content: "msg"})
		}()
	}
	wg.Wait()
}
