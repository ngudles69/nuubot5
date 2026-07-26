// Package ledger owns one Account's coherent Trades, Orders, and Fills.
package ledger

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"

	"nuubot/internal/order"
	"nuubot/internal/trade"
)

const (
	// None keeps Ledger evidence in memory until terminal publication.
	None = "none"
	// Max persists every accepted Ledger mutation.
	Max = "max"
)

// Config contains one Ledger's identity and persistence policy.
type Config struct {
	ID             uint64
	CycleNumber    int
	ExecutorNumber int
	Account        string
	Network        string
	Symbol         string
	PersistMode    string
	Path           string
}

// Plan reserves no state and describes the next admissible Trade batch identity.
type Plan struct {
	Trade    trade.Input
	OrderIDs []uint64
}

// SubmitOutcome contains one ordered Venue submission acknowledgement.
type SubmitOutcome struct {
	OrderID      uint64
	VenueOrderID uint64
	Error        string
	Raw          string
	Status       order.Status
	TimestampMS  uint64
}

// OrderEvidence contains one normalized canonical Venue Order observation.
type OrderEvidence struct {
	CLOID        string
	VenueOrderID uint64
	Status       order.Status
	RejectReason string
	TimestampMS  uint64
	Raw          string
}

// FillEvidence contains one normalized canonical Venue Fill observation.
type FillEvidence struct {
	CLOID        string
	VenueOrderID uint64
	VenueTID     uint64
	Side         string
	Quantity     decimal.Decimal
	Price        decimal.Decimal
	TimestampMS  uint64
	Fee          *decimal.Decimal
	Liquidity    string
	Raw          string
}

// ReconInput contains one complete validated reconciliation batch.
type ReconInput struct {
	Orders          []OrderEvidence
	Fills           []FillEvidence
	FillsThroughMS  uint64
	ObservedMS      uint64
	AccountStateRaw string
}

// Result contains one immutable terminal Ledger value.
type Result struct {
	Config
	FillsThroughMS uint64
	LastReconMS    uint64
	AccountState   string
	Trades         []trade.Snapshot
}

// Ledger owns one coherent local trading tree.
type Ledger struct {
	config          Config
	trades          map[uint64]*trade.Trade
	nextTradeID     uint64
	nextTradeNo     uint32
	nextOrderID     uint64
	fillsThroughMS  uint64
	lastReconMS     uint64
	accountStateRaw string
	store           *ledgerStore
	started         bool
	stopped         bool
}

// Section 1 - Program Flow

// Init prepares one empty Ledger.
func (l *Ledger) Init(cfg Config) error {
	// bind Ledger inputs
	if l.started || l.stopped {
		return fmt.Errorf("initialize ledger: invalid lifecycle state")
	}
	l.config = cfg

	// validate persistence mode
	if cfg.ID == 0 || cfg.CycleNumber <= 0 || cfg.ExecutorNumber <= 0 ||
		cfg.Account == "" || cfg.Symbol == "" {
		return fmt.Errorf("initialize ledger: complete identity is required")
	}
	if cfg.Network != "simnet" && cfg.Network != "testnet" && cfg.Network != "mainnet" {
		return fmt.Errorf("initialize ledger: invalid network %q", cfg.Network)
	}
	if cfg.PersistMode != None && cfg.PersistMode != Max {
		return fmt.Errorf("initialize ledger: invalid persistence mode %q", cfg.PersistMode)
	}
	if cfg.PersistMode == Max && cfg.Path == "" {
		return fmt.Errorf("initialize ledger: max persistence requires path")
	}

	// initialize Ledger
	l.trades = make(map[uint64]*trade.Trade)
	l.nextTradeID = 1
	l.nextTradeNo = 1
	l.nextOrderID = 1

	// open Ledger identity when configured
	if cfg.PersistMode == Max {
		var err error
		l.store, err = openLedgerStore(cfg.Path)
		if err != nil {
			return err
		}

		// load Ledger evidence when configured
		var loaded candidate
		var found bool
		loaded, found, err = l.store.load(cfg)
		if err != nil {
			l.store.close()
			return err
		}
		if found {
			l.publish(loaded)
		} else {
			err = l.store.save(cfg, l.currentCandidate())
			if err != nil {
				l.store.close()
				return err
			}
		}
	}

	// index active evidence
	l.started = true
	return nil
}

