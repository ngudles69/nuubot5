// Package botspec validates and shapes exact BotConfig into typed Bot specifications.
package botspec

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/shopspring/decimal"

	"nuubot/internal/signaler"
)

const (
	MacrossObserver = "macross_observer_bot"
	MacrossTrade    = "macross_trade_bot"
	MacrossGrid     = "macross_grid_bot"
	long            = "long"
	short           = "short"
)

type controllerConfig struct {
	MaxCycles uint64 `toml:"max_cycles"`
}

type macrossConfig struct {
	SignalTimeframe string `toml:"signal_timeframe"`
	RegimeTimeframe string `toml:"regime_timeframe"`
	FastMA          int    `toml:"fast_ma"`
	SlowMA          int    `toml:"slow_ma"`
	RegimeEMA       int    `toml:"regime_ema"`
}

type riskConfig struct {
	Kind string `toml:"kind"`
}

type tradeConfig struct {
	BotSpec    string                `toml:"bot_spec"`
	Controller controllerConfig      `toml:"controller"`
	Signaler   macrossConfig         `toml:"signaler"`
	Executors  []tradeExecutorConfig `toml:"executors"`
	Risks      []riskConfig          `toml:"risks"`
}

type tradeExecutorConfig struct {
	ID                string `toml:"id"`
	Role              string `toml:"role"`
	Kind              string `toml:"kind"`
	Side              string `toml:"side"`
	Venue             string `toml:"venue"`
	Network           string `toml:"network"`
	PhysicalAccountID string `toml:"physical_account_id"`
	Symbol            string `toml:"symbol"`
	CapitalUSDC       string `toml:"capital_usdc"`
	OrderSizeUSDC     string `toml:"order_size_usdc"`
	TakeProfitPct     string `toml:"take_profit_pct"`
	StopLossPct       string `toml:"stop_loss_pct"`
	FeePct            string `toml:"fee_pct"`
	SlippagePct       string `toml:"slippage_pct"`
	PersistMode       string `toml:"persist_mode"`
	Recon             string `toml:"recon"`
}

type observerConfig struct {
	BotSpec    string                   `toml:"bot_spec"`
	Controller controllerConfig         `toml:"controller"`
	Signaler   macrossConfig            `toml:"signaler"`
	Executors  []observerExecutorConfig `toml:"executors"`
	Risks      []riskConfig             `toml:"risks"`
}

type observerExecutorConfig struct {
	ID          string `toml:"id"`
	Kind        string `toml:"kind"`
	Side        string `toml:"side"`
	Venue       string `toml:"venue"`
	Network     string `toml:"network"`
	Symbol      string `toml:"symbol"`
	StopLossPct string `toml:"stop_loss_pct"`
}

type gridConfig struct {
	BotSpec    string               `toml:"bot_spec"`
	Controller controllerConfig     `toml:"controller"`
	Signaler   macrossConfig        `toml:"signaler"`
	Executors  []gridExecutorConfig `toml:"executors"`
	Risks      []riskConfig         `toml:"risks"`
}

type gridExecutorConfig struct {
	ID                 string `toml:"id"`
	Role               string `toml:"role"`
	Kind               string `toml:"kind"`
	Side               string `toml:"side"`
	Venue              string `toml:"venue"`
	Network            string `toml:"network"`
	PhysicalAccountID  string `toml:"physical_account_id"`
	Symbol             string `toml:"symbol"`
	CapitalUSDC        string `toml:"capital_usdc"`
	Levels             int    `toml:"levels"`
	RangePct           string `toml:"range_pct"`
	MinExpectedPnLUSDC string `toml:"min_expected_pnl_usdc"`
	FeePct             string `toml:"fee_pct"`
	SlippagePct        string `toml:"slippage_pct"`
	PersistMode        string `toml:"persist_mode"`
	Recon              string `toml:"recon"`
}

// Section 1 - Program Flow

// Validate checks one exact BotConfig.
func Validate(botSpecID, configTOML string) error {
	var _, err = Build(botSpecID, configTOML)
	return err
}

// ValidateReplaySymbol requires replay input for every configured Executor symbol.
func ValidateReplaySymbol(botSpecID, configTOML, replaySymbol string) error {
	var spec, err = Build(botSpecID, configTOML)
	if err != nil {
		return err
	}
	for _, current := range spec.Executors {
		if current.Resource.Symbol != replaySymbol {
			return fmt.Errorf(
				"executor symbol %s lacks replay input %s",
				current.Resource.Symbol,
				replaySymbol,
			)
		}
	}
	return nil
}

// Section 2 - Domain Helpers

func build(botSpecID, configTOML string) (Spec, error) {
	switch botSpecID {
	case MacrossTrade:
		return buildMacrossTrade(configTOML)
	case MacrossGrid:
		return buildMacrossGrid(configTOML)
	case MacrossObserver:
		return buildMacrossObserver(configTOML)
	default:
		return Spec{}, fmt.Errorf("unknown BotSpecID: %s", botSpecID)
	}
}

