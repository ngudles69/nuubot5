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

// Bot contains one exact stored Bot configuration and its replay input.
type Bot struct {
	SweepID    uint64
	BotID      uint64
	BotSpecID  string
	ConfigTOML string
	ConfigHash string
	Replay     ReplayInput
}

// Section 1 - Program Flow

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
