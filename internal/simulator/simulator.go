// Package simulator owns Hyperliquid-shaped simulated Venue truth.
package simulator

import (
	"fmt"
	"sort"

	"github.com/shopspring/decimal"

	"nuubot/internal/cloid"
	"nuubot/internal/market"
	"nuubot/internal/order"
)

const (
	// WaitingForFill identifies one inactive bracket child awaiting its entry.
	WaitingForFill = "waitingForFill"
	// WaitingForTrigger identifies one armed trigger Order.
	WaitingForTrigger = "waitingForTrigger"
)

// Config contains one Simulator's identity and policy.
type Config struct {
	LedgerID    uint64
	Name        string
	Account     string
	CycleNumber int
	Symbol      string
	Equity      decimal.Decimal
	FeePct      decimal.Decimal
	SlippagePct decimal.Decimal
	PersistMode string
	Path        string
}

// OrderRequest contains one validated Account-owned Venue request.
type OrderRequest struct {
	CLOID        string
	Symbol       string
	TradeID      uint64
	OrderID      uint64
	Role         string
	Side         string
	Type         string
	TimeInForce  string
	Quantity     decimal.Decimal
	Price        *decimal.Decimal
	TriggerPrice *decimal.Decimal
	ReduceOnly   bool
	TimestampMS  uint64
}

// SubmitStatus contains one ordered Hyperliquid-shaped submission outcome.
type SubmitStatus struct {
	Kind           string
	VenueOrderID   uint64
	CLOID          string
	FilledQuantity decimal.Decimal
	AveragePrice   decimal.Decimal
	Error          string
}

// SubmitResponse contains one complete ordered submission acknowledgement.
type SubmitResponse struct {
	Status   string
	Type     string
	Statuses []SubmitStatus
}

// CancelResponse contains one complete ordered cancellation acknowledgement.
type CancelResponse struct {
	Status   string
	Type     string
	Statuses []string
}

// OrderState contains one canonical simulated Venue Order observation.
type OrderState struct {
	CLOID             string
	VenueOrderID      uint64
	Symbol            string
	TradeID           uint64
	OrderID           uint64
	Role              string
	Side              string
	Status            order.Status
	Quantity          decimal.Decimal
	RemainingQuantity decimal.Decimal
	FilledQuantity    decimal.Decimal
	AverageFillPrice  decimal.Decimal
	HasAveragePrice   bool
	Price             decimal.Decimal
	HasPrice          bool
	TriggerPrice      decimal.Decimal
	HasTriggerPrice   bool
	ReduceOnly        bool
	TimestampMS       uint64
	Raw               string
}

// FillState contains one canonical simulated Venue Fill observation.
type FillState struct {
	CLOID         string
	VenueOrderID  uint64
	VenueTID      uint64
	Symbol        string
	Side          string
	Quantity      decimal.Decimal
	Price         decimal.Decimal
	TimestampMS   uint64
	StartPosition decimal.Decimal
	ClosedPnL     decimal.Decimal
	Direction     string
	Fee           decimal.Decimal
	Liquidity     string
}

// AccountState contains one canonical simulated clearinghouse observation.
type AccountState struct {
	ObservedMS    uint64
	AccountValue  decimal.Decimal
	Withdrawable  decimal.Decimal
	MarginUsed    decimal.Decimal
	PositionSize  decimal.Decimal
	EntryPrice    decimal.Decimal
	UnrealizedPnL decimal.Decimal
	RealizedPnL   decimal.Decimal
	Fees          decimal.Decimal
}

// Result contains one immutable Simulator result.
type Result struct {
	Config
	NextVenueOrderID uint64
	NextVenueTID     uint64
	Orders           []OrderState
	Fills            []FillState
}

type simOrder struct {
	request           OrderRequest
	venueOrderID      uint64
	status            string
	armed             bool
	remainingQuantity decimal.Decimal
	filledQuantity    decimal.Decimal
	averageFillPrice  decimal.Decimal
	fees              decimal.Decimal
	timestampMS       uint64
}

