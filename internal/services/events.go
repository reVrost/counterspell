package services

import (
	"log/slog"
	"sync"

	"github.com/revrost/code/counterspell/internal/models"
)

// EventBus handles pub/sub for real-time events via SSE.
type EventBus struct {
	subscribers map[chan models.Event]bool
	mu          sync.RWMutex
}

// NewEventBus creates a new event bus.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[chan models.Event]bool),
	}
}

// Publish sends an event to all subscribers.
func (b *EventBus) Publish(event models.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			// Channel full, skip this subscriber
			slog.Warn("Event channel full, dropping event", "subscribers", len(b.subscribers))
		}
	}
}

// Subscribe adds a new subscriber and returns the channel.
func (b *EventBus) Subscribe() chan models.Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan models.Event, 100)
	b.subscribers[ch] = true
	return ch
}

// Unsubscribe removes a subscriber.
func (b *EventBus) Unsubscribe(ch chan models.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
}
