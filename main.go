package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/rs/zerolog/log"
)

func onConnectionUp(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
        log.Info().Msg("MQTT connection up.")
        if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
                Subscriptions: map[string]paho.SubscribeOptions{
                        "pulseacceptord":   {QoS: 2, NoLocal: true},
                        "pulseacceptord-2": {QoS: 2, NoLocal: true},
                },
        }); err != nil {
                log.Error().Err(err).Msg("Failed to subscribe. This is likely to mean no messages will be received.")
                return
        }
        log.Info().Msg("MQTT subscription made.")
}

func onConnectionError(err error) {
	log.Error().Err(err).Msg("Error whilst attempting connection.")
}

func onClientError(err error) {
	log.Error().Err(err).Msg("Server requested disconnect.")
}

func onServerDisconnect(d *paho.Disconnect) {
	if d.Properties != nil {
		log.Warn().Str("reason", d.Properties.ReasonString).Msg("Server requested disconnect.")
	} else {
		log.Warn().Str("reason_code", string(d.ReasonCode)).Msg("Server requested disconnect.")
	}
}

func main() {
	cfg := loadConfig()

	cliCfg := autopaho.ClientConfig{
		BrokerUrls: cfg.Mqtt.BrokerUrls,
		OnConnectionUp: onConnectionUp,
		OnConnectError: onConnectionError,
		ClientConfig: paho.ClientConfig{
			ClientID: cfg.Mqtt.ClientId,
			Router: paho.NewSingleHandlerRouter(func(m *paho.Publish) {
				handleEvents(m)
			}),
			OnClientError: onClientError,
			OnServerDisconnect: onServerDisconnect,
		},
	}

	// Connect to the broker.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cm, err := autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		panic(err)
	}

	// This is where we send messages to ATM components.
	// This isn't where we reply.
	go func(broker *autopaho.ConnectionManager) {
		res := "false"
		for {
			pr, err := broker.Publish(ctx, &paho.Publish{
				QoS:     2,
				Topic:   "pulseacceptord-1",
				Payload: []byte(fmt.Sprintf(`{"accept": %s}`, res)),
			})
			if err != nil {
				log.Error().Err(err).Msg("Error publishing.")
			} else if pr.ReasonCode != 0 && pr.ReasonCode != 16 { // 16 = Server received message but there are no subscribers
				log.Warn().Int("reason_code", int(pr.ReasonCode)).Msg("")
			}
			log.Info().Msg("Sent message: state change.")
			time.Sleep(10 * time.Second)

			if rand.Intn(6)%2 == 0 {
				if res == "false" {
					res = "true"
				} else {
					res = "false"
				}
			}
		}
	}(cm)

	// Messages will be handled through the callback so we really just need to wait until a shutdown.
	// is requested
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	log.Info().Msg("Signal caught - exiting.")

	// We could cancel the context at this point but will call Disconnect instead (this waits for autopaho to shutdown).
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = cm.Disconnect(ctx)

	log.Info().Msg("Shutdown complete")
}