type position struct {
	size       decimal.Decimal
	entryPrice decimal.Decimal
	realized   decimal.Decimal
	fees       decimal.Decimal
}

// Simulator owns one Account's simulated exchange state.
type Simulator struct {
	config           Config
	nextVenueOrderID uint64
	nextVenueTID     uint64
	openOrders       map[string]*simOrder
	orderHistory     []OrderState
	fills            []FillState
	store            *simulatorStore
	lastPrice        decimal.Decimal
	lastTimestampMS  uint64
	hasBBO           bool
	started          bool
	stopped          bool
}

// Section 1 - Program Flow

// Init prepares one Simulator.
func (s *Simulator) Init(cfg Config) error {
	// bind Simulator inputs
	if s.started || s.stopped {
		return fmt.Errorf("initialize simulator: invalid lifecycle state")
	}
	s.config = cfg

	// validate Simulator identity
	if cfg.LedgerID == 0 || cfg.Name == "" || cfg.Account == "" ||
		cfg.CycleNumber <= 0 || cfg.Symbol == "" {
		return fmt.Errorf("initialize simulator: complete identity is required")
	}

	// validate Simulator policy
	if !cfg.Equity.IsPositive() || cfg.FeePct.IsNegative() ||
		cfg.SlippagePct.IsNegative() {
		return fmt.Errorf("initialize simulator: invalid equity, fee, or slippage")
	}

	// validate persistence mode
	if cfg.PersistMode != "none" && cfg.PersistMode != "max" {
		return fmt.Errorf("initialize simulator: invalid persistence mode %q", cfg.PersistMode)
	}
	if cfg.PersistMode == "max" && cfg.Path == "" {
		return fmt.Errorf("initialize simulator: max persistence requires path")
	}

	// initialize Simulator
	s.nextVenueOrderID = 1
	s.nextVenueTID = 1
	s.openOrders = make(map[string]*simOrder)

	// open Simulator state when configured
	if cfg.PersistMode == "max" {
		var err error
		s.store, err = openSimulatorStore(cfg.Path)
		if err != nil {
			return err
		}
		var stored storedState
		var found bool
		stored, found, err = s.store.load(cfg)
		if err != nil {
			s.store.close()
			return err
		}
		if found {
			err = s.restore(stored)
			if err != nil {
				s.store.close()
				return err
			}
		} else {
			err = s.store.save(cfg, s.storedState())
			if err != nil {
				s.store.close()
				return err
			}
		}
	}

	// admit Simulator lifecycle
	s.started = true
	return nil
}

// PlaceOrders admits one complete batch and returns ordered outcomes.
func (s *Simulator) PlaceOrders(requests []OrderRequest) (SubmitResponse, error) {
	if !s.started || s.stopped {
		return SubmitResponse{}, fmt.Errorf("place simulator Orders: invalid lifecycle state")
	}

	// validate Venue requests
	if len(requests) == 0 || len(requests) > 1000 {
		return SubmitResponse{}, fmt.Errorf("place simulator Orders: batch size must be from 1 to 1000")
	}
	var seen = make(map[string]struct{}, len(requests))
	for _, request := range requests {
		var err = s.validateRequest(request)
		if err != nil {
			return SubmitResponse{}, err
		}
		if _, exists := seen[request.CLOID]; exists || s.openOrders[request.CLOID] != nil {
			return SubmitResponse{}, fmt.Errorf("place simulator Orders: duplicate cloid %s", request.CLOID)
		}
		seen[request.CLOID] = struct{}{}
	}

	// stage recoverable Simulator mutation
	var staged = s.stage()

	// allocate Venue identities
	var added = make([]*simOrder, 0, len(requests))
	for _, request := range requests {
		var status = string(order.Open)
		var armed = true
		if request.Role == order.TakeProfit || request.Role == order.StopLoss {
			status = WaitingForFill
			armed = false
		}
		var row = &simOrder{
			request:           copyRequest(request),
			venueOrderID:      staged.nextVenueOrderID,
			status:            status,
			armed:             armed,
			remainingQuantity: request.Quantity,
			timestampMS:       request.TimestampMS,
		}
		staged.nextVenueOrderID++
		staged.openOrders[request.CLOID] = row
		staged.orderHistory = append(staged.orderHistory, staged.publicOrder(row))
		added = append(added, row)
	}

	// execute explicit market-like Orders
	if staged.hasBBO {
		staged.matchAdded(added, staged.lastTimestampMS)
	}

	// persist changed state when configured
	var err = staged.persist()
	if err != nil {
		return SubmitResponse{}, err
	}
	s.commit(staged)

	// return admitted SDK-shaped submit response
	var entryFilled = false
	for _, row := range added {
		var latest, _ = s.orderByCLOID(row.request.CLOID)
		if latest.request.Role == order.Entry && latest.status == string(order.Filled) {
			entryFilled = true
		}
	}
	var statuses = make([]SubmitStatus, 0, len(added))
	for _, row := range added {
		var latest, _ = s.orderByCLOID(row.request.CLOID)
		statuses = append(statuses, submitStatus(latest, entryFilled))
	}
	return SubmitResponse{Status: "ok", Type: "order", Statuses: statuses}, nil
}

