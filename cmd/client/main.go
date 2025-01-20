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
	fmt.Println("Starting Peril client...")

	conn, err := amqp.Dial(CONN_STRING)
	if err != nil {
		log.Fatalf("Error connection to rabittmq: %v", err)
	}
	defer conn.Close()
	log.Println("Connect to rabittmq successfully!")

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Cannot get username: %v", err)
	}
	exchangeName := fmt.Sprintf("%s.%s", routing.PauseKey, username)
	log.Printf("Using exchange name: %q", exchangeName)
	_, queue, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, exchangeName, routing.PauseKey, 1)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
