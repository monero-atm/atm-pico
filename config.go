package main

import (
	"log"
	"net/url"
	"os"
	"time"

	"github.com/eclipse/paho.golang/paho"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type brokerConfig struct {
	Brokers       []string `yaml:"brokers"`
	Topics        []string `yaml:"topics"`
	ClientId      string   `yaml:"client_id"`
	BrokerUrls    []*url.URL
	Subscriptions map[string]paho.SubscribeOptions
}

type backendConfig struct {
	Mqtt          brokerConfig  `yaml:"mqtt"`
	Mode          string        `yaml:"mode"`
	LogFormat     string        `yaml:"log_format"`
	LogFile       string        `yaml:"log_file"`
	Fee           float64       `yaml:"fee"`
	Moneropay     string        `yaml:"moneropay"`
	PricePollFreq time.Duration `yaml:"price_poll_frequency"`
	CurrencyShort string        `yaml:"currency_short"`
	Motd          string        `yaml:"motd"`
}

func loadConfig() backendConfig {
	var cfg backendConfig
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./atm-pico config.yaml")
	}
	file, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal("Failed to read config: ", err)
	}
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		log.Fatal("Failed to unmarshal yaml: ", err)
	}

	f, err := os.OpenFile(cfg.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}

	if cfg.LogFormat == "pretty" {
		zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: f,
			TimeFormat: time.RFC3339})
	}

	for _, urlStr := range cfg.Mqtt.Brokers {
		u, err := url.Parse(urlStr)
		if err != nil {
			log.Fatal("Failed to parse broker URL.")
		}
		cfg.Mqtt.BrokerUrls = append(cfg.Mqtt.BrokerUrls, u)
	}

	cfg.Mqtt.Subscriptions = make(map[string]paho.SubscribeOptions)
	for _, topic := range cfg.Mqtt.Topics {
		cfg.Mqtt.Subscriptions[topic] = paho.SubscribeOptions{QoS: 2, NoLocal: true}
	}

	return cfg
}