// CancelOrders cancels one complete owned Venue batch.
func (s *Simulator) CancelOrders(cloids []string, timestampMS uint64) (CancelResponse, error) {
	if !s.started || s.stopped {
		return CancelResponse{}, fmt.Errorf("cancel simulator Orders: invalid lifecycle state")
	}

	// cancel simulated Orders
	if len(cloids) == 0 || timestampMS == 0 {
		return CancelResponse{}, fmt.Errorf("cancel simulator Orders: cloids and timestamp are required")
	}
	for _, value := range cloids {
		var row = s.openOrders[value]
		if row == nil {
			return CancelResponse{}, fmt.Errorf("cancel simulator Orders: unknown active cloid %s", value)
		}
	}

	// stage recoverable Simulator mutation
	var staged = s.stage()
	for _, value := range cloids {
		var row = staged.openOrders[value]
		staged.cancel(row, timestampMS)
		if row.request.Role == order.Entry {
			staged.cancelChildren(row.request.TradeID, timestampMS)
		}
	}

	// persist changed state when configured
	var err = staged.persist()
	if err != nil {
		return CancelResponse{}, err
	}
	s.commit(staged)

	// return admitted SDK-shaped cancel response
	var statuses = make([]string, len(cloids))
	for index := range statuses {
		statuses[index] = "success"
	}
	return CancelResponse{Status: "ok", Type: "cancel", Statuses: statuses}, nil
}

// IngestBBO advances simulated exchange matching from one BBO.
func (s *Simulator) IngestBBO(bbo market.BBO) (bool, error) {
	if !s.started || s.stopped {
		return false, fmt.Errorf("ingest simulator BBO: invalid lifecycle state")
	}

	// validate BBO identity
	if bbo.TimestampMS == 0 || bbo.Price <= 0 ||
		(s.hasBBO && bbo.TimestampMS < s.lastTimestampMS) {
		return false, fmt.Errorf("ingest simulator BBO: invalid timestamp or price")
	}
	var price = decimal.NewFromFloat(bbo.Price)

	// warm transient market state
	if !s.hasBBO {
		s.lastPrice = price
		s.lastTimestampMS = bbo.TimestampMS
		s.hasBBO = true
		return false, nil
	}

	// stage recoverable Simulator mutation
	var staged = s.stage()

	// match eligible Orders
	var changed = staged.match(price, bbo.TimestampMS)
	staged.lastPrice = price
	staged.lastTimestampMS = bbo.TimestampMS

	// record simulated outcomes
	if !changed {
		s.commit(staged)
		return false, nil
	}

	// persist changed state when configured
	var err = staged.persist()
	if err != nil {
		return false, err
	}
	s.commit(staged)

	// report changed truth
	return true, nil
}

