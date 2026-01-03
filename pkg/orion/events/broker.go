package events

import (
	"sync"
)

// Event types used throughout the orion package.
const (
	// CreatedEvent indicates that an item was created.
	CreatedEvent = "created"

	// UpdatedEvent indicates that an item was updated.
	UpdatedEvent = "updated"

	// DeletedEvent indicates that an item was deleted.
	DeletedEvent = "deleted"

	// ErrorEvent indicates that an error occurred.
	ErrorEvent = "error"
)

// Broker implements a thread-safe pub/sub event distribution system.
// It allows components to subscribe to events and be notified
// when state changes occur.
type Broker[T any] struct {
	mu          sync.RWMutex
	subscribers []func(event string, data T)
}

// NewBroker creates a new event broker.
func NewBroker[T any]() *Broker[T] {
	return &Broker[T]{
		subscribers: make([]func(event string, data T), 0),
	}
}

// Publish sends an event to all registered subscribers.
// The event is dispatched in a separate goroutine to avoid
// blocking the publisher.
func (b *Broker[T]) Publish(event string, data T) {
	b.mu.RLock()
	subs := b.subscribers
	b.mu.RUnlock()

	for _, sub := range subs {
		go sub(event, data)
	}
}

// Subscribe adds a subscriber function that will be called
// for all published events.
func (b *Broker[T]) Subscribe(subscriber func(event string, data T)) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers = append(b.subscribers, subscriber)
}

// Clear removes all subscribers from the broker.
func (b *Broker[T]) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.subscribers = make([]func(event string, data T), 0)
}

// Count returns the number of registered subscribers.
func (b *Broker[T]) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subscribers)
}
