// Package ledger owns one Account's coherent Trades, Orders, and Fills.
package ledger

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"

	"nuubot/internal/account/fill"
	"nuubot/internal/account/order"
	"nuubot/internal/account/trade"
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

// Result contains one terminal Ledger summary without owned domain records.
type Result struct {
	Config
	FillsThroughMS uint64
	LastReconMS    uint64
	AccountState   string
	Summary        Summary
	Trades         int
	Orders         int
	Fills          int
	Cancellations  int
	StopOrders     int
}

// Summary contains cached Ledger totals for one Account observation.
type Summary struct {
	RealizedPnL   decimal.Decimal
	UnrealizedPnL decimal.Decimal
	GrossPnL      decimal.Decimal
	Fees          decimal.Decimal
	NetPnL        decimal.Decimal
	OpenTrades    int
	ActiveOrders  int
	Fills         int
	PendingOrders int
	PendingFills  int
}

// PendingFillAnchor identifies one stable missing-fee repair target.
type PendingFillAnchor struct {
	VenueTID    uint64
	TimestampMS uint64
}

// FillChangeKind identifies one staged Fill publication outcome.
type FillChangeKind string

const (
	// FillAdded identifies one newly admitted Fill.
	FillAdded FillChangeKind = "added"
	// FillFeeEnriched identifies one existing Fill whose fee became present.
	FillFeeEnriched FillChangeKind = "fee_enriched"
)

// FillChange contains one identity-bearing staged Fill change.
type FillChange struct {
	Kind     FillChangeKind
	VenueTID uint64
	HasFee   bool
	Fee      decimal.Decimal
}

type orderLocation struct {
	tradeID uint64
	orderID uint64
}

type fillLocation struct {
	tradeID  uint64
	orderID  uint64
	venueTID uint64
}

