package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

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
}
