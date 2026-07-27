package ledger

import (
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/order"
	"nuubot/internal/trade"
)

// ReconAttempt owns one unpublished canonical Ledger update.
type ReconAttempt struct {
	input          ReconInput
	trades         map[uint64]*trade.Trade
	touchedTrades  map[uint64]struct{}
	touchedOrders  map[uint64]struct{}
	touchedFills   map[uint64]struct{}
	tradeSummaries map[uint64]trade.Summary
	fillChanges    []FillChange
}

// Section 1 - Program Flow

// Init prepares one empty or restored Ledger.
func (l *Ledger) Init(cfg Config) error {
	// Step 1: bind Ledger inputs
	var err = l.bindInputs(cfg)
	if err != nil {
		return err
	}

	// Step 2: validate persistence mode
	err = l.validateConfig()
	if err != nil {
		return err
	}

	// Step 3: initialize Ledger
	l.initializeState()

	// Step 4: open Ledger identity when configured
	err = l.openState()
	if err != nil {
		return err
	}

	// Step 5: load Ledger evidence when configured
	err = l.loadState()
	if err != nil {
		l.closeStore()
		return err
	}

	// Step 6: rebuild indexes and cached Summary
	err = l.validateAndIndexState()
	if err != nil {
		l.closeStore()
		return err
	}
	l.started = true
	return nil
}

// Recon atomically applies one complete normalized Venue observation.
func (l *Ledger) Recon(input ReconInput) error {
	// Step 1: prepare Recon attempt
	var attempt, err = l.PrepareRecon(input)
	if err != nil {
		return err
	}

	// Step 2: update Fill records
	err = l.UpdateReconFills(attempt, input.Fills)
	if err != nil {
		return err
	}

	// Step 3: update Order records
	err = l.UpdateReconOrders(attempt, input.Orders)
	if err != nil {
		return err
	}

	// Step 4: update Trade records
	err = l.UpdateReconTrades(attempt, nil)
	if err != nil {
		return err
	}

	// Step 5: commit Recon attempt
	return l.CommitRecon(attempt)
}

// Section 2 - Domain Helpers

// Section 2.1 - Initialization

func (l *Ledger) bindInputs(cfg Config) error {
	if l.started || l.stopped {
		return fmt.Errorf("initialize ledger: invalid lifecycle state")
	}
	l.config = cfg
	return nil
}

func (l *Ledger) validateConfig() error {
	var cfg = l.config
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
	return nil
}

func (l *Ledger) initializeState() {
	l.trades = make(map[uint64]*trade.Trade)
	l.orders = make(map[uint64]orderLocation)
	l.cloids = make(map[string]orderLocation)
	l.venueOrders = make(map[uint64]orderLocation)
	l.fills = make(map[uint64]fillLocation)
	l.activeTrades = make(map[uint64]struct{})
	l.activeOrders = make(map[uint64]struct{})
	l.pendingOrders = make(map[uint64]struct{})
	l.pendingFills = make(map[uint64]struct{})
	l.nextTradeID = 1
	l.nextTradeNo = 1
	l.nextOrderID = 1
}

func (l *Ledger) openState() error {
	if l.config.PersistMode == None {
		return nil
	}
	var store, err = openLedgerStore(l.config.Path)
	if err != nil {
		return err
	}
	l.store = store
	return nil
}

func (l *Ledger) loadState() error {
	if l.store == nil {
		return nil
	}
	var loaded, found, err = l.store.load(l.config)
	if err != nil {
		return err
	}
	if found {
		l.publish(loaded)
		return nil
	}
	return l.store.save(l.config, l.currentCandidate())
}

func (l *Ledger) validateAndIndexState() error {
	var err = l.rebuildIndexes()
	if err != nil {
		return fmt.Errorf("initialize ledger: %w", err)
	}
	return nil
}

func (l *Ledger) closeStore() {
	if l.store != nil {
		l.store.close()
		l.store = nil
	}
}

// Section 2.2 - Reconciliation

