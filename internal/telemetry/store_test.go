package telemetry

import (
	"path/filepath"
	"testing"

	"github.com/shopspring/decimal"
)

// Section 1 - Program Flow

func TestStoreAppendsAndResumesSequence(t *testing.T) {
	var path = filepath.Join(t.TempDir(), "live.db")
	var first Store
	var lastSequence, err = first.Init(path)
	if err != nil {
		t.Fatal(err)
	}
	if lastSequence != 0 {
		t.Fatalf("initial sequence = %d, want 0", lastSequence)
	}
	var sample = telemetryStoreSample(1, 100, false)
	if err = first.Write(sample); err != nil {
		t.Fatal(err)
	}
	if err = first.Stop(); err != nil {
		t.Fatal(err)
	}

	var recovered Store
	lastSequence, err = recovered.Init(path)
	if err != nil {
		t.Fatal(err)
	}
	if lastSequence != 1 {
		t.Fatalf("recovered sequence = %d, want 1", lastSequence)
	}
	sample = telemetryStoreSample(2, 200, true)
	if err = recovered.Write(sample); err != nil {
		t.Fatal(err)
	}
	var count int
	var terminal bool
	err = recovered.db.QueryRow(`
		SELECT COUNT(*), MAX(terminal) FROM telemetry_sample
	`).Scan(&count, &terminal)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 || !terminal {
		t.Fatalf("count = %d terminal = %t, want 2/true", count, terminal)
	}
	if err = recovered.Stop(); err != nil {
		t.Fatal(err)
	}
}

// Section 2 - Domain Helpers

func telemetryStoreSample(sequence, timestampMS uint64, terminal bool) Sample {
	return Sample{
		Sequence:    sequence,
		TimestampMS: timestampMS,
		Terminal:    terminal,
		BotCapital:  decimal.NewFromInt(1000),
		BotBalance:  decimal.NewFromInt(1000),
		BotEquity:   decimal.NewFromInt(1000),
		PeakEquity:  decimal.NewFromInt(1000),
		Drawdown:    decimal.Zero,
		MaxDrawdown: decimal.Zero,
	}
}

// Section 3 - Generic Helpers
