package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

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
	for range signalCh {
		log.Println("Shutting Down!")
		os.Exit(1)
		return
	}

	amqpCh, err := conn.Channel()
	if err != nil {
		log.Println("Error opening channel", err)
	}

	if err := pubsub.PublishJSON(amqpCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true}); err != nil {
		log.Printf("Error PublishJSON %v", err)
	}
}
