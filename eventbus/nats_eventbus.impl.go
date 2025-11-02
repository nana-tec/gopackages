package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NatsEventBus[T any] struct {
	conn *nats.Conn
	js   jetstream.JetStream
	strm jetstream.Stream
}

func NewNatsEventBus[T any](url string, appname string) (*NatsEventBus[T], error) {
	appName = appname
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	js, err := jetstream.New(nc) // creating a jetstream instance for the above created nats connection
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create jetstream context: %w", err)
	}

	// Ensure stream exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	streamConf := jetstream.StreamConfig{
		Name:        appName,
		Description: fmt.Sprintf("Stores events for %s.%s", appname, "app_events"),
		Subjects:    []string{fmt.Sprintf("%s.>", appName)}, // Subject hierarchy
	}
	stream, err := js.Stream(ctx, appName)
	if err != nil {
		stream, err := js.CreateStream(ctx, streamConf)
		if err != nil {
			nc.Close()
			return nil, fmt.Errorf("failed to create stream '%s': %w", appName, err)
		}
		return &NatsEventBus[T]{conn: nc, js: js, strm: stream}, nil
	}

	return &NatsEventBus[T]{conn: nc, js: js, strm: stream}, nil
}

func (bus *NatsEventBus[T]) Subscribe(ctx context.Context, name string, subscriber Subscriber[T]) error {
	subject := fmt.Sprintf("%s.%s", appName, name)
	cons, err := bus.strm.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       name,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: subject,
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer for subject '%s': %w", subject, err)
	}

	// Consume messages
	_, err = cons.Consume(func(jsMsg jetstream.Msg) {

		var msg Event[T]
		if err := json.Unmarshal(jsMsg.Data(), &msg); err != nil {
			fmt.Printf("Error unmarshaling message from subject '%s': %v", jsMsg.Subject(), err)
			return
		}

		// Process the message using the provided handler
		subscriber(msg)
		jsMsg.Ack()
	})
	if err != nil {
		return fmt.Errorf("failed to start consuming from subject '%s': %w", subject, err)
	}

	return nil
}

func (bus *NatsEventBus[T]) Dispatch(ctx context.Context, event Event[T]) error {

	b, err := json.Marshal(event)
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return err
	}
	// Publish the event to the 'event.Type' subject

	_, err = bus.js.Publish(ctx, event.Type, b)
	if err != nil {
		return fmt.Errorf("failed to publish message to subject '%s': %w", event.Type, err)
	}
	return nil
}

func (bus *NatsEventBus[T]) Close() {

	if bus.conn != nil {
		bus.conn.Close()
	}

}