// PrepareRecon creates one unpublished canonical Ledger attempt.
func (l *Ledger) PrepareRecon(input ReconInput) (*ReconAttempt, error) {
	if !l.started || l.stopped {
		return nil, fmt.Errorf("reconcile ledger: invalid lifecycle state")
	}
	if input.ObservedMS == 0 || input.FillsThroughMS < l.fillsThroughMS {
		return nil, fmt.Errorf("reconcile ledger: invalid observation or backward Fill cursor")
	}
	return &ReconAttempt{
		input:          input,
		trades:         make(map[uint64]*trade.Trade),
		touchedTrades:  make(map[uint64]struct{}),
		touchedOrders:  make(map[uint64]struct{}),
		touchedFills:   make(map[uint64]struct{}),
		tradeSummaries: make(map[uint64]trade.Summary),
	}, nil
}

// UpdateReconFills stages selected Fill updates.
func (l *Ledger) UpdateReconFills(
	attempt *ReconAttempt,
	evidence []FillEvidence,
) (err error) {
	if attempt == nil {
		return fmt.Errorf("reconcile ledger Fills: attempt is required")
	}
	defer func() {
		if err != nil {
			l.refreshReconSummaries(attempt)
		}
	}()
	var owners = make(map[uint64]orderLocation, len(evidence))
	for _, current := range evidence {
		var location, exists, resolveErr = l.evidenceOrder(
			current.VenueOrderID,
			current.CLOID,
		)
		if resolveErr != nil {
			return fmt.Errorf("reconcile ledger Fills: %w", resolveErr)
		}
		if !exists {
			continue
		}
		if previous, seen := owners[current.VenueTID]; seen && previous != location {
			return fmt.Errorf("reconcile ledger Fills: Venue TID %d changed ownership", current.VenueTID)
		}
		if existing, known := l.fills[current.VenueTID]; known &&
			(existing.tradeID != location.tradeID || existing.orderID != location.orderID) {
			return fmt.Errorf("reconcile ledger Fills: Venue TID %d changed ownership", current.VenueTID)
		}
		owners[current.VenueTID] = location
	}
	for _, current := range evidence {
		var location, exists, resolveErr = l.evidenceOrder(
			current.VenueOrderID,
			current.CLOID,
		)
		if resolveErr != nil {
			return fmt.Errorf("reconcile ledger Fills: %w", resolveErr)
		}
		if !exists {
			continue
		}
		var owned, err = l.stageOrder(attempt, location)
		if err != nil {
			return fmt.Errorf("reconcile ledger Fills: %w", err)
		}
		var previousState = owned.ComparisonState()
		var input = owned.FillIdentity()
		var previousFill, known = owned.Fill(current.VenueTID)
		var previousHasFee bool
		if known {
			previousHasFee = previousFill.HasFee()
		}
		input.VenueOrderID = current.VenueOrderID
		input.VenueTID = current.VenueTID
		input.Side = current.Side
		input.Quantity = current.Quantity
		input.Price = current.Price
		input.TimestampMS = current.TimestampMS
		input.Fee = current.Fee
		input.Liquidity = current.Liquidity
		input.Raw = current.Raw
		err = owned.ApplyFill(input)
		if err != nil {
			return fmt.Errorf("reconcile ledger Fills: %w", err)
		}
		if owned.ComparisonState() == previousState {
			continue
		}
		attempt.touchedTrades[location.tradeID] = struct{}{}
		attempt.touchedOrders[location.orderID] = struct{}{}
		attempt.touchedFills[current.VenueTID] = struct{}{}
		l.fills[current.VenueTID] = fillLocation{
			tradeID: location.tradeID, orderID: location.orderID, venueTID: current.VenueTID,
		}
		var execution, _ = owned.Fill(current.VenueTID)
		var state = execution.State()
		if !known {
			attempt.fillChanges = append(attempt.fillChanges, FillChange{
				Kind: FillAdded, VenueTID: current.VenueTID, HasFee: state.HasFee, Fee: state.Fee,
			})
		} else if !previousHasFee && state.HasFee {
			attempt.fillChanges = append(attempt.fillChanges, FillChange{
				Kind: FillFeeEnriched, VenueTID: current.VenueTID, HasFee: true, Fee: state.Fee,
			})
		}
	}
	return nil
}

