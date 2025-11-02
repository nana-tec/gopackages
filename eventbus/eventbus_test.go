package eventbus

import (
	"testing"
	"time"
)

func TestEventBus(t *testing.T) {
	bus := NewInternalEventBus[string]()

	subscriberCalled := false
	subscriber := func(event Event[string]) error {
		subscriberCalled = true
		if event.Data != "testdata" {
			t.Errorf("Expected event data 'testdata', got '%s'", event.Data)
		}
		return nil
	}

	err := bus.Subscribe("testevent", subscriber)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	event := Event[string]{
		Type:      "testevent",
		Timestamp: time.Now(),
		Data:      "testdata",
	}

	err = bus.Dispatch(event)
	if err != nil {
		t.Fatalf("Failed to dispatch event: %v", err)
	}

	if !subscriberCalled {
		t.Error("Subscriber was not called")
	}

}
