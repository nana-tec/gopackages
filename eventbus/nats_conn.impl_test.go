package eventbus

import (
	"testing"
)

func TestNatsConnection(t *testing.T) {

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

}