// Result returns one immutable Simulator result.
func (s *Simulator) Result() Result {
	// copy identity policy and counters
	var result = Result{
		Config:           s.config,
		NextVenueOrderID: s.nextVenueOrderID,
		NextVenueTID:     s.nextVenueTID,
	}

	// copy Orders and Fill history
	result.Orders = append(result.Orders, s.orderHistory...)
	result.Fills = append(result.Fills, s.fills...)
	return result
}

// Stop stops Simulator admission.
func (s *Simulator) Stop() error {
	if s.stopped {
		return nil
	}

	// stop Simulator
	var err = s.persist()
	if err != nil {
		return err
	}
	if s.store != nil {
		err = s.store.close()
		if err != nil {
			return err
		}
		s.store = nil
	}
	s.started = false
	s.stopped = true
	return nil
}

// Section 2 - Domain Helpers

// OpenOrders returns current Hyperliquid-shaped open Order truth.
func (s *Simulator) OpenOrders() []OrderState {
	var rows = make([]OrderState, 0, len(s.openOrders))
	for _, row := range s.openOrders {
		rows = append(rows, s.publicOrder(row))
	}
	sort.Slice(rows, func(left int, right int) bool {
		return rows[left].VenueOrderID < rows[right].VenueOrderID
	})
	return rows
}

// Fills returns one inclusive bounded Fill range.
func (s *Simulator) Fills(startMS uint64, endMS uint64) []FillState {
	var rows []FillState
	for _, execution := range s.fills {
		if execution.TimestampMS >= startMS && execution.TimestampMS <= endMS {
			rows = append(rows, execution)
		}
	}
	return rows
}

// OrderStatus returns the latest canonical state for one CLOID.
func (s *Simulator) OrderStatus(value string) (OrderState, error) {
	var row = s.openOrders[value]
	if row != nil {
		return s.publicOrder(row), nil
	}
	for index := len(s.orderHistory) - 1; index >= 0; index-- {
		if s.orderHistory[index].CLOID == value {
			return s.orderHistory[index], nil
		}
	}
	return OrderState{}, fmt.Errorf("read simulator Order status: unknown cloid %s", value)
}

// AccountState returns one Hyperliquid-shaped clearinghouse snapshot.
func (s *Simulator) AccountState() (AccountState, error) {
	var current = s.position()
	if !current.size.IsZero() && !s.hasBBO {
		return AccountState{}, fmt.Errorf("read simulator Account state: fresh BBO is required")
	}
	var unrealized = decimal.Zero
	var positionValue = decimal.Zero
	if !current.size.IsZero() {
		unrealized = s.lastPrice.Sub(current.entryPrice).Mul(current.size)
		positionValue = current.size.Abs().Mul(s.lastPrice)
	}
	var accountValue = s.config.Equity.Add(current.realized).Add(unrealized).Sub(current.fees)
	var margin = positionValue.Div(decimal.NewFromInt(5))
	return AccountState{
		ObservedMS:    s.lastTimestampMS,
		AccountValue:  accountValue,
		Withdrawable:  accountValue.Sub(margin),
		MarginUsed:    margin,
		PositionSize:  current.size,
		EntryPrice:    current.entryPrice,
		UnrealizedPnL: unrealized,
		RealizedPnL:   current.realized,
		Fees:          current.fees,
	}, nil
}

func (s *Simulator) validateRequest(request OrderRequest) error {
	if request.Symbol != s.config.Symbol || request.TradeID == 0 ||
		request.OrderID == 0 || request.TimestampMS == 0 {
		return fmt.Errorf("place simulator Orders: invalid identity")
	}
	var _, err = cloid.Decode(request.CLOID)
	if err != nil {
		return fmt.Errorf("place simulator Orders: %w", err)
	}
	if request.Side != order.Buy && request.Side != order.Sell {
		return fmt.Errorf("place simulator Orders: invalid side %q", request.Side)
	}
	if !request.Quantity.IsPositive() {
		return fmt.Errorf("place simulator Orders: quantity must be positive")
	}
	if request.Price == nil || !request.Price.IsPositive() {
		return fmt.Errorf("place simulator Orders: positive price is required")
	}
	if (request.Role == order.TakeProfit || request.Role == order.StopLoss) &&
		(request.TriggerPrice == nil || !request.TriggerPrice.IsPositive()) {
		return fmt.Errorf("place simulator Orders: trigger price is required")
	}
	return nil
}

