package datastore

import "time"

// ReplayInput identifies one bounded historical input.
type ReplayInput struct {
	Symbol      string
	TicksPath   string
	ReplayStart time.Time
	ReplayEnd   time.Time
	StartAt     *time.Time
	EndAt       *time.Time
}

// BotStatus identifies one stored Bot lifecycle state.
type BotStatus string

const (
	BotConfigured BotStatus = "configured"
	BotStarting   BotStatus = "starting"
	BotRunning    BotStatus = "running"
	BotPaused     BotStatus = "paused"
	BotStopping   BotStatus = "stopping"
	BotStopped    BotStatus = "stopped"
	BotError      BotStatus = "error"
)

// Bot contains one exact stored Bot configuration, runtime status, and replay input.
type Bot struct {
	SweepID    uint64
	BotID      uint64
	BotSpecID  string
	ConfigTOML string
	ConfigHash string
	Status     BotStatus
	Replay     ReplayInput
}

// Section 1 - Program Flow

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
