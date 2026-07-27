package market

import (
	"errors"
	"fmt"
	"math"
	"sync"
)

// Key identifies one exact market stream.
type Key struct {
	Venue   string
	Network string
	Symbol  string
}

// Callback consumes one published update by reading MarketData's latest value.
type Callback func() error

type subscriptionEntry struct {
	id       uint64
	callback Callback
}

// MarketData owns latest BBO values and exact-key subscriptions.
type MarketData struct {
	mu            sync.RWMutex
	latest        map[Key]BBO
	subscriptions map[Key][]subscriptionEntry
	nextID        uint64
	stopped       bool
}

// Subscription owns one removable MarketData callback registration.
type Subscription struct {
	marketData *MarketData
	key        Key
	id         uint64
	once       sync.Once
}

// Section 1 - Program Flow

// CreateMarketData creates one empty shared MarketData service.
func CreateMarketData() *MarketData {
	return &MarketData{
		latest:        make(map[Key]BBO),
		subscriptions: make(map[Key][]subscriptionEntry),
	}
}

// IngestBBO validates, buffers, and publishes one exact-key BBO update.
func (m *MarketData) IngestBBO(key Key, bbo BBO) error {
	// Step 1: validate market update
	if err := validateKey(key); err != nil {
		return err
	}
	if bbo.Symbol != "" && bbo.Symbol != key.Symbol {
		return fmt.Errorf("ingest MarketData BBO: symbol %s does not match %s", bbo.Symbol, key.Symbol)
	}
	if bbo.TimestampMS == 0 || math.IsNaN(bbo.Price) || math.IsInf(bbo.Price, 0) || bbo.Price <= 0 {
		return fmt.Errorf("ingest MarketData BBO: invalid timestamp or price")
	}
	bbo.Symbol = key.Symbol

	// Step 2: publish latest BBO and copy subscribers
	m.mu.Lock()
	if m.stopped {
		m.mu.Unlock()
		return fmt.Errorf("ingest MarketData BBO: MarketData is stopped")
	}
	var previous, found = m.latest[key]
	if found && bbo.TimestampMS < previous.TimestampMS {
		m.mu.Unlock()
		return fmt.Errorf("ingest MarketData BBO: timestamp moved backward")
	}
	m.latest[key] = bbo
	var subscribers = append([]subscriptionEntry(nil), m.subscriptions[key]...)
	m.mu.Unlock()

	// Step 3: notify matching subscribers
	var failures []error
	for _, subscriber := range subscribers {
		if err := subscriber.callback(); err != nil {
			failures = append(failures, err)
		}
	}
	return errors.Join(failures...)
}

// SubscribeBBO registers one exact-key callback.
func (m *MarketData) SubscribeBBO(key Key, callback Callback) (*Subscription, error) {
	// Step 1: validate subscription
	if err := validateKey(key); err != nil {
		return nil, err
	}
	if callback == nil {
		return nil, fmt.Errorf("subscribe MarketData BBO: callback is required")
	}

	// Step 2: register subscription
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopped {
		return nil, fmt.Errorf("subscribe MarketData BBO: MarketData is stopped")
	}
	m.nextID++
	m.subscriptions[key] = append(m.subscriptions[key], subscriptionEntry{
		id:       m.nextID,
		callback: callback,
	})
	return &Subscription{marketData: m, key: key, id: m.nextID}, nil
}

// Stop removes all subscriptions and stops MarketData admission.
func (m *MarketData) Stop() error {
	// Step 1: lock MarketData
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.stopped {
		return nil
	}

	// Step 2: clear MarketData state
	m.subscriptions = nil
	m.latest = nil
	// Step 3: mark MarketData stopped
	m.stopped = true
	return nil
}

// Stop removes one subscription idempotently.
func (s *Subscription) Stop() error {
	if s == nil {
		return nil
	}
	s.once.Do(func() {
		if s.marketData != nil {
			s.marketData.unsubscribe(s.key, s.id)
		}
	})
	return nil
}

// Section 2 - Domain Helpers

// LatestBBO returns one detached latest BBO for an exact market key.
func (m *MarketData) LatestBBO(key Key) (BBO, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var bbo, found = m.latest[key]
	return bbo, found
}

func (m *MarketData) unsubscribe(key Key, id uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var current = m.subscriptions[key]
	for index, subscriber := range current {
		if subscriber.id != id {
			continue
		}
		current = append(current[:index], current[index+1:]...)
		if len(current) == 0 {
			delete(m.subscriptions, key)
		} else {
			m.subscriptions[key] = current
		}
		return
	}
}

// Section 3 - Generic Helpers

func validateKey(key Key) error {
	if key.Venue == "" || key.Network == "" || key.Symbol == "" {
		return fmt.Errorf("MarketData key requires Venue, network, and symbol")
	}
	return nil
}