func (s *Simulator) match(price decimal.Decimal, timestampMS uint64) bool {
	var changed = false
	var filledTrades = make(map[uint64]struct{})
	for {
		var candidates = s.sortedOpenOrders()
		var matched *simOrder
		for _, row := range candidates {
			if !row.armed {
				continue
			}
			if row.timestampMS >= timestampMS {
				continue
			}
			if _, filled := filledTrades[row.request.TradeID]; filled {
				continue
			}
			if crosses(row, price) {
				matched = row
				break
			}
		}
		if matched == nil {
			break
		}
		var quantity, executable = s.executableQuantity(matched)
		if !executable {
			s.cancel(matched, max(timestampMS, matched.timestampMS))
			changed = true
			continue
		}
		s.fill(matched, quantity, max(timestampMS, matched.timestampMS))
		filledTrades[matched.request.TradeID] = struct{}{}
		changed = true
	}
	return changed
}

func (s *Simulator) matchAdded(
	added []*simOrder,
	timestampMS uint64,
) {
	var filledTrades = make(map[uint64]struct{})
	for _, row := range added {
		if !row.armed {
			continue
		}
		if row.request.TimeInForce != order.IOC && !crosses(row, s.lastPrice) {
			continue
		}
		if _, filled := filledTrades[row.request.TradeID]; filled {
			continue
		}
		var quantity, executable = s.executableQuantity(row)
		if !executable {
			s.cancel(row, max(timestampMS, row.timestampMS))
			continue
		}
		s.fill(row, quantity, max(timestampMS, row.timestampMS))
		filledTrades[row.request.TradeID] = struct{}{}
	}
}

func (s *Simulator) fill(row *simOrder, quantity decimal.Decimal, timestampMS uint64) {
	var basis = *row.request.Price
	if row.request.TriggerPrice != nil {
		basis = *row.request.TriggerPrice
	}
	var rate = s.config.SlippagePct.Div(decimal.NewFromInt(100))
	var price = basis.Mul(decimal.NewFromInt(1).Add(rate))
	if row.request.Side == order.Sell {
		price = basis.Mul(decimal.NewFromInt(1).Sub(rate))
	}
	var before = s.position()
	var fee = quantity.Mul(price).Mul(s.config.FeePct).Div(decimal.NewFromInt(100))
	var closedPnL = closePnL(before, row.request.Side, quantity, price)
	var execution = FillState{
		CLOID:         row.request.CLOID,
		VenueOrderID:  row.venueOrderID,
		VenueTID:      s.nextVenueTID,
		Symbol:        row.request.Symbol,
		Side:          row.request.Side,
		Quantity:      quantity,
		Price:         price,
		TimestampMS:   timestampMS,
		StartPosition: before.size,
		ClosedPnL:     closedPnL,
		Direction:     fillDirection(before.size, row.request.Side),
		Fee:           fee,
		Liquidity:     "taker",
	}
	s.nextVenueTID++
	s.fills = append(s.fills, execution)
	delete(s.openOrders, row.request.CLOID)
	row.status = string(order.Filled)
	row.armed = false
	row.filledQuantity = quantity
	row.remainingQuantity = decimal.Zero
	row.averageFillPrice = price
	row.fees = fee
	row.timestampMS = timestampMS
	s.orderHistory = append(s.orderHistory, s.publicOrder(row))
	if row.request.Role == order.Entry {
		s.armChildren(row.request.TradeID, timestampMS)
	}
	if row.request.Role == order.TakeProfit || row.request.Role == order.StopLoss {
		s.cancelChildren(row.request.TradeID, timestampMS)
	}
}

