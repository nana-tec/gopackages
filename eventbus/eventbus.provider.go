package eventbus

import "context"

func NewEventBus[T any](ctx context.Context, cfg EventBusConfig) (EventBus[T], error) {

	if cfg.Provider == "nats" {
		return NewNatsEventBus[T](cfg.Url, cfg.Appname)
	}

	return NewInternalEventBus[T]()

}