// Ledger owns one coherent local trading tree.
type Ledger struct {
	config          Config
	trades          map[uint64]*trade.Trade
	orders          map[uint64]orderLocation
	cloids          map[string]orderLocation
	venueOrders     map[uint64]orderLocation
	fills           map[uint64]fillLocation
	activeTrades    map[uint64]struct{}
	activeOrders    map[uint64]struct{}
	pendingOrders   map[uint64]struct{}
	pendingFills    map[uint64]struct{}
	summary         Summary
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

// CreateTrade publishes one validated Trade and its initial Orders.
func (l *Ledger) CreateTrade(input trade.Input, orders []*order.Order) error {
	if !l.started || l.stopped {
		return fmt.Errorf("create ledger trade: invalid lifecycle state")
	}

	// Step 1: prepare Trade and initial Orders
	if input != l.nextTradeInput() || len(orders) == 0 {
		return fmt.Errorf("create ledger trade: unexpected Trade identity or empty Orders")
	}
	var staged, err = trade.New(input)
	if err != nil {
		return fmt.Errorf("create ledger trade: %w", err)
	}
	for index, created := range orders {
		if created == nil || created.ReconState().OrderID != l.nextOrderID+uint64(index) {
			return fmt.Errorf("create ledger trade: unexpected Order identity")
		}
		err = staged.AddOrder(created)
		if err != nil {
			return fmt.Errorf("create ledger trade: %w", err)
		}
	}
	err = l.validateNewOrderIndexes(nil, orders)
	if err != nil {
		return fmt.Errorf("create ledger trade: %w", err)
	}

	// Step 2: persist new Trade and Orders when configured
	if l.config.PersistMode == Max {
		var state = l.currentCandidate()
		state.nextTradeID++
		state.nextTradeNo++
		state.nextOrderID += uint64(len(orders))
		err = l.store.saveMutation(l.config, state, []*trade.Trade{staged}, orders)
		if err != nil {
			return err
		}
	}

	// Step 3: publish Trade, Orders, and exact Summary delta
	l.trades[input.TradeID] = staged
	l.replaceTradeSummary(nil, staged.Summary())
	l.nextTradeID++
	l.nextTradeNo++
	l.nextOrderID += uint64(len(orders))
	var indexErr = l.replaceTradeIndexes(input.TradeID, nil, staged)
	if indexErr != nil {
		return fmt.Errorf("create ledger trade: %w", indexErr)
	}
	return nil
}

// AddOrders publishes later Orders under one existing Trade.
func (l *Ledger) AddOrders(tradeID uint64, orders []*order.Order) error {
	if !l.started || l.stopped {
		return fmt.Errorf("add ledger orders: invalid lifecycle state")
	}

	// Step 1: validate new Orders
	var owned = l.trades[tradeID]
	if owned == nil || len(orders) == 0 {
		return fmt.Errorf("add ledger orders: unknown Trade or empty Orders")
	}
	var err = l.validateAddedOrders(tradeID, owned, orders)
	if err != nil {
		return fmt.Errorf("add ledger orders: %w", err)
	}
	// Step 2: attach validated Orders directly
	// Venue truth owns recovery; copying unchanged Ledger records cannot restore trust.
	var previousSummary = owned.Summary()
	defer func() {
		var currentSummary = owned.Summary()
		l.replaceTradeSummary(&previousSummary, currentSummary)
	}()
	for _, created := range orders {
		err = owned.AddOrder(created)
		if err != nil {
			return fmt.Errorf("add ledger orders: %w", err)
		}
	}

	// Step 3: persist touched Trade and Orders when configured
	if l.config.PersistMode == Max {
		var state = l.currentCandidate()
		state.nextOrderID += uint64(len(orders))
		err = l.store.saveMutation(l.config, state, []*trade.Trade{owned}, orders)
		if err != nil {
			return err
		}
	}

	// Step 4: publish Orders and exact Trade Summary delta
	l.nextOrderID += uint64(len(orders))
	l.addValidatedTradeIndexes(tradeID, owned)
	return nil
}

// RecordSubmit publishes one complete ordered submission batch.
func (l *Ledger) RecordSubmit(outcomes []SubmitOutcome) error {
	if !l.started || l.stopped {
		return fmt.Errorf("record ledger submit: invalid lifecycle state")
	}

	// Step 1: validate submitted Orders
	if len(outcomes) == 0 {
		return fmt.Errorf("record ledger submit: outcomes are required")
	}
	var submitted, touchedTrades, err = l.validateSubmitOutcomes(outcomes)
	if err != nil {
		return err
	}

	// Step 2: apply ordered outcomes directly
	// Venue truth owns recovery; copying unchanged Ledger records cannot restore trust.
	var previousSummaries = make(map[uint64]trade.Summary, len(touchedTrades))
	for _, ownedTrade := range touchedTrades {
		var state = ownedTrade.ReconState()
		previousSummaries[state.TradeID] = ownedTrade.Summary()
	}
	defer func() {
		for _, ownedTrade := range touchedTrades {
			var state = ownedTrade.ReconState()
			var previous = previousSummaries[state.TradeID]
			l.replaceTradeSummary(&previous, ownedTrade.Summary())
		}
	}()
	for index, outcome := range outcomes {
		var owned = submitted[index]
		err = owned.RecordSubmit(outcome.VenueOrderID, outcome.Error, outcome.Raw)
		if err != nil {
			return fmt.Errorf("record ledger submit: %w", err)
		}
		if outcome.Status != "" {
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
	}
	for _, ownedTrade := range touchedTrades {
		err = ownedTrade.Refresh()
		if err != nil {
			return fmt.Errorf("record ledger submit: %w", err)
		}
	}

	// Step 3: persist touched submission rows when configured
	if l.config.PersistMode == Max {
		err = l.store.saveMutation(l.config, l.currentCandidate(), touchedTrades, submitted)
		if err != nil {
			return err
		}
	}

	// Step 4: refresh touched indexes and exact Trade Summary deltas
	for _, ownedTrade := range touchedTrades {
		var state = ownedTrade.ReconState()
		l.addValidatedTradeIndexes(state.TradeID, ownedTrade)
	}
	return nil
}

// Result returns one terminal Ledger summary.
func (l *Ledger) Result() (Result, error) {
	// Step 1: validate result state
	if !l.started && !l.stopped {
		return Result{}, fmt.Errorf("read ledger result: ledger is not initialized")
	}

	// Step 2: aggregate terminal domain counts
	var orders int
	var fills int
	var cancellations int
	var stopOrders int
	for _, owned := range l.trades {
		var err = owned.EachOrder(func(ownedOrder *order.Order) error {
			var state = ownedOrder.ReconState()
			orders++
			fills += state.FillCount
			if state.Status == order.Canceled {
				cancellations++
			}
			if state.Role == order.Stop {
				stopOrders++
			}
			return nil
		})
		if err != nil {
			return Result{}, err
		}
	}

	// Step 3: return terminal Ledger result
	return Result{
		Config:         l.config,
		FillsThroughMS: l.fillsThroughMS,
		LastReconMS:    l.lastReconMS,
		AccountState:   l.accountStateRaw,
		Summary:        l.summary,
		Trades:         len(l.trades),
		Orders:         orders,
		Fills:          fills,
		Cancellations:  cancellations,
		StopOrders:     stopOrders,
	}, nil
}

// Stop stops Ledger admission.
func (l *Ledger) Stop() error {
	if l.stopped {
		return nil
	}

	// Step 1: stop Ledger
	var err error
	if l.store != nil {
		err = l.store.close()
	}
	l.started = false
	l.stopped = true
	return err
}

// Section 2 - Domain Helpers

// Section 2.1 - Identity Planning

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

// Section 2.2 - Domain Observation

// ActiveOrders returns focused current active Order evidence.
func (l *Ledger) ActiveOrders() []order.ActiveState {
	var ids = sortedSet(l.activeOrders)
	var active = make([]order.ActiveState, 0, len(ids))
	for _, orderID := range ids {
		var location = l.orders[orderID]
		var ownedTrade = l.trades[location.tradeID]
		var ownedOrder, _ = ownedTrade.Order(orderID)
		active = append(active, ownedOrder.ActiveState())
	}
	return active
}

// TradeState returns focused current Trade state.
func (l *Ledger) TradeState(tradeID uint64) (trade.ReconState, error) {
	var owned = l.trades[tradeID]
	if owned == nil {
		return trade.ReconState{}, fmt.Errorf("read ledger Trade: unknown Trade %d", tradeID)
	}
	return owned.ReconState(), nil
}

// NextBatchNo returns the next Order batch number for one Trade.
func (l *Ledger) NextBatchNo(tradeID uint64) (uint16, error) {
	var owned = l.trades[tradeID]
	if owned == nil {
		return 0, fmt.Errorf("read ledger Trade: unknown Trade %d", tradeID)
	}
	var batchNo uint16 = 1
	var err = owned.EachOrder(func(current *order.Order) error {
		var state = current.ReconState()
		if state.BatchNo >= batchNo {
			batchNo = state.BatchNo + 1
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	if batchNo > 1000 {
		return 0, fmt.Errorf("read ledger Trade: Trade exceeded 1000 batches")
	}
	return batchNo, nil
}

// OpenTrades returns focused state for current open exposure.
func (l *Ledger) OpenTrades() []trade.ReconState {
	var ids = sortedSet(l.activeTrades)
	var states = make([]trade.ReconState, 0, len(ids))
	for _, tradeID := range ids {
		var state = l.trades[tradeID].ReconState()
		if state.OpenQuantity.IsPositive() {
			states = append(states, state)
		}
	}
	return states
}

// CountOrders returns the number of Orders matching role and status.
func (l *Ledger) CountOrders(role string, status order.Status) uint64 {
	var count uint64
	for _, ownedTrade := range l.trades {
		ownedTrade.EachOrder(func(owned *order.Order) error {
			var current = owned.ReconState()
			if current.Role == role && current.Status == status {
				count++
			}
			return nil
		})
	}
	return count
}

// TradeCount returns the number of owned Trades.
func (l *Ledger) TradeCount() int {
	return len(l.trades)
}

// TradeOrders returns flat Order records linked by Trade identity.
func (l *Ledger) TradeOrders(tradeID uint64) ([]order.Record, error) {
	var owned = l.trades[tradeID]
	if owned == nil {
		return nil, fmt.Errorf("read ledger Trade Orders: unknown Trade %d", tradeID)
	}
	var records = make([]order.Record, 0)
	var err = owned.EachOrder(func(current *order.Order) error {
		records = append(records, current.Record())
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(records, func(left int, right int) bool {
		return records[left].OrderID < records[right].OrderID
	})
	return records, nil
}

// Orders returns selected flat Order records.
func (l *Ledger) Orders(orderIDs []uint64) ([]order.Record, error) {
	var records = make([]order.Record, 0, len(orderIDs))
	for _, orderID := range orderIDs {
		var location, exists = l.orders[orderID]
		if !exists {
			return nil, fmt.Errorf("read ledger Orders: unknown Order %d", orderID)
		}
		var owned, _ = l.trades[location.tradeID].Order(orderID)
		records = append(records, owned.Record())
	}
	return records, nil
}

// FillsThroughMS returns the inclusive Fill cursor.
func (l *Ledger) FillsThroughMS() uint64 {
	return l.fillsThroughMS
}

// Fill returns one current Fill by Venue identity.
func (l *Ledger) Fill(venueTID uint64) (fill.Record, bool) {
	var location, exists = l.fills[venueTID]
	if !exists {
		return fill.Record{}, false
	}
	var ownedTrade = l.trades[location.tradeID]
	var ownedOrder, _ = ownedTrade.Order(location.orderID)
	var execution, found = ownedOrder.Fill(venueTID)
	if !found {
		return fill.Record{}, false
	}
	return execution.State(), true
}

// PendingCounts returns current reconciliation-pending Order and Fill counts.
func (l *Ledger) PendingCounts() (int, int) {
	return len(l.pendingOrders), len(l.pendingFills)
}

// PendingFillAnchors returns stable missing-fee repair targets.
func (l *Ledger) PendingFillAnchors() []PendingFillAnchor {
	var anchors = make([]PendingFillAnchor, 0, len(l.pendingFills))
	for venueTID := range l.pendingFills {
		var execution, exists = l.Fill(venueTID)
		if exists {
			anchors = append(anchors, PendingFillAnchor{
				VenueTID: venueTID, TimestampMS: execution.TimestampMS,
			})
		}
	}
	sort.Slice(anchors, func(left int, right int) bool {
		if anchors[left].TimestampMS == anchors[right].TimestampMS {
			return anchors[left].VenueTID < anchors[right].VenueTID
		}
		return anchors[left].TimestampMS < anchors[right].TimestampMS
	})
	return anchors
}

// ReconFillChanges returns staged Fill additions and fee enrichments.
func (l *Ledger) ReconFillChanges(attempt *ReconAttempt) []FillChange {
	if attempt == nil {
		return nil
	}
	return append([]FillChange(nil), attempt.fillChanges...)
}

// HasPendingRecon reports selected unresolved Order or Fill work.
func (l *Ledger) HasPendingRecon() bool {
	return len(l.pendingOrders) != 0 || len(l.pendingFills) != 0
}

// Summary returns current cached finance and evidence totals.
func (l *Ledger) Summary() Summary {
	return l.summary
}

// Section 2.3 - Mutation Validation

func (l *Ledger) validateAddedOrders(
	tradeID uint64,
	owned *trade.Trade,
	orders []*order.Order,
) error {
	var tradeState = owned.ReconState()
	if !tradeIsActive(tradeState.Status) {
		return fmt.Errorf("add trade order: terminal trade cannot change")
	}
	for index, created := range orders {
		if created == nil {
			return fmt.Errorf("unexpected Order identity")
		}
		var state = created.ReconState()
		if state.OrderID != l.nextOrderID+uint64(index) {
			return fmt.Errorf("unexpected Order identity")
		}
		if state.LedgerID != l.config.ID || state.TradeID != tradeID ||
			state.Account != l.config.Account || state.CycleNumber != l.config.CycleNumber ||
			state.Symbol != l.config.Symbol {
			return fmt.Errorf("add trade order: ownership mismatch")
		}
	}
	return l.validateNewOrderIndexes(owned, orders)
}

func (l *Ledger) validateNewOrderIndexes(
	owned *trade.Trade,
	orders []*order.Order,
) error {
	type position struct {
		batchNo  uint16
		orderPos uint16
	}
	var positions = make(map[position]struct{})
	if owned != nil {
		owned.EachOrder(func(existing *order.Order) error {
			var state = existing.ReconState()
			positions[position{batchNo: state.BatchNo, orderPos: state.OrderPos}] = struct{}{}
			return nil
		})
	}
	var seenCLOIDs = make(map[string]struct{}, len(orders))
	for _, created := range orders {
		var state = created.ReconState()
		if _, exists := l.orders[state.OrderID]; exists {
			return fmt.Errorf("add trade order: duplicate order id %d", state.OrderID)
		}
		if _, exists := l.cloids[state.CLOID]; exists {
			return fmt.Errorf("add trade order: duplicate cloid %s", state.CLOID)
		}
		if _, exists := seenCLOIDs[state.CLOID]; exists {
			return fmt.Errorf("add trade order: duplicate cloid %s", state.CLOID)
		}
		var current = position{batchNo: state.BatchNo, orderPos: state.OrderPos}
		if _, exists := positions[current]; exists {
			return fmt.Errorf(
				"add trade order: duplicate batch %d position %d",
				state.BatchNo,
				state.OrderPos,
			)
		}
		seenCLOIDs[state.CLOID] = struct{}{}
		positions[current] = struct{}{}
	}
	return nil
}

func (l *Ledger) validateSubmitOutcomes(
	outcomes []SubmitOutcome,
) ([]*order.Order, []*trade.Trade, error) {
	var submitted = make([]*order.Order, 0, len(outcomes))
	var tradesByID = make(map[uint64]*trade.Trade)
	var seenOrders = make(map[uint64]struct{}, len(outcomes))
	var seenVenueOrders = make(map[uint64]uint64, len(outcomes))
	for _, outcome := range outcomes {
		if _, duplicate := seenOrders[outcome.OrderID]; duplicate {
			return nil, nil, fmt.Errorf("record ledger submit: duplicate Order %d", outcome.OrderID)
		}
		var location, exists = l.orders[outcome.OrderID]
		if !exists {
			return nil, nil, fmt.Errorf("record ledger submit: unknown Order %d", outcome.OrderID)
		}
		var ownedTrade = l.trades[location.tradeID]
		var owned, _ = ownedTrade.Order(outcome.OrderID)
		var state = owned.ReconState()
		if state.Status != order.Created && state.Status != order.Submitted {
			return nil, nil, fmt.Errorf("record ledger submit: order is already %s", state.Status)
		}
		if outcome.VenueOrderID != 0 && state.VenueOrderID != 0 &&
			state.VenueOrderID != outcome.VenueOrderID {
			return nil, nil, fmt.Errorf("record ledger submit: changed venue order identity")
		}
		if outcome.Status != "" &&
			((outcome.Status != order.Rejected && outcome.Status != order.Error) ||
				outcome.TimestampMS == 0) {
			return nil, nil, fmt.Errorf("record ledger submit: invalid terminal outcome")
		}
		if outcome.VenueOrderID != 0 {
			var indexed, indexedExists = l.venueOrders[outcome.VenueOrderID]
			if indexedExists && indexed.orderID != outcome.OrderID {
				return nil, nil, fmt.Errorf("record ledger submit: duplicate Venue Order %d", outcome.VenueOrderID)
			}
			var seenOrderID, duplicate = seenVenueOrders[outcome.VenueOrderID]
			if duplicate && seenOrderID != outcome.OrderID {
				return nil, nil, fmt.Errorf("record ledger submit: duplicate Venue Order %d", outcome.VenueOrderID)
			}
			seenVenueOrders[outcome.VenueOrderID] = outcome.OrderID
		}
		submitted = append(submitted, owned)
		tradesByID[location.tradeID] = ownedTrade
		seenOrders[outcome.OrderID] = struct{}{}
	}
	var tradeIDs = make([]uint64, 0, len(tradesByID))
	for tradeID := range tradesByID {
		tradeIDs = append(tradeIDs, tradeID)
	}
	sort.Slice(tradeIDs, func(left int, right int) bool {
		return tradeIDs[left] < tradeIDs[right]
	})
	var touchedTrades = make([]*trade.Trade, 0, len(tradeIDs))
	for _, tradeID := range tradeIDs {
		touchedTrades = append(touchedTrades, tradesByID[tradeID])
	}
	return submitted, touchedTrades, nil
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

// Section 2.4 - State and Indexes

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

func (l *Ledger) rebuildIndexes() error {
	clear(l.orders)
	clear(l.cloids)
	clear(l.venueOrders)
	clear(l.fills)
	clear(l.activeTrades)
	clear(l.activeOrders)
	clear(l.pendingOrders)
	clear(l.pendingFills)
	l.summary = Summary{}
	for tradeID, owned := range l.trades {
		var err = l.addTradeIndexes(tradeID, owned)
		if err != nil {
			return err
		}
		l.replaceTradeSummary(nil, owned.Summary())
	}
	return nil
}

func (l *Ledger) replaceTradeIndexes(
	tradeID uint64,
	previous *trade.Trade,
	current *trade.Trade,
) error {
	if previous != nil {
		l.removeTradeIndexes(tradeID, previous)
	}
	var err = l.addTradeIndexes(tradeID, current)
	if err != nil {
		if previous != nil {
			var restoreErr = l.addTradeIndexes(tradeID, previous)
			if restoreErr != nil {
				return fmt.Errorf("replace Trade %d indexes: %v; restore: %w", tradeID, err, restoreErr)
			}
		}
		return err
	}
	return nil
}

func (l *Ledger) addTradeIndexes(tradeID uint64, owned *trade.Trade) error {
	var state = owned.ReconState()
	if state.TradeID != tradeID {
		return fmt.Errorf("index ledger: Trade identity mismatch")
	}
	if tradeIsActive(state.Status) {
		l.activeTrades[tradeID] = struct{}{}
	}
	return owned.EachOrder(func(ownedOrder *order.Order) error {
		var current = ownedOrder.ReconState()
		var location = orderLocation{tradeID: tradeID, orderID: current.OrderID}
		if _, exists := l.orders[current.OrderID]; exists {
			return fmt.Errorf("index ledger: duplicate Order %d", current.OrderID)
		}
		if _, exists := l.cloids[current.CLOID]; exists {
			return fmt.Errorf("index ledger: duplicate cloid %s", current.CLOID)
		}
		l.orders[current.OrderID] = location
		l.cloids[current.CLOID] = location
		if current.VenueOrderID != 0 {
			if _, exists := l.venueOrders[current.VenueOrderID]; exists {
				return fmt.Errorf("index ledger: duplicate Venue Order %d", current.VenueOrderID)
			}
			l.venueOrders[current.VenueOrderID] = location
		}
		if current.Active {
			l.activeOrders[current.OrderID] = struct{}{}
		}
		var orderPending = current.ReconciliationPending
		var err = ownedOrder.EachFill(func(execution *fill.Fill) error {
			var observed = execution.State()
			if _, exists := l.fills[observed.VenueTID]; exists {
				return fmt.Errorf("index ledger: duplicate Venue TID %d", observed.VenueTID)
			}
			l.fills[observed.VenueTID] = fillLocation{
				tradeID: tradeID, orderID: current.OrderID, venueTID: observed.VenueTID,
			}
			if !observed.HasFee {
				l.pendingFills[observed.VenueTID] = struct{}{}
				orderPending = true
			}
			return nil
		})
		if err != nil {
			return err
		}
		if orderPending {
			l.pendingOrders[current.OrderID] = struct{}{}
		}
		return nil
	})
}

func (l *Ledger) refreshTradeIndexes(tradeID uint64) {
	var owned = l.trades[tradeID]
	var state = owned.ReconState()
	if tradeIsActive(state.Status) {
		l.activeTrades[tradeID] = struct{}{}
	} else {
		delete(l.activeTrades, tradeID)
	}
	owned.EachOrder(func(ownedOrder *order.Order) error {
		var current = ownedOrder.ReconState()
		var location = orderLocation{tradeID: tradeID, orderID: current.OrderID}
		l.orders[current.OrderID] = location
		l.cloids[current.CLOID] = location
		if current.VenueOrderID != 0 {
			l.venueOrders[current.VenueOrderID] = location
		}
		if current.Active {
			l.activeOrders[current.OrderID] = struct{}{}
		} else {
			delete(l.activeOrders, current.OrderID)
		}
		var orderPending = current.ReconciliationPending
		ownedOrder.EachFill(func(execution *fill.Fill) error {
			var observed = execution.State()
			l.fills[observed.VenueTID] = fillLocation{
				tradeID: tradeID, orderID: current.OrderID, venueTID: observed.VenueTID,
			}
			if observed.HasFee {
				delete(l.pendingFills, observed.VenueTID)
			} else {
				l.pendingFills[observed.VenueTID] = struct{}{}
				orderPending = true
			}
			return nil
		})
		if orderPending {
			l.pendingOrders[current.OrderID] = struct{}{}
		} else {
			delete(l.pendingOrders, current.OrderID)
		}
		return nil
	})
}

func (l *Ledger) addValidatedTradeIndexes(tradeID uint64, _ *trade.Trade) {
	l.refreshTradeIndexes(tradeID)
}

// Cached totals avoid repeated all-Trade aggregation while owned Trades remain authoritative.
func (l *Ledger) replaceTradeSummary(previous *trade.Summary, current trade.Summary) {
	if previous != nil {
		l.summary.RealizedPnL = l.summary.RealizedPnL.Sub(previous.RealizedPnL)
		l.summary.UnrealizedPnL = l.summary.UnrealizedPnL.Sub(previous.UnrealizedPnL)
		l.summary.GrossPnL = l.summary.GrossPnL.Sub(previous.GrossPnL)
		l.summary.Fees = l.summary.Fees.Sub(previous.Fees)
		l.summary.NetPnL = l.summary.NetPnL.Sub(previous.NetPnL)
		if tradeIsActive(previous.Status) {
			l.summary.OpenTrades--
		}
		l.summary.ActiveOrders -= previous.ActiveOrders
		l.summary.Fills -= previous.Fills
		l.summary.PendingOrders -= previous.PendingOrders
		l.summary.PendingFills -= previous.PendingFills
	}
	l.summary.RealizedPnL = l.summary.RealizedPnL.Add(current.RealizedPnL)
	l.summary.UnrealizedPnL = l.summary.UnrealizedPnL.Add(current.UnrealizedPnL)
	l.summary.GrossPnL = l.summary.GrossPnL.Add(current.GrossPnL)
	l.summary.Fees = l.summary.Fees.Add(current.Fees)
	l.summary.NetPnL = l.summary.NetPnL.Add(current.NetPnL)
	if tradeIsActive(current.Status) {
		l.summary.OpenTrades++
	}
	l.summary.ActiveOrders += current.ActiveOrders
	l.summary.Fills += current.Fills
	l.summary.PendingOrders += current.PendingOrders
	l.summary.PendingFills += current.PendingFills
}

func (l *Ledger) removeTradeIndexes(tradeID uint64, owned *trade.Trade) {
	delete(l.activeTrades, tradeID)
	owned.EachOrder(func(ownedOrder *order.Order) error {
		var current = ownedOrder.ReconState()
		delete(l.orders, current.OrderID)
		delete(l.cloids, current.CLOID)
		delete(l.activeOrders, current.OrderID)
		delete(l.pendingOrders, current.OrderID)
		if current.VenueOrderID != 0 {
			delete(l.venueOrders, current.VenueOrderID)
		}
		ownedOrder.EachFill(func(execution *fill.Fill) error {
			var observed = execution.State()
			delete(l.fills, observed.VenueTID)
			delete(l.pendingFills, observed.VenueTID)
			return nil
		})
		return nil
	})
}

// Section 3 - Generic Helpers

func cloneTrades(source map[uint64]*trade.Trade) map[uint64]*trade.Trade {
	var cloned = make(map[uint64]*trade.Trade, len(source))
	for id, owned := range source {
		cloned[id] = owned.Clone()
	}
	return cloned
}

func indexCLOIDs(trades map[uint64]*trade.Trade) map[string]*order.Order {
	var indexed = make(map[string]*order.Order)
	for _, ownedTrade := range trades {
		ownedTrade.EachOrder(func(owned *order.Order) error {
			var state = owned.ReconState()
			indexed[state.CLOID] = owned
			return nil
		})
	}
	return indexed
}

func sortedSet(values map[uint64]struct{}) []uint64 {
	var ids = make([]uint64, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})
	return ids
}

func tradeIsActive(status trade.Status) bool {
	return status != trade.Closed && status != trade.Canceled && status != trade.Error
}
