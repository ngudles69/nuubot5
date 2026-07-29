// Package ledger owns one Account's flat Trades, Orders, and Fills.
package ledger

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"

	"nuubot/internal/account/fill"
	"nuubot/internal/account/order"
	"nuubot/internal/account/trade"
)

// Config contains one Ledger's identity.
type Config struct {
	SweepID        uint64
	BotID          uint64
	ID             uint64
	CycleNumber    int
	ExecutorNumber int
	Venue          string
	Network        string
	Account        string
	Symbol         string
}

// Plan contains the next Trade and Order identities.
type Plan struct {
	Trade    trade.Trade
	OrderIDs []uint64
}

// OrderUpdate identifies one trusted update for an existing Order.
type OrderUpdate struct {
	OrderID uint64
	Update  order.Update
}

// Result contains one terminal Ledger summary.
type Result struct {
	Config
	AccountRawJSON string
	Summary        Summary
	Trades         int
	Orders         int
	Fills          int
	Cancellations  int
	StopOrders     int
}

// Summary contains cached Ledger totals.
type Summary struct {
	RealizedPnL   decimal.Decimal
	UnrealizedPnL decimal.Decimal
	GrossPnL      decimal.Decimal
	Fees          decimal.Decimal
	NetPnL        decimal.Decimal
	OpenTrades    int
	ClosedTrades  int
	ActiveOrders  int
	Fills         int
	PendingOrders int
	PendingFills  int
}

// PendingFillAnchor identifies one missing-fee Fill.
type PendingFillAnchor struct {
	VenueTID    uint64
	TimestampMS uint64
}

// OrderRef locates one Order without duplicating it.
type OrderRef struct {
	TradeID uint64
	OrderID uint64
}

// StoreState exposes current records selected for one persistence transaction.
type StoreState struct {
	Config
	NextTradeID    uint64
	NextOrderID    uint64
	NextFillID     uint64
	AccountRawJSON string
	LedgerDirty    bool
	Trades         []*trade.Trade
	Orders         []*order.Order
	Fills          []*fill.Fill
}

// Ledger owns one flat set of accounting records and relationship indexes.
type Ledger struct {
	config          Config
	trades          map[uint64]*trade.Trade
	orders          map[uint64]*order.Order
	fills           map[uint64]*fill.Fill
	orderByCLOID    map[string]OrderRef
	orderByOID      map[uint64]OrderRef
	orderIDsByTrade map[uint64][]uint64
	fillIDByTID     map[uint64]uint64
	fillIDsByOrder  map[uint64][]uint64
	activeTradeIDs  map[uint64]struct{}
	activeOrderIDs  map[uint64]struct{}
	summary         Summary
	nextTradeID     uint64
	nextOrderID     uint64
	nextFillID      uint64
	accountRawJSON  string
	dirtyTrades     map[uint64]struct{}
	dirtyOrders     map[uint64]struct{}
	dirtyFills      map[uint64]struct{}
	ledgerDirty     bool
	started         bool
	stopped         bool
}

// Section 1 - Program Flow

// Init initializes one empty flat Ledger.
func (l *Ledger) Init(cfg Config) error {
	// Step 1: validate Ledger configuration
	if cfg.SweepID == 0 || cfg.BotID == 0 || cfg.ID == 0 || cfg.CycleNumber <= 0 ||
		cfg.Venue == "" || cfg.Network == "" || cfg.Account == "" ||
		cfg.Symbol == "" {
		return fmt.Errorf("initialize ledger: invalid configuration")
	}
	if l.started || l.stopped {
		return fmt.Errorf("initialize ledger: invalid lifecycle state")
	}

	// Step 2: initialize flat records and indexes
	l.config = cfg
	l.trades = make(map[uint64]*trade.Trade)
	l.orders = make(map[uint64]*order.Order)
	l.fills = make(map[uint64]*fill.Fill)
	l.orderByCLOID = make(map[string]OrderRef)
	l.orderByOID = make(map[uint64]OrderRef)
	l.orderIDsByTrade = make(map[uint64][]uint64)
	l.fillIDByTID = make(map[uint64]uint64)
	l.fillIDsByOrder = make(map[uint64][]uint64)
	l.activeTradeIDs = make(map[uint64]struct{})
	l.activeOrderIDs = make(map[uint64]struct{})
	l.dirtyTrades = make(map[uint64]struct{})
	l.dirtyOrders = make(map[uint64]struct{})
	l.dirtyFills = make(map[uint64]struct{})
	l.nextTradeID = 1
	l.nextOrderID = 1
	l.nextFillID = 1
	l.ledgerDirty = true
	l.started = true
	return nil
}

