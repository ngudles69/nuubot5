package botspec

import (
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/bot"
	"nuubot/internal/datastore"
	"nuubot/internal/meta"
	"nuubot/internal/risk"
	"nuubot/internal/signaler"
	"nuubot/internal/toolkit/logging"
)

// BuildInput contains admitted external values needed to build one Bot.
type BuildInput struct {
	Bot             datastore.Bot
	Meta            meta.Instrument
	MinNotionalUSDC decimal.Decimal
	ResultPath      string
}

// Section 1 - Program Flow

// Build constructs one exact compiled BotSpec.
func Build(log *logging.Logger, input BuildInput) (bot.Definition, error) {
	var cfg, err = admit(input.Bot.BotSpecID, input.Bot.ConfigTOML)
	if err != nil {
		return bot.Definition{}, err
	}
	var replay = input.Bot.Replay
	var end = replay.ReplayEnd
	if replay.EndAt != nil && replay.EndAt.Before(end) {
		end = *replay.EndAt
	}

	var signals signaler.Signaler
	signals, err = signaler.Create(
		log,
		cfg.signaler,
		replay.Symbol,
		replay.TicksPath,
		replay.ReplayStart,
		end,
	)
	if err != nil {
		return bot.Definition{}, fmt.Errorf("build %s Signaler: %w", input.Bot.BotSpecID, err)
	}

	var risks = make([]risk.Risk, 0, len(cfg.risks))
	for index, kind := range cfg.risks {
		var created risk.Risk
		created, err = risk.Create(log, index+1, kind)
		if err != nil {
			signals.Stop()
			return bot.Definition{}, fmt.Errorf(
				"build %s Risk %d: %w",
				input.Bot.BotSpecID,
				index+1,
				err,
			)
		}
		risks = append(risks, created)
	}

	for index := range cfg.executors {
		if cfg.executors[index].Resource.Symbol != replay.Symbol {
			stopRisks(risks)
			signals.Stop()
			return bot.Definition{}, fmt.Errorf(
				"build %s: Executor symbol %s lacks replay input",
				input.Bot.BotSpecID,
				cfg.executors[index].Resource.Symbol,
			)
		}
		cfg.executors[index].Meta = input.Meta
		cfg.executors[index].MinNotionalUSDC = input.MinNotionalUSDC
		cfg.executors[index].ResultPath = input.ResultPath
	}

	return bot.Definition{
		Identity: bot.Identity{
			SweepID:    input.Bot.SweepID,
			BotID:      input.Bot.BotID,
			BotSpecID:  input.Bot.BotSpecID,
			ConfigTOML: input.Bot.ConfigTOML,
			ConfigHash: input.Bot.ConfigHash,
		},
		SignalSymbol: replay.Symbol,
		MaxCycles:    cfg.maxCycles,
		Signaler:     signals,
		Risks:        risks,
		Executors:    cfg.executors,
	}, nil
}

// Section 2 - Domain Helpers

func stopRisks(risks []risk.Risk) {
	for index := len(risks) - 1; index >= 0; index-- {
		risks[index].Stop()
	}
}

// Section 3 - Generic Helpers
