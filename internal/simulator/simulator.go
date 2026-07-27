// Package simulator owns Hyperliquid-shaped simulated Venue truth.
package simulator

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/shopspring/decimal"

	"nuubot/internal/hyperliquid"
	"nuubot/internal/market"
)

const (
	orderOpen     = "open"
	orderFilled   = "filled"
	orderCanceled = "canceled"

	kindLimit = "limit"
	kindTP    = "tp"
	kindSL    = "sl"
)

// Config contains one Simulator's official identity and policy.
type Config struct {
	Account     string
	Asset       int
	Symbol      string
	Equity      decimal.Decimal
	FeePct      decimal.Decimal
	SlippagePct decimal.Decimal
	PersistMode string
	Path        string
}

type comparisonKey struct {
	digits        string
	integerDigits int64
}

type simOrder struct {
	venueOrderID      uint64
	cloid             string
	asset             int
	symbol            string
	batchID           uint64
	kind              string
	isBuy             bool
	price             decimal.Decimal
	comparisonKey     comparisonKey
	quantity          decimal.Decimal
	reduceOnly        bool
	timeInForce       string
	triggerPrice      decimal.Decimal
	hasTriggerPrice   bool
	status            string
	armed             bool
	remainingQuantity decimal.Decimal
	filledQuantity    decimal.Decimal
	averageFillPrice  decimal.Decimal
	fees              decimal.Decimal
	timestampMS       uint64
}

type simFill struct {
	venueOrderID  uint64
	venueTID      uint64
	symbol        string
	isBuy         bool
	quantity      decimal.Decimal
	price         decimal.Decimal
	timestampMS   uint64
	startPosition decimal.Decimal
	closedPnL     decimal.Decimal
	direction     string
	fee           decimal.Decimal
	hasFee        bool
	liquidity     string
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
	nextBatchID      uint64
	orders           []*simOrder
	ordersByCLOID    map[string]*simOrder
	activeOrders     map[uint64]*simOrder
	fills            []simFill
	currentPosition  position
	store            *simulatorStore
	lastPrice        decimal.Decimal
	lastPriceKey     comparisonKey
	lastTimestampMS  uint64
	observedMS       uint64
	hasBBO           bool
	started          bool
	stopped          bool
}

type exchangeResponse struct {
	Status   string               `json:"status"`
	Response exchangeResponseBody `json:"response"`
}

type exchangeResponseBody struct {
	Type string               `json:"type"`
	Data exchangeResponseData `json:"data"`
}

type exchangeResponseData struct {
	Statuses []any `json:"statuses"`
}

type restingStatus struct {
	Resting restingStatusBody `json:"resting"`
}

type restingStatusBody struct {
	OID   uint64 `json:"oid"`
	CLOID string `json:"cloid"`
}

type filledStatus struct {
	Filled filledStatusBody `json:"filled"`
}

type filledStatusBody struct {
	TotalSize string `json:"totalSz"`
	AveragePx string `json:"avgPx"`
	OID       uint64 `json:"oid"`
	CLOID     string `json:"cloid"`
}

type openOrderResponse struct {
	Symbol      string `json:"coin"`
	LimitPrice  string `json:"limitPx"`
	OID         uint64 `json:"oid"`
	Side        string `json:"side"`
	Size        string `json:"sz"`
	TimestampMS uint64 `json:"timestamp"`
	CLOID       string `json:"cloid"`
}

type orderStatusResponse struct {
	Status string                  `json:"status"`
	Order  orderStatusResponseBody `json:"order"`
}

type orderStatusResponseBody struct {
	Order           orderStatusOrder `json:"order"`
	Status          string           `json:"status"`
	StatusTimestamp uint64           `json:"statusTimestamp"`
}

type orderStatusOrder struct {
	Symbol      string `json:"coin"`
	Side        string `json:"side"`
	LimitPrice  string `json:"limitPx"`
	Size        string `json:"sz"`
	OID         uint64 `json:"oid"`
	TimestampMS uint64 `json:"timestamp"`
	OriginalSz  string `json:"origSz"`
	CLOID       string `json:"cloid"`
}

type fillResponse struct {
	Symbol        string  `json:"coin"`
	Price         string  `json:"px"`
	Size          string  `json:"sz"`
	Side          string  `json:"side"`
	TimestampMS   uint64  `json:"time"`
	StartPosition string  `json:"startPosition"`
	Direction     string  `json:"dir"`
	ClosedPnL     string  `json:"closedPnl"`
	Hash          string  `json:"hash"`
	OID           uint64  `json:"oid"`
	Crossed       bool    `json:"crossed"`
	Fee           *string `json:"fee"`
	TID           uint64  `json:"tid"`
	FeeToken      string  `json:"feeToken"`
	BuilderFee    *string `json:"builderFee"`
}

type clearinghouseResponse struct {
	AssetPositions             []assetPositionResponse `json:"assetPositions"`
	CrossMaintenanceMarginUsed string                  `json:"crossMaintenanceMarginUsed"`
	CrossMarginSummary         marginSummaryResponse   `json:"crossMarginSummary"`
	MarginSummary              marginSummaryResponse   `json:"marginSummary"`
	Time                       uint64                  `json:"time"`
	Withdrawable               string                  `json:"withdrawable"`
}

type marginSummaryResponse struct {
	AccountValue    string `json:"accountValue"`
	TotalMarginUsed string `json:"totalMarginUsed"`
	TotalNtlPos     string `json:"totalNtlPos"`
	TotalRawUSD     string `json:"totalRawUsd"`
}

