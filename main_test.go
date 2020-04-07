package main

import (
	"context"
	"testing"

	"github.com/streadway/amqp"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/rabbitpubsub"
)

func TestRabbit(t *testing.T) {

	ctx := context.Background()
	rabbitConn, err := amqp.Dial("amqp://local:locallocallocal@localhost:5672/")
	if err != nil {
		t.Fatal(err)
	}
	defer rabbitConn.Close()
	subscription := rabbitpubsub.OpenSubscription(rabbitConn, "atl.orthanc_download", nil)
	defer subscription.Shutdown(ctx)

	topic := rabbitpubsub.OpenTopic(rabbitConn, "atl.orthanc_download", nil)
	defer topic.Shutdown(ctx)
	if err = topic.Send(ctx, &pubsub.Message{
		Body: []byte("Hello, World!\n"),
		// Metadata is optional and can be nil.
		Metadata: map[string]string{
			// These are examples of metadata.
			// There is nothing special about the key names.
			"language":   "en",
			"importance": "high",
		},
		BeforeSend: func(asFunc func(interface{}) bool) error {
			var pub *amqp.Publishing
			ok := asFunc(&pub)
			if !ok {
				t.Fatal("expected ok")
			}

			pub.DeliveryMode = 2

			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}

	msg, err := subscription.Receive(ctx)
	if err != nil {
		t.Fatal(err)
	}

	defer msg.Ack()
	t.Logf("%+v", msg)
}
