package hyperliquid

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testAddress = "0x0000000000000000000000000000000000000001"

// Section 1 - Program Flow

func TestNewRejectsInvalidConfig(t *testing.T) {
	var cases = []struct {
		name    string
		network string
		timeout time.Duration
	}{
		{name: "network", network: "invalid", timeout: time.Second},
		{name: "timeout", network: "testnet", timeout: 0},
	}
	var testCase struct {
		name    string
		network string
		timeout time.Duration
	}
	for _, testCase = range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var _, err = New(testCase.network, testCase.timeout)
			if err == nil {
				t.Fatalf("actual error nil, expected invalid configuration")
			}
		})
	}
}

func TestClearinghouseStateTranslatesResponse(t *testing.T) {
	var server = httptest.NewServer(http.HandlerFunc(func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		if request.Method != http.MethodPost {
			t.Fatalf("actual method %q, expected %q", request.Method, http.MethodPost)
		}
		if request.URL.Path != "/info" {
			t.Fatalf("actual path %q, expected %q", request.URL.Path, "/info")
		}
		var payload map[string]any
		var err = json.NewDecoder(request.Body).Decode(&payload)
		if err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload["type"] != "clearinghouseState" {
			t.Fatalf("actual type %v, expected clearinghouseState", payload["type"])
		}
		if payload["user"] != testAddress {
			t.Fatalf("actual user %v, expected %s", payload["user"], testAddress)
		}
		if payload["dex"] != "" {
			t.Fatalf("actual dex %v, expected empty string", payload["dex"])
		}
		response.Header().Set("Content-Type", "application/json")
		var _, _ = response.Write([]byte(testClearinghouseResponse))
	}))
	defer server.Close()

	var client = &Client{
		baseURL: server.URL,
		http:    &http.Client{Timeout: time.Second},
	}
	var state, err = client.ClearinghouseState(context.Background(), testAddress)
	if err != nil {
		t.Fatalf("read clearinghouse state: %v", err)
	}
	if state.TimeMS != 1733968369395 {
		t.Fatalf("actual time %d, expected 1733968369395", state.TimeMS)
	}
	if state.Margin.Equity.String() != "13109.482328" {
		t.Fatalf("actual equity %s, expected 13109.482328", state.Margin.Equity)
	}
	if state.MaintenanceMargin.String() != "0.5" {
		t.Fatalf("actual maintenance margin %s, expected 0.5", state.MaintenanceMargin)
	}
	if len(state.Positions) != 1 {
		t.Fatalf("actual positions %d, expected 1", len(state.Positions))
	}
	var position = state.Positions[0]
	if position.Symbol != "ETH" {
		t.Fatalf("actual symbol %q, expected ETH", position.Symbol)
	}
	if position.SignedSize.String() != "0.0335" {
		t.Fatalf("actual signed size %s, expected 0.0335", position.SignedSize)
	}
	if position.EntryPrice == nil || position.EntryPrice.String() != "2986.3" {
		t.Fatalf("actual entry price %v, expected 2986.3", position.EntryPrice)
	}
	if position.Leverage.RawUSD == nil || position.Leverage.RawUSD.String() != "-95.059824" {
		t.Fatalf("actual leverage raw USD %v, expected -95.059824", position.Leverage.RawUSD)
	}
	if position.MaxLeverage != 50 {
		t.Fatalf("actual max leverage %d, expected 50", position.MaxLeverage)
	}
}

func TestClearinghouseStateRejectsInvalidInput(t *testing.T) {
	var client = &Client{
		baseURL: "http://invalid",
		http:    &http.Client{Timeout: time.Second},
	}
	var _, err = client.ClearinghouseState(context.Background(), "not-an-address")
	if err == nil || !strings.Contains(err.Error(), "expected 42-character hexadecimal address") {
		t.Fatalf("actual error %v, expected address validation", err)
	}
}

func TestClearinghouseStateRejectsInvalidDecimal(t *testing.T) {
	var response = strings.Replace(
		testClearinghouseResponse,
		`"accountValue": "13109.482328"`,
		`"accountValue": "invalid"`,
		1,
	)
	var client, closeServer = testClient(t, http.StatusOK, response)
	defer closeServer()

	var _, err = client.ClearinghouseState(context.Background(), testAddress)
	if err == nil || !strings.Contains(err.Error(), "marginSummary.accountValue") {
		t.Fatalf("actual error %v, expected accountValue decimal error", err)
	}
}

