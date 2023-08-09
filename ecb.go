package main

import (
	"encoding/xml"
	"net/http"
	"time"
)

type ecbRates struct {
	Rates []struct {
		Currency string `xml:"currency,attr"`
		Rate     string `xml:"rate,attr"`
	} `xml:"Cube>Cube>Cube"`
}

const ecbDailyUrl = "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml"

// Get daily conversion rates from ECB
func fetchEcbDaily() (map[string]string, error) {
	req, err := http.NewRequest("GET", ecbDailyUrl, nil)
	if err != nil {
		return nil, err
	}
	cl := &http.Client{Timeout: 15 * time.Second}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rates ecbRates
	if err := xml.NewDecoder(resp.Body).Decode(&rates); err != nil {
		return nil, err
	}

	mappedRates := make(map[string]string)
	for _, r := range rates.Rates {
		mappedRates[r.Currency] = r.Rate
	}
	return mappedRates, nil
}
