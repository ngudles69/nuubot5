package meta

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	_ "modernc.org/sqlite"

	"nuubot/internal/hyperliquid"
)

type source struct {
	calls int
	meta  hyperliquid.PerpetualMeta
}

// Section 1 - Program Flow

func TestEnsureFreshStoresAndReusesPerpetualMeta(t *testing.T) {
	var fetcher = &source{meta: testMeta()}
	var path = filepath.Join(t.TempDir(), "nuubot.db")
	var now = time.Unix(1_700_000_000, 0).UTC()
	var first, err = EnsureFresh(
		context.Background(),
		path,
		"BTC",
		now,
		fetcher,
	)
	if err != nil {
		t.Fatalf("ensure fresh meta: %v", err)
	}
	var second Instrument
	second, err = EnsureFresh(
		context.Background(),
		path,
		"BTC",
		now.Add(23*time.Hour),
		fetcher,
	)
	if err != nil {
		t.Fatalf("reuse fresh meta: %v", err)
	}
	if fetcher.calls != 1 {
		t.Fatalf("actual fetches %d, expected 1", fetcher.calls)
	}
	if first != second || first.AssetID != 0 || first.PriceDecimals != 1 {
		t.Fatalf("unexpected instruments first=%+v second=%+v", first, second)
	}
}

func TestEnsureFreshRefreshesStaleMeta(t *testing.T) {
	var fetcher = &source{meta: testMeta()}
	var path = filepath.Join(t.TempDir(), "nuubot.db")
	var now = time.Unix(1_700_000_000, 0).UTC()
	var _, err = EnsureFresh(
		context.Background(),
		path,
		"BTC",
		now,
		fetcher,
	)
	if err != nil {
		t.Fatalf("seed meta: %v", err)
	}
	_, err = EnsureFresh(
		context.Background(),
		path,
		"BTC",
		now.Add(24*time.Hour),
		fetcher,
	)
	if err != nil {
		t.Fatalf("refresh stale meta: %v", err)
	}
	if fetcher.calls != 2 {
		t.Fatalf("actual fetches %d, expected 2", fetcher.calls)
	}
}

func TestEnsureFreshRefreshesMissingRows(t *testing.T) {
	var fetcher = &source{meta: testMeta()}
	var path = filepath.Join(t.TempDir(), "nuubot.db")
	var now = time.Unix(1_700_000_000, 0).UTC()
	var _, err = EnsureFresh(
		context.Background(),
		path,
		"BTC",
		now,
		fetcher,
	)
	if err != nil {
		t.Fatalf("seed meta: %v", err)
	}
	var db *sql.DB
	db, err = sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open meta database: %v", err)
	}
	if _, err = db.Exec(`DELETE FROM meta`); err != nil {
		t.Fatalf("delete meta rows: %v", err)
	}
	if err = db.Close(); err != nil {
		t.Fatalf("close meta database: %v", err)
	}
	_, err = EnsureFresh(
		context.Background(),
		path,
		"BTC",
		now.Add(time.Hour),
		fetcher,
	)
	if err != nil {
		t.Fatalf("refresh missing meta: %v", err)
	}
	if fetcher.calls != 2 {
		t.Fatalf("actual fetches %d, expected 2", fetcher.calls)
	}
}

func TestInstrumentRoundsHyperliquidValues(t *testing.T) {
	var instrument = Instrument{SizeDecimals: 5, PriceDecimals: 1}
	var size = instrument.RoundSize(decimal.RequireFromString("0.00001999"))
	if size.String() != "0.00001" {
		t.Fatalf("actual size %s, expected 0.00001", size)
	}
	var price = instrument.RoundPrice(decimal.RequireFromString("1234.56"))
	if price.String() != "1234.6" {
		t.Fatalf("actual price %s, expected 1234.6", price)
	}
	price = instrument.RoundPrice(decimal.RequireFromString("123456"))
	if price.String() != "123456" {
		t.Fatalf("actual integer price %s, expected 123456", price)
	}
}

// Section 2 - Domain Helpers

func (s *source) PerpetualMeta(context.Context) (hyperliquid.PerpetualMeta, error) {
	s.calls++
	return s.meta, nil
}

func testMeta() hyperliquid.PerpetualMeta {
	return hyperliquid.PerpetualMeta{
		Universe: []hyperliquid.PerpetualAsset{{
			Name:          "BTC",
			SizeDecimals:  5,
			MaxLeverage:   40,
			MarginTableID: 20,
			Raw:           `{"name":"BTC"}`,
		}},
		MarginTables: []hyperliquid.MarginTable{{
			ID: 20,
			Tiers: []hyperliquid.MarginTier{{
				LowerBound:  "0.0",
				MaxLeverage: 40,
			}},
		}},
	}
}

// Section 3 - Generic Helpers