// CreateTrade stores one Trade and its initial Orders.
func (l *Ledger) CreateTrade(created *trade.Trade, orders []*order.Order) error {
	// Step 1: validate Trade and Order identities
	if !l.ready() || created == nil || len(orders) == 0 {
		return fmt.Errorf("create ledger trade: invalid state or empty records")
	}
	if created.SweepID != l.config.SweepID ||
		created.BotID != l.config.BotID ||
		created.Venue != l.config.Venue ||
		created.Network != l.config.Network ||
		created.Account != l.config.Account ||
		created.LedgerID != l.config.ID ||
		created.TradeID != l.nextTradeID ||
		created.CycleNumber != l.config.CycleNumber ||
		created.Symbol != l.config.Symbol {
		return fmt.Errorf("create ledger trade: unexpected Trade identity")
	}
	var err = l.validateOrders(created.TradeID, orders)
	if err != nil {
		return fmt.Errorf("create ledger trade: %w", err)
	}

	// Step 2: store flat records and indexes
	l.trades[created.TradeID] = created
	l.activeTradeIDs[created.TradeID] = struct{}{}
	l.dirtyTrades[created.TradeID] = struct{}{}
	l.nextTradeID++
	l.ledgerDirty = true
	for _, current := range orders {
		l.addOrder(current)
	}

	// Step 3: refresh cached accounting
	return l.refreshTrade(created.TradeID)
}

// AddOrders stores later Orders under one existing Trade.
func (l *Ledger) AddOrders(tradeID uint64, orders []*order.Order) error {
	// Step 1: validate parent and child records
	if !l.ready() || l.trades[tradeID] == nil || len(orders) == 0 {
		return fmt.Errorf("add ledger Orders: invalid state or empty records")
	}
	var err = l.validateOrders(tradeID, orders)
	if err != nil {
		return fmt.Errorf("add ledger Orders: %w", err)
	}

	// Step 2: store Orders and refresh accounting
	for _, current := range orders {
		l.addOrder(current)
	}
	err = l.refreshTrade(tradeID)
	if err != nil {
		return fmt.Errorf("add ledger Orders: %w", err)
	}
	l.dirtyTrades[tradeID] = struct{}{}
	return nil
}

