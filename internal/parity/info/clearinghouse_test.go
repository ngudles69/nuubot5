package info

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"nuubot/internal/hyperliquid"
)

type clearinghouseClientStub struct {
	response hyperliquid.Response
	err      error
}

func (c clearinghouseClientStub) ClearinghouseStatePayload(
	context.Context,
	string,
) (hyperliquid.Response, error) {
	return c.response, c.err
}

// Section 1 - Program Flow

func TestWriteClearinghouseSuccessPreservesRawPayload(t *testing.T) {
	var evidenceDir = filepath.Join(t.TempDir(), "capture")
	var payload = []byte(`{"withdrawable":"10"}`)
	var state = hyperliquid.AccountState{
		Margin:       hyperliquid.Margin{Equity: decimal.RequireFromString("10")},
		Withdrawable: decimal.RequireFromString("10"),
	}
	var err = reserveEvidenceDir(evidenceDir)
	if err != nil {
		t.Fatalf("reserve evidence directory: %v", err)
	}
	err = writeRawPayload(evidenceDir, payload)
	if err != nil {
		t.Fatalf("write raw payload: %v", err)
	}
	err = writeClearinghouseSuccess(
		evidenceDir,
		Target{Network: "testnet", Account: "test"},
		hyperliquid.Response{StatusCode: 200, Payload: payload},
		state,
		time.Millisecond,
	)
	if err != nil {
		t.Fatalf("actual error %v, expected nil", err)
	}
	var actual []byte
	actual, err = os.ReadFile(filepath.Join(evidenceDir, "raw.json"))
	if err != nil {
		t.Fatalf("read raw payload: %v", err)
	}
	if string(actual) != string(payload) {
		t.Fatalf("actual payload %s, expected %s", actual, payload)
	}
	for _, name := range []string{"normalized.json", "report.json"} {
		var _, statErr = os.Stat(filepath.Join(evidenceDir, name))
		if statErr != nil {
			t.Fatalf("actual %s error %v, expected file", name, statErr)
		}
	}
	err = reserveEvidenceDir(evidenceDir)
	if err == nil {
		t.Fatalf("actual error nil, expected existing capture rejection")
	}
}

func TestClearinghouseStatePreservesMalformedPayload(t *testing.T) {
	var evidenceRoot = t.TempDir()
	var payload = []byte(`{"assetPositions":`)
	var client = clearinghouseClientStub{
		response: hyperliquid.Response{StatusCode: 200, Payload: payload},
	}
	var _, err = ClearinghouseState(
		context.Background(),
		client,
		Target{
			Network:     "testnet",
			Account:     "test",
			Address:     "unused",
			EvidenceDir: evidenceRoot,
		},
		"malformed",
	)
	if err == nil {
		t.Fatalf("actual error nil, expected malformed response error")
	}
	var evidenceDir = filepath.Join(
		evidenceRoot,
		"info",
		"clearinghouse-state",
		"malformed",
		"testnet",
		"test",
	)
	var actual []byte
	actual, err = os.ReadFile(filepath.Join(evidenceDir, "raw.json"))
	if err != nil {
		t.Fatalf("read raw payload: %v", err)
	}
	if string(actual) != string(payload) {
		t.Fatalf("actual payload %s, expected %s", actual, payload)
	}
	var reportPayload []byte
	reportPayload, err = os.ReadFile(filepath.Join(evidenceDir, "report.json"))
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var actualReport report
	err = json.Unmarshal(reportPayload, &actualReport)
	if err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if actualReport.Error == "" {
		t.Fatalf("actual report error empty, expected decode failure")
	}
	var _, statErr = os.Stat(filepath.Join(evidenceDir, "normalized.json"))
	if !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("actual normalized error %v, expected missing file", statErr)
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