type assetPositionResponse struct {
	Position positionResponse `json:"position"`
	Type     string           `json:"type"`
}

type positionResponse struct {
	Symbol           string           `json:"coin"`
	Cumulative       fundingResponse  `json:"cumFunding"`
	EntryPrice       *string          `json:"entryPx"`
	Leverage         leverageResponse `json:"leverage"`
	LiquidationPrice *string          `json:"liquidationPx"`
	MarginUsed       string           `json:"marginUsed"`
	MaxLeverage      uint32           `json:"maxLeverage"`
	PositionValue    string           `json:"positionValue"`
	ReturnOnEquity   string           `json:"returnOnEquity"`
	Size             string           `json:"szi"`
	UnrealizedPnL    string           `json:"unrealizedPnl"`
}

type leverageResponse struct {
	Type  string `json:"type"`
	Value uint32 `json:"value"`
}

type fundingResponse struct {
	AllTime     string `json:"allTime"`
	SinceChange string `json:"sinceChange"`
	SinceOpen   string `json:"sinceOpen"`
}

// Section 1 - Program Flow

// Init prepares one Simulator.
func (s *Simulator) Init(cfg Config) error {
	// Step 1: validate Simulator config
	if s.started || s.stopped {
		return fmt.Errorf("initialize simulator: invalid lifecycle state")
	}
	if cfg.Account == "" || cfg.Asset < 0 || cfg.Symbol == "" {
		return fmt.Errorf("initialize simulator: complete official identity is required")
	}
	if !cfg.Equity.IsPositive() || cfg.FeePct.IsNegative() || cfg.SlippagePct.IsNegative() {
		return fmt.Errorf("initialize simulator: invalid equity, fee, or slippage")
	}
	if cfg.PersistMode != "none" && cfg.PersistMode != "max" {
		return fmt.Errorf("initialize simulator: invalid persistence mode %q", cfg.PersistMode)
	}
	if cfg.PersistMode == "max" && cfg.Path == "" {
		return fmt.Errorf("initialize simulator: max persistence requires path")
	}

	// Step 2: initialize Simulator state
	s.config = cfg
	s.nextVenueOrderID = 1
	s.nextVenueTID = 1
	s.nextBatchID = 1
	s.ordersByCLOID = make(map[string]*simOrder)
	s.activeOrders = make(map[uint64]*simOrder)

	// Step 3: restore durable Simulator state when configured
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
		} else {
			err = s.store.save(cfg, s.storedState())
		}
		if err != nil {
			s.store.close()
			return err
		}
	}

	// Step 4: mark Simulator started
	s.started = true
	return nil
}

// PlaceOrders admits one official Hyperliquid order action.
func (s *Simulator) PlaceOrders(
	action hyperliquid.PlaceOrderAction,
	timestampMS uint64,
) ([]byte, error) {
	if !s.started || s.stopped {
		return nil, fmt.Errorf("place simulator Orders: invalid lifecycle state")
	}
	if action.Type != "order" || len(action.Orders) == 0 ||
		len(action.Orders) > 1000 || timestampMS == 0 ||
		(action.Grouping != "na" && action.Grouping != "normalTpsl" &&
			action.Grouping != "positionTpsl") {
		return nil, fmt.Errorf("place simulator Orders: invalid action")
	}

	// Step 1: validate official Order action
	var seen = make(map[string]struct{}, len(action.Orders))
	for _, request := range action.Orders {
		if err := s.validateRequest(request); err != nil {
			return nil, err
		}
		if _, exists := seen[request.CLOID]; exists ||
			s.ordersByCLOID[request.CLOID] != nil {
			return nil, fmt.Errorf("place simulator Orders: duplicate cloid %s", request.CLOID)
		}
		seen[request.CLOID] = struct{}{}
	}

	// Step 2: stage Order mutation
	var staged = s.stage()
	staged.observedMS = max(staged.observedMS, timestampMS)
	var batchID = staged.nextBatchID
	staged.nextBatchID++
	var hasEntry = false
	for _, request := range action.Orders {
		if request.Type.Limit != nil {
			hasEntry = true
		}
	}
	var added = make([]*simOrder, 0, len(action.Orders))
	for _, request := range action.Orders {
		var row, err = staged.newOrder(request, batchID, timestampMS)
		if err != nil {
			return nil, err
		}
		if action.Grouping == "normalTpsl" && hasEntry && row.kind != kindLimit {
			row.armed = false
		}
		staged.nextVenueOrderID++
		staged.orders = append(staged.orders, row)
		staged.ordersByCLOID[row.cloid] = row
		staged.activeOrders[row.venueOrderID] = row
		added = append(added, row)
	}

	// Step 3: match marketable Orders
	if staged.hasBBO {
		staged.matchAdded(added, staged.lastTimestampMS)
	}

	// Step 4: persist and publish Order mutation
	if err := staged.persist(); err != nil {
		return nil, err
	}
	s.commit(staged)

	// Step 5: return official Order response
	var statuses = make([]any, 0, len(added))
	for _, original := range added {
		var row = s.ordersByCLOID[original.cloid]
		if row.status == orderFilled {
			statuses = append(statuses, filledStatus{Filled: filledStatusBody{
				TotalSize: row.filledQuantity.String(),
				AveragePx: row.averageFillPrice.String(),
				OID:       row.venueOrderID,
				CLOID:     row.cloid,
			}})
		} else {
			statuses = append(statuses, restingStatus{
				Resting: restingStatusBody{
					OID:   row.venueOrderID,
					CLOID: row.cloid,
				},
			})
		}
	}
	return hyperliquid.Encode(exchangeResponse{
		Status: "ok",
		Response: exchangeResponseBody{
			Type: "order",
			Data: exchangeResponseData{Statuses: statuses},
		},
	})
}

