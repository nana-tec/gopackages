package eventbus

import (
	"context"
	"time"
)

type IntergrationPubEvent struct {
	EventName          string
	EventTimestamp     time.Time
	EventData          map[string]any
	EventPublisherName string
}
type IntergrationSubscriber struct {
	SubscriberName string
	EventName      string
	handler        func(event IntergrationPubEvent) error
}

type IntergrationEventBroker interface {
	Publish(ctx context.Context, pubEvent IntergrationPubEvent) error
	Subscribe(ctx context.Context, subscriber IntergrationSubscriber) error
}

// idea save event on intergrationQueue before publishing ...and on msg processed by consumer update

type IntergrationEventRepo interface {
	SaveEvent(event IntergrationPubEvent) error
}
