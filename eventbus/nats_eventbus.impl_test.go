package eventbus

import (
	"context"
	"testing"
	"time"
)

func TestNatsEventBus(t *testing.T) {
	// Create a root context
	rootCtx := context.Background()
	bus, err := NewNatsEventBus[string]("nats://localhost:4222", "teststream1")
	if err != nil {
		t.Fatalf("Failed to create event bus: %v", err)
	}
	defer bus.Close()

	subscriberCalled := false
	subscriber := func(event Event[string]) error {
		subscriberCalled = true
		if event.Data != "testdata" {
			t.Errorf("Expected event data 'testdata', got '%s'", event.Data)
		}
		return nil
	}
	//timeoutCtx, cancelTimeout := context.WithTimeout(rootCtx, 15*time.Second)
	//defer cancelTimeout() // Ensure resources are released

	err = bus.Subscribe(rootCtx, "testevent", subscriber)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	event := NewEvent("testevent", "testdata", time.Now())

	err = bus.Dispatch(rootCtx, event)
	if err != nil {
		t.Fatalf("Failed to dispatch event: %v", err)
	}
	time.Sleep(2 * time.Second) // wait for the message to be processed

	if !subscriberCalled {
		t.Error("Subscriber was not called")
	}
}