// UpdateOrders updates existing Orders and affected Trades.
func (l *Ledger) UpdateOrders(updates []OrderUpdate) error {
	// Step 1: validate complete Order update identities
	if !l.ready() || len(updates) == 0 {
		return fmt.Errorf("update ledger Orders: invalid state or empty updates")
	}
	var proposedOwners = make(map[uint64]uint64)
	var proposedOrderOIDs = make(map[uint64]uint64)
	for _, current := range updates {
		var owned = l.orders[current.OrderID]
		if owned == nil {
			return fmt.Errorf("update ledger Orders: unknown Order %d", current.OrderID)
		}
		var venueOrderID = current.Update.VenueOrderID
		if venueOrderID == 0 {
			continue
		}
		if owned.VenueOrderID != 0 && owned.VenueOrderID != venueOrderID {
			return fmt.Errorf("update ledger Orders: changed Venue Order identity")
		}
		var proposedOID, proposed = proposedOrderOIDs[current.OrderID]
		if proposed && proposedOID != venueOrderID {
			return fmt.Errorf("update ledger Orders: changed Venue Order identity")
		}
		var indexed, exists = l.orderByOID[venueOrderID]
		if exists && indexed.OrderID != current.OrderID {
			return fmt.Errorf("update ledger Orders: duplicate Venue Order %d", venueOrderID)
		}
		var proposedOwner, duplicate = proposedOwners[venueOrderID]
		if duplicate && proposedOwner != current.OrderID {
			return fmt.Errorf("update ledger Orders: duplicate Venue Order %d", venueOrderID)
		}
		proposedOrderOIDs[current.OrderID] = venueOrderID
		proposedOwners[venueOrderID] = current.OrderID
	}

	// Step 2: apply trusted Order updates
	var touched = make(map[uint64]struct{})
	for _, current := range updates {
		var owned = l.orders[current.OrderID]
		var changed, err = owned.Update(current.Update)
		if err != nil {
			return fmt.Errorf("update ledger Orders: %w", err)
		}
		if !changed {
			continue
		}
		if owned.VenueOrderID != 0 {
			l.orderByOID[owned.VenueOrderID] = OrderRef{TradeID: owned.TradeID, OrderID: owned.OrderID}
		}
		l.refreshActiveOrder(owned)
		l.dirtyOrders[owned.OrderID] = struct{}{}
		touched[owned.TradeID] = struct{}{}
	}

	// Step 3: refresh affected Trades
	for tradeID := range touched {
		var err = l.refreshTrade(tradeID)
		if err != nil {
			return fmt.Errorf("update ledger Orders: %w", err)
		}
		l.dirtyTrades[tradeID] = struct{}{}
	}
	return nil
}

// AddFill stores one new Fill or applies later evidence to an existing Fill.
func (l *Ledger) AddFill(input fill.Fill) (bool, error) {
	// Step 1: update existing Fill evidence
	if !l.ready() {
		return false, fmt.Errorf("add ledger Fill: invalid state")
	}
	var existingID = l.fillIDByTID[input.VenueTID]
	var existing = l.fills[existingID]
	if existing != nil {
		var hadFee = existing.HasFee()
		var changed, err = existing.Update(input.Fee, input.RawJSON)
		if err != nil {
			return false, fmt.Errorf("add ledger Fill: %w", err)
		}
		if !changed {
			return false, nil
		}
		var owned = l.orders[existing.OrderID]
		if !hadFee && existing.HasFee() {
			owned.Fees = owned.Fees.Add(*existing.Fee)
			owned.PendingFeeCount--
			l.refreshActiveOrder(owned)
		}
		err = l.refreshTrade(owned.TradeID)
		l.dirtyFills[existing.FillID] = struct{}{}
		l.dirtyOrders[owned.OrderID] = struct{}{}
		l.dirtyTrades[owned.TradeID] = struct{}{}
		return true, err
	}

	// Step 2: resolve the existing parent Order
	var reference, err = l.resolveOrder(input.CLOID, input.VenueOrderID)
	if err != nil {
		return false, fmt.Errorf("add ledger Fill: %w", err)
	}
	var owned = l.orders[reference.OrderID]
	input.SweepID = l.config.SweepID
	input.BotID = l.config.BotID
	input.Venue = l.config.Venue
	input.Network = l.config.Network
	input.Account = l.config.Account
	input.LedgerID = l.config.ID
	input.TradeID = owned.TradeID
	input.OrderID = owned.OrderID
	input.FillID = l.nextFillID
	input.CycleNumber = owned.CycleNumber
	if input.Symbol != owned.Symbol || input.Side != owned.Side {
		return false, fmt.Errorf("add ledger Fill: Exchange Fill does not match Order")
	}

	// Step 3: store the new Fill and update its parent records
	var created *fill.Fill
	created, err = fill.New(input)
	if err != nil {
		return false, fmt.Errorf("add ledger Fill: %w", err)
	}
	var total = owned.FilledQuantity.Add(created.Quantity)
	if total.GreaterThan(owned.SubmittedQuantity) {
		return false, fmt.Errorf("add ledger Fill: quantity exceeds Order")
	}
	l.fills[created.FillID] = created
	l.fillIDByTID[created.VenueTID] = created.FillID
	l.fillIDsByOrder[created.OrderID] = append(l.fillIDsByOrder[created.OrderID], created.FillID)
	l.nextFillID++
	l.ledgerDirty = true
	l.applyFillTotals(owned, created)
	err = l.refreshTrade(owned.TradeID)
	if err != nil {
		return false, fmt.Errorf("add ledger Fill: %w", err)
	}
	l.dirtyFills[created.FillID] = struct{}{}
	l.dirtyOrders[owned.OrderID] = struct{}{}
	l.dirtyTrades[owned.TradeID] = struct{}{}
	return true, nil
}