func (s *Simulator) cancel(row *simOrder, timestampMS uint64) {
	delete(s.openOrders, row.request.CLOID)
	row.status = string(order.Canceled)
	row.armed = false
	row.timestampMS = timestampMS
	s.orderHistory = append(s.orderHistory, s.publicOrder(row))
}

func (s *Simulator) armChildren(tradeID uint64, timestampMS uint64) {
	for _, child := range s.openOrders {
		if child.request.TradeID == tradeID &&
			(child.request.Role == order.TakeProfit || child.request.Role == order.StopLoss) {
			child.status = WaitingForTrigger
			child.armed = true
			child.timestampMS = timestampMS
		}
	}
}

func (s *Simulator) cancelChildren(tradeID uint64, timestampMS uint64) {
	var candidates = s.sortedOpenOrders()
	for _, child := range candidates {
		if child.request.TradeID == tradeID &&
			(child.request.Role == order.TakeProfit || child.request.Role == order.StopLoss) {
			s.cancel(child, timestampMS)
		}
	}
}

func (s *Simulator) executableQuantity(row *simOrder) (decimal.Decimal, bool) {
	if !row.request.ReduceOnly {
		return row.remainingQuantity, true
	}
	var size = s.position().size
	var available = size
	if row.request.Side == order.Buy {
		available = size.Neg()
	}
	if !available.IsPositive() {
		return decimal.Zero, false
	}
	return decimal.Min(row.remainingQuantity, available), true
}

func (s *Simulator) position() position {
	var result position
	for _, execution := range s.fills {
		var delta = execution.Quantity
		if execution.Side == order.Sell {
			delta = delta.Neg()
		}
		result.fees = result.fees.Add(execution.Fee)
		if result.size.IsZero() || sameSign(result.size, delta) {
			var total = result.size.Abs().Add(execution.Quantity)
			result.entryPrice = result.size.Abs().Mul(result.entryPrice).
				Add(execution.Quantity.Mul(execution.Price)).
				Div(total)
			result.size = result.size.Add(delta)
			continue
		}
		result.realized = result.realized.Add(
			closePnL(result, execution.Side, execution.Quantity, execution.Price),
		)
		var before = result.size
		result.size = result.size.Add(delta)
		if result.size.IsZero() {
			result.entryPrice = decimal.Zero
		} else if !sameSign(before, result.size) {
			result.entryPrice = execution.Price
		}
	}
	return result
}

func (s *Simulator) publicOrder(row *simOrder) OrderState {
	var status = order.Status(row.status)
	if row.status == WaitingForFill || row.status == WaitingForTrigger {
		status = order.Open
	}
	var result = OrderState{
		CLOID:             row.request.CLOID,
		VenueOrderID:      row.venueOrderID,
		Symbol:            row.request.Symbol,
		TradeID:           row.request.TradeID,
		OrderID:           row.request.OrderID,
		Role:              row.request.Role,
		Side:              row.request.Side,
		Status:            status,
		Quantity:          row.request.Quantity,
		RemainingQuantity: row.remainingQuantity,
		FilledQuantity:    row.filledQuantity,
		AverageFillPrice:  row.averageFillPrice,
		HasAveragePrice:   row.filledQuantity.IsPositive(),
		ReduceOnly:        row.request.ReduceOnly,
		TimestampMS:       row.timestampMS,
	}
	if row.request.Price != nil {
		result.Price = *row.request.Price
		result.HasPrice = true
	}
	if row.request.TriggerPrice != nil {
		result.TriggerPrice = *row.request.TriggerPrice
		result.HasTriggerPrice = true
	}
	return result
}

func (s *Simulator) sortedOpenOrders() []*simOrder {
	var rows = make([]*simOrder, 0, len(s.openOrders))
	for _, row := range s.openOrders {
		rows = append(rows, row)
	}
	sort.Slice(rows, func(left int, right int) bool {
		return rows[left].venueOrderID < rows[right].venueOrderID
	})
	return rows
}