// UpdateReconOrders stages selected Order updates.
func (l *Ledger) UpdateReconOrders(
	attempt *ReconAttempt,
	evidence []OrderEvidence,
) (err error) {
	if attempt == nil {
		return fmt.Errorf("reconcile ledger Orders: attempt is required")
	}
	defer func() {
		if err != nil {
			l.refreshReconSummaries(attempt)
		}
	}()
	for _, current := range evidence {
		var location, exists, resolveErr = l.evidenceOrder(
			current.VenueOrderID,
			current.CLOID,
		)
		if resolveErr != nil {
			return fmt.Errorf("reconcile ledger Orders: %w", resolveErr)
		}
		if !exists {
			continue
		}
		if current.VenueOrderID != 0 {
			if existing, known := l.venueOrders[current.VenueOrderID]; known && existing != location {
				return fmt.Errorf(
					"reconcile ledger Orders: Venue Order %d changed ownership",
					current.VenueOrderID,
				)
			}
		}
		var owned, err = l.stageOrder(attempt, location)
		if err != nil {
			return fmt.Errorf("reconcile ledger Orders: %w", err)
		}
		var previous = owned.ComparisonState()
		err = owned.ApplyVenueState(order.VenueState{
			VenueOrderID: current.VenueOrderID,
			Status:       current.Status,
			RejectReason: current.RejectReason,
			TimestampMS:  current.TimestampMS,
			Raw:          current.Raw,
		})
		if err != nil {
			return fmt.Errorf("reconcile ledger Orders: %w", err)
		}
		owned.RefreshRecon()
		if current.VenueOrderID != 0 {
			l.venueOrders[current.VenueOrderID] = location
		}
		if owned.ComparisonState() == previous {
			continue
		}
		attempt.touchedTrades[location.tradeID] = struct{}{}
		attempt.touchedOrders[location.orderID] = struct{}{}
	}
	return nil
}

func (l *Ledger) evidenceOrder(
	venueOrderID uint64,
	cloid string,
) (orderLocation, bool, error) {
	var byVenue, venueKnown = l.venueOrders[venueOrderID]
	var byCLOID, cloidKnown = l.cloids[cloid]
	if cloid != "" && cloidKnown {
		if venueOrderID != 0 && venueKnown && byVenue != byCLOID {
			return orderLocation{}, false, fmt.Errorf(
				"Venue Order %d and cloid %s identify different Orders",
				venueOrderID,
				cloid,
			)
		}
		return byCLOID, true, nil
	}
	if venueOrderID != 0 && venueKnown {
		return byVenue, true, nil
	}
	return orderLocation{}, false, nil
}

// UpdateReconTrades refreshes touched structure and marks active Trade exposure.
func (l *Ledger) UpdateReconTrades(
	attempt *ReconAttempt,
	markPrice *decimal.Decimal,
) error {
	if attempt == nil {
		return fmt.Errorf("reconcile ledger Trades: attempt is required")
	}
	for tradeID := range l.activeTrades {
		if attempt.trades[tradeID] == nil {
			var ownedTrade = l.trades[tradeID]
			attempt.trades[tradeID] = ownedTrade
			attempt.tradeSummaries[tradeID] = markedTradeSummary(
				trade.Summary{},
				ownedTrade.ReconState(),
			)
		}
	}
	for tradeID, ownedTrade := range attempt.trades {
		if _, touched := attempt.touchedTrades[tradeID]; !touched {
			continue
		}
		var previous = ownedTrade.ReconState()
		var previousSummary = attempt.tradeSummaries[tradeID]
		delete(attempt.touchedTrades, tradeID)
		var err = ownedTrade.RefreshRecon(nil)
		if err != nil {
			return fmt.Errorf("reconcile ledger Trades: %w", err)
		}
		if !sameTradeReconState(previous, ownedTrade.ReconState()) {
			attempt.touchedTrades[tradeID] = struct{}{}
		}
		var currentSummary = ownedTrade.Summary()
		// Cached totals avoid repeated all-Trade aggregation while owned Trades remain authoritative.
		l.replaceTradeSummary(&previousSummary, currentSummary)
		attempt.tradeSummaries[tradeID] = currentSummary
	}
	for tradeID := range l.activeTrades {
		var ownedTrade = attempt.trades[tradeID]
		var previous = ownedTrade.ReconState()
		var previousSummary = attempt.tradeSummaries[tradeID]
		var err = ownedTrade.RefreshMark(markPrice)
		if err != nil {
			return fmt.Errorf("reconcile ledger Trades: %w", err)
		}
		if !sameTradeReconState(previous, ownedTrade.ReconState()) {
			attempt.touchedTrades[tradeID] = struct{}{}
		}
		var currentSummary = markedTradeSummary(
			previousSummary,
			ownedTrade.ReconState(),
		)
		// Cached totals avoid repeated all-Trade aggregation while owned Trades remain authoritative.
		l.replaceTradeSummary(&previousSummary, currentSummary)
		attempt.tradeSummaries[tradeID] = currentSummary
	}
	return nil
}

