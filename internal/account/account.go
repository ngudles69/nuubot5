// Package account owns one Executor's trading boundary, Ledger, and Venue.
package account

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/cloid"
	"nuubot/internal/fill"
	"nuubot/internal/hyperliquid"
	"nuubot/internal/ledger"
	"nuubot/internal/market"
	"nuubot/internal/order"
	"nuubot/internal/setup"
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
	Infrastructure setup.Infrastructure
	LedgerID       uint64
	CycleNumber    int
	ExecutorNumber int
	Name           string
	Venue          string
	Network        string
	Symbol         string
	EquityUSDC     decimal.Decimal
	FeePct         decimal.Decimal
	SlippagePct    decimal.Decimal
	PersistMode    string
	Recon          string
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
	Orders  []order.Record
}

// Snapshot contains one immutable coherent post-recon Account value.
type Snapshot struct {
	Generation       uint64
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
	RealizedPnL      decimal.Decimal
	UnrealizedPnL    decimal.Decimal
	GrossPnL         decimal.Decimal
	Fees             decimal.Decimal
	NetPnL           decimal.Decimal
	OpenTrades       int
	ActiveOrders     int
	Fills            int
	PendingOrders    int
	PendingFills     int
}

// Result contains one immutable terminal Account result.
type Result struct {
	Config
	Snapshot Snapshot
	Recon    ReconStats
	Ledger   ledger.Result
}

// ReconStats contains cumulative reconciliation outcomes.
type ReconStats struct {
	Calls        uint64
	SkippedClean uint64
	Executed     uint64
	Succeeded    uint64
	Failed       uint64
}

// ReconOutcome identifies one complete canonical reconciliation result.
type ReconOutcome string

const (
	// ReconSkipped retained one trusted Snapshot without Venue work.
	ReconSkipped ReconOutcome = "skipped"
	// ReconSucceeded published one new trusted Snapshot generation.
	ReconSucceeded ReconOutcome = "succeeded"
	// ReconFailed published no decision Snapshot.
	ReconFailed ReconOutcome = "failed"
)

// FillQueryTelemetry contains one physical Fill-history request observation.
type FillQueryTelemetry struct {
	Kind           string
	StartMS        uint64
	EndMS          uint64
	Rows           int
	DurationMS     int64
	FillsAdded     int
	FillsUnchanged int
	FeesEnriched   int
	PendingMatched int
	Error          string
}

// ReconTelemetry contains one canonical reconciliation observation.
type ReconTelemetry struct {
	Kind                string
	Outcome             ReconOutcome
	Stage               string
	ObservedMS          uint64
	DurationMS          int64
	ConsecutiveFailures uint64
	Orders              int
	OrderStatusQueries  int
	Fills               int
	PendingOrdersBefore int
	PendingFillsBefore  int
	PendingOrders       int
	PendingFills        int
	FillQueries         []FillQueryTelemetry
	Error               string
}

type stats struct {
	ordersPlaced uint64
	reconciles   uint64
	bbosIngested uint64
}

type venue interface {
	PlaceOrders(hyperliquid.PlaceOrderAction, uint64) ([]byte, error)
	CancelOrders(hyperliquid.CancelByCLOIDAction, uint64) ([]byte, error)
	IngestBBO(market.BBO) (bool, error)
	OpenOrders(string) ([]byte, error)
	Fills(string, uint64, uint64) ([]byte, error)
	OrderStatus(string, string) ([]byte, error)
	AccountState(string) ([]byte, error)
	Stop() error
}

// Account owns one Venue and one Ledger.
type Account struct {
	log            *logging.Logger
	config         Config
	ledger         ledger.Ledger
	venue          venue
	lastBBO        market.BBO
	lastSnapshot   Snapshot
	reconTelemetry ReconTelemetry
	reconStats     ReconStats
	stats          stats
	generation     uint64
	failureCount   uint64
	dirty          bool
	hasBBO         bool
	started        bool
	stopped        bool
}