// UpdateAccountPayload stores the latest raw Account payload.
func (l *Ledger) UpdateAccountPayload(rawJSON string) bool {
	if rawJSON == "" || rawJSON == l.accountRawJSON {
		return false
	}
	l.accountRawJSON = rawJSON
	l.ledgerDirty = true
	return true
}

// UpdateMark refreshes marked finance for active Trades.
func (l *Ledger) UpdateMark(markPrice *decimal.Decimal) error {
	if !l.ready() {
		return fmt.Errorf("update ledger mark: invalid state")
	}
	for _, tradeID := range sortedSet(l.activeTradeIDs) {
		var owned = l.trades[tradeID]
		var previous = tradeSummary(owned)
		var changed, err = owned.UpdateMark(markPrice)
		if err != nil {
			return fmt.Errorf("update ledger mark: %w", err)
		}
		if changed {
			l.replaceTradeSummary(previous, tradeSummary(owned))
			l.dirtyTrades[tradeID] = struct{}{}
		}
	}
	return nil
}

// Result returns one terminal Ledger summary.
func (l *Ledger) Result() (Result, error) {
	if !l.started && !l.stopped {
		return Result{}, fmt.Errorf("read ledger result: ledger is not initialized")
	}
	var cancellations int
	var stopOrders int
	for _, current := range l.orders {
		if current.Status == order.Canceled {
			cancellations++
		}
		if current.Role == order.Stop {
			stopOrders++
		}
	}
	return Result{
		Config:         l.config,
		AccountRawJSON: l.accountRawJSON,
		Summary:        l.summary,
		Trades:         len(l.trades),
		Orders:         len(l.orders),
		Fills:          len(l.fills),
		Cancellations:  cancellations,
		StopOrders:     stopOrders,
	}, nil
}

// Stop stops Ledger admission.
func (l *Ledger) Stop() error {
	if l.stopped {
		return nil
	}
	l.started = false
	l.stopped = true
	return nil
}

// StoreChanges returns direct pointers for dirty or complete persistence.
func (l *Ledger) StoreChanges(all bool) StoreState {
	var state = StoreState{
		Config:         l.config,
		NextTradeID:    l.nextTradeID,
		NextOrderID:    l.nextOrderID,
		NextFillID:     l.nextFillID,
		AccountRawJSON: l.accountRawJSON,
		LedgerDirty:    all || l.ledgerDirty,
	}
	for _, id := range selectedIDs(l.trades, l.dirtyTrades, all) {
		state.Trades = append(state.Trades, l.trades[id])
	}
	for _, id := range selectedIDs(l.orders, l.dirtyOrders, all) {
		state.Orders = append(state.Orders, l.orders[id])
	}
	for _, id := range selectedIDs(l.fills, l.dirtyFills, all) {
		state.Fills = append(state.Fills, l.fills[id])
	}
	return state
}

// StoreCommitted clears only rows included in one successful transaction.
func (l *Ledger) StoreCommitted(state StoreState) {
	if state.LedgerDirty {
		l.ledgerDirty = false
	}
	for _, current := range state.Trades {
		delete(l.dirtyTrades, current.TradeID)
	}
	for _, current := range state.Orders {
		delete(l.dirtyOrders, current.OrderID)
	}
	for _, current := range state.Fills {
		delete(l.dirtyFills, current.FillID)
	}
}

// Section 2 - Domain Helpers

// Section 2.1 - Identity Planning