// CreateTrade publishes one validated Trade and its initial Orders.
func (l *Ledger) CreateTrade(input trade.Input, orders []*order.Order) error {
	if !l.started || l.stopped {
		return fmt.Errorf("create ledger trade: invalid lifecycle state")
	}

	// stage Trade and initial Orders
	if input != l.nextTradeInput() || len(orders) == 0 {
		return fmt.Errorf("create ledger trade: unexpected Trade identity or empty Orders")
	}
	var staged, err = trade.New(input)
	if err != nil {
		return fmt.Errorf("create ledger trade: %w", err)
	}
	for index, created := range orders {
		if created == nil || created.Snapshot().OrderID != l.nextOrderID+uint64(index) {
			return fmt.Errorf("create ledger trade: unexpected Order identity")
		}
		err = staged.AddOrder(created)
		if err != nil {
			return fmt.Errorf("create ledger trade: %w", err)
		}
	}

	// persist staged tree when configured
	err = l.persistCandidate(candidate{
		trades:          withTrade(l.trades, staged),
		nextTradeID:     l.nextTradeID + 1,
		nextTradeNo:     l.nextTradeNo + 1,
		nextOrderID:     l.nextOrderID + uint64(len(orders)),
		fillsThroughMS:  l.fillsThroughMS,
		lastReconMS:     l.lastReconMS,
		accountStateRaw: l.accountStateRaw,
	})
	if err != nil {
		return err
	}

	// publish Trade and Orders
	l.trades[input.TradeID] = staged
	l.nextTradeID++
	l.nextTradeNo++
	l.nextOrderID += uint64(len(orders))
	return nil
}

// AddOrders publishes later Orders under one existing Trade.
func (l *Ledger) AddOrders(tradeID uint64, orders []*order.Order) error {
	if !l.started || l.stopped {
		return fmt.Errorf("add ledger orders: invalid lifecycle state")
	}

	// stage existing Trade
	var staged = cloneTrades(l.trades)
	var owned, exists = staged[tradeID]
	if !exists || len(orders) == 0 {
		return fmt.Errorf("add ledger orders: unknown Trade or empty Orders")
	}
	for index, created := range orders {
		if created == nil || created.Snapshot().OrderID != l.nextOrderID+uint64(index) {
			return fmt.Errorf("add ledger orders: unexpected Order identity")
		}
		var err = owned.AddOrder(created)
		if err != nil {
			return fmt.Errorf("add ledger orders: %w", err)
		}
	}

	// persist staged Orders when configured
	var err = l.persistCandidate(candidate{
		trades:          staged,
		nextTradeID:     l.nextTradeID,
		nextTradeNo:     l.nextTradeNo,
		nextOrderID:     l.nextOrderID + uint64(len(orders)),
		fillsThroughMS:  l.fillsThroughMS,
		lastReconMS:     l.lastReconMS,
		accountStateRaw: l.accountStateRaw,
	})
	if err != nil {
		return err
	}

	// publish Orders
	l.trades = staged
	l.nextOrderID += uint64(len(orders))
	return nil
}

