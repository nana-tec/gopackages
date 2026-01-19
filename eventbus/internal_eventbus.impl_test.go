package eventbus

import (
	"context"
	"testing"
	"time"
)

func TestInternalEventBus(t *testing.T) {
	rootCtx := context.Background()

	println("Running internal event bus test")
	bus, err := NewInternalEventBus()

	if err != nil {
		t.Fatalf("Failed to start internal event bus: %v", err)
	}

	subscriberCalled := false
	subscriber := func(event Event) error {
		subscriberCalled = true
		return nil
	}

	err = bus.Subscribe(rootCtx, "testevent", subscriber)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	event := Event{
		Type:      "testevent",
		Timestamp: time.Now(),
		Data:      map[string]any{"myname": "testname"},
	}

	err = bus.Dispatch(rootCtx, event)
	if err != nil {
		t.Fatalf("Failed to dispatch event: %v", err)
	}

	if !subscriberCalled {
		t.Error("Subscriber was not called")
	}

	println("Internal event test finished")

}
