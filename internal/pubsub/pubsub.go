package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	// Create a channel to start the process
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	args := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}
	queue, err := ch.QueueDeclare(
		queueName,                             // name
		simpleQueueType == SimpleQueueDurable, // durable
		simpleQueueType != SimpleQueueDurable, // delete when unused
		simpleQueueType != SimpleQueueDurable, // exclusive
		false,                                 // no-wait
		args,                                  // arguments
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not declare queue: %v", err)
	}

	// Bind the queue to an exchange
	err = ch.QueueBind(
		queue.Name, // queue name
		key,        // routing key
		exchange,   // exchange
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not bind queue: %v", err)
	}
	return ch, queue, nil
}

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	ctx := context.Background()

	jsonVal, err := json.Marshal(val)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{Body: jsonVal, ContentType: "application/json"}
	return ch.PublishWithContext(ctx, exchange, key, false, false, msg)
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType, handler func(T) AckType) error {
	// Make sure the queue exists and bound to the exchange
	amqpCh, amqpQueue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		return err
	}

	deliveryCh, err := amqpCh.Consume(amqpQueue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for delivery := range deliveryCh {
			body := delivery.Body
			var msg T
			err := json.Unmarshal(body, &msg)
			if err != nil {
				fmt.Printf("failed to unmarshal message: %v\n", err)
				continue
			}
			// Call the provided handler with the unmarshaled message
			actType := handler(msg)
			// Acknowledge the message
			switch actType {
			case Ack:
				log.Println("Ack")
				err = delivery.Ack(false)
				if err != nil {
					fmt.Printf("failed to acknowledge message: %v\n", err)
				}
			case NackRequeue:
				log.Println("Nack re queue")
				err = delivery.Nack(false, true)
				if err != nil {
					fmt.Printf("failed to nack and re queue message: %v\n", err)
				}
			case NackDiscard:
				log.Println("Nack discard")
				err = delivery.Nack(false, false)
				if err != nil {
					fmt.Printf("failed to nack and re queue message: %v\n", err)
				}
			}
		}
	}()

	return nil
}