func (s *Simulator) orderByCLOID(value string) (*simOrder, bool) {
	var current = s.openOrders[value]
	if current != nil {
		return current, true
	}
	for index := len(s.orderHistory) - 1; index >= 0; index-- {
		if s.orderHistory[index].CLOID != value {
			continue
		}
		var state = s.orderHistory[index]
		return &simOrder{
			request: OrderRequest{
				CLOID:       state.CLOID,
				Symbol:      state.Symbol,
				TradeID:     state.TradeID,
				OrderID:     state.OrderID,
				Role:        state.Role,
				Side:        state.Side,
				Quantity:    state.Quantity,
				ReduceOnly:  state.ReduceOnly,
				TimestampMS: state.TimestampMS,
			},
			venueOrderID:      state.VenueOrderID,
			status:            string(state.Status),
			remainingQuantity: state.RemainingQuantity,
			filledQuantity:    state.FilledQuantity,
			averageFillPrice:  state.AverageFillPrice,
			timestampMS:       state.TimestampMS,
		}, true
	}
	return nil, false
}

func (s *Simulator) persist() error {
	if s.config.PersistMode == "none" {
		return nil
	}
	return s.store.save(s.config, s.storedState())
}

func (s *Simulator) storedState() storedState {
	var state = storedState{
		SchemaVersion:    simulatorSchemaVersion,
		LedgerID:         s.config.LedgerID,
		Name:             s.config.Name,
		Account:          s.config.Account,
		CycleNumber:      s.config.CycleNumber,
		Symbol:           s.config.Symbol,
		Equity:           s.config.Equity.String(),
		FeePct:           s.config.FeePct.String(),
		SlippagePct:      s.config.SlippagePct.String(),
		NextVenueOrderID: s.nextVenueOrderID,
		NextVenueTID:     s.nextVenueTID,
		OrderHistory:     append([]OrderState(nil), s.orderHistory...),
		Fills:            append([]FillState(nil), s.fills...),
	}
	for _, row := range s.sortedOpenOrders() {
		state.OpenOrders = append(state.OpenOrders, storedOrder{
			Request:           copyRequest(row.request),
			VenueOrderID:      row.venueOrderID,
			Status:            row.status,
			Armed:             row.armed,
			RemainingQuantity: row.remainingQuantity.String(),
			FilledQuantity:    row.filledQuantity.String(),
			AverageFillPrice:  row.averageFillPrice.String(),
			Fees:              row.fees.String(),
			TimestampMS:       row.timestampMS,
		})
	}
	return state
}

func (s *Simulator) stage() *Simulator {
	if s.config.PersistMode != "max" {
		return s
	}
	var staged = *s
	staged.openOrders = make(map[string]*simOrder, len(s.openOrders))
	for value, row := range s.openOrders {
		var copied = *row
		copied.request = copyRequest(row.request)
		staged.openOrders[value] = &copied
	}
	staged.orderHistory = append([]OrderState(nil), s.orderHistory...)
	staged.fills = append([]FillState(nil), s.fills...)
	return &staged
}

func (s *Simulator) commit(staged *Simulator) {
	if staged != s {
		*s = *staged
	}
}

