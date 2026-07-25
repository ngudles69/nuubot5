// Package account owns one Executor's trading boundary, Ledger, and Venue.
package account

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/cloid"
	"nuubot/internal/fill"
	"nuubot/internal/ledger"
	"nuubot/internal/market"
	"nuubot/internal/meta"
	"nuubot/internal/order"
	"nuubot/internal/simulator"
	"nuubot/internal/toolkit/logging"
	"nuubot/internal/trade"
)

var networkCodes = map[string]uint8{
	"mainnet": 0,
	"testnet": 1,
	"simnet":  2,
}

var purposeCodes = map[string]uint8{
	order.Entry:      1,
	order.TakeProfit: 2,
	order.StopLoss:   3,
	order.Exit:       4,
	order.Cleanup:    6,
	order.Stop:       7,
}

// ErrNotSubmitted proves the Venue did not commit the requested Order batch.
var ErrNotSubmitted = errors.New("account Order batch was not submitted")

// Config contains one Account's identity, policy, and direct-child inputs.
type Config struct {
	LedgerID        uint64
	CycleNumber     int
	ExecutorNumber  int
	Name            string
	Venue           string
	Network         string
	Symbol          string
	Meta            meta.Instrument
	MinNotionalUSDC decimal.Decimal
	EquityUSDC      decimal.Decimal
	FeePct          decimal.Decimal
	SlippagePct     decimal.Decimal
	PersistMode     string
	ResultPath      string
}

