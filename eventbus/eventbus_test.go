package eventbus

import (
	"context"
	"testing"
	"time"
)

func TestEventBus(t *testing.T) {
	rootCtx := context.Background()
	bus, err := NewInternalEventBus[string]()

	if err != nil {
		t.Fatalf("Failed to start internal event bus: %v", err)
	}

	subscriberCalled := false
	subscriber := func(event Event[string]) error {
		subscriberCalled = true
		if event.Data != "testdata" {
			t.Errorf("Expected event data 'testdata', got '%s'", event.Data)
		}
		return nil
	}

	err = bus.Subscribe(rootCtx, "testevent", subscriber)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	event := Event[string]{
		Type:      "testevent",
		Timestamp: time.Now(),
		Data:      "testdata",
	}

	err = bus.Dispatch(rootCtx, event)
	if err != nil {
		t.Fatalf("Failed to dispatch event: %v", err)
	}

	if !subscriberCalled {
		t.Error("Subscriber was not called")
	}

}
