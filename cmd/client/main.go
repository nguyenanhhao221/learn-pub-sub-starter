package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/nguyenanhhao221/learn-pub-sub-starter/internal/gamelogic"
	"github.com/nguyenanhhao221/learn-pub-sub-starter/internal/pubsub"
	"github.com/nguyenanhhao221/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

const CONN_STRING = "amqp://guest:guest@localhost:5672/"

func main() {
	fmt.Println("Starting Peril client...")

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go func() {
		<-signalChan
		fmt.Println("\nShutdown signal received!")
		fmt.Println("Goodbye!")
		os.Exit(0)
	}()

	conn, err := amqp.Dial(CONN_STRING)
	if err != nil {
		log.Fatalf("Error connection to rabittmq: %v", err)
	}
	defer conn.Close()

	log.Println("Connect to rabittmq successfully!")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Cannot get username: %v", err)
	}
	exchangeName := fmt.Sprintf("%s.%s", routing.PauseKey, username)
	log.Printf("Using exchange name: %q", exchangeName)
	_, pauseQueue, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, exchangeName, routing.PauseKey, pubsub.SimpleQueueTransient)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", pauseQueue.Name)

	// Creating a new game state
	gs := gamelogic.NewGameState(username)

	// Subscribe to the ExchangePerilDirectDirect and Pause queue
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, pauseQueue.Name, routing.PauseKey, pubsub.SimpleQueueTransient, handlerPause(gs))
	if err != nil {
		log.Fatalf("fail to SubscribeJSON : %v", err)
	}

	// Declare Subscribe to the ExchangePerilTopic and the move queue
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+gs.GetUsername(), routing.ArmyMovesPrefix+".*", pubsub.SimpleQueueTransient, handlerMove(gs, publishCh))
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	// Declare Subscribe to the ExchangePerilTopic and the war queue
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+".*", pubsub.SimpleQueueDurable, handlerWar(gs, publishCh))
	if err != nil {
		log.Fatalf("could not subscribe to war queue: %v", err)
	}

	// keep the main process running
	for {
		input := gamelogic.GetInput()
		if len(input) <= 0 {
			continue
		}
		log.Println("User input:", input)
		switch input[0] {
		case "spawn":
			err := gs.CommandSpawn(input)
			if err != nil {
				log.Printf("Error: %v", err)
			}
		case "move":
			move, err := gs.CommandMove(input)
			if err != nil {
				log.Printf("Error: %v", err)
			}
			err = pubsub.PublishJSON(publishCh, string(routing.ExchangePerilTopic), string(routing.ArmyMovesPrefix)+"."+move.Player.Username, move)
			if err != nil {
				log.Printf("Error: %v\n", err)
				continue
			}
			log.Printf("Moved %v units to %q \n", len(move.Units), move.ToLocation)

		case "status":
			gs.CommandStatus()
		case "spam":
			log.Println("spam command")
			if len(input) < 2 {
				log.Println("spam command  expect more more input")
			}
			timesOfSpam := input[1]
			n, err := strconv.Atoi(timesOfSpam)
			if err != nil {
				log.Printf("Error: %v", err)
			}
			for i := 0; i < n; i++ {
				msg := gamelogic.GetMaliciousLog()
				err := publishGameLog(publishCh, username, msg)
				if err != nil {
					log.Printf("Error: %v\n", err)
				}
			}
			log.Printf("Published %v malicious logs\n", n)
		case "help":
			gamelogic.PrintClientHelp()
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			log.Println("Don't understand the command")
			continue
		}
	}
}

func publishGameLog(publishCh *amqp.Channel, username, msg string) error {
	return pubsub.PublishGob(
		publishCh,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+username,
		routing.GameLog{
			Username:    username,
			CurrentTime: time.Now(),
			Message:     msg,
		},
	)
}
