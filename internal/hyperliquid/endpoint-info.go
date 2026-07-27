// Package hyperliquid owns Nuubot's Hyperliquid protocol boundary.
package hyperliquid

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	mainnetURL      = "https://api.hyperliquid.xyz"
	testnetURL      = "https://api.hyperliquid-testnet.xyz"
	maxResponseSize = 4 << 20
)

// Info sends bounded public Hyperliquid Info requests.
type Info struct {
	baseURL string
	http    *http.Client
}

// Response contains one admitted Hyperliquid HTTP response.
type Response struct {
	StatusCode int
	Payload    []byte
}

// Section 1 - Program Flow

// NewInfo constructs one Hyperliquid Info endpoint.
func NewInfo(network string, timeout time.Duration) (*Info, error) {
	// validate network
	var baseURL string
	switch network {
	case "mainnet":
		baseURL = mainnetURL
	case "testnet":
		baseURL = testnetURL
	default:
		return nil, fmt.Errorf("create Hyperliquid Info: unknown network %q", network)
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("create Hyperliquid Info: timeout must be positive")
	}

	// configure HTTP client
	return &Info{
		baseURL: baseURL,
		http:    &http.Client{Timeout: timeout},
	}, nil
}

// ClearinghouseState reads one user's default perpetual clearinghouse state.
func (c *Info) ClearinghouseState(ctx context.Context, address string) (AccountState, error) {
	// request clearinghouse payload
	var response, err = c.ClearinghouseStatePayload(ctx, address)
	if err != nil {
		return AccountState{}, err
	}

	// decode clearinghouse payload
	var state AccountState
	state, err = DecodeClearinghouseState(response.Payload)
	if err != nil {
		return AccountState{}, fmt.Errorf("read Hyperliquid clearinghouse state: %w", err)
	}
	return state, nil
}

// ClearinghouseStatePayload reads one raw perpetual clearinghouse payload.
func (c *Info) ClearinghouseStatePayload(
	ctx context.Context,
	address string,
) (Response, error) {
	// validate address
	var err = validateAddress(address)
	if err != nil {
		return Response{}, fmt.Errorf("read Hyperliquid clearinghouse state: %w", err)
	}

	// encode request payload
	var requestPayload []byte
	requestPayload, err = json.Marshal(map[string]any{
		"type": "clearinghouseState",
		"user": address,
		"dex":  "",
	})
	if err != nil {
		return Response{}, fmt.Errorf(
			"read Hyperliquid clearinghouse state: encode request payload: %v",
			err,
		)
	}

	// post request payload
	var response Response
	response, err = c.Post(ctx, "/info", requestPayload)
	if err != nil {
		return response, fmt.Errorf("read Hyperliquid clearinghouse state: %w", err)
	}
	return response, nil
}

// Post sends one JSON payload to a Hyperliquid endpoint.
func (c *Info) Post(
	ctx context.Context,
	endpoint string,
	payload []byte,
) (Response, error) {
	// create request
	var request, err = http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return Response{}, fmt.Errorf("post Hyperliquid payload: create request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")

	// send request
	var response *http.Response
	response, err = c.http.Do(request)
	if err != nil {
		return Response{}, fmt.Errorf("post Hyperliquid payload: send request: %w", err)
	}
	defer response.Body.Close()

	// read response payload
	var responsePayload []byte
	responsePayload, err = io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return Response{}, fmt.Errorf("post Hyperliquid payload: read response: %w", err)
	}
	if len(responsePayload) > maxResponseSize {
		return Response{}, fmt.Errorf(
			"post Hyperliquid payload: response exceeds %d bytes",
			maxResponseSize,
		)
	}

	// validate response
	var result = Response{
		StatusCode: response.StatusCode,
		Payload:    responsePayload,
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return result, fmt.Errorf(
			"post Hyperliquid payload: http status %d",
			response.StatusCode,
		)
	}
	return result, nil
}

// Stop releases idle Hyperliquid Info connections.
func (c *Info) Stop() {
	// stop Info endpoint
	c.http.CloseIdleConnections()
}

// Section 2 - Domain Helpers

func validateAddress(address string) error {
	if len(address) != 42 || !strings.HasPrefix(address, "0x") {
		return fmt.Errorf("validate Hyperliquid address: expected 42-character hexadecimal address")
	}
	var _, err = hex.DecodeString(address[2:])
	if err != nil {
		return fmt.Errorf("validate Hyperliquid address: invalid hexadecimal")
	}
	return nil
}

// Section 3 - Generic Helpers