func TestDecodeClearinghouseStateRejectsInvalidSemantics(t *testing.T) {
	var cases = []struct {
		name string
		old  string
		new  string
	}{
		{
			name: "time",
			old:  `"time": 1733968369395`,
			new:  `"time": 0`,
		},
		{
			name: "asset positions",
			old:  `"assetPositions": [`,
			new:  `"missingAssetPositions": [`,
		},
		{
			name: "position mode",
			old:  `"type": "oneWay"`,
			new:  `"type": "unknown"`,
		},
		{
			name: "leverage type",
			old:  `"type": "isolated"`,
			new:  `"type": "unknown"`,
		},
		{
			name: "leverage value",
			old:  `"value": 20`,
			new:  `"value": 0`,
		},
		{
			name: "maximum leverage",
			old:  `"maxLeverage": 50`,
			new:  `"maxLeverage": 0`,
		},
	}
	var testCase struct {
		name string
		old  string
		new  string
	}
	for _, testCase = range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var payload = strings.Replace(
				testClearinghouseResponse,
				testCase.old,
				testCase.new,
				1,
			)
			var _, err = DecodeClearinghouseState([]byte(payload))
			if err == nil || !strings.Contains(err.Error(), "validate clearinghouse payload") {
				t.Fatalf("actual error %v, expected semantic validation", err)
			}
		})
	}
}

func TestClearinghouseStateRejectsHTTPError(t *testing.T) {
	var client, closeServer = testClient(t, http.StatusTooManyRequests, `{}`)
	defer closeServer()

	var _, err = client.ClearinghouseState(context.Background(), testAddress)
	if err == nil || !strings.Contains(err.Error(), "http status 429") {
		t.Fatalf("actual error %v, expected http 429", err)
	}
}

func TestClearinghouseStateRejectsOversizedResponse(t *testing.T) {
	var response = strings.Repeat("x", maxResponseSize+1)
	var client, closeServer = testClient(t, http.StatusOK, response)
	defer closeServer()

	var _, err = client.ClearinghouseState(context.Background(), testAddress)
	if err == nil || !strings.Contains(err.Error(), "response exceeds") {
		t.Fatalf("actual error %v, expected oversized response", err)
	}
}

func TestClearinghouseStateHonorsContext(t *testing.T) {
	var server = httptest.NewServer(http.HandlerFunc(func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		<-request.Context().Done()
	}))
	defer server.Close()
	var client = &Client{
		baseURL: server.URL,
		http:    &http.Client{Timeout: time.Second},
	}
	var ctx, cancel = context.WithCancel(context.Background())
	cancel()

	var _, err = client.ClearinghouseState(ctx, testAddress)
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("actual error %v, expected context canceled", err)
	}
}

// Section 2 - Domain Helpers

func testClient(t *testing.T, status int, body string) (*Client, func()) {
	t.Helper()
	var server = httptest.NewServer(http.HandlerFunc(func(
		response http.ResponseWriter,
		request *http.Request,
	) {
		response.WriteHeader(status)
		var _, _ = response.Write([]byte(body))
	}))
	return &Client{
		baseURL: server.URL,
		http:    &http.Client{Timeout: time.Second},
	}, server.Close
}

// Section 3 - Generic Helpers

const testClearinghouseResponse = `{
  "assetPositions": [{
    "position": {
      "coin": "ETH",
      "cumFunding": {
        "allTime": "514.085417",
        "sinceChange": "0.0",
        "sinceOpen": "0.0"
      },
      "entryPx": "2986.3",
      "leverage": {
        "rawUsd": "-95.059824",
        "type": "isolated",
        "value": 20
      },
      "liquidationPx": "2866.26936529",
      "marginUsed": "4.967826",
      "maxLeverage": 50,
      "positionValue": "100.02765",
      "returnOnEquity": "-0.0026789",
      "szi": "0.0335",
      "unrealizedPnl": "-0.0134"
    },
    "type": "oneWay"
  }],
  "crossMaintenanceMarginUsed": "0.5",
  "crossMarginSummary": {
    "accountValue": "13104.514502",
    "totalMarginUsed": "0.0",
    "totalNtlPos": "0.0",
    "totalRawUsd": "13104.514502"
  },
  "marginSummary": {
    "accountValue": "13109.482328",
    "totalMarginUsed": "4.967826",
    "totalNtlPos": "100.02765",
    "totalRawUsd": "13104.514502"
  },
  "time": 1733968369395,
  "withdrawable": "13104.514502"
}`
