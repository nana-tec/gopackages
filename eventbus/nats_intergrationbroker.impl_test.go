package eventbus

import (
	"context"
	"testing"
	"time"
)

func TestNatsEIntergrationBroker(t *testing.T) {

	//rootCtx := context.Background()

	var natsConf = &NatsConfig{
		natsUrl:             "nats://localhost:4222",
		appName:             "eventbus",
		requiresCredentials: false,
		username:            "",
		password:            "",
	}
	println("Connecting to nats ")
	bus, err := NewNatsConnection(*natsConf)
	if err != nil {
		t.Fatalf("Failed to create nats connection : %v", err)
	}

	println("Nats connected ")

	// get status of connection
	status := bus.Status()
	if status != Active {
		t.Errorf("Expected active status, got %s", status)
	}

	println("Connection active ")

	natsbroker, errr := NewNatsIntergrationBroker(bus, "testeventbus")
	if errr != nil {
		t.Fatalf("Failed to create nats connection : %v", err)
	}
	println("Nats Intergration Broker created  ")
	// setup subscriber

	subscriberCalled := false
	err = natsbroker.Subscribe(context.Background(), IntergrationSubscriber{
		EventName:      "testevent",
		SubscriberName: "testsubscriber",
		handler: func(event IntergrationPubEvent) error {
			println("Event Recieved ")
			subscriberCalled = true
			return nil
		},
	})

	// publish an event

	pubEvent := IntergrationPubEvent{
		EventName:          "testevent",
		EventData:          map[string]any{"myname": "testdata"},
		EventTimestamp:     time.Now(),
		EventPublisherName: "testpublisher",
	}
	err = natsbroker.Publish(context.Background(), pubEvent)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	println("Test Event published ")

	time.Sleep(3 * time.Second) // wait for the message to be processed

	if !subscriberCalled {
		t.Error("Subscriber was not called")
	}

}