// PlanTrade returns the next Trade and Order identities.
func (l *Ledger) PlanTrade(orderCount int) (Plan, error) {
	if !l.ready() || orderCount <= 0 {
		return Plan{}, fmt.Errorf("plan ledger Trade: invalid state or Order count")
	}
	var ids = make([]uint64, orderCount)
	for index := range ids {
		ids[index] = l.nextOrderID + uint64(index)
	}
	return Plan{
		Trade: trade.Trade{
			SweepID: l.config.SweepID, BotID: l.config.BotID,
			Venue: l.config.Venue, Network: l.config.Network, Account: l.config.Account,
			LedgerID: l.config.ID, TradeID: l.nextTradeID,
			CycleNumber: l.config.CycleNumber, Symbol: l.config.Symbol,
		},
		OrderIDs: ids,
	}, nil
}

// PlanOrders returns the next Order identities.
func (l *Ledger) PlanOrders(orderCount int) ([]uint64, error) {
	if !l.ready() || orderCount <= 0 {
		return nil, fmt.Errorf("plan ledger Orders: invalid state or Order count")
	}
	var ids = make([]uint64, orderCount)
	for index := range ids {
		ids[index] = l.nextOrderID + uint64(index)
	}
	return ids, nil
}

// Section 2.2 - Domain Observation

// ActiveOrders returns current active Orders.
func (l *Ledger) ActiveOrders() []order.Order {
	var ids = sortedSet(l.activeOrderIDs)
	var values = make([]order.Order, 0, len(ids))
	for _, orderID := range ids {
		values = append(values, *l.orders[orderID])
	}
	return values
}

// ActiveTrades returns current active Trades.
func (l *Ledger) ActiveTrades() []trade.Trade {
	var ids = sortedSet(l.activeTradeIDs)
	var values = make([]trade.Trade, 0, len(ids))
	for _, tradeID := range ids {
		values = append(values, *l.trades[tradeID])
	}
	return values
}

// Trade returns one Trade.
func (l *Ledger) Trade(tradeID uint64) (trade.Trade, error) {
	var owned = l.trades[tradeID]
	if owned == nil {
		return trade.Trade{}, fmt.Errorf("read ledger Trade: unknown Trade %d", tradeID)
	}
	return *owned, nil
}

// Order returns one Order.
func (l *Ledger) Order(orderID uint64) (order.Order, error) {
	var owned = l.orders[orderID]
	if owned == nil {
		return order.Order{}, fmt.Errorf("read ledger Order: unknown Order %d", orderID)
	}
	return *owned, nil
}

// Fill returns one Fill by Venue identity.
func (l *Ledger) Fill(venueTID uint64) (fill.Fill, bool) {
	var fillID, exists = l.fillIDByTID[venueTID]
	if !exists {
		return fill.Fill{}, false
	}
	var owned = l.fills[fillID]
	return *owned, true
}

// TradeOrders returns Orders linked to one Trade.
func (l *Ledger) TradeOrders(tradeID uint64) ([]order.Order, error) {
	if l.trades[tradeID] == nil {
		return nil, fmt.Errorf("read ledger Trade Orders: unknown Trade %d", tradeID)
	}
	var ids = append([]uint64(nil), l.orderIDsByTrade[tradeID]...)
	sort.Slice(ids, func(left int, right int) bool { return ids[left] < ids[right] })
	var values = make([]order.Order, 0, len(ids))
	for _, orderID := range ids {
		values = append(values, *l.orders[orderID])
	}
	return values, nil
}

// PendingCounts returns unresolved Order and Fill counts.
func (l *Ledger) PendingCounts() (int, int) {
	var orders int
	var fills int
	for _, current := range l.orders {
		if current.PendingFeeCount > 0 ||
			(current.Status == order.Filled && !current.IsClosed()) {
			orders++
		}
		fills += current.PendingFeeCount
	}
	return orders, fills
}