// CancelOrders applies one official Hyperliquid cancel-by-CLOID action.
func (s *Simulator) CancelOrders(
	action hyperliquid.CancelByCLOIDAction,
	timestampMS uint64,
) ([]byte, error) {
	if !s.started || s.stopped {
		return nil, fmt.Errorf("cancel simulator Orders: invalid lifecycle state")
	}
	if action.Type != "cancelByCloid" || len(action.Cancels) == 0 || timestampMS == 0 {
		return nil, fmt.Errorf("cancel simulator Orders: invalid action")
	}

	// Step 1: validate official cancel action
	var seen = make(map[string]struct{}, len(action.Cancels))
	for _, cancel := range action.Cancels {
		var row = s.ordersByCLOID[cancel.CLOID]
		if cancel.Asset != s.config.Asset || row == nil || row.status != orderOpen {
			return nil, fmt.Errorf("cancel simulator Orders: unknown active cloid %s", cancel.CLOID)
		}
		if _, duplicate := seen[cancel.CLOID]; duplicate {
			return nil, fmt.Errorf("cancel simulator Orders: duplicate cloid %s", cancel.CLOID)
		}
		seen[cancel.CLOID] = struct{}{}
	}

	// Step 2: stage cancel mutation
	var staged = s.stage()
	staged.observedMS = max(staged.observedMS, timestampMS)
	for _, cancel := range action.Cancels {
		var row = staged.ordersByCLOID[cancel.CLOID]
		staged.cancel(row, timestampMS)
		if row.kind == kindLimit {
			staged.cancelChildren(row.batchID, timestampMS)
		}
	}

	// Step 3: persist and publish cancel mutation
	if err := staged.persist(); err != nil {
		return nil, err
	}
	s.commit(staged)

	// Step 4: return official cancel response
	var statuses = make([]any, len(action.Cancels))
	for index := range statuses {
		statuses[index] = "success"
	}
	return hyperliquid.Encode(exchangeResponse{
		Status: "ok",
		Response: exchangeResponseBody{
			Type: "cancel",
			Data: exchangeResponseData{Statuses: statuses},
		},
	})
}

// IngestBBO advances simulated exchange matching from one BBO.
func (s *Simulator) IngestBBO(bbo market.BBO) (bool, error) {
	if !s.started || s.stopped {
		return false, fmt.Errorf("ingest simulator BBO: invalid lifecycle state")
	}
	if bbo.TimestampMS == 0 || bbo.Price <= 0 ||
		(s.hasBBO && bbo.TimestampMS < s.lastTimestampMS) {
		return false, fmt.Errorf("ingest simulator BBO: invalid timestamp or price")
	}

	// Step 1: validate and normalize BBO
	var price = decimal.NewFromFloat(bbo.Price)
	var priceKey = newComparisonKey(price)

	// Step 2: warm initial BBO state
	if !s.hasBBO {
		s.lastPrice = price
		s.lastPriceKey = priceKey
		s.lastTimestampMS = bbo.TimestampMS
		s.observedMS = max(s.observedMS, bbo.TimestampMS)
		s.hasBBO = true
		return false, nil
	}

	// Step 3: stage BBO matching
	var staged = s.stage()
	var changed = staged.match(priceKey, bbo.TimestampMS)
	staged.lastPrice = price
	staged.lastPriceKey = priceKey
	staged.lastTimestampMS = bbo.TimestampMS
	staged.observedMS = max(staged.observedMS, bbo.TimestampMS)

	// Step 4: persist changed Venue truth
	if changed {
		if err := staged.persist(); err != nil {
			return false, err
		}
	}

	// Step 5: publish BBO outcome
	s.commit(staged)
	return changed, nil
}

// Stop stops Simulator admission.
func (s *Simulator) Stop() error {
	// Step 1: ignore repeated stop
	if s.stopped {
		return nil
	}

	// Step 2: persist Simulator state
	if err := s.persist(); err != nil {
		return err
	}

	// Step 3: close Simulator store
	if s.store != nil {
		if err := s.store.close(); err != nil {
			return err
		}
		s.store = nil
	}

	// Step 4: mark Simulator stopped
	s.started = false
	s.stopped = true
	return nil
}

// Section 2 - Domain Helpers

// Section 2.1 - Official Queries

// OpenOrders returns fresh detached official Hyperliquid JSON.
func (s *Simulator) OpenOrders(account string) ([]byte, error) {
	if err := s.validateAccount(account); err != nil {
		return nil, err
	}
	var active = s.sortedActiveOrders()
	var rows = make([]openOrderResponse, 0, len(active))
	for _, row := range active {
		rows = append(rows, openOrderResponse{
			Symbol:      row.symbol,
			LimitPrice:  orderPrice(row).String(),
			OID:         row.venueOrderID,
			Side:        sideCode(row.isBuy),
			Size:        row.remainingQuantity.String(),
			TimestampMS: row.timestampMS,
			CLOID:       row.cloid,
		})
	}
	return hyperliquid.Encode(rows)
}

