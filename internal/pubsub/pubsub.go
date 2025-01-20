package pubsub

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	ctx := context.Background()

	jsonVal, err := json.Marshal(val)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{Body: jsonVal, ContentType: "application/json"}
	return ch.PublishWithContext(ctx, exchange, key, false, false, msg)
}
