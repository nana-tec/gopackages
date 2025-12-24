package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

type NatsIntergrationBroker struct {
	natsConn               *NatsConnInstance
	js                     jetstream.JetStream
	strm                   jetstream.Stream
	appname                string
	intergrationStreamSubj string
}

func NewNatsIntergrationBroker(natsConn *NatsConnInstance, appname string) (*NatsIntergrationBroker, error) {

	intergrationStreamSubj := fmt.Sprintf("%s.intergration.>", appname)
	intergrationStream := fmt.Sprintf("%s.intergration", appname)

	if natsConn.status != Active {
		return nil, fmt.Errorf("nats connection not active: %s", natsConn.status)
	}

	nc := natsConn.conn

	js, err := jetstream.New(nc) // creating a jetstream instance for the above created nats connection
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream context: %w", err)
	}

	// Ensure stream exists
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	streamConf := jetstream.StreamConfig{
		Name:        appname,
		Description: fmt.Sprintf("Stores events for %s", appname),
		Retention:   jetstream.WorkQueuePolicy,        //
		Subjects:    []string{intergrationStreamSubj}, // Subject hierarchy
	}
	stream, err := js.Stream(ctx, appname)
	if err != nil {
		stream, err := js.CreateStream(ctx, streamConf)
		if err != nil {
			nc.Close()
			return nil, fmt.Errorf("failed to create stream '%s': %w", appname, err)
		}
		return &NatsIntergrationBroker{natsConn: natsConn, js: js, strm: stream, appname: appname, intergrationStreamSubj: intergrationStream}, nil
	}

	return &NatsIntergrationBroker{natsConn: natsConn, js: js, strm: stream, appname: appname, intergrationStreamSubj: intergrationStream}, nil

}

func (ntib *NatsIntergrationBroker) Publish(ctx context.Context, pubEvent IntergrationPubEvent) error {
	// Marshal the struct into a JSON byte slice
	b, err := json.Marshal(pubEvent)
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return err
	}
	// Publish the event to the 'appname.intergration.eventname' subject
	intersub := fmt.Sprintf("%s.%s", ntib.intergrationStreamSubj, pubEvent.EventName)

	_, err = ntib.js.Publish(ctx, intersub, b)
	if err != nil {
		return fmt.Errorf("failed to publish message to subject '%s': %w", intersub, err)
	}
	return nil
}

func (ntib *NatsIntergrationBroker) Subscribe(ctx context.Context, subscriber IntergrationSubscriber) error {
	// subscriber to 'intergration.eventname'
	subject := fmt.Sprintf("%s.%s", ntib.intergrationStreamSubj, subscriber.EventName)
	cons, err := ntib.strm.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       subscriber.EventName,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: subject,
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer for subject '%s': %w", subject, err)
	}

	// Consume messages
	_, err = cons.Consume(func(jsMsg jetstream.Msg) {

		//fmt.Printf("Received message on subject %s: %s\n", jsMsg.Subject(), string(jsMsg.Data()))

		var msg IntergrationPubEvent
		// Unmarshal the JSON data into the struct address
		if err := json.Unmarshal(jsMsg.Data(), &msg); err != nil {
			fmt.Printf("Error unmarshaling message from subject '%s': %v", jsMsg.Subject(), err)
			return
		}

		// Process the message using the provided handler
		subscriber.handler(msg)
		jsMsg.Ack()
	})
	if err != nil {
		return fmt.Errorf("failed to start consuming from subject '%s': %w", subject, err)
	}

	return nil

}