// PendingFillAnchors returns missing-fee Fill anchors.
func (l *Ledger) PendingFillAnchors() []PendingFillAnchor {
	var anchors = make([]PendingFillAnchor, 0)
	for _, current := range l.fills {
		if !current.HasFee() {
			anchors = append(anchors, PendingFillAnchor{
				VenueTID: current.VenueTID, TimestampMS: current.TimestampMS,
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

// HasPendingRecon reports unresolved Order or Fill evidence.
func (l *Ledger) HasPendingRecon() bool {
	var orders, fills = l.PendingCounts()
	return orders != 0 || fills != 0
}

// Summary returns current cached accounting totals.
func (l *Ledger) Summary() Summary {
	return l.summary
}

// Section 2.3 - Flat Record Ownership

func (l *Ledger) validateOrders(tradeID uint64, orders []*order.Order) error {
	var orderIDs = make(map[uint64]struct{}, len(orders))
	var cloids = make(map[string]struct{}, len(orders))
	for index, current := range orders {
		if current == nil ||
			current.SweepID != l.config.SweepID ||
			current.BotID != l.config.BotID ||
			current.Venue != l.config.Venue ||
			current.Network != l.config.Network ||
			current.Account != l.config.Account ||
			current.LedgerID != l.config.ID ||
			current.TradeID != tradeID ||
			current.OrderID != l.nextOrderID+uint64(index) ||
			current.CycleNumber != l.config.CycleNumber ||
			current.Symbol != l.config.Symbol {
			return fmt.Errorf("unexpected Order identity")
		}
		if _, exists := l.orders[current.OrderID]; exists {
			return fmt.Errorf("duplicate Order %d", current.OrderID)
		}
		if _, exists := l.orderByCLOID[current.CLOID]; exists {
			return fmt.Errorf("duplicate cloid %s", current.CLOID)
		}
		if _, exists := orderIDs[current.OrderID]; exists {
			return fmt.Errorf("duplicate Order %d", current.OrderID)
		}
		if _, exists := cloids[current.CLOID]; exists {
			return fmt.Errorf("duplicate cloid %s", current.CLOID)
		}
		orderIDs[current.OrderID] = struct{}{}
		cloids[current.CLOID] = struct{}{}
	}
	return nil
}

func (l *Ledger) addOrder(current *order.Order) {
	var reference = OrderRef{TradeID: current.TradeID, OrderID: current.OrderID}
	l.orders[current.OrderID] = current
	l.orderByCLOID[current.CLOID] = reference
	if current.VenueOrderID != 0 {
		l.orderByOID[current.VenueOrderID] = reference
	}
	l.orderIDsByTrade[current.TradeID] = append(l.orderIDsByTrade[current.TradeID], current.OrderID)
	l.refreshActiveOrder(current)
	l.dirtyOrders[current.OrderID] = struct{}{}
	l.nextOrderID++
	l.ledgerDirty = true
}

func (l *Ledger) resolveOrder(cloid string, venueOrderID uint64) (OrderRef, error) {
	if cloid == "" && venueOrderID == 0 {
		return OrderRef{}, fmt.Errorf("unknown Fill parent")
	}
	var byCLOID OrderRef
	var hasCLOID bool
	if cloid != "" {
		byCLOID, hasCLOID = l.orderByCLOID[cloid]
		if !hasCLOID {
			return OrderRef{}, fmt.Errorf("unknown Fill cloid %s", cloid)
		}
	}
	var byOID OrderRef
	var hasOID bool
	if venueOrderID != 0 {
		byOID, hasOID = l.orderByOID[venueOrderID]
		if !hasOID {
			return OrderRef{}, fmt.Errorf("unknown Fill Venue Order %d", venueOrderID)
		}
	}
	if hasCLOID && hasOID && byCLOID != byOID {
		return OrderRef{}, fmt.Errorf("Fill identity resolves different Orders")
	}
	if hasCLOID {
		return byCLOID, nil
	}
	if hasOID {
		return byOID, nil
	}
	return OrderRef{}, fmt.Errorf("unknown Fill parent")
}

func (l *Ledger) applyFillTotals(owned *order.Order, execution *fill.Fill) {
	owned.FilledQuantity = owned.FilledQuantity.Add(execution.Quantity)
	owned.FilledNotional = owned.FilledNotional.Add(execution.Quantity.Mul(execution.Price))
	owned.AverageFillPrice = owned.FilledNotional.Div(owned.FilledQuantity)
	owned.RemainingQuantity = owned.SubmittedQuantity.Sub(owned.FilledQuantity)
	owned.FillCount++
	if execution.HasFee() {
		owned.Fees = owned.Fees.Add(*execution.Fee)
	} else {
		owned.PendingFeeCount++
	}
	if execution.TimestampMS > owned.LastFillMS {
		owned.LastFillMS = execution.TimestampMS
	}
	l.refreshActiveOrder(owned)
}

func (l *Ledger) refreshTrade(tradeID uint64) error {
	var owned = l.trades[tradeID]
	var orders = make([]*order.Order, 0, len(l.orderIDsByTrade[tradeID]))
	var fills = make([]*fill.Fill, 0)
	for _, orderID := range l.orderIDsByTrade[tradeID] {
		orders = append(orders, l.orders[orderID])
		for _, fillID := range l.fillIDsByOrder[orderID] {
			fills = append(fills, l.fills[fillID])
		}
	}
	var previous = tradeSummary(owned)
	var changed, err = owned.Update(orders, fills)
	if err != nil {
		return err
	}
	_ = changed
	l.replaceTradeSummary(previous, tradeSummary(owned))
	if owned.IsClosed() {
		delete(l.activeTradeIDs, tradeID)
	} else {
		l.activeTradeIDs[tradeID] = struct{}{}
	}
	return nil
}

func (l *Ledger) refreshActiveOrder(owned *order.Order) {
	if owned.IsClosed() {
		delete(l.activeOrderIDs, owned.OrderID)
	} else {
		l.activeOrderIDs[owned.OrderID] = struct{}{}
	}
}

func (l *Ledger) replaceTradeSummary(previous Summary, current Summary) {
	l.summary.RealizedPnL = l.summary.RealizedPnL.Sub(previous.RealizedPnL).Add(current.RealizedPnL)
	l.summary.UnrealizedPnL = l.summary.UnrealizedPnL.Sub(previous.UnrealizedPnL).Add(current.UnrealizedPnL)
	l.summary.GrossPnL = l.summary.GrossPnL.Sub(previous.GrossPnL).Add(current.GrossPnL)
	l.summary.Fees = l.summary.Fees.Sub(previous.Fees).Add(current.Fees)
	l.summary.NetPnL = l.summary.NetPnL.Sub(previous.NetPnL).Add(current.NetPnL)
	l.recountSummary()
}

func (l *Ledger) recountSummary() {
	l.summary.OpenTrades = len(l.activeTradeIDs)
	l.summary.ClosedTrades = 0
	for _, current := range l.trades {
		if current.Status == trade.Closed {
			l.summary.ClosedTrades++
		}
	}
	l.summary.ActiveOrders = len(l.activeOrderIDs)
	l.summary.Fills = len(l.fills)
	l.summary.PendingOrders, l.summary.PendingFills = l.PendingCounts()
}

func tradeSummary(current *trade.Trade) Summary {
	if current == nil {
		return Summary{}
	}
	return Summary{
		RealizedPnL:   current.RealizedPnL,
		UnrealizedPnL: current.UnrealizedPnL,
		GrossPnL:      current.GrossPnL,
		Fees:          current.Fees,
		NetPnL:        current.NetPnL,
	}
}

func (l *Ledger) ready() bool {
	return l.started && !l.stopped
}

// Section 3 - Generic Helpers

func sortedSet(values map[uint64]struct{}) []uint64 {
	var ids = make([]uint64, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool { return ids[left] < ids[right] })
	return ids
}

func selectedIDs[T any](
	values map[uint64]T,
	dirty map[uint64]struct{},
	all bool,
) []uint64 {
	if all {
		var ids = make([]uint64, 0, len(values))
		for id := range values {
			ids = append(ids, id)
		}
		return ids
	}
	var ids = make([]uint64, 0, len(dirty))
	for id := range dirty {
		ids = append(ids, id)
	}
	return ids
}