// Section 1 - Program Flow

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
		var current trade.ReconState
		current, err = a.ledger.TradeState(normalized[0].TradeID)
		if err != nil {
			return PlaceResult{}, fmt.Errorf("place Account Orders: %w", err)
		}
		tradeInput = current.Input
		batchNo, err = a.ledger.NextBatchNo(normalized[0].TradeID)
		if err != nil {
			return PlaceResult{}, fmt.Errorf("place Account Orders: %w", err)
		}
		orderIDs, err = a.ledger.PlanOrders(len(normalized))
		if err != nil {
			return PlaceResult{}, fmt.Errorf("place Account Orders: %w", err)
		}
	}

	// create CLOIDs
	var created = make([]*order.Order, 0, len(normalized))
	var requests = make([]hyperliquid.OrderRequest, 0, len(normalized))
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
		var request hyperliquid.OrderRequest
		request, err = a.venueOrderRequest(spec, value)
		if err != nil {
			return PlaceResult{}, err
		}
		requests = append(requests, request)
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
	var payload []byte
	payload, err = a.venue.PlaceOrders(hyperliquid.PlaceOrderAction{
		Type:     "order",
		Orders:   requests,
		Grouping: venueGrouping(normalized),
	}, normalized[0].TimestampMS)
	if err != nil {
		var submitErr = fmt.Errorf("place Account Orders: submit Venue batch: %w", err)
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
	var response hyperliquid.SubmitResponse
	response, err = hyperliquid.DecodeSubmitResponse(payload)
	if err != nil {
		return PlaceResult{TradeID: tradeInput.TradeID}, fmt.Errorf(
			"place Account Orders: %w",
			err,
		)
	}
	var partial = PlaceResult{
		TradeID: tradeInput.TradeID,
	}
	if len(response.Statuses) != len(created) {
		return partial, fmt.Errorf("place Account Orders: malformed Venue response")
	}
	var outcomes = make([]ledger.SubmitOutcome, 0, len(response.Statuses))
	var rejected = false
	for index, status := range response.Statuses {
		var expectedCLOID = created[index].ReconState().CLOID
		if status.CLOID != "" && status.CLOID != expectedCLOID {
			return partial, fmt.Errorf(
				"place Account Orders: response %d changed CLOID",
				index,
			)
		}
		var hasIdentity = status.CLOID == expectedCLOID || status.VenueOrderID != 0
		switch status.Kind {
		case "resting":
			if !hasIdentity {
				return partial, fmt.Errorf(
					"place Account Orders: response %d lacks Venue identity",
					index,
				)
			}
		case "filled":
			if !hasIdentity {
				return partial, fmt.Errorf(
					"place Account Orders: response %d lacks Venue identity",
					index,
				)
			}
			var total, totalErr = decimal.NewFromString(status.TotalSize)
			var average, averageErr = decimal.NewFromString(status.AveragePrice)
			if totalErr != nil || averageErr != nil ||
				!total.IsPositive() || !average.IsPositive() {
				return partial, fmt.Errorf(
					"place Account Orders: response %d has invalid Fill",
					index,
				)
			}
			if status.Fee != nil {
				var _, feeErr = decimal.NewFromString(*status.Fee)
				if feeErr != nil {
					return partial, fmt.Errorf(
						"place Account Orders: response %d has invalid fee",
						index,
					)
				}
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
			Raw:          response.Raw,
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
	var records []order.Record
	records, err = a.ledger.Orders(orderIDs)
	if err != nil {
		return partial, fmt.Errorf("place Account Orders: %w", err)
	}
	var result = PlaceResult{
		TradeID: tradeInput.TradeID,
		Orders:  records,
	}
	if rejected {
		return result, fmt.Errorf("place Account Orders: Venue rejected one or more Orders")
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
	var cancels = make([]hyperliquid.CancelByCLOIDRequest, 0, len(cloids))
	for _, value := range cloids {
		cancels = append(cancels, hyperliquid.CancelByCLOIDRequest{
			Asset: int(a.config.Infrastructure.Meta.AssetID),
			CLOID: value,
		})
	}
	var payload, err = a.venue.CancelOrders(hyperliquid.CancelByCLOIDAction{
		Type:    "cancelByCloid",
		Cancels: cancels,
	}, timestampMS)
	if err != nil {
		return fmt.Errorf("cancel Account Orders: %w", err)
	}

	// validate cancel response
	var response hyperliquid.CancelResponse
	response, err = hyperliquid.DecodeCancelResponse(payload)
	if err != nil {
		return fmt.Errorf("cancel Account Orders: %w", err)
	}
	if len(response.Statuses) != len(cloids) {
		return fmt.Errorf("cancel Account Orders: malformed Venue response")
	}
	for index, status := range response.Statuses {
		if !status.Success {
			return fmt.Errorf(
				"cancel Account Orders: Venue rejected Order %d: %s",
				index,
				status.Error,
			)
		}
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
	var changed, err = a.venue.IngestBBO(bbo)
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

// Result returns one immutable terminal Account result.
func (a *Account) Result() (Result, error) {
	// get immutable Ledger result
	var ledgerResult, err = a.ledger.Result()
	if err != nil {
		return Result{}, fmt.Errorf("read Account result: %w", err)
	}

	// return immutable Account result
	return Result{
		Config:   a.config,
		Snapshot: a.lastSnapshot,
		Recon:    a.reconStats,
		Ledger:   ledgerResult,
	}, nil
}

// Telemetry returns the latest observed immutable Account snapshot.
func (a *Account) Telemetry() (Snapshot, bool) {
	if a.lastSnapshot.ObservedMS == 0 {
		return Snapshot{}, false
	}
	return a.lastSnapshot, true
}

// ReconciliationTelemetry returns the latest canonical Recon outcome.
func (a *Account) ReconciliationTelemetry() ReconTelemetry {
	var current = a.reconTelemetry
	current.FillQueries = append([]FillQueryTelemetry(nil), current.FillQueries...)
	return current
}

// Clone returns one independently owned Account result.
func (r Result) Clone() Result {
	return r
}

// Stop releases the owned Venue and Ledger.
func (a *Account) Stop() error {
	if a.stopped {
		return nil
	}

	// stop Venue
	var venueErr = a.venue.Stop()

	// stop Ledger
	var ledgerErr = a.ledger.Stop()

	// stop Account
	a.started = false
	a.stopped = true
	a.log.Info(fmt.Sprintf(
		"account stopped cycle=%d executor=%d account=%s orders=%d reconciles=%d recon_calls=%d recon_skipped_clean=%d recon_executed=%d recon_succeeded=%d recon_failed=%d ingested_bbos=%d",
		a.config.CycleNumber,
		a.config.ExecutorNumber,
		a.config.Name,
		a.stats.ordersPlaced,
		a.stats.reconciles,
		a.reconStats.Calls,
		a.reconStats.SkippedClean,
		a.reconStats.Executed,
		a.reconStats.Succeeded,
		a.reconStats.Failed,
		a.stats.bbosIngested,
	))
	return errors.Join(venueErr, ledgerErr)
}

// Section 2 - Domain Helpers

// ActiveOrders returns current active local Order snapshots.
func (a *Account) ActiveOrders() []order.ActiveState {
	return a.ledger.ActiveOrders()
}

// Trade returns focused current Trade state.
func (a *Account) Trade(tradeID uint64) (trade.ReconState, error) {
	return a.ledger.TradeState(tradeID)
}

// OpenTrades returns focused state for current open exposure.
func (a *Account) OpenTrades() []trade.ReconState {
	return a.ledger.OpenTrades()
}

// CountOrders returns the number of Orders matching role and status.
func (a *Account) CountOrders(role string, status order.Status) uint64 {
	return a.ledger.CountOrders(role, status)
}

// TradeOrders returns flat Order records linked by Trade identity.
func (a *Account) TradeOrders(tradeID uint64) ([]order.Record, error) {
	return a.ledger.TradeOrders(tradeID)
}

// Order returns one flat Order record by local identity.
func (a *Account) Order(orderID uint64) (order.Record, error) {
	var records, err = a.ledger.Orders([]uint64{orderID})
	if err != nil {
		return order.Record{}, err
	}
	return records[0], nil
}

// Fill returns one flat Fill record by Venue identity.
func (a *Account) Fill(venueTID uint64) (fill.Record, bool) {
	return a.ledger.Fill(venueTID)
}

// PositionQuantity returns the latest reconciled signed exposure.
func (a *Account) PositionQuantity() decimal.Decimal {
	return a.lastSnapshot.PositionQuantity
}

// HasPendingRecon reports unresolved Order or Fill evidence.
func (a *Account) HasPendingRecon() bool {
	return a.ledger.HasPendingRecon()
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
		if spec.TradeID != tradeID || spec.TimestampMS != specs[0].TimestampMS ||
			spec.Quantity.IsNegative() || spec.Quantity.IsZero() ||
			spec.Price == nil || !spec.Price.IsPositive() || spec.TimestampMS == 0 {
			return nil, fmt.Errorf("place Account Orders: invalid or mixed batch")
		}
		var roundedPrice = a.config.Infrastructure.Meta.RoundPrice(*spec.Price)
		normalized[index].Price = &roundedPrice
		if spec.TriggerPrice != nil {
			var trigger = a.config.Infrastructure.Meta.RoundPrice(*spec.TriggerPrice)
			normalized[index].TriggerPrice = &trigger
		}
		normalized[index].Quantity = a.config.Infrastructure.Meta.RoundSize(spec.Quantity)
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
		var minNotionalUSDC = decimal.NewFromInt(
			int64(a.config.Infrastructure.App.Hyperliquid.MinOrderNotionalUSDC),
		)
		var quantity = decimal.Zero
		var step = decimal.New(1, -a.config.Infrastructure.Meta.SizeDecimals)
		for _, spec := range normalized {
			var required = minNotionalUSDC.Div(*spec.Price).
				Truncate(a.config.Infrastructure.Meta.SizeDecimals)
			if required.Mul(*spec.Price).LessThan(minNotionalUSDC) {
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
	if a.config.Infrastructure.Meta.AssetID > 0xffff ||
		spec.TimestampMS/1000 > 0x7fffffff {
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
		SymbolID:   uint16(a.config.Infrastructure.Meta.AssetID),
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

func (a *Account) venueOrderRequest(
	spec OrderSpec,
	value string,
) (hyperliquid.OrderRequest, error) {
	var request = hyperliquid.OrderRequest{
		Asset:      int(a.config.Infrastructure.Meta.AssetID),
		IsBuy:      spec.Side == order.Buy,
		Price:      spec.Price.String(),
		Size:       spec.Quantity.String(),
		ReduceOnly: spec.ReduceOnly,
		CLOID:      value,
	}
	switch spec.Type {
	case order.Limit:
		request.Type.Limit = &hyperliquid.LimitOrderType{
			TimeInForce: spec.TimeInForce,
		}
	case order.Market:
		request.Type.Limit = &hyperliquid.LimitOrderType{
			TimeInForce: order.IOC,
		}
	case order.Trigger:
		var tpsl string
		switch spec.Role {
		case order.TakeProfit:
			tpsl = "tp"
		case order.StopLoss:
			tpsl = "sl"
		default:
			return hyperliquid.OrderRequest{}, fmt.Errorf(
				"place Account Orders: trigger role %q is invalid",
				spec.Role,
			)
		}
		if spec.TriggerPrice == nil || !spec.TriggerPrice.IsPositive() {
			return hyperliquid.OrderRequest{}, fmt.Errorf(
				"place Account Orders: positive trigger price is required",
			)
		}
		request.Type.Trigger = &hyperliquid.TriggerOrderType{
			IsMarket:     false,
			TriggerPrice: spec.TriggerPrice.String(),
			TPSL:         tpsl,
		}
	default:
		return hyperliquid.OrderRequest{}, fmt.Errorf(
			"place Account Orders: unknown Order type %q",
			spec.Type,
		)
	}
	return request, nil
}

func venueGrouping(specs []OrderSpec) string {
	for _, spec := range specs {
		if spec.Role == order.TakeProfit || spec.Role == order.StopLoss {
			return "normalTpsl"
		}
	}
	return "na"
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