func (s *Simulator) restore(state storedState) error {
	if state.NextVenueOrderID == 0 || state.NextVenueTID == 0 {
		return fmt.Errorf("load Simulator: invalid counters")
	}
	var restored = make(map[string]*simOrder, len(state.OpenOrders))
	for _, stored := range state.OpenOrders {
		if restored[stored.Request.CLOID] != nil {
			return fmt.Errorf("load Simulator: duplicate active cloid %s", stored.Request.CLOID)
		}
		var remaining, err = decimal.NewFromString(stored.RemainingQuantity)
		if err != nil {
			return fmt.Errorf("load Simulator: invalid remaining quantity: %v", err)
		}
		var filled decimal.Decimal
		filled, err = decimal.NewFromString(stored.FilledQuantity)
		if err != nil {
			return fmt.Errorf("load Simulator: invalid filled quantity: %v", err)
		}
		var average decimal.Decimal
		average, err = decimal.NewFromString(stored.AverageFillPrice)
		if err != nil {
			return fmt.Errorf("load Simulator: invalid average fill price: %v", err)
		}
		var fees decimal.Decimal
		fees, err = decimal.NewFromString(stored.Fees)
		if err != nil {
			return fmt.Errorf("load Simulator: invalid fees: %v", err)
		}
		var row = &simOrder{
			request:           copyRequest(stored.Request),
			venueOrderID:      stored.VenueOrderID,
			status:            stored.Status,
			armed:             stored.Armed,
			remainingQuantity: remaining,
			filledQuantity:    filled,
			averageFillPrice:  average,
			fees:              fees,
			timestampMS:       stored.TimestampMS,
		}
		if err = s.validateRequest(row.request); err != nil {
			return fmt.Errorf("load Simulator: %w", err)
		}
		restored[row.request.CLOID] = row
	}
	s.nextVenueOrderID = state.NextVenueOrderID
	s.nextVenueTID = state.NextVenueTID
	s.openOrders = restored
	s.orderHistory = append([]OrderState(nil), state.OrderHistory...)
	s.fills = append([]FillState(nil), state.Fills...)
	return nil
}

// Section 3 - Generic Helpers

func copyRequest(request OrderRequest) OrderRequest {
	var copied = request
	if request.Price != nil {
		var value = *request.Price
		copied.Price = &value
	}
	if request.TriggerPrice != nil {
		var value = *request.TriggerPrice
		copied.TriggerPrice = &value
	}
	return copied
}

func crosses(row *simOrder, price decimal.Decimal) bool {
	if row.request.TimeInForce == order.IOC {
		return true
	}
	var threshold = *row.request.Price
	if row.request.TriggerPrice != nil {
		threshold = *row.request.TriggerPrice
	}
	switch row.request.Role {
	case order.TakeProfit:
		if row.request.Side == order.Sell {
			return price.GreaterThanOrEqual(threshold)
		}
		return price.LessThanOrEqual(threshold)
	case order.StopLoss:
		if row.request.Side == order.Sell {
			return price.LessThanOrEqual(threshold)
		}
		return price.GreaterThanOrEqual(threshold)
	default:
		if row.request.Side == order.Buy {
			return price.LessThanOrEqual(threshold)
		}
		return price.GreaterThanOrEqual(threshold)
	}
}

func submitStatus(row *simOrder, entryFilled bool) SubmitStatus {
	var result = SubmitStatus{
		VenueOrderID: row.venueOrderID,
		CLOID:        row.request.CLOID,
	}
	if row.status == string(order.Filled) {
		result.Kind = "filled"
		result.FilledQuantity = row.filledQuantity
		result.AveragePrice = row.averageFillPrice
		return result
	}
	if row.request.Role == order.TakeProfit || row.request.Role == order.StopLoss {
		if entryFilled {
			result.Kind = WaitingForTrigger
		} else {
			result.Kind = WaitingForFill
		}
		return result
	}
	result.Kind = "resting"
	return result
}

func closePnL(
	current position,
	side string,
	quantity decimal.Decimal,
	price decimal.Decimal,
) decimal.Decimal {
	if current.size.IsPositive() && side == order.Sell {
		return price.Sub(current.entryPrice).Mul(decimal.Min(current.size.Abs(), quantity))
	}
	if current.size.IsNegative() && side == order.Buy {
		return current.entryPrice.Sub(price).Mul(decimal.Min(current.size.Abs(), quantity))
	}
	return decimal.Zero
}

func fillDirection(size decimal.Decimal, side string) string {
	if side == order.Buy {
		if size.IsNegative() {
			return "Close Short"
		}
		return "Open Long"
	}
	if size.IsPositive() {
		return "Close Long"
	}
	return "Open Short"
}

func sameSign(left decimal.Decimal, right decimal.Decimal) bool {
	return left.IsPositive() && right.IsPositive() ||
		left.IsNegative() && right.IsNegative()
}
