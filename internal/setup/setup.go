package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"nuubot/internal/botspec"
	"nuubot/internal/config"
	"nuubot/internal/datastore"
	"nuubot/internal/hyperliquid"
	"nuubot/internal/market"
	"nuubot/internal/meta"
	"nuubot/internal/toolkit/clock"
	"nuubot/internal/toolkit/logging"
)

// Nuubot contains shared application infrastructure prepared for one Bot.
type Nuubot struct {
	Log         *logging.Logger
	App         config.App
	Runtime     config.Runtime
	Bot         datastore.Bot
	BotSpec     botspec.Spec
	Clock       clock.Clock
	MarketData  *market.MarketData
	Info        *hyperliquid.Info
	WebSocket   *hyperliquid.WebSocket
	Meta        meta.Instrument
	ResultPath  string
	RuntimePath string
}

// Section 1 - Program Flow

// Setup prepares shared application infrastructure for one Bot.
func Setup(
	caller context.Context,
	log *logging.Logger,
	sweepID,
	botID uint64,
) (*Nuubot, error) {
	// Step 1: validate caller context
	if caller == nil {
		return nil, fmt.Errorf("setup requires caller context")
	}

	// Step 2: resolve repository root
	var root, err = os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	// Step 3: load App Config
	var configPath = os.Getenv("NUUBOT_CONFIG")
	if configPath == "" {
		configPath = filepath.Join(root, "workspace", "config", "config.toml")
	} else if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(root, configPath)
	}
	var app config.App
	app, err = config.LoadApp(configPath)
	if err != nil {
		return nil, fmt.Errorf("load setup AppConfig: %w", err)
	}

	// Step 4: load Bot input
	var botInput datastore.Bot
	botInput, err = datastore.LoadBot(
		config.Rooted(root, app.Paths.Database),
		sweepID,
		botID,
	)
	if err != nil {
		return nil, fmt.Errorf("prepare datastore: %w", err)
	}

	// Step 5: resolve replay input path
	botInput.Replay.TicksPath, err = config.ResolveDataPath(
		config.Rooted(root, app.Paths.SharedData),
		botInput.Replay.TicksPath,
	)
	if err != nil {
		return nil, fmt.Errorf("validate ticks path: %w", err)
	}

	// Step 6: build BotSpec
	var botSpec botspec.Spec
	botSpec, err = botspec.Build(botInput.BotSpecID, botInput.ConfigTOML)
	if err != nil {
		return nil, fmt.Errorf("build BotSpec: %w", err)
	}

	// Step 7: validate Executor replay symbols
	for _, executorSpec := range botSpec.Executors {
		if executorSpec.Resource.Symbol != botInput.Replay.Symbol {
			return nil, fmt.Errorf(
				"controller Executor symbol %s lacks replay input",
				executorSpec.Resource.Symbol,
			)
		}
	}

	// Step 8: load mainnet Meta
	var timeout = time.Duration(app.Process.RequestTimeoutSeconds) * time.Second
	var client *hyperliquid.Info
	client, err = hyperliquid.NewInfo("mainnet", timeout)
	if err != nil {
		return nil, fmt.Errorf("load Meta: %w", err)
	}
	var requestContext, cancel = context.WithTimeout(caller, timeout)
	defer cancel()
	var instrument meta.Instrument
	instrument, err = meta.EnsureFresh(
		requestContext,
		config.Rooted(root, app.Paths.Database),
		botInput.Replay.Symbol,
		time.Now().UTC(),
		client,
	)
	if err != nil {
		return nil, fmt.Errorf("load Meta: %w", err)
	}

	// Shared WebSocket ownership remains TBD. Setup starts no background work.

	// Step 9: prepare Nuubot
	var resultPath = filepath.Join(
		root,
		"workspace",
		"db",
		"sweeps",
		fmt.Sprintf("sweep_%d", sweepID),
		fmt.Sprintf("bot_%d.db", botID),
	)
	var nuubot = &Nuubot{
		Log:         log,
		App:         app,
		Bot:         botInput,
		BotSpec:     botSpec,
		Meta:        instrument,
		ResultPath:  resultPath,
		RuntimePath: resultPath,
	}

	// Step 10: log setup completed
	log.Info(fmt.Sprintf(
		"setup initialized bot_spec=%s symbol=%s",
		botInput.BotSpecID,
		botInput.Replay.Symbol,
	))
	return nuubot, nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