func buildMacrossGrid(configTOML string) (Spec, error) {
	var cfg gridConfig
	if _, err := toml.Decode(configTOML, &cfg); err != nil {
		return Spec{}, fmt.Errorf("decode %s Config: %w", MacrossGrid, err)
	}
	var result, err = buildMacross(
		MacrossGrid,
		cfg.BotSpec,
		cfg.Controller,
		cfg.Signaler,
		cfg.Risks,
	)
	if err != nil {
		return Spec{}, err
	}
	if len(cfg.Executors) != 1 {
		return Spec{}, fmt.Errorf("%s requires one Grid Executor", MacrossGrid)
	}
	var spec ExecutorSpec
	spec, err = buildGridExecutor(cfg.Executors[0])
	if err != nil {
		return Spec{}, err
	}
	result.Executors = []ExecutorSpec{spec}
	return result, nil
}

func buildMacrossTrade(configTOML string) (Spec, error) {
	var cfg tradeConfig
	if _, err := toml.Decode(configTOML, &cfg); err != nil {
		return Spec{}, fmt.Errorf("decode %s Config: %w", MacrossTrade, err)
	}
	var result, err = buildMacross(
		MacrossTrade,
		cfg.BotSpec,
		cfg.Controller,
		cfg.Signaler,
		cfg.Risks,
	)
	if err != nil {
		return Spec{}, err
	}
	if len(cfg.Executors) == 0 {
		return Spec{}, fmt.Errorf("%s requires at least one Executor", MacrossTrade)
	}
	var resources = make(map[Resource]bool, len(cfg.Executors))
	for _, raw := range cfg.Executors {
		var spec ExecutorSpec
		spec, err = buildTradeExecutor(raw)
		if err != nil {
			return Spec{}, err
		}
		if resources[spec.Resource] {
			return Spec{}, fmt.Errorf(
				"duplicate Executor resource: %s",
				spec.Resource.Key(),
			)
		}
		resources[spec.Resource] = true
		result.Executors = append(result.Executors, spec)
	}
	return result, nil
}

func buildMacrossObserver(configTOML string) (Spec, error) {
	var cfg observerConfig
	if _, err := toml.Decode(configTOML, &cfg); err != nil {
		return Spec{}, fmt.Errorf("decode %s Config: %w", MacrossObserver, err)
	}
	var result, err = buildMacross(
		MacrossObserver,
		cfg.BotSpec,
		cfg.Controller,
		cfg.Signaler,
		cfg.Risks,
	)
	if err != nil {
		return Spec{}, err
	}
	if len(cfg.Executors) != 1 {
		return Spec{}, fmt.Errorf("%s requires one Observer Executor", MacrossObserver)
	}
	var raw = cfg.Executors[0]
	if raw.Venue == "" && raw.Network == "" {
		raw.Venue = "simulator"
		raw.Network = "simnet"
	}
	var stopLoss decimal.Decimal
	stopLoss, err = decimal.NewFromString(raw.StopLossPct)
	if err != nil || raw.ID == "" || raw.Kind != "observer" ||
		(raw.Side != long && raw.Side != short) ||
		!validMarket(raw.Venue, raw.Network) ||
		raw.Symbol == "" ||
		!stopLoss.IsPositive() ||
		stopLoss.GreaterThanOrEqual(decimal.NewFromInt(1)) {
		return Spec{}, fmt.Errorf("invalid %s Observer Executor", MacrossObserver)
	}
	result.Executors = []ExecutorSpec{{
		ID:   raw.ID,
		Kind: raw.Kind,
		Side: raw.Side,
		Resource: Resource{
			Venue:   raw.Venue,
			Network: raw.Network,
			Symbol:  raw.Symbol,
		},
		StopLossPct: stopLoss,
	}}
	return result, nil
}

func buildMacross(
	botSpecID string,
	configBotSpecID string,
	controller controllerConfig,
	macross macrossConfig,
	risks []riskConfig,
) (Spec, error) {
	if configBotSpecID != botSpecID {
		return Spec{}, fmt.Errorf(
			"Config BotSpecID %q does not match %q",
			configBotSpecID,
			botSpecID,
		)
	}
	if controller.MaxCycles == 0 ||
		macross.SignalTimeframe == "" ||
		macross.RegimeTimeframe == "" ||
		macross.FastMA <= 0 ||
		macross.FastMA >= macross.SlowMA ||
		macross.RegimeEMA <= 0 {
		return Spec{}, fmt.Errorf("invalid %s Controller or Signaler Config", botSpecID)
	}
	if len(risks) == 0 {
		return Spec{}, fmt.Errorf("%s requires Risk Config", botSpecID)
	}
	var result = Spec{
		Controller: ControllerSpec{MaxCycles: controller.MaxCycles},
		Signaler: signaler.Config{
			Kind:            "macross",
			SignalTimeframe: macross.SignalTimeframe,
			RegimeTimeframe: macross.RegimeTimeframe,
			FastMA:          macross.FastMA,
			SlowMA:          macross.SlowMA,
			RegimeEMA:       macross.RegimeEMA,
		},
	}
	for _, current := range risks {
		if current.Kind != "balanced" {
			return Spec{}, fmt.Errorf("unknown Risk: %s", current.Kind)
		}
		result.Risks = append(result.Risks, RiskSpec{Kind: current.Kind})
	}
	return result, nil
}

