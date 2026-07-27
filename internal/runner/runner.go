// Package runner owns one live Bot process.
package runner

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"nuubot/internal/controller"
	"nuubot/internal/datastore"
	"nuubot/internal/hyperliquid"
	"nuubot/internal/setup"
	"nuubot/internal/toolkit/clock"
	"nuubot/internal/toolkit/logging"
)

// Runner owns one live Bot runtime.
type Runner struct {
	caller        context.Context
	log           *logging.Logger
	clock         clock.Clock
	info          *hyperliquid.Info
	webSocket     *hyperliquid.WebSocket
	controller    controller.Controller
	pollInterval  time.Duration
	stopRequested chan struct{}
	stopOnce      sync.Once
	started       bool
	stopped       bool
}

// Section 1 - Program Flow

// Init prepares one live Bot runtime.
func (r *Runner) Init(
	caller context.Context,
	log *logging.Logger,
	sweepID,
	botID uint64,
) error {
	r.log = log

	// Step 1: general app global setup
	var nuubot, err = setup.Setup(caller, log, sweepID, botID)
	if err != nil {
		return fmt.Errorf("prepare setup: %w", err)
	}

	// Step 2: reject terminal Bot
	switch nuubot.Bot.Status {
	case datastore.BotError, datastore.BotStopped:
		return fmt.Errorf("runner cannot initialize terminal bot status %q", nuubot.Bot.Status)
	}

	// Step 3: retain runtime inputs
	if nuubot.App.Process.PollSeconds == 0 {
		return fmt.Errorf("runner poll interval must be positive")
	}
	r.caller = caller
	r.pollInterval = time.Duration(nuubot.App.Process.PollSeconds) * time.Second
	r.stopRequested = make(chan struct{})
	var timeout = time.Duration(nuubot.App.Process.RequestTimeoutSeconds) * time.Second

	// Step 4: create clock
	r.clock, err = clock.Create(clock.Wall)
	if err != nil {
		return fmt.Errorf("create clock: %w", err)
	}

	// Step 5: initialize clock
	err = r.clock.Init(log, uint64(time.Now().UnixMilli()))
	if err != nil {
		return fmt.Errorf("initialize clock: %w", err)
	}

	// Step 6: attach clock to Nuubot
	nuubot.Clock = r.clock

	// Step 7: initialize Info endpoint
	r.info, err = hyperliquid.NewInfo(nuubot.App.Network.Default, timeout)
	if err != nil {
		return fmt.Errorf("initialize Info endpoint: %w", err)
	}
	nuubot.Info = r.info

	// Step 8: initialize WebSocket endpoint
	r.webSocket, err = hyperliquid.NewWebSocket(nuubot.App.Network.Default, timeout)
	if err != nil {
		return fmt.Errorf("initialize WebSocket endpoint: %w", err)
	}
	nuubot.WebSocket = r.webSocket

	// Step 9: initialize Controller
	err = r.controller.Init(nuubot)
	if err != nil {
		return fmt.Errorf("initialize Controller: %w", err)
	}

	// Step 10: register Controller timer
	err = r.clock.RegisterTimer(clock.Timer{
		Name:       "controller",
		IntervalMS: uint64(r.pollInterval.Milliseconds()),
	}, r.controllerRun)
	if err != nil {
		return fmt.Errorf("register Controller timer: %w", err)
	}

	// Step 11: log init completed
	log.Info(fmt.Sprintf(
		"runner initialized bot_spec=%s symbol=%s",
		nuubot.Bot.BotSpecID,
		nuubot.Bot.Replay.Symbol,
	))
	return nil
}

// Start starts the owned endpoints, Controller, and Clock.
func (r *Runner) Start() error {
	if r.started || r.stopped {
		return fmt.Errorf("runner cannot start from current state")
	}

	// Step 1: start WebSocket endpoint
	var err = r.webSocket.Start(r.caller)
	if err != nil {
		return fmt.Errorf("start WebSocket endpoint: %w", err)
	}

	// Step 2: start Info endpoint
	var _, infoErr = r.info.PerpetualMeta(r.caller)
	if infoErr != nil {
		r.webSocket.Stop()
		return fmt.Errorf("start Info endpoint: %w", infoErr)
	}

	// Step 3: start Controller
	err = r.controller.Start()
	if err != nil {
		r.webSocket.Stop()
		return fmt.Errorf("start Controller: %w", err)
	}

	// Step 4: start clock
	err = r.clock.Start()
	if err != nil {
		var controllerErr = r.controller.Stop("start_error")
		var webSocketErr = r.webSocket.Stop()
		return errors.Join(fmt.Errorf("start clock: %w", err), controllerErr, webSocketErr)
	}

	// Step 5: log start completed
	r.started = true
	r.log.Info("runner started")
	return nil
}

// Loop supervises the live Bot runtime.
func (r *Runner) Loop() error {
	if !r.started || r.stopped {
		return fmt.Errorf("runner cannot loop from current state")
	}
	r.log.Info("runner loop started")
	var poll = time.NewTicker(r.pollInterval)
	defer poll.Stop()

	for {
		// Step 1: wait for runtime event
		select {
		case <-r.caller.Done():
			return nil
		case <-r.stopRequested:
			return nil
		case <-poll.C:
		}

		// Step 2: check clock failure
		var err = r.clock.Err()
		if err != nil {
			return fmt.Errorf("run clock: %w", err)
		}
	}
}

// Stop releases owned resources and reports final results.
func (r *Runner) Stop() error {
	// Step 1: log stop started
	r.log.Info("runner stop started")

	// Step 2: ignore repeated stop request
	if r.stopped {
		r.log.Info("runner stopping - ignoring stop request")
		return nil
	}

	// Step 3: mark Runner stopped
	r.started = false
	r.stopped = true

	// Step 4: stop clock
	r.clock.Stop()

	// Step 5: stop WebSocket endpoint
	var webSocketErr = r.webSocket.Stop()
	if webSocketErr != nil {
		webSocketErr = fmt.Errorf("stop WebSocket endpoint: %w", webSocketErr)
	}

	// Step 6: stop Info endpoint
	r.info.Stop()

	// Step 7: stop Controller
	var controllerErr = r.controller.Stop("parent_stop")
	if controllerErr != nil {
		controllerErr = fmt.Errorf("stop Controller: %w", controllerErr)
	}

	// Step 8: log stop results and stats
	var result = "complete"
	if webSocketErr != nil || controllerErr != nil {
		result = "failed"
	}
	r.log.Info(fmt.Sprintf("runner stopped result=%s", result))

	// Step 9: return stop errors
	if controllerErr != nil {
		return controllerErr
	}
	if webSocketErr != nil {
		return webSocketErr
	}

	// Step 10: log stop completed
	r.log.Info("runner stopped.")
	return nil
}

// Section 2 - Domain Helpers

func (r *Runner) controllerRun(_ uint64) error {
	// run Controller - triggered by Timer
	var stop, err = r.controller.Run()
	if err != nil {
		return fmt.Errorf("run Controller: %w", err)
	}
	// remember stop request
	if stop {
		r.stopOnce.Do(func() {
			close(r.stopRequested)
		})
	}
	return nil
}

// Section 3 - Generic Helpers
