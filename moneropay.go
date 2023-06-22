package main

import (
	"bytes"
	"encoding/json"
	"errors"
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
	if resp.StatusCode != 200 {
		var errResp mpay.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, err
		}
		return nil, errors.New(errResp.Message)
	}
	var respData mpay.TransferPostResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return nil, err
	}
	return &respData, nil
}
