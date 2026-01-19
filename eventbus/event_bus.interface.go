package eventbus

import (
	"context"
	"time"
)

// Event represents a generic event structure
type Event struct {
	Type      string
	Timestamp time.Time
	Data      map[string]any
}

type Subscriber func(event Event) error

// EventBus is a simple event bus for publishing and subscribing to events

type EventBus interface {
	Subscribe(ctx context.Context, name string, subscriber Subscriber) error
	Dispatch(ctx context.Context, event Event) error
	Close()
}

func NewEvent(name string, data map[string]any, timestamp time.Time) Event {

	return Event{
		Type:      name,
		Timestamp: timestamp,
		Data:      data,
	}
}
