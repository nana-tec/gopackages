package eventbus

import (
	"context"
	"sync"
)

type InternalEventBus[T any] struct {
	mu          sync.RWMutex
	subscribers map[string][]Subscriber[T]
}

func NewInternalEventBus[T any]() (*InternalEventBus[T], error) {
	return &InternalEventBus[T]{
		subscribers: make(map[string][]Subscriber[T]),
	}, nil
}
func (bus *InternalEventBus[T]) Subscribe(ctx context.Context, name string, subscriber Subscriber[T]) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.subscribers[name] = append(bus.subscribers[name], subscriber)
	return nil
}

func (bus *InternalEventBus[T]) Dispatch(ctx context.Context, event Event[T]) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	for _, subscriber := range bus.subscribers[event.Type] {
		if err := subscriber(event); err != nil {
			return err
		}
	}
	return nil
}

func (bus *InternalEventBus[T]) Close() {
	bus.mu.Lock()
	defer bus.mu.Unlock()
	bus.subscribers = make(map[string][]Subscriber[T])
}
