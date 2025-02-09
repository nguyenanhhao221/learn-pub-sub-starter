package main

import (
	"fmt"

	"github.com/nguyenanhhao221/learn-pub-sub-starter/internal/gamelogic"
	"github.com/nguyenanhhao221/learn-pub-sub-starter/internal/pubsub"
	"github.com/nguyenanhhao221/learn-pub-sub-starter/internal/routing"
)

func handlerLog() func(gl routing.GameLog) pubsub.AckType {
	return func(gl routing.GameLog) pubsub.AckType {
		defer fmt.Printf("> ")
		err := gamelogic.WriteLog(gl)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
