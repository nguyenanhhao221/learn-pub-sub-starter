package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/nguyenanhhao221/learn-pub-sub-starter/internal/gamelogic"
	"github.com/nguyenanhhao221/learn-pub-sub-starter/internal/pubsub"
	"github.com/nguyenanhhao221/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	CONN_STRING := "amqp://guest:guest@localhost:5672/"
	fmt.Println("Starting Peril server...")
	conn, err := amqp.Dial(CONN_STRING)
	if err != nil {
		log.Fatalf("Error connection to rabittmq: %v", err)
	}
	defer conn.Close()
	log.Println("Connect to rabittmq successfully!")

	// Capture exit signal and print exit message
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	go func() {
		<-signalCh
		fmt.Println("\nShutdown signal received!")
		fmt.Println("Goodbye!")
		os.Exit(0)
	}()

	publishCh, err := conn.Channel()
	if err != nil {
		log.Println("Error opening channel", err)
	}

	// Declare and bind a queue to the new peril_topic exchange. It should be a durable queue named game_logs. The routing key should be game_logs.*. We'll go into detail on the routing key later.
	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+"."+"*", pubsub.SimpleQueueDurable)
	if err != nil {
		log.Fatalf("error when declare and bind to exchange: %v", err)
	}

	gamelogic.PrintServerHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) <= 0 {
			continue
		}
		switch input[0] {
		case "pause":
			log.Println("Pausing the game")
			// Publish the pausing message to the broker
			if err := pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true}); err != nil {
				log.Printf("Error PublishJSON %v", err)
			}
			log.Println("Pause message published successfully")
		case "resume":
			log.Println("Resuming the game")
			if err := pubsub.PublishJSON(publishCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false}); err != nil {
				log.Printf("Error PublishJSON %v", err)
			}
			log.Println("Resume game successfully")
		case "quit":
			log.Println("Quitting game")
			return
		default:
			log.Println("Don't understand the command")
			continue
		}
	}
}
