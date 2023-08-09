package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

var krakenUrl = "https://api.kraken.com/0/public/Ticker?pair=XMR"

type krakenPair struct {
	Result struct {
		XmrEur struct {
			C []string `json:"c"`
		} `json:"XXMRZEUR"`
		XmrUsd struct {
			C []string `json:"c"`
		} `json:"XXMRZUSD"`
	} `json:"result"`
}

// Kraken only has XMR/EUR and XMR/USD pairs. When another fiat currency is
// specified, this function will calculate the value based on the daily rate
// provided by European Central Bank. "fiatEurRate" contains this rate.
// It's ignored when "currency" is set to EUR or USD.
func getXmrPrice(currency string, fiatEurRate float64) (float64, error) {
	// Target fiat currency to get rate of from Kraken
	target := "EUR"
	if currency == "USD" {
		target = "USD"
	}

	req, err := http.NewRequest("GET", krakenUrl+target, nil)
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

	var rate string
	if target == "EUR" {
		rate = kp.Result.XmrEur.C[0]
	} else {
		rate = kp.Result.XmrUsd.C[0]
	}

	rateFloat, err := strconv.ParseFloat(rate, 64)
	if err != nil {
		return 0, err
	}

	if currency == "EUR" || currency == "USD" {
		return rateFloat, nil
	}

	return rateFloat * fiatEurRate, nil
}

type priceEvent float64

func pricePoll(currency string, fiatEurRate float64) {
	pause := false
	for {
		select {
		case p := <-pricePause:
			pause = p
		case <-time.After(cfg.PricePollFreq):
			if !pause {
				price, err := getXmrPrice(currency, fiatEurRate)
				if err != nil {
					log.Error().Err(err).Msg("Failed to get XMR price")
				} else {
					priceUpdate <- priceEvent(price)

					log.Info().Msg("Sent price update!")
				}
			}
		}
	}
}