// Fills returns fresh detached official Hyperliquid JSON for one inclusive range.
func (s *Simulator) Fills(account string, startMS uint64, endMS uint64) ([]byte, error) {
	if err := s.validateAccount(account); err != nil {
		return nil, err
	}
	if endMS < startMS {
		return nil, fmt.Errorf("read simulator Fills: invalid range")
	}
	var rows = make([]fillResponse, 0)
	for _, execution := range s.fills {
		if execution.timestampMS < startMS || execution.timestampMS > endMS {
			continue
		}
		var fee *string
		if execution.hasFee {
			var value = execution.fee.String()
			fee = &value
		}
		rows = append(rows, fillResponse{
			Symbol:        execution.symbol,
			Price:         execution.price.String(),
			Size:          execution.quantity.String(),
			Side:          sideCode(execution.isBuy),
			TimestampMS:   execution.timestampMS,
			StartPosition: execution.startPosition.String(),
			Direction:     execution.direction,
			ClosedPnL:     execution.closedPnL.String(),
			Hash:          fmt.Sprintf("0x%064x", execution.venueTID),
			OID:           execution.venueOrderID,
			Crossed:       true,
			Fee:           fee,
			TID:           execution.venueTID,
			FeeToken:      "USDC",
		})
	}
	return hyperliquid.Encode(rows)
}

// OrderStatus returns fresh detached official Hyperliquid JSON for one CLOID.
func (s *Simulator) OrderStatus(account string, value string) ([]byte, error) {
	if err := s.validateAccount(account); err != nil {
		return nil, err
	}
	var row = s.ordersByCLOID[value]
	if row == nil {
		return hyperliquid.Encode(struct {
			Status string `json:"status"`
		}{Status: "unknownOid"})
	}
	return hyperliquid.Encode(orderStatusResponse{
		Status: "order",
		Order: orderStatusResponseBody{
			Order: orderStatusOrder{
				Symbol:      row.symbol,
				Side:        sideCode(row.isBuy),
				LimitPrice:  orderPrice(row).String(),
				Size:        row.quantity.String(),
				OID:         row.venueOrderID,
				TimestampMS: row.timestampMS,
				OriginalSz:  row.quantity.String(),
				CLOID:       row.cloid,
			},
			Status:          row.status,
			StatusTimestamp: row.timestampMS,
		},
	})
}

// AccountState returns fresh detached official Hyperliquid clearinghouse JSON.
func (s *Simulator) AccountState(account string) ([]byte, error) {
	if err := s.validateAccount(account); err != nil {
		return nil, err
	}
	var current = s.currentPosition
	if !current.size.IsZero() && !s.hasBBO {
		return nil, fmt.Errorf("read simulator Account state: fresh BBO is required")
	}
	var unrealized = decimal.Zero
	var positionValue = decimal.Zero
	if !current.size.IsZero() {
		unrealized = s.lastPrice.Sub(current.entryPrice).Mul(current.size)
		positionValue = current.size.Abs().Mul(s.lastPrice)
	}
	var accountValue = s.config.Equity.Add(current.realized).Add(unrealized).Sub(current.fees)
	var margin = positionValue.Div(decimal.NewFromInt(5))
	var summary = marginSummaryResponse{
		AccountValue:    accountValue.String(),
		TotalMarginUsed: margin.String(),
		TotalNtlPos:     positionValue.String(),
		TotalRawUSD:     accountValue.String(),
	}
	var positions = make([]assetPositionResponse, 0, 1)
	if !current.size.IsZero() {
		var entryPrice = current.entryPrice.String()
		positions = append(positions, assetPositionResponse{
			Type: "oneWay",
			Position: positionResponse{
				Symbol:           s.config.Symbol,
				Cumulative:       fundingResponse{"0", "0", "0"},
				EntryPrice:       &entryPrice,
				Leverage:         leverageResponse{Type: "cross", Value: 5},
				MarginUsed:       margin.String(),
				MaxLeverage:      5,
				PositionValue:    positionValue.String(),
				ReturnOnEquity:   unrealized.Div(margin).String(),
				Size:             current.size.String(),
				UnrealizedPnL:    unrealized.String(),
				LiquidationPrice: nil,
			},
		})
	}
	return hyperliquid.Encode(clearinghouseResponse{
		AssetPositions:             positions,
		CrossMaintenanceMarginUsed: "0",
		CrossMarginSummary:         summary,
		MarginSummary:              summary,
		Time:                       s.observedMS,
		Withdrawable:               accountValue.Sub(margin).String(),
	})
}

// SetFillFeeAvailableForTest controls delayed-fee evidence in focused tests.
func (s *Simulator) SetFillFeeAvailableForTest(venueTID uint64, available bool) error {
	for index := range s.fills {
		if s.fills[index].venueTID == venueTID {
			s.fills[index].hasFee = available
			return nil
		}
	}
	return fmt.Errorf("set simulator Fill fee availability: unknown Venue TID %d", venueTID)
}

// Section 2.2 - Validation and Construction

func (s *Simulator) validateAccount(account string) error {
	if !s.started || s.stopped {
		return fmt.Errorf("query simulator: invalid lifecycle state")
	}
	if account != s.config.Account {
		return fmt.Errorf("query simulator: unknown account %q", account)
	}
	return nil
}

