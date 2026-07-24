// Package parity owns the permanent parity-probe harness.
package parity

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"nuubot/internal/config"
	"nuubot/internal/hyperliquid"
	"nuubot/internal/parity/info"
	"nuubot/internal/toolkit/logging"
)

const clearinghouseState = "clearinghouse-state"

// Input contains one admitted parity-probe request.
type Input struct {
	Network   string
	Account   string
	Operation string
	Arguments []string
}

// Probe owns one parity-probe execution.
type Probe struct {
	log     *logging.Logger
	input   Input
	capture string
	target  info.Target
	client  *hyperliquid.Client
}

// Section 1 - Program Flow

// ParseInput parses one parity-probe command.
func ParseInput(args []string) (Input, error) {
	if len(args) < 3 {
		return Input{}, fmt.Errorf(
			"usage: parity-probe <testnet|simnet> <account> <operation> [arguments...]",
		)
	}
	var input = Input{
		Network:   args[0],
		Account:   args[1],
		Operation: args[2],
		Arguments: args[3:],
	}
	if input.Network != "testnet" && input.Network != "simnet" {
		return Input{}, fmt.Errorf("parse parity input: unsupported network %q", input.Network)
	}
	var err = validatePathSegment("account", input.Account)
	if err != nil {
		return Input{}, err
	}
	if input.Operation != clearinghouseState {
		return Input{}, fmt.Errorf(
			"parse parity input: unsupported operation %q",
			input.Operation,
		)
	}
	if len(input.Arguments) > 1 {
		return Input{}, fmt.Errorf(
			"parse parity input: clearinghouse-state accepts only an optional capture name",
		)
	}
	if len(input.Arguments) == 1 {
		err = validatePathSegment("capture", input.Arguments[0])
		if err != nil {
			return Input{}, err
		}
	}
	return input, nil
}

// Init prepares one parity-probe execution.
func (p *Probe) Init(
	log *logging.Logger,
	root string,
	input Input,
) error {
	p.log = log
	p.input = input

	// select target
	if input.Network == "simnet" {
		return fmt.Errorf(
			"initialize parity probe: simulator clearinghouse state is not implemented",
		)
	}

	// load shared config
	var cfg, err = config.Load(
		filepath.Join(root, "workspace", "config", "config.toml"),
	)
	if err != nil {
		return fmt.Errorf("initialize parity probe: load config: %w", err)
	}

	// load credentials
	var credentials config.Credentials
	credentials, err = config.LoadCredentials(
		filepath.Join(root, "workspace", "config", "credentials.toml"),
	)
	if err != nil {
		return fmt.Errorf("initialize parity probe: load credentials: %w", err)
	}

	// select account
	var account config.HyperliquidAccountCredentials
	account, err = selectAccount(credentials, input.Network, input.Account)
	if err != nil {
		return fmt.Errorf("initialize parity probe: %w", err)
	}

	// initialize Hyperliquid client
	var timeout = time.Duration(cfg.Process.RequestTimeoutSeconds) * time.Second
	p.client, err = hyperliquid.New(input.Network, timeout)
	if err != nil {
		return fmt.Errorf("initialize parity probe: %w", err)
	}
	p.target = info.Target{
		Network: input.Network,
		Account: input.Account,
		Address: account.Address,
		EvidenceDir: filepath.Join(
			root,
			"wiki",
			"design",
			"hyperliquid",
			"json",
		),
	}

	// select capture
	p.capture = time.Now().UTC().Format("20060102T150405Z")
	if len(input.Arguments) == 1 {
		p.capture = input.Arguments[0]
	}
	return nil
}

// Run executes one parity-probe operation.
func (p *Probe) Run(ctx context.Context) error {
	switch p.input.Operation {
	case clearinghouseState:
		var result, err = info.ClearinghouseState(
			ctx,
			p.client,
			p.target,
			p.capture,
		)
		if err != nil {
			return fmt.Errorf("run parity probe: %w", err)
		}
		var message = fmt.Sprintf(
			"parity probe completed network=%s account=%s operation=%s "+
				"equity=%s margin_used=%s maintenance_margin=%s "+
				"withdrawable=%s positions=%d duration_ms=%d evidence=%s",
			p.input.Network,
			p.input.Account,
			p.input.Operation,
			result.Equity,
			result.MarginUsed,
			result.MaintenanceMargin,
			result.Withdrawable,
			result.Positions,
			result.Duration.Milliseconds(),
			result.EvidenceDir,
		)
		p.log.Info(message)
		return nil
	default:
		return fmt.Errorf(
			"run parity probe: unsupported operation %q",
			p.input.Operation,
		)
	}
}

// Section 2 - Domain Helpers

func selectAccount(
	credentials config.Credentials,
	network string,
	name string,
) (config.HyperliquidAccountCredentials, error) {
	var found *config.HyperliquidAccountCredentials
	var index int
	for index = range credentials.Hyperliquid.Accounts {
		var account = &credentials.Hyperliquid.Accounts[index]
		if account.Network != network || account.Name != name {
			continue
		}
		if found != nil {
			return config.HyperliquidAccountCredentials{}, fmt.Errorf(
				"select account: duplicate %s account %q",
				network,
				name,
			)
		}
		found = account
	}
	if found == nil {
		return config.HyperliquidAccountCredentials{}, fmt.Errorf(
			"select account: %s account %q not found",
			network,
			name,
		)
	}
	return *found, nil
}

// Section 3 - Generic Helpers

func validatePathSegment(name, value string) error {
	if value == "" || strings.ContainsAny(value, `/\`) || value == "." || value == ".." {
		return fmt.Errorf("parse parity input: invalid %s name %q", name, value)
	}
	var character rune
	for _, character = range value {
		var admitted = character >= 'a' && character <= 'z' ||
			character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' ||
			character == '-' ||
			character == '_'
		if !admitted {
			return fmt.Errorf("parse parity input: invalid %s name %q", name, value)
		}
	}
	return nil
}