// RecordSubmit publishes one complete ordered submission batch.
func (l *Ledger) RecordSubmit(outcomes []SubmitOutcome) error {
	if !l.started || l.stopped {
		return fmt.Errorf("record ledger submit: invalid lifecycle state")
	}

	// stage submitted Orders
	if len(outcomes) == 0 {
		return fmt.Errorf("record ledger submit: outcomes are required")
	}
	var staged = cloneTrades(l.trades)
	var orders = indexOrders(staged)
	var seen = make(map[uint64]struct{}, len(outcomes))

	// apply ordered outcomes
	for _, outcome := range outcomes {
		if _, duplicate := seen[outcome.OrderID]; duplicate {
			return fmt.Errorf("record ledger submit: duplicate Order %d", outcome.OrderID)
		}
		var owned = orders[outcome.OrderID]
		if owned == nil {
			return fmt.Errorf("record ledger submit: unknown Order %d", outcome.OrderID)
		}
		var err = owned.RecordSubmit(outcome.VenueOrderID, outcome.Error, outcome.Raw)
		if err != nil {
			return fmt.Errorf("record ledger submit: %w", err)
		}
		if outcome.Status != "" {
			if (outcome.Status != order.Rejected && outcome.Status != order.Error) ||
				outcome.TimestampMS == 0 {
				return fmt.Errorf("record ledger submit: invalid terminal outcome")
			}
			err = owned.ApplyVenueState(order.VenueState{
				VenueOrderID: outcome.VenueOrderID,
				Status:       outcome.Status,
				RejectReason: outcome.Error,
				TimestampMS:  outcome.TimestampMS,
				Raw:          outcome.Raw,
			})
			if err != nil {
				return fmt.Errorf("record ledger submit: %w", err)
			}
		}
		seen[outcome.OrderID] = struct{}{}
	}
	for _, ownedTrade := range staged {
		var err = ownedTrade.Refresh()
		if err != nil {
			return fmt.Errorf("record ledger submit: %w", err)
		}
	}

	// persist submission evidence when configured
	var err = l.persistCandidate(candidate{
		trades:          staged,
		nextTradeID:     l.nextTradeID,
		nextTradeNo:     l.nextTradeNo,
		nextOrderID:     l.nextOrderID,
		fillsThroughMS:  l.fillsThroughMS,
		lastReconMS:     l.lastReconMS,
		accountStateRaw: l.accountStateRaw,
	})
	if err != nil {
		return err
	}

	// publish submit result
	l.trades = staged
	return nil
}

// Result returns one immutable terminal Ledger value.
func (l *Ledger) Result(markPrice *decimal.Decimal) (Result, error) {
	if !l.started && !l.stopped {
		return Result{}, fmt.Errorf("read ledger result: ledger is not initialized")
	}

	// copy Trades Orders and Fills
	var trades = make([]trade.Snapshot, 0, len(l.trades))
	for _, owned := range l.trades {
		var snapshot, err = owned.Snapshot(markPrice)
		if err != nil {
			return Result{}, fmt.Errorf("read ledger result: %w", err)
		}
		trades = append(trades, snapshot)
	}
	sort.Slice(trades, func(left int, right int) bool {
		return trades[left].TradeID < trades[right].TradeID
	})

	// copy reconciliation cursor and snapshot
	return Result{
		Config:         l.config,
		FillsThroughMS: l.fillsThroughMS,
		LastReconMS:    l.lastReconMS,
		AccountState:   l.accountStateRaw,
		Trades:         trades,
	}, nil
}

// Stop stops Ledger admission.
func (l *Ledger) Stop() error {
	if l.stopped {
		return nil
	}

	// stop Ledger
	var err error
	if l.store != nil {
		err = l.store.close()
	}
	l.started = false
	l.stopped = true
	return err
}

// Section 2 - Domain Helpers

// PlanTrade returns the next synchronous Trade and Order identities.
func (l *Ledger) PlanTrade(orderCount int) (Plan, error) {
	if !l.started || l.stopped || orderCount <= 0 || orderCount > 1000 {
		return Plan{}, fmt.Errorf("plan ledger trade: invalid state or Order count")
	}
	var ids = make([]uint64, orderCount)
	for index := range ids {
		ids[index] = l.nextOrderID + uint64(index)
	}
	return Plan{Trade: l.nextTradeInput(), OrderIDs: ids}, nil
}

// PlanOrders returns the next synchronous Order identities.
func (l *Ledger) PlanOrders(orderCount int) ([]uint64, error) {
	if !l.started || l.stopped || orderCount <= 0 || orderCount > 1000 {
		return nil, fmt.Errorf("plan ledger Orders: invalid state or Order count")
	}
	var ids = make([]uint64, orderCount)
	for index := range ids {
		ids[index] = l.nextOrderID + uint64(index)
	}
	return ids, nil
}

// ActiveOrders returns current active Order snapshots.
func (l *Ledger) ActiveOrders() []order.Snapshot {
	var active []order.Snapshot
	for _, ownedTrade := range l.trades {
		for _, ownedOrder := range ownedTrade.Orders() {
			if ownedOrder.Active {
				active = append(active, ownedOrder)
			}
		}
	}
	sort.Slice(active, func(left int, right int) bool {
		return active[left].OrderID < active[right].OrderID
	})
	return active
}

