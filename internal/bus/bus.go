package bus

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// MessageBus routes messages between connectors and the agent runtime,
// and broadcasts in-process events to subscribers.
type MessageBus struct {
	inbound  chan InboundMessage
	outbound chan OutboundMessage

	// Connector message handlers: connector/channel name -> handler.
	handlers map[string]MessageHandler
	handlerMu sync.RWMutex

	// Event subscribers: subscriber ID -> handler.
	subscribers map[string]EventHandler
	subMu       sync.RWMutex

	closeOnce sync.Once
}

func New() *MessageBus {
	return &MessageBus{
		inbound:     make(chan InboundMessage, 1000),
		outbound:    make(chan OutboundMessage, 1000),
		handlers:    make(map[string]MessageHandler),
		subscribers: make(map[string]EventHandler),
	}
}

// PublishInbound queues an inbound message from a connector/channel.
// It blocks if the inbound buffer is full.
func (mb *MessageBus) PublishInbound(msg InboundMessage) {
	mb.inbound <- msg
}

// TryPublishInbound attempts to queue an inbound message without blocking.
// It returns false if the inbound buffer is full and the message was dropped.
func (mb *MessageBus) TryPublishInbound(msg InboundMessage) bool {
	select {
	case mb.inbound <- msg:
		return true
	default:
		return false
	}
}

// ConsumeInbound blocks until an inbound message is available or ctx is cancelled.
func (mb *MessageBus) ConsumeInbound(ctx context.Context) (InboundMessage, bool) {
	select {
	case msg, ok := <-mb.inbound:
		if !ok {
			return InboundMessage{}, false
		}
		return msg, true
	case <-ctx.Done():
		return InboundMessage{}, false
	}
}

// PublishOutbound queues an outbound message to a connector/channel.
// It blocks if the outbound buffer is full.
func (mb *MessageBus) PublishOutbound(msg OutboundMessage) {
	mb.outbound <- msg
}

// TryPublishOutbound attempts to queue an outbound message without blocking.
// It returns false if the outbound buffer is full and the message was dropped.
func (mb *MessageBus) TryPublishOutbound(msg OutboundMessage) bool {
	select {
	case mb.outbound <- msg:
		return true
	default:
		return false
	}
}

// SubscribeOutbound blocks until an outbound message is available or ctx is cancelled.
func (mb *MessageBus) SubscribeOutbound(ctx context.Context) (OutboundMessage, bool) {
	select {
	case msg, ok := <-mb.outbound:
		if !ok {
			return OutboundMessage{}, false
		}
		return msg, true
	case <-ctx.Done():
		return OutboundMessage{}, false
	}
}

// RegisterHandler registers a message handler for a connector/channel.
func (mb *MessageBus) RegisterHandler(channel string, handler MessageHandler) {
	mb.handlerMu.Lock()
	defer mb.handlerMu.Unlock()
	mb.handlers[channel] = handler
}

// GetHandler returns the message handler for a connector/channel.
func (mb *MessageBus) GetHandler(channel string) (MessageHandler, bool) {
	mb.handlerMu.RLock()
	defer mb.handlerMu.RUnlock()
	handler, ok := mb.handlers[channel]
	return handler, ok
}

// Subscribe registers an event subscriber.
func (mb *MessageBus) Subscribe(id string, handler EventHandler) {
	mb.subMu.Lock()
	defer mb.subMu.Unlock()
	mb.subscribers[id] = handler
}

// Unsubscribe removes an event subscriber.
func (mb *MessageBus) Unsubscribe(id string) {
	mb.subMu.Lock()
	defer mb.subMu.Unlock()
	delete(mb.subscribers, id)
}

// Broadcast sends an event to all subscribers.
// Panicking handlers are caught and logged so one bad subscriber cannot crash PAW.
func (mb *MessageBus) Broadcast(event Event) {
	mb.subMu.RLock()
	handlers := make(map[string]EventHandler, len(mb.subscribers))
	for id, handler := range mb.subscribers {
		handlers[id] = handler
	}
	mb.subMu.RUnlock()

	for id, handler := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("bus: subscriber panicked",
						"subscriber", id,
						"event", event.Name,
						"panic", fmt.Sprint(r),
					)
				}
			}()
			handler(event)
		}()
	}
}

// Close shuts down the message bus.
func (mb *MessageBus) Close() {
	mb.closeOnce.Do(func() {
		close(mb.inbound)
		close(mb.outbound)
	})
}
