package eventbus

import "context"

func NewEventBus(ctx context.Context) (EventBus, error) {

	return NewInternalEventBus()

}
