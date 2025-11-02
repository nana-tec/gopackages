package eventbus

import (
	"context"
	"fmt"
	"time"
)

// Event represents a generic event structure
type Event[T any] struct {
	Type      string
	Timestamp time.Time
	Data      T
}

type Subscriber[T any] func(event Event[T]) error

var appName string = "eventbus"

// EventBus is a simple event bus for publishing and subscribing to events

type EventBus[T any] interface {
	Subscribe(ctx context.Context, name string, subscriber Subscriber[T]) error
	Dispatch(ctx context.Context, event Event[T]) error
	Close()
}

func NewEvent[T any](name string, data T, timestamp time.Time) Event[T] {
	typename := fmt.Sprintf("%s.%s", appName, name)
	return Event[T]{
		Type:      typename,
		Timestamp: timestamp,
		Data:      data,
	}
}