func (s *Simulator) validateRequest(request hyperliquid.OrderRequest) error {
	if request.Asset != s.config.Asset || !validCLOID(request.CLOID) {
		return fmt.Errorf("place simulator Orders: invalid official identity")
	}
	var price, err = decimal.NewFromString(request.Price)
	if err != nil || !price.IsPositive() {
		return fmt.Errorf("place simulator Orders: positive price is required")
	}
	var quantity decimal.Decimal
	quantity, err = decimal.NewFromString(request.Size)
	if err != nil || !quantity.IsPositive() {
		return fmt.Errorf("place simulator Orders: quantity must be positive")
	}
	if (request.Type.Limit == nil) == (request.Type.Trigger == nil) {
		return fmt.Errorf("place simulator Orders: exactly one order type is required")
	}
	if request.Type.Limit != nil &&
		request.Type.Limit.TimeInForce != "Gtc" &&
		request.Type.Limit.TimeInForce != "Ioc" {
		return fmt.Errorf("place simulator Orders: unsupported time in force")
	}
	if request.Type.Trigger != nil {
		var trigger decimal.Decimal
		trigger, err = decimal.NewFromString(request.Type.Trigger.TriggerPrice)
		if err != nil || !trigger.IsPositive() ||
			(request.Type.Trigger.TPSL != kindTP &&
				request.Type.Trigger.TPSL != kindSL) {
			return fmt.Errorf("place simulator Orders: invalid trigger type")
		}
	}
	return nil
}

func (s *Simulator) newOrder(
	request hyperliquid.OrderRequest,
	batchID uint64,
	timestampMS uint64,
) (*simOrder, error) {
	var price, err = decimal.NewFromString(request.Price)
	if err != nil {
		return nil, err
	}
	var quantity decimal.Decimal
	quantity, err = decimal.NewFromString(request.Size)
	if err != nil {
		return nil, err
	}
	var row = &simOrder{
		venueOrderID:      s.nextVenueOrderID,
		cloid:             request.CLOID,
		asset:             request.Asset,
		symbol:            s.config.Symbol,
		batchID:           batchID,
		kind:              kindLimit,
		isBuy:             request.IsBuy,
		price:             price,
		quantity:          quantity,
		reduceOnly:        request.ReduceOnly,
		status:            orderOpen,
		armed:             true,
		remainingQuantity: quantity,
		timestampMS:       timestampMS,
	}
	if request.Type.Limit != nil {
		row.timeInForce = request.Type.Limit.TimeInForce
	} else {
		row.kind = request.Type.Trigger.TPSL
		row.timeInForce = "Gtc"
		row.hasTriggerPrice = true
		row.triggerPrice, err = decimal.NewFromString(request.Type.Trigger.TriggerPrice)
		if err != nil {
			return nil, err
		}
	}
	row.comparisonKey = newComparisonKey(orderPrice(row))
	return row, nil
}

// Section 2.3 - Matching and Position