// ReconSummary returns current cached totals after direct domain updates.
func (l *Ledger) ReconSummary(attempt *ReconAttempt) (Summary, error) {
	if attempt == nil {
		return Summary{}, fmt.Errorf("read ledger recon summary: attempt is required")
	}
	return l.summary, nil
}

// CommitRecon persists and publishes one validated Ledger attempt.
func (l *Ledger) CommitRecon(attempt *ReconAttempt) error {
	if attempt == nil {
		return fmt.Errorf("commit ledger recon: attempt is required")
	}
	var err error
	if l.config.PersistMode == Max {
		err = l.store.saveRecon(l.config, attempt)
		if err != nil {
			return err
		}
	}
	for tradeID := range attempt.trades {
		l.refreshTradeIndexes(tradeID)
	}
	l.fillsThroughMS = attempt.input.FillsThroughMS
	l.lastReconMS = attempt.input.ObservedMS
	l.accountStateRaw = attempt.input.AccountStateRaw
	return nil
}

func (l *Ledger) stageOrder(
	attempt *ReconAttempt,
	location orderLocation,
) (*order.Order, error) {
	var ownedTrade = attempt.trades[location.tradeID]
	if ownedTrade == nil {
		ownedTrade = l.trades[location.tradeID]
		if ownedTrade == nil {
			return nil, fmt.Errorf("unknown Trade %d", location.tradeID)
		}
		attempt.trades[location.tradeID] = ownedTrade
		attempt.tradeSummaries[location.tradeID] = ownedTrade.Summary()
	}
	var owned, exists = ownedTrade.Order(location.orderID)
	if !exists {
		return nil, fmt.Errorf("unknown Order %d", location.orderID)
	}
	return owned, nil
}

func (l *Ledger) refreshReconSummaries(attempt *ReconAttempt) {
	for tradeID, ownedTrade := range attempt.trades {
		var previous = attempt.tradeSummaries[tradeID]
		var current = ownedTrade.Summary()
		l.replaceTradeSummary(&previous, current)
		attempt.tradeSummaries[tradeID] = current
	}
}

func sameTradeReconState(left trade.ReconState, right trade.ReconState) bool {
	return left.LedgerID == right.LedgerID && left.TradeID == right.TradeID &&
		left.TradeNo == right.TradeNo && left.Account == right.Account &&
		left.CycleNumber == right.CycleNumber && left.Symbol == right.Symbol &&
		left.Status == right.Status && left.Side == right.Side &&
		left.OpenQuantity.Equal(right.OpenQuantity) &&
		left.AverageEntryPrice.Equal(right.AverageEntryPrice) &&
		left.HasAveragePrice == right.HasAveragePrice &&
		left.RealizedPnL.Equal(right.RealizedPnL) &&
		left.UnrealizedPnL.Equal(right.UnrealizedPnL) &&
		left.GrossPnL.Equal(right.GrossPnL) && left.Fees.Equal(right.Fees) &&
		left.NetPnL.Equal(right.NetPnL) &&
		left.OpenedMS == right.OpenedMS && left.ClosedMS == right.ClosedMS &&
		left.UpdatedMS == right.UpdatedMS
}

func markedTradeSummary(previous trade.Summary, current trade.ReconState) trade.Summary {
	previous.Status = current.Status
	previous.RealizedPnL = current.RealizedPnL
	previous.UnrealizedPnL = current.UnrealizedPnL
	previous.GrossPnL = current.GrossPnL
	previous.Fees = current.Fees
	previous.NetPnL = current.NetPnL
	return previous
}

// Section 3 - Generic Helpers
