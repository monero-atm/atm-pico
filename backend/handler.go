package main

import (
	"encoding/json"

	"github.com/eclipse/paho.golang/paho"
	"github.com/rs/zerolog/log"
)

// This is a demo packet that belongs to pulseacceptord.
// We need to design a protocol.
type Message struct {
	Amount uint64
	Data   interface{}
}

// This is called when a message is received
func handleEvents(msg *paho.Publish) {
	var m Message
	if err := json.Unmarshal(msg.Payload, &m); err != nil {
		log.Error().Err(err).Str("payload", string(msg.Payload)).Msg("Message could not be parsed")
	}
	log.Info().Str("payload", string(msg.Payload)).Str("topic", msg.Topic).Msg("Received message")
}