func (s *Simulator) match(price comparisonKey, timestampMS uint64) bool {
	var changed = false
	var filledBatches = make(map[uint64]struct{})
	for {
		var matched *simOrder
		for _, row := range s.sortedActiveOrders() {
			if !row.armed || row.timestampMS >= timestampMS {
				continue
			}
			if _, filled := filledBatches[row.batchID]; filled {
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
		} else {
			s.fill(matched, quantity, max(timestampMS, matched.timestampMS))
			filledBatches[matched.batchID] = struct{}{}
		}
		changed = true
	}
	return changed
}

func (s *Simulator) matchAdded(added []*simOrder, timestampMS uint64) {
	var filledBatches = make(map[uint64]struct{})
	for _, row := range added {
		if !row.armed {
			continue
		}
		if row.timeInForce != "Ioc" && !crosses(row, s.lastPriceKey) {
			continue
		}
		if _, filled := filledBatches[row.batchID]; filled {
			continue
		}
		var quantity, executable = s.executableQuantity(row)
		if !executable {
			s.cancel(row, max(timestampMS, row.timestampMS))
			continue
		}
		s.fill(row, quantity, max(timestampMS, row.timestampMS))
		filledBatches[row.batchID] = struct{}{}
	}
}

func (s *Simulator) fill(row *simOrder, quantity decimal.Decimal, timestampMS uint64) {
	var basis = orderPrice(row)
	var rate = s.config.SlippagePct.Div(decimal.NewFromInt(100))
	var price = basis.Mul(decimal.NewFromInt(1).Add(rate))
	if !row.isBuy {
		price = basis.Mul(decimal.NewFromInt(1).Sub(rate))
	}
	var before = s.currentPosition
	var fee = quantity.Mul(price).Mul(s.config.FeePct).Div(decimal.NewFromInt(100))
	var closedPnL = closePnL(before, row.isBuy, quantity, price)
	s.fills = append(s.fills, simFill{
		venueOrderID:  row.venueOrderID,
		venueTID:      s.nextVenueTID,
		symbol:        row.symbol,
		isBuy:         row.isBuy,
		quantity:      quantity,
		price:         price,
		timestampMS:   timestampMS,
		startPosition: before.size,
		closedPnL:     closedPnL,
		direction:     fillDirection(before.size, row.isBuy),
		fee:           fee,
		hasFee:        true,
		liquidity:     "taker",
	})
	s.nextVenueTID++

	var delta = quantity
	if !row.isBuy {
		delta = delta.Neg()
	}
	var current = before
	current.fees = current.fees.Add(fee)
	if before.size.IsZero() || sameSign(before.size, delta) {
		var total = before.size.Abs().Add(quantity)
		current.entryPrice = before.size.Abs().Mul(before.entryPrice).
			Add(quantity.Mul(price)).
			Div(total)
		current.size = before.size.Add(delta)
	} else {
		current.realized = current.realized.Add(closedPnL)
		current.size = before.size.Add(delta)
		if current.size.IsZero() {
			current.entryPrice = decimal.Zero
		} else if !sameSign(before.size, current.size) {
			current.entryPrice = price
		}
	}
	s.currentPosition = current
	delete(s.activeOrders, row.venueOrderID)
	row.status = orderFilled
	row.armed = false
	row.filledQuantity = quantity
	row.remainingQuantity = decimal.Zero
	row.averageFillPrice = price
	row.fees = fee
	row.timestampMS = timestampMS
	if row.kind == kindLimit {
		s.armChildren(row.batchID, timestampMS)
	} else {
		s.cancelChildren(row.batchID, timestampMS)
	}
}

func (s *Simulator) cancel(row *simOrder, timestampMS uint64) {
	delete(s.activeOrders, row.venueOrderID)
	row.status = orderCanceled
	row.armed = false
	row.timestampMS = timestampMS
}

func (s *Simulator) armChildren(batchID uint64, timestampMS uint64) {
	for _, child := range s.activeOrders {
		if child.batchID == batchID && child.kind != kindLimit {
			child.armed = true
			child.timestampMS = timestampMS
		}
	}
}

func (s *Simulator) cancelChildren(batchID uint64, timestampMS uint64) {
	for _, child := range s.sortedActiveOrders() {
		if child.batchID == batchID && child.kind != kindLimit {
			s.cancel(child, timestampMS)
		}
	}
}

func (s *Simulator) executableQuantity(row *simOrder) (decimal.Decimal, bool) {
	if !row.reduceOnly {
		return row.remainingQuantity, true
	}
	var available = s.currentPosition.size
	if row.isBuy {
		available = available.Neg()
	}
	if !available.IsPositive() {
		return decimal.Zero, false
	}
	return decimal.Min(row.remainingQuantity, available), true
}

func (s *Simulator) sortedActiveOrders() []*simOrder {
	var rows = make([]*simOrder, 0, len(s.activeOrders))
	for _, row := range s.activeOrders {
		rows = append(rows, row)
	}
	sort.Slice(rows, func(left int, right int) bool {
		return rows[left].venueOrderID < rows[right].venueOrderID
	})
	return rows
}

func (s *Simulator) position() (position, error) {
	var result position
	for index, execution := range s.fills {
		if execution.venueOrderID == 0 || execution.venueTID != uint64(index+1) ||
			execution.symbol != s.config.Symbol || !execution.quantity.IsPositive() ||
			!execution.price.IsPositive() || execution.timestampMS == 0 {
			return position{}, fmt.Errorf("load Simulator: invalid Fill %d", index+1)
		}
		var expectedClosed = closePnL(
			result,
			execution.isBuy,
			execution.quantity,
			execution.price,
		)
		if !execution.startPosition.Equal(result.size) ||
			!execution.closedPnL.Equal(expectedClosed) ||
			execution.direction != fillDirection(result.size, execution.isBuy) {
			return position{}, fmt.Errorf("load Simulator: invalid derived Fill %d", index+1)
		}
		var delta = execution.quantity
		if !execution.isBuy {
			delta = delta.Neg()
		}
		result.fees = result.fees.Add(execution.fee)
		if result.size.IsZero() || sameSign(result.size, delta) {
			var total = result.size.Abs().Add(execution.quantity)
			result.entryPrice = result.size.Abs().Mul(result.entryPrice).
				Add(execution.quantity.Mul(execution.price)).
				Div(total)
			result.size = result.size.Add(delta)
			continue
		}
		result.realized = result.realized.Add(expectedClosed)
		var before = result.size
		result.size = result.size.Add(delta)
		if result.size.IsZero() {
			result.entryPrice = decimal.Zero
		} else if !sameSign(before, result.size) {
			result.entryPrice = execution.price
		}
	}
	if s.nextVenueTID != uint64(len(s.fills))+1 {
		return position{}, fmt.Errorf("load Simulator: invalid Fill counter")
	}
	return result, nil
}

// Section 2.4 - Persistence State

func (s *Simulator) persist() error {
	if s.config.PersistMode == "none" {
		return nil
	}
	return s.store.save(s.config, s.storedState())
}

func (s *Simulator) storedState() storedState {
	var state = storedState{
		SchemaVersion:    simulatorSchemaVersion,
		Account:          s.config.Account,
		Asset:            s.config.Asset,
		Symbol:           s.config.Symbol,
		Equity:           s.config.Equity.String(),
		FeePct:           s.config.FeePct.String(),
		SlippagePct:      s.config.SlippagePct.String(),
		NextVenueOrderID: s.nextVenueOrderID,
		NextVenueTID:     s.nextVenueTID,
		NextBatchID:      s.nextBatchID,
		ObservedMS:       s.observedMS,
	}
	for _, row := range s.orders {
		state.Orders = append(state.Orders, storedOrder{
			VenueOrderID:      row.venueOrderID,
			CLOID:             row.cloid,
			Asset:             row.asset,
			Symbol:            row.symbol,
			BatchID:           row.batchID,
			Kind:              row.kind,
			IsBuy:             row.isBuy,
			Price:             row.price.String(),
			Quantity:          row.quantity.String(),
			ReduceOnly:        row.reduceOnly,
			TimeInForce:       row.timeInForce,
			TriggerPrice:      row.triggerPrice.String(),
			HasTriggerPrice:   row.hasTriggerPrice,
			Status:            row.status,
			Armed:             row.armed,
			RemainingQuantity: row.remainingQuantity.String(),
			FilledQuantity:    row.filledQuantity.String(),
			AverageFillPrice:  row.averageFillPrice.String(),
			Fees:              row.fees.String(),
			TimestampMS:       row.timestampMS,
		})
	}
	for _, execution := range s.fills {
		state.Fills = append(state.Fills, storedFill{
			VenueOrderID:  execution.venueOrderID,
			VenueTID:      execution.venueTID,
			Symbol:        execution.symbol,
			IsBuy:         execution.isBuy,
			Quantity:      execution.quantity.String(),
			Price:         execution.price.String(),
			TimestampMS:   execution.timestampMS,
			StartPosition: execution.startPosition.String(),
			ClosedPnL:     execution.closedPnL.String(),
			Direction:     execution.direction,
			Fee:           execution.fee.String(),
			HasFee:        execution.hasFee,
			Liquidity:     execution.liquidity,
		})
	}
	return state
}

func (s *Simulator) stage() *Simulator {
	if s.config.PersistMode != "max" {
		return s
	}
	var staged = *s
	staged.orders = make([]*simOrder, 0, len(s.orders))
	staged.ordersByCLOID = make(map[string]*simOrder, len(s.ordersByCLOID))
	staged.activeOrders = make(map[uint64]*simOrder, len(s.activeOrders))
	for _, row := range s.orders {
		var copied = *row
		staged.orders = append(staged.orders, &copied)
		staged.ordersByCLOID[copied.cloid] = &copied
		if copied.status == orderOpen {
			staged.activeOrders[copied.venueOrderID] = &copied
		}
	}
	staged.fills = append([]simFill(nil), s.fills...)
	return &staged
}

func (s *Simulator) commit(staged *Simulator) {
	if staged != s {
		*s = *staged
	}
}

func (s *Simulator) restore(state storedState) error {
	if state.NextVenueOrderID == 0 || state.NextVenueTID == 0 || state.NextBatchID == 0 {
		return fmt.Errorf("load Simulator: invalid counters")
	}
	var orders = make([]*simOrder, 0, len(state.Orders))
	var byCLOID = make(map[string]*simOrder, len(state.Orders))
	var active = make(map[uint64]*simOrder)
	var maxBatchID uint64
	for index, stored := range state.Orders {
		var row, err = restoreOrder(stored)
		if err != nil {
			return fmt.Errorf("load Simulator: invalid Order %d: %v", index+1, err)
		}
		if row.venueOrderID != uint64(index+1) ||
			row.asset != s.config.Asset ||
			row.symbol != s.config.Symbol ||
			byCLOID[row.cloid] != nil {
			return fmt.Errorf("load Simulator: invalid Order identity %d", index+1)
		}
		maxBatchID = max(maxBatchID, row.batchID)
		orders = append(orders, row)
		byCLOID[row.cloid] = row
		if row.status == orderOpen {
			active[row.venueOrderID] = row
		}
	}
	if state.NextVenueOrderID != uint64(len(orders))+1 ||
		state.NextBatchID <= maxBatchID {
		return fmt.Errorf("load Simulator: invalid Order counters")
	}
	var fills = make([]simFill, 0, len(state.Fills))
	for _, stored := range state.Fills {
		var execution, err = restoreFill(stored)
		if err != nil {
			return fmt.Errorf("load Simulator: invalid Fill: %v", err)
		}
		if execution.symbol != s.config.Symbol ||
			execution.venueOrderID == 0 ||
			execution.venueOrderID > uint64(len(orders)) ||
			orders[execution.venueOrderID-1].status != orderFilled {
			return fmt.Errorf("load Simulator: invalid Fill identity")
		}
		fills = append(fills, execution)
	}
	s.nextVenueOrderID = state.NextVenueOrderID
	s.nextVenueTID = state.NextVenueTID
	s.nextBatchID = state.NextBatchID
	s.observedMS = state.ObservedMS
	s.orders = orders
	s.ordersByCLOID = byCLOID
	s.activeOrders = active
	s.fills = fills
	var err error
	s.currentPosition, err = s.position()
	return err
}

// Section 3 - Generic Helpers

func restoreOrder(stored storedOrder) (*simOrder, error) {
	var price, err = decimal.NewFromString(stored.Price)
	if err != nil {
		return nil, err
	}
	var quantity decimal.Decimal
	quantity, err = decimal.NewFromString(stored.Quantity)
	if err != nil {
		return nil, err
	}
	var trigger decimal.Decimal
	trigger, err = decimal.NewFromString(stored.TriggerPrice)
	if err != nil {
		return nil, err
	}
	var remaining decimal.Decimal
	remaining, err = decimal.NewFromString(stored.RemainingQuantity)
	if err != nil {
		return nil, err
	}
	var filled decimal.Decimal
	filled, err = decimal.NewFromString(stored.FilledQuantity)
	if err != nil {
		return nil, err
	}
	var average decimal.Decimal
	average, err = decimal.NewFromString(stored.AverageFillPrice)
	if err != nil {
		return nil, err
	}
	var fees decimal.Decimal
	fees, err = decimal.NewFromString(stored.Fees)
	if err != nil {
		return nil, err
	}
	if stored.CLOID == "" || stored.Symbol == "" || stored.BatchID == 0 ||
		!price.IsPositive() || !quantity.IsPositive() || stored.TimestampMS == 0 ||
		(stored.Kind != kindLimit && stored.Kind != kindTP && stored.Kind != kindSL) ||
		(stored.Kind == kindLimit && stored.HasTriggerPrice) ||
		(stored.Kind != kindLimit && !stored.HasTriggerPrice) ||
		(stored.Status != orderOpen && stored.Status != orderFilled &&
			stored.Status != orderCanceled) {
		return nil, fmt.Errorf("invalid fields")
	}
	var row = &simOrder{
		venueOrderID:      stored.VenueOrderID,
		cloid:             stored.CLOID,
		asset:             stored.Asset,
		symbol:            stored.Symbol,
		batchID:           stored.BatchID,
		kind:              stored.Kind,
		isBuy:             stored.IsBuy,
		price:             price,
		quantity:          quantity,
		reduceOnly:        stored.ReduceOnly,
		timeInForce:       stored.TimeInForce,
		triggerPrice:      trigger,
		hasTriggerPrice:   stored.HasTriggerPrice,
		status:            stored.Status,
		armed:             stored.Armed,
		remainingQuantity: remaining,
		filledQuantity:    filled,
		averageFillPrice:  average,
		fees:              fees,
		timestampMS:       stored.TimestampMS,
	}
	row.comparisonKey = newComparisonKey(orderPrice(row))
	return row, nil
}

func restoreFill(stored storedFill) (simFill, error) {
	var quantity, err = decimal.NewFromString(stored.Quantity)
	if err != nil {
		return simFill{}, err
	}
	var price decimal.Decimal
	price, err = decimal.NewFromString(stored.Price)
	if err != nil {
		return simFill{}, err
	}
	var start decimal.Decimal
	start, err = decimal.NewFromString(stored.StartPosition)
	if err != nil {
		return simFill{}, err
	}
	var closed decimal.Decimal
	closed, err = decimal.NewFromString(stored.ClosedPnL)
	if err != nil {
		return simFill{}, err
	}
	var fee decimal.Decimal
	fee, err = decimal.NewFromString(stored.Fee)
	if err != nil {
		return simFill{}, err
	}
	return simFill{
		venueOrderID:  stored.VenueOrderID,
		venueTID:      stored.VenueTID,
		symbol:        stored.Symbol,
		isBuy:         stored.IsBuy,
		quantity:      quantity,
		price:         price,
		timestampMS:   stored.TimestampMS,
		startPosition: start,
		closedPnL:     closed,
		direction:     stored.Direction,
		fee:           fee,
		hasFee:        stored.HasFee,
		liquidity:     stored.Liquidity,
	}, nil
}

func orderPrice(row *simOrder) decimal.Decimal {
	if row.hasTriggerPrice {
		return row.triggerPrice
	}
	return row.price
}

func crosses(row *simOrder, price comparisonKey) bool {
	if row.timeInForce == "Ioc" {
		return true
	}
	var comparison = compareComparisonKeys(price, row.comparisonKey)
	switch row.kind {
	case kindTP:
		if !row.isBuy {
			return comparison >= 0
		}
		return comparison <= 0
	case kindSL:
		if !row.isBuy {
			return comparison <= 0
		}
		return comparison >= 0
	default:
		if row.isBuy {
			return comparison <= 0
		}
		return comparison >= 0
	}
}

func newComparisonKey(value decimal.Decimal) comparisonKey {
	var digits = value.Coefficient().String()
	return comparisonKey{
		digits:        digits,
		integerDigits: int64(len(digits)) + int64(value.Exponent()),
	}
}

func compareComparisonKeys(left comparisonKey, right comparisonKey) int {
	if left.integerDigits < right.integerDigits {
		return -1
	}
	if left.integerDigits > right.integerDigits {
		return 1
	}
	var length = max(len(left.digits), len(right.digits))
	for index := 0; index < length; index++ {
		var leftDigit byte = '0'
		if index < len(left.digits) {
			leftDigit = left.digits[index]
		}
		var rightDigit byte = '0'
		if index < len(right.digits) {
			rightDigit = right.digits[index]
		}
		if leftDigit < rightDigit {
			return -1
		}
		if leftDigit > rightDigit {
			return 1
		}
	}
	return 0
}

func closePnL(
	current position,
	isBuy bool,
	quantity decimal.Decimal,
	price decimal.Decimal,
) decimal.Decimal {
	if current.size.IsPositive() && !isBuy {
		return price.Sub(current.entryPrice).Mul(decimal.Min(current.size.Abs(), quantity))
	}
	if current.size.IsNegative() && isBuy {
		return current.entryPrice.Sub(price).Mul(decimal.Min(current.size.Abs(), quantity))
	}
	return decimal.Zero
}

func fillDirection(size decimal.Decimal, isBuy bool) string {
	if isBuy {
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

func sideCode(isBuy bool) string {
	if isBuy {
		return "B"
	}
	return "A"
}

func sameSign(left decimal.Decimal, right decimal.Decimal) bool {
	return left.IsPositive() && right.IsPositive() ||
		left.IsNegative() && right.IsNegative()
}

func validCLOID(value string) bool {
	if len(value) != 34 || !strings.HasPrefix(value, "0x") {
		return false
	}
	var _, err = hex.DecodeString(value[2:])
	return err == nil
}
