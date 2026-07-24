package parity

import (
	"io"
	"strings"
	"testing"

	"nuubot/internal/config"
	"nuubot/internal/toolkit/logging"
)

// Section 1 - Program Flow

func TestParseInputAcceptsClearinghouseState(t *testing.T) {
	var input, err = ParseInput([]string{
		"testnet",
		"tgrid",
		"clearinghouse-state",
		"baseline",
	})
	if err != nil {
		t.Fatalf("actual error %v, expected nil", err)
	}
	if input.Network != "testnet" {
		t.Fatalf("actual network %q, expected testnet", input.Network)
	}
	if len(input.Arguments) != 1 || input.Arguments[0] != "baseline" {
		t.Fatalf("actual arguments %v, expected [baseline]", input.Arguments)
	}
}

func TestParseInputRejectsUnsafeCapture(t *testing.T) {
	var cases = []struct {
		name string
		args []string
	}{
		{
			name: "account",
			args: []string{"testnet", "../tgrid", "clearinghouse-state"},
		},
		{
			name: "capture",
			args: []string{"testnet", "tgrid", "clearinghouse-state", ".."},
		},
	}
	var testCase struct {
		name string
		args []string
	}
	for _, testCase = range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var _, err = ParseInput(testCase.args)
			if err == nil || !strings.Contains(err.Error(), "invalid") {
				t.Fatalf("actual error %v, expected invalid path segment", err)
			}
		})
	}
}

func TestParseInputRejectsUnsupportedValues(t *testing.T) {
	var cases = []struct {
		name string
		args []string
	}{
		{
			name: "network",
			args: []string{"mainnet", "grid", "clearinghouse-state"},
		},
		{
			name: "operation",
			args: []string{"testnet", "tgrid", "unknown"},
		},
	}
	var testCase struct {
		name string
		args []string
	}
	for _, testCase = range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var _, err = ParseInput(testCase.args)
			if err == nil {
				t.Fatalf("actual error nil, expected unsupported value")
			}
		})
	}
}

func TestSelectAccountRequiresExactNetworkAndName(t *testing.T) {
	var credentials = config.Credentials{
		Hyperliquid: config.HyperliquidCredentials{
			Accounts: []config.HyperliquidAccountCredentials{
				{Network: "testnet", Name: "tgrid", Address: "test-address"},
				{Network: "mainnet", Name: "grid", Address: "main-address"},
			},
		},
	}
	var account, err = selectAccount(credentials, "testnet", "tgrid")
	if err != nil {
		t.Fatalf("actual error %v, expected nil", err)
	}
	if account.Address != "test-address" {
		t.Fatalf("actual address %q, expected test-address", account.Address)
	}
}

func TestSelectAccountRejectsMissingAccount(t *testing.T) {
	var credentials config.Credentials
	var _, err = selectAccount(credentials, "testnet", "missing")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("actual error %v, expected missing account", err)
	}
}

func TestInitRejectsSimnetWithoutLoadingCredentials(t *testing.T) {
	var probe Probe
	var err = probe.Init(
		logging.Create(io.Discard),
		t.TempDir(),
		Input{
			Network:   "simnet",
			Account:   "sgrid",
			Operation: "clearinghouse-state",
		},
	)
	if err == nil || !strings.Contains(err.Error(), "simulator clearinghouse state") {
		t.Fatalf("actual error %v, expected unavailable Simulator", err)
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