// OrderSpec contains one Executor-owned Order intent.
type OrderSpec struct {
	TradeID      uint64
	OrderLevel   uint16
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

// PlaceResult contains one admitted local Trade and ordered submission evidence.
type PlaceResult struct {
	TradeID uint64
	Orders  []order.Snapshot
	Submit  simulator.SubmitResponse
}

// Snapshot contains one immutable coherent post-recon Account value.
type Snapshot struct {
	CycleNumber      int
	ExecutorNumber   int
	Account          string
	Venue            string
	Network          string
	Symbol           string
	ObservedMS       uint64
	AccountValue     decimal.Decimal
	Withdrawable     decimal.Decimal
	PositionQuantity decimal.Decimal
	EntryPrice       decimal.Decimal
	UnrealizedPnL    decimal.Decimal
	GrossPnL         decimal.Decimal
	Fees             decimal.Decimal
	NetPnL           decimal.Decimal
	OpenTrades       int
	ActiveOrders     int
	Fills            int
}

// Result contains one immutable terminal Account result.
type Result struct {
	Config
	Snapshot  Snapshot
	Ledger    ledger.Result
	Simulator *simulator.Result
}

type stats struct {
	ordersPlaced uint64
	reconciles   uint64
	bbosIngested uint64
}

// Account owns one Simulator and one Ledger.
type Account struct {
	log          *logging.Logger
	config       Config
	ledger       ledger.Ledger
	simulator    simulator.Simulator
	lastBBO      market.BBO
	lastSnapshot Snapshot
	stats        stats
	dirty        bool
	hasBBO       bool
	started      bool
	stopped      bool
}

// Section 1 - Program Flow

// Init prepares one Simulator-backed Account.
func (a *Account) Init(log *logging.Logger, cfg Config) error {
	a.log = log
	a.config = cfg

	// validate Account identity
	if log == nil || cfg.LedgerID == 0 || cfg.CycleNumber <= 0 ||
		cfg.ExecutorNumber <= 0 || cfg.Name == "" || cfg.Symbol == "" {
		return fmt.Errorf("initialize Account: complete identity is required")
	}
	if cfg.Venue != "simulator" || cfg.Network != "simnet" {
		return fmt.Errorf("initialize Account: first trading tranche requires simulator simnet")
	}
	if cfg.Meta.Symbol != cfg.Symbol || cfg.Meta.IsDelisted || cfg.Meta.Retired {
		return fmt.Errorf("initialize Account: symbol Meta is unavailable")
	}
	if !cfg.MinNotionalUSDC.IsPositive() || !cfg.EquityUSDC.IsPositive() {
		return fmt.Errorf("initialize Account: notional floor and equity must be positive")
	}

	// validate persistence mode
	if cfg.PersistMode != ledger.None && cfg.PersistMode != ledger.Max {
		return fmt.Errorf("initialize Account: invalid persistence mode %q", cfg.PersistMode)
	}

	// initialize Ledger with persistence mode
	var err = a.ledger.Init(ledger.Config{
		ID:             cfg.LedgerID,
		CycleNumber:    cfg.CycleNumber,
		ExecutorNumber: cfg.ExecutorNumber,
		Account:        cfg.Name,
		Network:        cfg.Network,
		Symbol:         cfg.Symbol,
		PersistMode:    cfg.PersistMode,
		Path:           cfg.ResultPath,
	})
	if err != nil {
		return fmt.Errorf("initialize Account: %w", err)
	}

	// initialize Venue with persistence mode
	err = a.simulator.Init(simulator.Config{
		LedgerID:    cfg.LedgerID,
		Name:        cfg.Name,
		Account:     cfg.Name,
		CycleNumber: cfg.CycleNumber,
		Symbol:      cfg.Symbol,
		Equity:      cfg.EquityUSDC,
		FeePct:      cfg.FeePct,
		SlippagePct: cfg.SlippagePct,
		PersistMode: cfg.PersistMode,
		Path:        cfg.ResultPath,
	})
	if err != nil {
		a.ledger.Stop()
		return fmt.Errorf("initialize Account: %w", err)
	}

	// initialize Account
	a.dirty = true
	a.started = true
	return nil
}

// PlaceOrders submits one complete validated Order batch.
func (a *Account) PlaceOrders(specs []OrderSpec) (PlaceResult, error) {
	if !a.started || a.stopped {
		return PlaceResult{}, fmt.Errorf("place Account Orders: invalid lifecycle state")
	}

	// validate complete order batch
	var normalized, err = a.normalizeSpecs(specs)
	if err != nil {
		return PlaceResult{}, err
	}

	// resolve Trade ownership
	var newTrade = normalized[0].Role == order.Entry
	var tradeInput trade.Input
	var orderIDs []uint64
	var batchNo uint16 = 1
	if newTrade {
		var plan ledger.Plan
		plan, err = a.ledger.PlanTrade(len(normalized))
		if err != nil {
			return PlaceResult{}, fmt.Errorf("place Account Orders: %w", err)
		}
		tradeInput = plan.Trade
		orderIDs = plan.OrderIDs
	} else {
		var current trade.Snapshot
		current, err = a.ledger.Trade(normalized[0].TradeID, a.markPrice())
		if err != nil {
			return PlaceResult{}, fmt.Errorf("place Account Orders: %w", err)
		}
		tradeInput = current.Input
		for _, existing := range current.Orders {
			if existing.BatchNo >= batchNo {
				batchNo = existing.BatchNo + 1
			}
		}
		if batchNo > 1000 {
			return PlaceResult{}, fmt.Errorf("place Account Orders: Trade exceeded 1000 batches")
		}
		orderIDs, err = a.ledger.PlanOrders(len(normalized))
		if err != nil {
			return PlaceResult{}, fmt.Errorf("place Account Orders: %w", err)
		}
	}

	// create CLOIDs
	var created = make([]*order.Order, 0, len(normalized))
	var requests = make([]simulator.OrderRequest, 0, len(normalized))
	for index, spec := range normalized {
		var value string
		value, err = a.createCLOID(
			tradeInput.TradeNo,
			batchNo,
			spec.OrderLevel,
			spec,
		)
		if err != nil {
			return PlaceResult{}, err
		}
		var createdOrder *order.Order
		createdOrder, err = order.New(order.Input{
			LedgerID:          a.config.LedgerID,
			TradeID:           tradeInput.TradeID,
			OrderID:           orderIDs[index],
			Account:           a.config.Name,
			CycleNumber:       a.config.CycleNumber,
			Symbol:            a.config.Symbol,
			BatchNo:           batchNo,
			OrderPos:          uint16(index + 1),
			CLOID:             value,
			Role:              spec.Role,
			Side:              spec.Side,
			Type:              spec.Type,
			TimeInForce:       spec.TimeInForce,
			RequestedQuantity: spec.Quantity,
			RequestedPrice:    spec.Price,
			TriggerPrice:      spec.TriggerPrice,
			ReduceOnly:        spec.ReduceOnly,
			TimestampMS:       spec.TimestampMS,
		})
		if err != nil {
			return PlaceResult{}, fmt.Errorf("place Account Orders: %w", err)
		}
		created = append(created, createdOrder)
		requests = append(requests, simulator.OrderRequest{
			CLOID:        value,
			Symbol:       a.config.Symbol,
			TradeID:      tradeInput.TradeID,
			OrderID:      orderIDs[index],
			Role:         spec.Role,
			Side:         spec.Side,
			Type:         spec.Type,
			TimeInForce:  spec.TimeInForce,
			Quantity:     spec.Quantity,
			Price:        spec.Price,
			TriggerPrice: spec.TriggerPrice,
			ReduceOnly:   spec.ReduceOnly,
			TimestampMS:  spec.TimestampMS,
		})
	}

	// commit created Trade and Orders
	if newTrade {
		err = a.ledger.CreateTrade(tradeInput, created)
	} else {
		err = a.ledger.AddOrders(tradeInput.TradeID, created)
	}
	if err != nil {
		return PlaceResult{}, fmt.Errorf("place Account Orders: %w", err)
	}

	// submit Venue batch
	a.dirty = true
	var response simulator.SubmitResponse
	response, err = a.simulator.PlaceOrders(requests)
	if err != nil {
		var submitErr = fmt.Errorf("place Account Orders: submit Simulator batch: %w", err)
		var outcomes = make([]ledger.SubmitOutcome, len(orderIDs))
		for index, orderID := range orderIDs {
			outcomes[index] = ledger.SubmitOutcome{
				OrderID:     orderID,
				Error:       submitErr.Error(),
				Status:      order.Error,
				TimestampMS: normalized[index].TimestampMS,
			}
		}
		var ledgerErr = a.ledger.RecordSubmit(outcomes)
		return PlaceResult{
			TradeID: tradeInput.TradeID,
		}, errors.Join(ErrNotSubmitted, submitErr, ledgerErr)
	}

	// validate submit response
	var partial = PlaceResult{
		TradeID: tradeInput.TradeID,
		Submit:  response,
	}
	if response.Status != "ok" || response.Type != "order" ||
		len(response.Statuses) != len(created) {
		return partial, fmt.Errorf("place Account Orders: malformed Simulator response")
	}
	var outcomes = make([]ledger.SubmitOutcome, 0, len(response.Statuses))
	var rejected = false
	for index, status := range response.Statuses {
		switch status.Kind {
		case "resting", "filled", simulator.WaitingForFill, simulator.WaitingForTrigger:
			if status.VenueOrderID == 0 {
				return partial, fmt.Errorf(
					"place Account Orders: response %d lacks Venue identity",
					index,
				)
			}
		case "error":
			if status.Error == "" {
				return partial, fmt.Errorf(
					"place Account Orders: response %d has empty error",
					index,
				)
			}
			rejected = true
		default:
			return partial, fmt.Errorf(
				"place Account Orders: unknown response %q",
				status.Kind,
			)
		}
		var terminalStatus order.Status
		var timestampMS uint64
		if status.Kind == "error" {
			terminalStatus = order.Rejected
			timestampMS = normalized[index].TimestampMS
		}
		outcomes = append(outcomes, ledger.SubmitOutcome{
			OrderID:      orderIDs[index],
			VenueOrderID: status.VenueOrderID,
			Error:        status.Error,
			Status:       terminalStatus,
			TimestampMS:  timestampMS,
		})
	}

	// commit submit outcomes
	err = a.ledger.RecordSubmit(outcomes)
	if err != nil {
		return partial, fmt.Errorf("place Account Orders: %w", err)
	}

	// mark Account dirty
	a.dirty = true
	a.stats.ordersPlaced += uint64(len(created))
	var snapshots []order.Snapshot
	snapshots, err = a.ledger.Orders(orderIDs)
	if err != nil {
		return partial, fmt.Errorf("place Account Orders: %w", err)
	}
	var result = PlaceResult{
		TradeID: tradeInput.TradeID,
		Orders:  snapshots,
		Submit:  response,
	}
	if rejected {
		return result, fmt.Errorf("place Account Orders: Simulator rejected one or more Orders")
	}
	return result, nil
}

// CancelOrders requests cancellation of active owned Orders.
func (a *Account) CancelOrders(cloids []string, timestampMS uint64) error {
	if !a.started || a.stopped {
		return fmt.Errorf("cancel Account Orders: invalid lifecycle state")
	}

	// validate owned active Orders
	var active = make(map[string]struct{})
	for _, current := range a.ledger.ActiveOrders() {
		active[current.CLOID] = struct{}{}
	}
	for _, value := range cloids {
		if _, exists := active[value]; !exists {
			return fmt.Errorf("cancel Account Orders: unknown active cloid %s", value)
		}
	}

	// cancel Venue batch
	var response, err = a.simulator.CancelOrders(cloids, timestampMS)
	if err != nil {
		return fmt.Errorf("cancel Account Orders: %w", err)
	}

	// validate cancel response
	if response.Status != "ok" || response.Type != "cancel" ||
		len(response.Statuses) != len(cloids) {
		return fmt.Errorf("cancel Account Orders: malformed Simulator response")
	}

	// mark Account dirty
	a.dirty = true
	return nil
}

// IngestBBO advances Simulator truth before normal Executor BBO handling.
func (a *Account) IngestBBO(bbo market.BBO) error {
	if !a.started || a.stopped {
		return fmt.Errorf("ingest Account BBO: invalid lifecycle state")
	}
	if bbo.Symbol != "" && bbo.Symbol != a.config.Symbol {
		return fmt.Errorf(
			"ingest Account BBO: symbol %s does not match %s",
			bbo.Symbol,
			a.config.Symbol,
		)
	}

	// ingest Venue BBO
	var changed, err = a.simulator.IngestBBO(bbo)
	if err != nil {
		return fmt.Errorf("ingest Account BBO: %w", err)
	}
	a.lastBBO = bbo
	a.hasBBO = true
	a.stats.bbosIngested++

	// mark Account dirty when Venue or open-position marks change
	if changed || !a.lastSnapshot.PositionQuantity.IsZero() {
		a.dirty = true
	}
	return nil
}

// Reconcile creates one coherent Account snapshot.
func (a *Account) Reconcile(nowMS uint64, forced bool) (Snapshot, bool, error) {
	if !a.started || a.stopped || nowMS == 0 {
		return Snapshot{}, false, fmt.Errorf("reconcile Account: invalid state or timestamp")
	}
	if !a.dirty && !forced {
		return a.lastSnapshot, false, nil
	}

	// claim dirty state
	a.dirty = false
	var failed = true
	defer func() {
		if failed {
			a.dirty = true
		}
	}()

	// read open Venue Orders
	var openOrders = a.simulator.OpenOrders()
	var openCLOIDs = make(map[string]struct{}, len(openOrders))
	var evidence = make([]ledger.OrderEvidence, 0, len(openOrders))
	for _, current := range openOrders {
		openCLOIDs[current.CLOID] = struct{}{}
		evidence = append(evidence, orderEvidence(current))
	}

	// read bounded Venue Fills
	var fills = a.simulator.Fills(a.ledger.FillsThroughMS(), nowMS)
	var fillEvidence = make([]ledger.FillEvidence, 0, len(fills))
	for _, execution := range fills {
		var fee = execution.Fee
		fillEvidence = append(fillEvidence, ledger.FillEvidence{
			CLOID:        execution.CLOID,
			VenueOrderID: execution.VenueOrderID,
			VenueTID:     execution.VenueTID,
			Side:         execution.Side,
			Quantity:     execution.Quantity,
			Price:        execution.Price,
			TimestampMS:  execution.TimestampMS,
			Fee:          &fee,
			Liquidity:    execution.Liquidity,
		})
	}

	// read missing active Order statuses
	for _, active := range a.ledger.ActiveOrders() {
		if _, open := openCLOIDs[active.CLOID]; open {
			continue
		}
		var current, err = a.simulator.OrderStatus(active.CLOID)
		if err != nil {
			if active.VenueOrderID == 0 &&
				(active.Status == order.Created || active.Status == order.Submitted) {
				evidence = append(evidence, ledger.OrderEvidence{
					CLOID:        active.CLOID,
					Status:       order.Error,
					RejectReason: "Simulator submission is absent",
					TimestampMS:  nowMS,
				})
				continue
			}
			return Snapshot{}, false, fmt.Errorf("reconcile Account: %w", err)
		}
		evidence = append(evidence, orderEvidence(current))
	}

	// read Venue account state
	var state, err = a.simulator.AccountState()
	if err != nil {
		return Snapshot{}, false, fmt.Errorf("reconcile Account: %w", err)
	}

	// validate complete Venue evidence
	var raw []byte
	raw, err = json.Marshal(state)
	if err != nil {
		return Snapshot{}, false, fmt.Errorf("reconcile Account: encode Account state: %v", err)
	}

	// reconcile Ledger
	err = a.ledger.Recon(ledger.ReconInput{
		Orders:          evidence,
		Fills:           fillEvidence,
		FillsThroughMS:  nowMS,
		ObservedMS:      nowMS,
		AccountStateRaw: string(raw),
	})
	if err != nil {
		return Snapshot{}, false, fmt.Errorf("reconcile Account: %w", err)
	}

	// publish Account snapshot
	var result ledger.Result
	result, err = a.ledger.Result(a.markPrice())
	if err != nil {
		return Snapshot{}, false, fmt.Errorf("reconcile Account: %w", err)
	}
	var snapshot = accountSnapshot(a.config, nowMS, state, result)
	a.lastSnapshot = snapshot
	a.stats.reconciles++
	failed = false
	return snapshot, true, nil
}

// Result returns one immutable terminal Account result.
func (a *Account) Result() (Result, error) {
	// get immutable Ledger result
	var ledgerResult, err = a.ledger.Result(a.markPrice())
	if err != nil {
		return Result{}, fmt.Errorf("read Account result: %w", err)
	}

	// get immutable Simulator result
	var simulatorResult = a.simulator.Result()

	// return immutable Account result
	return Result{
		Config:    a.config,
		Snapshot:  a.lastSnapshot,
		Ledger:    ledgerResult,
		Simulator: &simulatorResult,
	}, nil
}

// Telemetry returns the latest observed immutable Account snapshot.
func (a *Account) Telemetry() (Snapshot, bool) {
	if a.lastSnapshot.ObservedMS == 0 {
		return Snapshot{}, false
	}
	return a.lastSnapshot, true
}

// Clone returns one independently owned Account result.
func (r Result) Clone() Result {
	var clone = r
	clone.Ledger.Trades = make([]trade.Snapshot, len(r.Ledger.Trades))
	for tradeIndex, ownedTrade := range r.Ledger.Trades {
		var tradeClone = ownedTrade
		tradeClone.Orders = make([]order.Snapshot, len(ownedTrade.Orders))
		for orderIndex, ownedOrder := range ownedTrade.Orders {
			var orderClone = ownedOrder
			orderClone.Fills = append([]fill.Snapshot(nil), ownedOrder.Fills...)
			tradeClone.Orders[orderIndex] = orderClone
		}
		clone.Ledger.Trades[tradeIndex] = tradeClone
	}
	if r.Simulator != nil {
		var simulatorClone = *r.Simulator
		simulatorClone.Orders = append([]simulator.OrderState(nil), r.Simulator.Orders...)
		simulatorClone.Fills = append([]simulator.FillState(nil), r.Simulator.Fills...)
		clone.Simulator = &simulatorClone
	}
	return clone
}

// Stop releases the owned Simulator and Ledger.
func (a *Account) Stop() error {
	if a.stopped {
		return nil
	}

	// stop Venue
	var venueErr = a.simulator.Stop()

	// stop Ledger
	var ledgerErr = a.ledger.Stop()

	// stop Account
	a.started = false
	a.stopped = true
	a.log.Info(fmt.Sprintf(
		"account stopped cycle=%d executor=%d account=%s orders=%d reconciles=%d ingested_bbos=%d",
		a.config.CycleNumber,
		a.config.ExecutorNumber,
		a.config.Name,
		a.stats.ordersPlaced,
		a.stats.reconciles,
		a.stats.bbosIngested,
	))
	return errors.Join(venueErr, ledgerErr)
}

// Section 2 - Domain Helpers

// ActiveOrders returns current active local Order snapshots.
func (a *Account) ActiveOrders() []order.Snapshot {
	return a.ledger.ActiveOrders()
}

// Trade returns one current owned Trade snapshot.
func (a *Account) Trade(tradeID uint64) (trade.Snapshot, error) {
	return a.ledger.Trade(tradeID, a.markPrice())
}

// PositionQuantity returns the latest reconciled signed exposure.
func (a *Account) PositionQuantity() decimal.Decimal {
	return a.lastSnapshot.PositionQuantity
}

func (a *Account) normalizeSpecs(specs []OrderSpec) ([]OrderSpec, error) {
	if len(specs) == 0 || len(specs) > 1000 {
		return nil, fmt.Errorf("place Account Orders: batch size must be from 1 to 1000")
	}
	var normalized = make([]OrderSpec, len(specs))
	var entries int
	var tradeID = specs[0].TradeID
	for index, spec := range specs {
		normalized[index] = copySpec(spec)
		if spec.Role == order.Entry {
			entries++
		}
		if spec.TradeID != tradeID || spec.Quantity.IsNegative() || spec.Quantity.IsZero() ||
			spec.Price == nil || !spec.Price.IsPositive() || spec.TimestampMS == 0 {
			return nil, fmt.Errorf("place Account Orders: invalid or mixed batch")
		}
		var roundedPrice = a.config.Meta.RoundPrice(*spec.Price)
		normalized[index].Price = &roundedPrice
		if spec.TriggerPrice != nil {
			var trigger = a.config.Meta.RoundPrice(*spec.TriggerPrice)
			normalized[index].TriggerPrice = &trigger
		}
		normalized[index].Quantity = a.config.Meta.RoundSize(spec.Quantity)
		if !normalized[index].Quantity.IsPositive() {
			return nil, fmt.Errorf("place Account Orders: quantity rounds to zero")
		}
	}
	var newTrade = normalized[0].Role == order.Entry
	if newTrade && (entries != 1 || tradeID != 0) {
		return nil, fmt.Errorf("place Account Orders: entry batch requires one new Trade")
	}
	if !newTrade && (entries != 0 || tradeID == 0) {
		return nil, fmt.Errorf("place Account Orders: existing Trade batch has invalid identity")
	}
	if newTrade {
		var quantity = decimal.Zero
		var step = decimal.New(1, -a.config.Meta.SizeDecimals)
		for _, spec := range normalized {
			var required = a.config.MinNotionalUSDC.Div(*spec.Price).
				Truncate(a.config.Meta.SizeDecimals)
			if required.Mul(*spec.Price).LessThan(a.config.MinNotionalUSDC) {
				required = required.Add(step)
			}
			if spec.Quantity.GreaterThan(quantity) {
				quantity = spec.Quantity
			}
			if required.GreaterThan(quantity) {
				quantity = required
			}
		}
		for index := range normalized {
			normalized[index].Quantity = quantity
		}
	}
	return normalized, nil
}

func (a *Account) createCLOID(
	tradeNo uint32,
	batchNo uint16,
	orderLevel uint16,
	spec OrderSpec,
) (string, error) {
	if a.config.Meta.AssetID > 0xffff || spec.TimestampMS/1000 > 0x7fffffff {
		return "", fmt.Errorf("place Account Orders: CLOID identity exceeds fixed range")
	}
	var side uint8
	if spec.Side == order.Buy {
		side = 1
	}
	var purpose, exists = purposeCodes[spec.Role]
	if !exists {
		return "", fmt.Errorf("place Account Orders: unknown Order role %q", spec.Role)
	}
	var network, networkExists = networkCodes[a.config.Network]
	if !networkExists {
		return "", fmt.Errorf("place Account Orders: unknown network %q", a.config.Network)
	}
	var value, err = cloid.Encode(cloid.Fields{
		BotCycleID: uint32(a.config.CycleNumber),
		SymbolID:   uint16(a.config.Meta.AssetID),
		Exchange:   1,
		Network:    network,
		Side:       side,
		ReduceOnly: spec.ReduceOnly,
		Purpose:    purpose,
		TradeNo:    tradeNo,
		BatchNo:    batchNo,
		OrderLevel: orderLevel,
		TimestampS: uint32(spec.TimestampMS / 1000),
	})
	if err != nil {
		return "", fmt.Errorf("place Account Orders: %w", err)
	}
	return value, nil
}

func (a *Account) markPrice() *decimal.Decimal {
	if !a.hasBBO {
		return nil
	}
	var price = decimal.NewFromFloat(a.lastBBO.Price)
	return &price
}

func orderEvidence(state simulator.OrderState) ledger.OrderEvidence {
	return ledger.OrderEvidence{
		CLOID:        state.CLOID,
		VenueOrderID: state.VenueOrderID,
		Status:       state.Status,
		TimestampMS:  state.TimestampMS,
		Raw:          state.Raw,
	}
}

func accountSnapshot(
	cfg Config,
	nowMS uint64,
	state simulator.AccountState,
	result ledger.Result,
) Snapshot {
	var gross = decimal.Zero
	var fees = decimal.Zero
	var net = decimal.Zero
	var openTrades int
	var fills int
	var activeOrders int
	for _, ownedTrade := range result.Trades {
		gross = gross.Add(ownedTrade.GrossPnL)
		fees = fees.Add(ownedTrade.Fees)
		net = net.Add(ownedTrade.NetPnL)
		if ownedTrade.Status != trade.Closed &&
			ownedTrade.Status != trade.Canceled &&
			ownedTrade.Status != trade.Error {
			openTrades++
		}
		for _, ownedOrder := range ownedTrade.Orders {
			if ownedOrder.Active {
				activeOrders++
			}
			fills += len(ownedOrder.Fills)
		}
	}
	return Snapshot{
		CycleNumber:      cfg.CycleNumber,
		ExecutorNumber:   cfg.ExecutorNumber,
		Account:          cfg.Name,
		Venue:            cfg.Venue,
		Network:          cfg.Network,
		Symbol:           cfg.Symbol,
		ObservedMS:       nowMS,
		AccountValue:     state.AccountValue,
		Withdrawable:     state.Withdrawable,
		PositionQuantity: state.PositionSize,
		EntryPrice:       state.EntryPrice,
		UnrealizedPnL:    state.UnrealizedPnL,
		GrossPnL:         gross,
		Fees:             fees,
		NetPnL:           net,
		OpenTrades:       openTrades,
		ActiveOrders:     activeOrders,
		Fills:            fills,
	}
}

// Section 3 - Generic Helpers

func copySpec(spec OrderSpec) OrderSpec {
	var copied = spec
	if spec.Price != nil {
		var value = *spec.Price
		copied.Price = &value
	}
	if spec.TriggerPrice != nil {
		var value = *spec.TriggerPrice
		copied.TriggerPrice = &value
	}
	return copied
}
