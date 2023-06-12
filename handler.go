package main

import (
	"encoding/json"

	"github.com/eclipse/paho.golang/paho"
	"github.com/rs/zerolog/log"
	"gitlab.com/openkiosk/proto"
)

func handleEvents(msg *paho.Publish) {
	var m proto.Event
	if err := json.Unmarshal(msg.Payload, &m); err != nil {
		log.Error().Err(err).Str("payload", string(msg.Payload)).Msg("Message could not be parsed")
	}
	log.Info().Str("payload", string(msg.Payload)).Str("topic", msg.Topic).Msg("Received event")
	sub <- m
}
