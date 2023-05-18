package main

import (
	"log"
	"os"
	"time"
	"net/url"

	"gopkg.in/yaml.v3"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

type brokerConfig struct {
	Brokers    []string `yaml:"brokers"`
	Topics     []string `yaml:"topics"`
	ClientId   string   `yaml:"client_id"`
	BrokerUrls []*url.URL
}

type backendConfig struct {
	Mqtt      brokerConfig `yaml:"mqtt"`
	LogFormat string       `yaml:"log_format"`
}

func loadConfig() backendConfig {
	var cfg backendConfig
	file, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Failed to read config: ", err)
	}
	if err := yaml.Unmarshal(file, &cfg); err != nil {
		log.Fatal("Failed to unmarshal yaml: ", err)
	}

	if cfg.LogFormat == "pretty" {
		zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr,
		    TimeFormat: time.RFC3339})
	}

	for _, urlStr := range cfg.Mqtt.Brokers {
		u, err := url.Parse(urlStr)
		if err != nil {
			log.Fatal("Failed to parse broker URL.")
		}
		cfg.Mqtt.BrokerUrls = append(cfg.Mqtt.BrokerUrls, u)
	}

	return cfg
}