// Trade returns one current marked Trade snapshot.
func (l *Ledger) Trade(tradeID uint64, markPrice *decimal.Decimal) (trade.Snapshot, error) {
	var owned = l.trades[tradeID]
	if owned == nil {
		return trade.Snapshot{}, fmt.Errorf("read ledger Trade: unknown Trade %d", tradeID)
	}
	var snapshot, err = owned.Snapshot(markPrice)
	if err != nil {
		return trade.Snapshot{}, fmt.Errorf("read ledger Trade: %w", err)
	}
	return snapshot, nil
}

// TradeCount returns the number of owned Trades.
func (l *Ledger) TradeCount() int {
	return len(l.trades)
}

// Orders returns selected current Order snapshots.
func (l *Ledger) Orders(orderIDs []uint64) ([]order.Snapshot, error) {
	var indexed = indexOrders(l.trades)
	var snapshots = make([]order.Snapshot, 0, len(orderIDs))
	for _, orderID := range orderIDs {
		var owned = indexed[orderID]
		if owned == nil {
			return nil, fmt.Errorf("read ledger Orders: unknown Order %d", orderID)
		}
		snapshots = append(snapshots, owned.Snapshot())
	}
	return snapshots, nil
}

// FillsThroughMS returns the inclusive Fill cursor.
func (l *Ledger) FillsThroughMS() uint64 {
	return l.fillsThroughMS
}

func (l *Ledger) nextTradeInput() trade.Input {
	return trade.Input{
		LedgerID:    l.config.ID,
		TradeID:     l.nextTradeID,
		TradeNo:     l.nextTradeNo,
		Account:     l.config.Account,
		CycleNumber: l.config.CycleNumber,
		Symbol:      l.config.Symbol,
	}
}

type candidate struct {
	trades          map[uint64]*trade.Trade
	nextTradeID     uint64
	nextTradeNo     uint32
	nextOrderID     uint64
	fillsThroughMS  uint64
	lastReconMS     uint64
	accountStateRaw string
}

func (l *Ledger) persistCandidate(state candidate) error {
	if l.config.PersistMode == None {
		return nil
	}
	return l.store.save(l.config, state)
}

func (l *Ledger) currentCandidate() candidate {
	return candidate{
		trades:          l.trades,
		nextTradeID:     l.nextTradeID,
		nextTradeNo:     l.nextTradeNo,
		nextOrderID:     l.nextOrderID,
		fillsThroughMS:  l.fillsThroughMS,
		lastReconMS:     l.lastReconMS,
		accountStateRaw: l.accountStateRaw,
	}
}

func (l *Ledger) publish(state candidate) {
	l.trades = state.trades
	l.nextTradeID = state.nextTradeID
	l.nextTradeNo = state.nextTradeNo
	l.nextOrderID = state.nextOrderID
	l.fillsThroughMS = state.fillsThroughMS
	l.lastReconMS = state.lastReconMS
	l.accountStateRaw = state.accountStateRaw
}

// Section 3 - Generic Helpers

func cloneTrades(source map[uint64]*trade.Trade) map[uint64]*trade.Trade {
	var cloned = make(map[uint64]*trade.Trade, len(source))
	for id, owned := range source {
		cloned[id] = owned.Clone()
	}
	return cloned
}

func withTrade(
	source map[uint64]*trade.Trade,
	created *trade.Trade,
) map[uint64]*trade.Trade {
	var cloned = cloneTrades(source)
	var snapshot = created.State()
	cloned[snapshot.TradeID] = created
	return cloned
}

func indexOrders(trades map[uint64]*trade.Trade) map[uint64]*order.Order {
	var indexed = make(map[uint64]*order.Order)
	for _, ownedTrade := range trades {
		for _, snapshot := range ownedTrade.Orders() {
			var owned, _ = ownedTrade.Order(snapshot.OrderID)
			indexed[snapshot.OrderID] = owned
		}
	}
	return indexed
}

func indexCLOIDs(trades map[uint64]*trade.Trade) map[string]*order.Order {
	var indexed = make(map[string]*order.Order)
	for _, ownedTrade := range trades {
		for _, snapshot := range ownedTrade.Orders() {
			var owned, _ = ownedTrade.Order(snapshot.OrderID)
			indexed[snapshot.CLOID] = owned
		}
	}
	return indexed
}
