// Package live owns one live Bot process.
package live

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"nuubot/internal/controller"
	"nuubot/internal/datastore"
	"nuubot/internal/hyperliquid"
	"nuubot/internal/market"
	"nuubot/internal/setup"
	"nuubot/internal/telemetry"
	"nuubot/internal/toolkit/clock"
	"nuubot/internal/toolkit/logging"
)

// Run owns one live Bot runtime.
type Run struct {
	caller                  context.Context
	log                     *logging.Logger
	clock                   clock.Clock
	marketData              *market.MarketData
	info                    *hyperliquid.Info
	webSocket               *hyperliquid.WebSocket
	controller              controller.Controller
	telemetryStore          telemetry.Store
	telemetry               []telemetry.Sample
	telemetryIntervalMS     uint64
	lastTelemetryMS         uint64
	telemetrySequence       uint64
	telemetryWriteOnCollect bool
	pollInterval            time.Duration
	stopRequested           chan struct{}
	stopOnce                sync.Once
	started                 bool
	stopped                 bool
}

// Section 1 - Program Flow

// Init prepares one live Bot runtime.
func (r *Run) Init(
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

	// Step 2: select Live runtime policy
	nuubot.Runtime = nuubot.App.Live
	r.telemetryIntervalMS = nuubot.Runtime.TelemetryIntervalMS
	r.telemetryWriteOnCollect = nuubot.Runtime.TelemetryWriteOnCollect

	// Step 3: reject terminal Bot
	switch nuubot.Bot.Status {
	case datastore.BotError, datastore.BotStopped:
		return fmt.Errorf("runner cannot initialize terminal bot status %q", nuubot.Bot.Status)
	}

	// Step 4: retain runtime inputs
	if nuubot.App.Process.PollSeconds == 0 {
		return fmt.Errorf("runner poll interval must be positive")
	}
	r.caller = caller
	r.pollInterval = time.Duration(nuubot.App.Process.PollSeconds) * time.Second
	r.stopRequested = make(chan struct{})
	var timeout = time.Duration(nuubot.App.Process.RequestTimeoutSeconds) * time.Second

	// Step 5: create clock
	r.clock, err = clock.Create(clock.Wall)
	if err != nil {
		return fmt.Errorf("create clock: %w", err)
	}

	// Step 6: initialize clock
	err = r.clock.Init(log, uint64(time.Now().UnixMilli()))
	if err != nil {
		return fmt.Errorf("initialize clock: %w", err)
	}

	// Step 7: attach clock to Nuubot
	nuubot.Clock = r.clock

	// Step 8: create and attach MarketData to Nuubot
	r.marketData = market.CreateMarketData()
	nuubot.MarketData = r.marketData

	// Step 9: initialize Info endpoint
	r.info, err = hyperliquid.NewInfo(nuubot.App.Network.Default, timeout)
	if err != nil {
		return fmt.Errorf("initialize Info endpoint: %w", err)
	}
	nuubot.Info = r.info

	// Step 10: initialize WebSocket endpoint
	r.webSocket, err = hyperliquid.NewWebSocket(nuubot.App.Network.Default, timeout)
	if err != nil {
		return fmt.Errorf("initialize WebSocket endpoint: %w", err)
	}
	nuubot.WebSocket = r.webSocket

	// Step 11: initialize Controller
	err = r.controller.Init(nuubot)
	if err != nil {
		return fmt.Errorf("initialize Controller: %w", err)
	}

	// Step 12: register Controller timer
	err = r.clock.RegisterTimer(clock.Timer{
		Name:       "controller",
		IntervalMS: nuubot.Runtime.ControllerIntervalMS,
	}, r.controllerRun)
	if err != nil {
		return fmt.Errorf("register Controller timer: %w", err)
	}

	// Step 13: initialize telemetry persistence
	if r.telemetryWriteOnCollect {
		r.telemetrySequence, err = r.telemetryStore.Init(nuubot.RuntimePath)
		if err != nil {
			return fmt.Errorf("initialize telemetry persistence: %w", err)
		}
	}

	// Step 14: log init completed
	log.Info(fmt.Sprintf(
		"runner initialized bot_spec=%s symbol=%s",
		nuubot.Bot.BotSpecID,
		nuubot.Bot.Replay.Symbol,
	))
	return nil
}

