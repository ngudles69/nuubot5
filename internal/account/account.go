// Package account owns one Executor's trading boundary, Ledger, and Venue.
package account

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/account/ledger"
	"nuubot/internal/account/order"
	"nuubot/internal/account/trade"
	"nuubot/internal/cloid"
	"nuubot/internal/hyperliquid"
	"nuubot/internal/market"
	"nuubot/internal/setup"
	"nuubot/internal/toolkit/logging"
	"nuubot/internal/venue"
)

// ErrNotSubmitted proves the Venue did not commit the requested Order batch.
var ErrNotSubmitted = errors.New("account Order batch was not submitted")

// Config contains one Account's identity, policy, and direct-child inputs.
type Config struct {
	Nuubot         *setup.Nuubot
	SweepID        uint64
	BotID          uint64
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
	Level        uint16
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
	Orders  []order.Order
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
	OIDSearchOrders     int
	OIDSearchFills      int
	FillQueries         []FillQueryTelemetry
	Error               string
}

type stats struct {
	ordersPlaced uint64
	reconciles   uint64
}

// Account owns one Venue and one Ledger.
type Account struct {
	log            *logging.Logger
	config         Config
	ledger         ledger.Ledger
	venue          venue.Venue
	store          *store
	lastSnapshot   Snapshot
	reconTelemetry ReconTelemetry
	reconStats     ReconStats
	stats          stats
	generation     uint64
	failureCount   uint64
	lastReconMS    uint64
	dirty          bool
	started        bool
	stopped        bool
}

// Section 1 - Program Flow

