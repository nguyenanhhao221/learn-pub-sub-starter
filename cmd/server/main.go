package main

import (
	"fmt"
	"log"

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
	// signalCh := make(chan os.Signal, 1)
	// signal.Notify(signalCh, os.Interrupt)
	// for range signalCh {
	// 	log.Println("Shutting Down!")
	// 	os.Exit(1)
	// 	return
	// }

	publishCh, err := conn.Channel()
	if err != nil {
		log.Println("Error opening channel", err)
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
