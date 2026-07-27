package market

import (
	"errors"
	"testing"
)

// Section 1 - Program Flow

func TestMarketDataBuffersAndPublishesExactKey(t *testing.T) {
	var data = CreateMarketData()
	var key = Key{Venue: "simulator", Network: "simnet", Symbol: "BTC"}
	var calls int
	var subscription, err = data.SubscribeBBO(key, func() error {
		calls++
		var latest, found = data.LatestBBO(key)
		if !found || latest.TimestampMS != 1000 || latest.Price != 100 {
			t.Fatalf("callback read latest=%+v found=%t", latest, found)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	var bbo, _ = CreateBBO(1000, 100)
	if err = data.IngestBBO(key, bbo); err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if calls != 1 {
		t.Fatalf("callbacks=%d want=1", calls)
	}
	if err = subscription.Stop(); err != nil {
		t.Fatalf("stop subscription: %v", err)
	}
	if err = subscription.Stop(); err != nil {
		t.Fatalf("repeat stop subscription: %v", err)
	}
	var next, _ = CreateBBO(2000, 101)
	if err = data.IngestBBO(key, next); err != nil {
		t.Fatalf("ingest without subscriber: %v", err)
	}
	var latest, found = data.LatestBBO(key)
	if !found || latest != (BBO{Symbol: "BTC", TimestampMS: 2000, Price: 101}) {
		t.Fatalf("latest=%+v found=%t", latest, found)
	}
	if calls != 1 {
		t.Fatalf("stopped callback calls=%d want=1", calls)
	}
}

func TestMarketDataRejectsInvalidUpdateWithoutMutation(t *testing.T) {
	var data = CreateMarketData()
	var key = Key{Venue: "simulator", Network: "simnet", Symbol: "BTC"}
	var initial, _ = CreateBBO(2000, 100)
	if err := data.IngestBBO(key, initial); err != nil {
		t.Fatalf("ingest initial: %v", err)
	}
	var calls int
	var _, err = data.SubscribeBBO(key, func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	var stale, _ = CreateBBO(1000, 90)
	if err = data.IngestBBO(key, stale); err == nil {
		t.Fatal("stale update succeeded")
	}
	var latest, found = data.LatestBBO(key)
	if !found || latest.TimestampMS != 2000 || latest.Price != 100 {
		t.Fatalf("latest changed=%+v found=%t", latest, found)
	}
	if calls != 0 {
		t.Fatalf("invalid update callbacks=%d want=0", calls)
	}
}

func TestMarketDataReturnsAllCallbackFailures(t *testing.T) {
	var data = CreateMarketData()
	var key = Key{Venue: "simulator", Network: "simnet", Symbol: "BTC"}
	var first = errors.New("first")
	var second = errors.New("second")
	var calls int
	for _, failure := range []error{first, second} {
		var current = failure
		if _, err := data.SubscribeBBO(key, func() error {
			calls++
			return current
		}); err != nil {
			t.Fatalf("subscribe: %v", err)
		}
	}
	var bbo, _ = CreateBBO(1000, 100)
	var err = data.IngestBBO(key, bbo)
	if !errors.Is(err, first) || !errors.Is(err, second) {
		t.Fatalf("callback errors=%v", err)
	}
	if calls != 2 {
		t.Fatalf("callbacks=%d want=2", calls)
	}
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
