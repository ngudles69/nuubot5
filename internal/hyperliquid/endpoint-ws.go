package hyperliquid

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrWebSocketNotImplemented identifies the reserved Hyperliquid WebSocket boundary.
var ErrWebSocketNotImplemented = errors.New("Hyperliquid WebSocket is not implemented")

// WebSocket owns one Hyperliquid public WebSocket connection.
type WebSocket struct {
	network string
	timeout time.Duration
	started bool
	stopped bool
}

// Section 1 - Program Flow

// NewWebSocket creates one stopped Hyperliquid WebSocket endpoint.
func NewWebSocket(network string, timeout time.Duration) (*WebSocket, error) {
	// Step 1: validate WebSocket config
	if network != "mainnet" && network != "testnet" {
		return nil, fmt.Errorf("create Hyperliquid WebSocket: unknown network %q", network)
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("create Hyperliquid WebSocket: timeout must be positive")
	}

	// Step 2: create WebSocket endpoint
	return &WebSocket{
		network: network,
		timeout: timeout,
	}, nil
}

// Start connects and proves the Hyperliquid WebSocket endpoint.
func (w *WebSocket) Start(ctx context.Context) error {
	// Step 1: validate start state
	if ctx == nil || w.started || w.stopped {
		return fmt.Errorf("start Hyperliquid WebSocket: invalid state")
	}

	// Step 2: start WebSocket endpoint
	return ErrWebSocketNotImplemented
}

// Stop stops the Hyperliquid WebSocket endpoint.
func (w *WebSocket) Stop() error {
	// Step 1: ignore repeated stop request
	if w.stopped {
		return nil
	}

	// Step 2: stop WebSocket endpoint
	w.started = false
	w.stopped = true
	return nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
