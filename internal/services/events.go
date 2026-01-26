package services

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/revrost/counterspell/internal/models"
)

// EventType represents the type of event being published.
type EventType string

const (
	EventTypeAgentRunStarted EventType = "agent_run_started"
	EventTypeAgentRunUpdated EventType = "agent_run_updated"
	EventTypeTaskUpdated     EventType = "task_updated"
	EventTypeTaskStarted     EventType = "task_started"
	EventTypeLog             EventType = "log"
	EventTypeAgentUpdate     EventType = "agent_update"
)

// EventBus handles pub/sub for real-time events via SSE.
type EventBus struct {
	subscribers map[chan models.Event]bool
	mu          sync.RWMutex

	// Event sequence for deduplication
	sequence int64

	// eventLog stores recent events per task for reconnection replay
	// Key: taskID, Value: slice of events (capped at maxEventsPerTask)
	eventLog   map[string][]models.Event
	eventLogMu sync.RWMutex

	// lastAgentState stores the most recent agent_update for each task
	// This is the full message history JSON for quick reconnection
	lastAgentState   map[string]string
	lastAgentStateMu sync.RWMutex
}

const (
	maxEventsPerTask = 100 // Keep last 100 events per task
	eventLogTTL      = 30 * time.Minute
)

// NewEventBus creates a new event bus.
func NewEventBus() *EventBus {
	eb := &EventBus{
		subscribers:    make(map[chan models.Event]bool),
		eventLog:       make(map[string][]models.Event),
		lastAgentState: make(map[string]string),
	}

	// Start cleanup goroutine
	go eb.cleanupLoop()

	return eb
}

// cleanupLoop periodically removes old event logs
func (b *EventBus) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		b.eventLogMu.Lock()
		// Remove event logs for tasks with no recent activity
		// In practice, logs are cleared on status_change anyway
		for taskID, events := range b.eventLog {
			if len(events) == 0 {
				delete(b.eventLog, taskID)
			}
		}
		b.eventLogMu.Unlock()
	}
}

// Publish sends an event to all subscribers.
func (b *EventBus) Publish(event models.Event) {
	// Assign sequence ID for deduplication
	event.ID = atomic.AddInt64(&b.sequence, 1)

	// Store event in log for reconnection replay
	if event.TaskID != "" {
		b.eventLogMu.Lock()
		events := b.eventLog[event.TaskID]
		events = append(events, event)
		// Cap at maxEventsPerTask
		if len(events) > maxEventsPerTask {
			events = events[len(events)-maxEventsPerTask:]
		}
		b.eventLog[event.TaskID] = events
		b.eventLogMu.Unlock()
	}

	// Cache agent_update for quick state recovery
	if event.Type == string(EventTypeAgentUpdate) && event.TaskID != "" {
		b.lastAgentStateMu.Lock()
		b.lastAgentState[event.TaskID] = event.Data
		b.lastAgentStateMu.Unlock()
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			// Channel full, skip this subscriber
			slog.Warn("Event channel full, dropping event", "subscribers", len(b.subscribers), "event_id", event.ID)
		}
	}
}

// GetLiveHistory returns the cached live message history for a task (if any).
// Returns empty string if no live history is cached.
func (b *EventBus) GetLiveHistory(taskID string) string {
	b.lastAgentStateMu.RLock()
	defer b.lastAgentStateMu.RUnlock()
	return b.lastAgentState[taskID]
}

// GetEventsSince returns all events for a task since the given event ID.
// Used for SSE reconnection to replay missed events.
func (b *EventBus) GetEventsSince(taskID string, lastEventID int64) []models.Event {
	b.eventLogMu.RLock()
	defer b.eventLogMu.RUnlock()

	events := b.eventLog[taskID]
	if len(events) == 0 {
		return nil
	}

	// Find events after lastEventID
	var result []models.Event
	for _, e := range events {
		if e.ID > lastEventID {
			result = append(result, e)
		}
	}
	return result
}

// GetLastEventID returns the most recent event ID for a task.
func (b *EventBus) GetLastEventID(taskID string) int64 {
	b.eventLogMu.RLock()
	defer b.eventLogMu.RUnlock()

	events := b.eventLog[taskID]
	if len(events) == 0 {
		return 0
	}
	return events[len(events)-1].ID
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

// Shutdown stops the event bus and cleanup goroutine.
func (b *EventBus) Shutdown() {
	slog.Info("[EVENTBUS] Shutting down")
	// Close all subscriber channels to unblock them
	b.mu.Lock()
	for ch := range b.subscribers {
		close(ch)
		delete(b.subscribers, ch)
	}
	b.mu.Unlock()
	slog.Info("[EVENTBUS] Shutdown complete")
}
