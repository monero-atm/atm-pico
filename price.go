package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

var endpoint = "https://api.kraken.com/0/public/Ticker?pair=XMREUR"

type krakenPair struct {
	Result struct {
		XmrEur struct {
			C []string `json:"c"`
		} `json:"XXMRZEUR"`
	} `json:"result"`
}

func getXmrPrice() (float64, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return 0, err
	}
	cl := &http.Client{Timeout: 15 * time.Second}
	resp, err := cl.Do(req)
	if err != nil {
		return 0, err
	}

	var kp krakenPair
	if err := json.NewDecoder(resp.Body).Decode(&kp); err != nil {
		return 0, err
	}

	priceFloat, err := strconv.ParseFloat(kp.Result.XmrEur.C[0], 64)
	if err != nil {
		return 0, err
	}
	return priceFloat, nil
}

type priceEvent float64

func pricePoll() {
	pause := false
	for {
		select {
		case p := <-pricePause:
			pause = p
		case <-time.After(cfg.PricePollFreq):
			if !pause {
				price, err := getXmrPrice()
				if err != nil {
					log.Error().Err(err).Msg("Failed to get XMR/EUR price")
				} else {
					priceUpdate <- priceEvent(price)

					log.Info().Msg("Sent price update!")
				}
			}
		}
	}
}
