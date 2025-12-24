package eventbus

import (
	"context"
	"sync"
)

type InternalEventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]Subscriber
}

func NewInternalEventBus() (*InternalEventBus, error) {
	return &InternalEventBus{
		subscribers: make(map[string][]Subscriber),
	}, nil
}

// Subscribe adds a subscriber to the given event name. The subscriber will be
// called with the published event when the event is published to the
// event bus. The subscriber must be safe to be called concurrently.
// The subscriber will not be called if the event is published after the
// subscriber is unsubscribed or the event bus is closed.
func (bus *InternalEventBus) Subscribe(ctx context.Context, name string, subscriber Subscriber) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.subscribers[name] = append(bus.subscribers[name], subscriber)
	return nil
}

func (bus *InternalEventBus) Dispatch(ctx context.Context, event Event) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	for _, subscriber := range bus.subscribers[event.Type] {
		if err := subscriber(event); err != nil {
			return err
		}
	}
	return nil
}

func (bus *InternalEventBus) Close() {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.subscribers = make(map[string][]Subscriber)
}
