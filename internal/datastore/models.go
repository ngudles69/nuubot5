package datastore

import "time"

type BotSpec struct {
	Symbol      string
	TicksPath   string
	ReplayStart time.Time
	ReplayEnd   time.Time
	StartAt     *time.Time
	EndAt       *time.Time
}
