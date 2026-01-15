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

	// liveHistory stores the most recent message history for in-progress tasks
	// This allows SSE reconnections to get the latest state even before DB persistence
	liveHistory   map[string]string // taskID -> JSON message history
	liveHistoryMu sync.RWMutex
}

// NewEventBus creates a new event bus.
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[chan models.Event]bool),
		liveHistory: make(map[string]string),
	}
}

// Publish sends an event to all subscribers.
func (b *EventBus) Publish(event models.Event) {
	// Cache agent_update events for SSE reconnections
	if event.Type == "agent_update" && event.TaskID != "" {
		b.liveHistoryMu.Lock()
		b.liveHistory[event.TaskID] = event.HTMLPayload
		b.liveHistoryMu.Unlock()
	}

	// Clear cache when task completes
	if event.Type == "status_change" && event.TaskID != "" {
		b.liveHistoryMu.Lock()
		delete(b.liveHistory, event.TaskID)
		b.liveHistoryMu.Unlock()
	}

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

// GetLiveHistory returns the cached live message history for a task (if any).
// Returns empty string if no live history is cached.
func (b *EventBus) GetLiveHistory(taskID string) string {
	b.liveHistoryMu.RLock()
	defer b.liveHistoryMu.RUnlock()
	return b.liveHistory[taskID]
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