// Start starts the owned endpoints, Controller, and Clock.
func (r *Run) Start() error {
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
func (r *Run) Loop() error {
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
func (r *Run) Stop() error {
	// Step 1: log stop started
	r.log.Info("runner stop started")

	// Step 2: ignore repeated stop request
	if r.stopped {
		r.log.Info("runner stopping - ignoring stop request")
		return nil
	}

	// Step 3: mark Run stopped
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

	// Step 8: stop MarketData
	var marketDataErr = r.marketData.Stop()
	if marketDataErr != nil {
		marketDataErr = fmt.Errorf("stop MarketData: %w", marketDataErr)
	}

	// Step 9: collect terminal telemetry
	var telemetryErr = r.collectTelemetry(r.clock.NowMS(), true)
	if telemetryErr != nil {
		telemetryErr = fmt.Errorf("collect terminal telemetry: %w", telemetryErr)
	}

	// Step 10: stop telemetry persistence
	var telemetryStopErr = r.telemetryStore.Stop()
	if telemetryStopErr != nil {
		telemetryStopErr = fmt.Errorf("stop telemetry persistence: %w", telemetryStopErr)
	}

	// Step 11: log stop results and stats
	var result = "complete"
	if webSocketErr != nil || controllerErr != nil || marketDataErr != nil ||
		telemetryErr != nil || telemetryStopErr != nil {
		result = "failed"
	}
	r.log.Info(fmt.Sprintf("runner stopped result=%s", result))

	// Step 12: return stop errors
	if controllerErr != nil {
		return controllerErr
	}
	if webSocketErr != nil {
		return webSocketErr
	}
	if marketDataErr != nil {
		return marketDataErr
	}
	if telemetryErr != nil {
		return telemetryErr
	}
	if telemetryStopErr != nil {
		return telemetryStopErr
	}

	// Step 13: log stop completed
	r.log.Info("runner stopped.")
	return nil
}

// Section 2 - Domain Helpers

func (r *Run) controllerRun(_ uint64) error {
	// Step 1: run Controller - triggered by Timer
	var stop, err = r.controller.Run()
	if err != nil {
		return fmt.Errorf("run Controller: %w", err)
	}

	// Step 2: collect due telemetry
	var nowMS = r.clock.NowMS()
	if r.lastTelemetryMS == 0 || nowMS-r.lastTelemetryMS >= r.telemetryIntervalMS {
		err = r.collectTelemetry(nowMS, false)
		if err != nil {
			return err
		}
		r.lastTelemetryMS = nowMS
	}

	// Step 3: remember stop request
	if stop {
		r.stopOnce.Do(func() {
			close(r.stopRequested)
		})
	}
	return nil
}

func (r *Run) collectTelemetry(nowMS uint64, terminal bool) error {
	var current = r.controller.Telemetry()
	var sample = telemetry.Sample{
		Sequence:            r.telemetrySequence + 1,
		TimestampMS:         nowMS,
		Terminal:            terminal,
		TicksServed:         current.Ticks,
		ControllerRuns:      current.Runs,
		SignalPackages:      current.SignalPackagesRead,
		StartActionsSkipped: current.StartActionsSkipped,
		CyclesStarted:       current.CyclesStarted,
		CyclesRejected:      current.CyclesRejected,
		CyclesClosed:        current.CyclesClosed,
		ActiveCycle:         current.ActiveCycle,
		BotCapital:          current.BotCapital,
		BotBalance:          current.BotBalance,
		BotEquity:           current.BotEquity,
		NetPnL:              current.NetPnL,
		PeakEquity:          current.PeakEquity,
		Drawdown:            current.Drawdown,
		MaxDrawdown:         current.MaxDrawdown,
	}
	if r.telemetryWriteOnCollect {
		var err = r.telemetryStore.Write(sample)
		if err != nil {
			return err
		}
	} else {
		r.telemetry = append(r.telemetry, sample)
	}
	r.telemetrySequence = sample.Sequence
	return nil
}

// Section 3 - Generic Helpers