func buildTradeExecutor(raw tradeExecutorConfig) (ExecutorSpec, error) {
	var values = make([]decimal.Decimal, 0, 6)
	for _, text := range []string{
		raw.CapitalUSDC,
		raw.OrderSizeUSDC,
		raw.TakeProfitPct,
		raw.StopLossPct,
		raw.FeePct,
		raw.SlippagePct,
	} {
		var value, err = decimal.NewFromString(text)
		if err != nil {
			return ExecutorSpec{}, fmt.Errorf("invalid Executor decimal: %w", err)
		}
		values = append(values, value)
	}
	var resource = Resource{
		Venue:             raw.Venue,
		Network:           raw.Network,
		PhysicalAccountID: raw.PhysicalAccountID,
		Symbol:            raw.Symbol,
	}
	if raw.ID == "" || raw.Role == "" || raw.Kind != "trade" ||
		(raw.Side != long && raw.Side != short) ||
		resource.Venue != "simulator" ||
		resource.Network != "simnet" ||
		resource.PhysicalAccountID == "" ||
		resource.Symbol == "" ||
		!values[0].IsPositive() ||
		!values[1].IsPositive() ||
		values[1].GreaterThan(values[0]) ||
		!values[2].IsPositive() ||
		values[2].GreaterThanOrEqual(decimal.NewFromInt(1)) ||
		!values[3].IsPositive() ||
		values[3].GreaterThanOrEqual(decimal.NewFromInt(1)) ||
		values[4].IsNegative() ||
		values[5].IsNegative() ||
		(raw.PersistMode != "none" && raw.PersistMode != "max") ||
		(raw.Recon != "" && raw.Recon != "recon") {
		return ExecutorSpec{}, fmt.Errorf("invalid %s Trade Executor", MacrossTrade)
	}
	return ExecutorSpec{
		ID:            raw.ID,
		Role:          raw.Role,
		Kind:          raw.Kind,
		Side:          raw.Side,
		Resource:      resource,
		CapitalUSDC:   values[0],
		OrderSizeUSDC: values[1],
		TakeProfitPct: values[2],
		StopLossPct:   values[3],
		FeePct:        values[4],
		SlippagePct:   values[5],
		PersistMode:   raw.PersistMode,
		Recon:         raw.Recon,
	}, nil
}

func buildGridExecutor(raw gridExecutorConfig) (ExecutorSpec, error) {
	var values = make([]decimal.Decimal, 0, 5)
	for _, text := range []string{
		raw.CapitalUSDC,
		raw.RangePct,
		raw.MinExpectedPnLUSDC,
		raw.FeePct,
		raw.SlippagePct,
	} {
		var value, err = decimal.NewFromString(text)
		if err != nil {
			return ExecutorSpec{}, fmt.Errorf("invalid Executor decimal: %w", err)
		}
		values = append(values, value)
	}
	var resource = Resource{
		Venue:             raw.Venue,
		Network:           raw.Network,
		PhysicalAccountID: raw.PhysicalAccountID,
		Symbol:            raw.Symbol,
	}
	if raw.ID == "" || raw.Role == "" || raw.Kind != "grid" ||
		(raw.Side != long && raw.Side != short) ||
		resource.Venue != "simulator" ||
		resource.Network != "simnet" ||
		resource.PhysicalAccountID == "" ||
		resource.Symbol == "" ||
		!values[0].IsPositive() ||
		raw.Levels < 3 || raw.Levels > 1024 ||
		!values[1].IsPositive() ||
		values[1].GreaterThanOrEqual(decimal.NewFromInt(1)) ||
		values[2].IsNegative() ||
		values[3].IsNegative() ||
		values[4].IsNegative() ||
		(raw.PersistMode != "none" && raw.PersistMode != "max") ||
		(raw.Recon != "" && raw.Recon != "recon") {
		return ExecutorSpec{}, fmt.Errorf("invalid %s Grid Executor", MacrossGrid)
	}
	return ExecutorSpec{
		ID:             raw.ID,
		Role:           raw.Role,
		Kind:           raw.Kind,
		Side:           raw.Side,
		Resource:       resource,
		CapitalUSDC:    values[0],
		GridLevels:     raw.Levels,
		RangePct:       values[1],
		MinExpectedPnL: values[2],
		FeePct:         values[3],
		SlippagePct:    values[4],
		PersistMode:    raw.PersistMode,
		Recon:          raw.Recon,
	}, nil
}

func validMarket(venue, network string) bool {
	return venue == "simulator" && network == "simnet" ||
		venue == "hyperliquid" && (network == "testnet" || network == "mainnet")
}

// Section 3 - Generic Helpers
