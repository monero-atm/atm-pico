package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"gitlab.com/moneropay/go-monero/walletrpc"
	mpay "gitlab.com/moneropay/moneropay/v2/pkg/model"
)

func mpayTransfer(amount uint64, address string) (*mpay.TransferPostResponse, error) {
	endpoint, err := url.JoinPath(cfg.Moneropay, "/transfer")
	if err != nil {
		return nil, err
	}
	reqData := mpay.TransferPostRequest{
		Destinations: []walletrpc.Destination{
			{Amount: amount, Address: address},
		}}
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(reqData); err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", endpoint, b)
	if err != nil {
		return nil, err
	}
	cl := &http.Client{Timeout: 15 * time.Second}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	var respData mpay.TransferPostResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return nil, err
	}
	return &respData, nil
}