// PlaceOrders submits one complete validated Order batch.
func (a *Account) PlaceOrders(specs []OrderSpec) (
	result PlaceResult,
	err error,
) {
	if !a.started || a.stopped {
		return PlaceResult{}, fmt.Errorf("place Account Orders: invalid lifecycle state")
	}
	defer func() {
		if a.config.PersistMode == "max" {
			err = errors.Join(err, a.persist(false))
		}
	}()

	// Step 1: validate complete Order batch
	var normalized []OrderSpec
	normalized, err = a.prepareOrderBatch(specs)
	if err != nil {
		return PlaceResult{}, err
	}

	// Step 2: resolve Trade ownership
	var newTrade = normalized[0].TradeID == 0
	var tradeRecord trade.Trade
	var orderIDs []uint64
	if newTrade {
		var plan ledger.Plan
		plan, err = a.ledger.PlanTrade(len(normalized))
		if err != nil {
			return PlaceResult{}, fmt.Errorf("place Account Orders: %w", err)
		}
		tradeRecord = plan.Trade
		orderIDs = plan.OrderIDs
	} else {
		var current trade.Trade
		current, err = a.ledger.Trade(normalized[0].TradeID)
		if err != nil {
			return PlaceResult{}, fmt.Errorf("place Account Orders: %w", err)
		}
		tradeRecord = current
		orderIDs, err = a.ledger.PlanOrders(len(normalized))
		if err != nil {
			return PlaceResult{}, fmt.Errorf("place Account Orders: %w", err)
		}
	}

	// Step 3: create CLOIDs
	var created = make([]*order.Order, 0, len(normalized))
	var requests = make([]hyperliquid.OrderRequest, 0, len(normalized))
	for index, spec := range normalized {
		var value string
		value, err = cloid.Encode(a.config.LedgerID, orderIDs[index])
		if err != nil {
			return PlaceResult{}, fmt.Errorf("place Account Orders: %w", err)
		}
		var createdOrder *order.Order
		createdOrder, err = order.New(order.Order{
			SweepID:           a.config.SweepID,
			BotID:             a.config.BotID,
			Venue:             a.config.Venue,
			Network:           a.config.Network,
			LedgerID:          a.config.LedgerID,
			TradeID:           tradeRecord.TradeID,
			OrderID:           orderIDs[index],
			Account:           a.config.Name,
			CycleNumber:       a.config.CycleNumber,
			Symbol:            a.config.Symbol,
			Level:             spec.Level,
			CLOID:             value,
			Role:              spec.Role,
			Side:              spec.Side,
			Type:              spec.Type,
			TimeInForce:       spec.TimeInForce,
			SubmittedQuantity: spec.Quantity,
			SubmittedPrice:    spec.Price,
			TriggerPrice:      spec.TriggerPrice,
			ReduceOnly:        spec.ReduceOnly,
			SubmittedMS:       spec.TimestampMS,
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

	// Step 4: commit created Trade and Orders
	if newTrade {
		var createdTrade *trade.Trade
		createdTrade, err = trade.New(tradeRecord)
		if err == nil {
			err = a.ledger.CreateTrade(createdTrade, created)
		}
	} else {
		err = a.ledger.AddOrders(tradeRecord.TradeID, created)
	}
	if err != nil {
		return PlaceResult{}, fmt.Errorf("place Account Orders: %w", err)
	}

	// Step 5: submit Venue batch
	a.dirty = true
	var payload []byte
	payload, err = a.venue.PlaceOrders(hyperliquid.PlaceOrderAction{
		Type:     "order",
		Orders:   requests,
		Grouping: venueGrouping(normalized),
	}, normalized[0].TimestampMS)
	if err != nil {
		var submitErr = fmt.Errorf("place Account Orders: submit Venue batch: %w", err)
		var outcomes = make([]ledger.OrderUpdate, len(orderIDs))
		for index, orderID := range orderIDs {
			outcomes[index] = ledger.OrderUpdate{
				OrderID: orderID,
				Update: order.Update{
					Status: order.Error, RejectReason: submitErr.Error(),
					UpdatedMS: normalized[index].TimestampMS,
				},
			}
		}
		var ledgerErr = a.ledger.UpdateOrders(outcomes)
		return PlaceResult{
			TradeID: tradeRecord.TradeID,
		}, errors.Join(ErrNotSubmitted, submitErr, ledgerErr)
	}

	// Step 6: validate submit response
	var response hyperliquid.SubmitResponse
	response, err = hyperliquid.DecodeSubmitResponse(payload)
	if err != nil {
		return PlaceResult{TradeID: tradeRecord.TradeID}, fmt.Errorf(
			"place Account Orders: %w",
			err,
		)
	}
	var partial = PlaceResult{
		TradeID: tradeRecord.TradeID,
	}
	if len(response.Statuses) != len(created) {
		return partial, fmt.Errorf("place Account Orders: malformed Venue response")
	}
	var outcomes = make([]ledger.OrderUpdate, 0, len(response.Statuses))
	var rejected = false
	for index, status := range response.Statuses {
		var expectedCLOID = created[index].CLOID
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
		var updateStatus = terminalStatus
		if updateStatus == "" {
			updateStatus = order.Submitted
			timestampMS = normalized[index].TimestampMS
		}
		outcomes = append(outcomes, ledger.OrderUpdate{
			OrderID: orderIDs[index],
			Update: order.Update{
				VenueOrderID: status.VenueOrderID,
				Status:       updateStatus, RejectReason: status.Error,
				UpdatedMS: timestampMS, RawJSON: response.Raw,
			},
		})
	}

	// Step 7: commit submit outcomes
	err = a.ledger.UpdateOrders(outcomes)
	if err != nil {
		return partial, fmt.Errorf("place Account Orders: %w", err)
	}

	// Step 8: mark Account dirty
	a.dirty = true
	a.stats.ordersPlaced += uint64(len(created))
	var records = make([]order.Order, 0, len(orderIDs))
	for _, orderID := range orderIDs {
		var current order.Order
		current, err = a.ledger.Order(orderID)
		if err != nil {
			return partial, fmt.Errorf("place Account Orders: %w", err)
		}
		records = append(records, current)
	}
	result = PlaceResult{
		TradeID: tradeRecord.TradeID,
		Orders:  records,
	}
	if rejected {
		return result, fmt.Errorf("place Account Orders: Venue rejected one or more Orders")
	}
	return result, nil
}

// CancelOrders requests cancellation of active owned Orders.
func (a *Account) CancelOrders(
	cloids []string,
	timestampMS uint64,
) (err error) {
	if !a.started || a.stopped {
		return fmt.Errorf("cancel Account Orders: invalid lifecycle state")
	}
	defer func() {
		if a.config.PersistMode == "max" {
			err = errors.Join(err, a.persist(false))
		}
	}()

	// Step 1: validate owned active Orders
	var active = make(map[string]struct{})
	for _, current := range a.ledger.ActiveOrders() {
		active[current.CLOID] = struct{}{}
	}
	for _, value := range cloids {
		if _, exists := active[value]; !exists {
			return fmt.Errorf("cancel Account Orders: unknown active cloid %s", value)
		}
	}

	// Step 2: cancel Venue batch
	var cancels = make([]hyperliquid.CancelByCLOIDRequest, 0, len(cloids))
	for _, value := range cloids {
		cancels = append(cancels, hyperliquid.CancelByCLOIDRequest{
			Asset: int(a.config.Nuubot.Meta.AssetID),
			CLOID: value,
		})
	}
	var payload []byte
	payload, err = a.venue.CancelOrders(hyperliquid.CancelByCLOIDAction{
		Type:    "cancelByCloid",
		Cancels: cancels,
	}, timestampMS)
	if err != nil {
		return fmt.Errorf("cancel Account Orders: %w", err)
	}

	// Step 3: validate cancel response
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

	// Step 4: mark Account dirty
	a.dirty = true
	return nil
}

// Result returns one immutable terminal Account result.
func (a *Account) Result() (Result, error) {
	// Step 1: get immutable Ledger result
	var ledgerResult, err = a.ledger.Result()
	if err != nil {
		return Result{}, fmt.Errorf("read Account result: %w", err)
	}

	// Step 2: return immutable Account result
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

// Stop releases the owned Venue and Ledger.
func (a *Account) Stop() error {
	if a.stopped {
		return nil
	}

	// Step 1: disconnect Venue
	var venueErr = a.venue.Disconnect()

	// Step 2: persist final Account evidence
	var storeErr = a.persist(a.config.PersistMode == "none")
	var closeErr error
	if a.store != nil {
		closeErr = a.store.close()
		a.store = nil
	}

	// Step 3: stop Ledger
	var ledgerErr = a.ledger.Stop()

	// Step 4: stop Account
	a.started = false
	a.stopped = true
	a.log.Info(fmt.Sprintf(
		"account stopped cycle=%d executor=%d account=%s orders=%d reconciles=%d recon_calls=%d recon_skipped_clean=%d recon_executed=%d recon_succeeded=%d recon_failed=%d",
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
	))
	return errors.Join(venueErr, storeErr, closeErr, ledgerErr)
}

// Section 2 - Domain Helpers

// Section 2.1 - Ledger Observation

// ActiveOrders returns current active local Order snapshots.
func (a *Account) ActiveOrders() []order.Order {
	return a.ledger.ActiveOrders()
}

// Trade returns focused current Trade state.
func (a *Account) Trade(tradeID uint64) (trade.Trade, error) {
	return a.ledger.Trade(tradeID)
}

// OpenTrades returns focused state for current open exposure.
func (a *Account) OpenTrades() []trade.Trade {
	return a.ledger.ActiveTrades()
}

// Section 2.2 - Order Preparation

func (a *Account) prepareOrderBatch(specs []OrderSpec) ([]OrderSpec, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("place Account Orders: Order set must not be empty")
	}
	var normalized = make([]OrderSpec, len(specs))
	var tradeID = specs[0].TradeID
	var minNotionalUSDC = decimal.NewFromInt(
		int64(a.config.Nuubot.App.Hyperliquid.MinOrderNotionalUSDC),
	)
	for index, spec := range specs {
		normalized[index] = spec
		if spec.Price != nil {
			var price = *spec.Price
			normalized[index].Price = &price
		}
		if spec.TriggerPrice != nil {
			var trigger = *spec.TriggerPrice
			normalized[index].TriggerPrice = &trigger
		}
		if spec.TradeID != tradeID || spec.TimestampMS != specs[0].TimestampMS ||
			spec.Quantity.IsNegative() || spec.Quantity.IsZero() ||
			spec.Price == nil || !spec.Price.IsPositive() || spec.TimestampMS == 0 {
			return nil, fmt.Errorf("place Account Orders: invalid or mixed batch")
		}
		var roundedPrice = a.config.Nuubot.Meta.RoundPrice(*spec.Price)
		normalized[index].Price = &roundedPrice
		if spec.TriggerPrice != nil {
			var trigger = a.config.Nuubot.Meta.RoundPrice(*spec.TriggerPrice)
			normalized[index].TriggerPrice = &trigger
		}
		normalized[index].Quantity = a.config.Nuubot.Meta.RoundSize(spec.Quantity)
		if !normalized[index].Quantity.IsPositive() {
			return nil, fmt.Errorf("place Account Orders: quantity rounds to zero")
		}
		var notional = normalized[index].Quantity.Mul(*normalized[index].Price)
		if notional.LessThan(minNotionalUSDC) {
			return nil, fmt.Errorf(
				"place Account Orders: Order %d notional is below minimum",
				index,
			)
		}
	}
	return normalized, nil
}

func (a *Account) markPrice() *decimal.Decimal {
	var bbo, found = a.config.Nuubot.MarketData.LatestBBO(market.Key{
		Venue:   a.config.Venue,
		Network: a.config.Network,
		Symbol:  a.config.Symbol,
	})
	if !found {
		return nil
	}
	var price = decimal.NewFromFloat(bbo.Price)
	return &price
}

func (a *Account) venueOrderRequest(
	spec OrderSpec,
	value string,
) (hyperliquid.OrderRequest, error) {
	var request = hyperliquid.OrderRequest{
		Asset:      int(a.config.Nuubot.Meta.AssetID),
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
