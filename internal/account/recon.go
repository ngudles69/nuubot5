package account

import (
	"fmt"
	"sort"
	"time"

	"github.com/shopspring/decimal"

	"nuubot/internal/account/fill"
	"nuubot/internal/account/ledger"
	"nuubot/internal/account/order"
	"nuubot/internal/hyperliquid"
	"nuubot/internal/market"
	"nuubot/internal/venue"
)

const feeRepairWindowMS = 1000

type reconAttempt struct {
	nowMS              uint64
	started            time.Time
	stage              string
	orders             []ledger.OrderUpdate
	orderStatusQueries int
	fills              []fill.Fill
	deferredFills      []fill.Fill
	fillQueries        []FillQueryTelemetry
	oidSearchOrders    int
	oidSearchFills     int
	pendingOrders      int
	pendingFills       int
	accountState       hyperliquid.AccountState
	accountRaw         string
	ledgerSummary      ledger.Summary
	snapshot           Snapshot
}

// Section 1 - Program Flow

// Init prepares one Venue-backed Account.
func (a *Account) Init(cfg Config) error {
	// Step 1: bind Account inputs
	a.bindInputs(cfg)

	// Step 2: validate Account identity
	var err = a.validateIdentity()
	if err != nil {
		return err
	}

	// Step 3: initialize Ledger
	err = a.initializeLedger()
	if err != nil {
		return err
	}

	// Step 4: initialize Store
	if cfg.PersistMode == "max" {
		a.store, err = openStore(cfg.Nuubot.RuntimePath)
		if err != nil {
			a.ledger.Stop()
			return err
		}
	}

	// Step 5: initialize Venue
	err = a.initializeVenue()
	if err != nil {
		if a.store != nil {
			a.store.close()
		}
		a.ledger.Stop()
		return err
	}

	// Step 6: initialize Account
	a.initializeAccount()
	if err = a.persist(false); err != nil {
		a.venue.Stop()
		a.store.close()
		a.ledger.Stop()
		return err
	}
	return nil
}

// Reconcile creates one coherent Account snapshot and reports consecutive failures.
func (a *Account) Reconcile(nowMS uint64, forced bool) (Snapshot, bool, uint64, error) {
	// Step 1: record reconciliation call
	a.reconStats.Calls++

	// Step 2: execute reconciliation
	var snapshot, refreshed, err = a.reconcile(nowMS, forced)
	if err != nil {
		a.failureCount++
	} else if refreshed {
		a.failureCount = 0
	}

	// Step 3: publish reconciliation outcome
	var telemetry = a.reconTelemetry
	telemetry.ConsecutiveFailures = a.failureCount
	a.reconTelemetry = telemetry
	a.recordReconOutcome(refreshed, err)
	return snapshot, refreshed, a.failureCount, err
}

func (a *Account) reconcile(nowMS uint64, forced bool) (Snapshot, bool, error) {
	// Step 1: prepare reconciliation
	var attempt, skipped, err = a.prepareRecon(nowMS, forced)
	if err != nil || skipped {
		return a.finalizeRecon(attempt, err)
	}

	// Step 2: download current Order evidence
	err = a.downloadOrderEvidence(attempt)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 3: download Fill history
	err = a.downloadFillEvidence(attempt)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 4: download current Account state
	err = a.downloadAccountState(attempt)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 5 - Update Fill Records
	attempt.deferredFills, err = a.updateFillRecords(attempt, attempt.fills)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 6 - Update Order Records
	err = a.updateOrderRecords(attempt, attempt.orders)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 7 - Search Fills by Updated Order OIDs
	err = a.searchFillOIDs(attempt)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 8 - Update Trade Records
	err = a.updateTradeRecords(attempt)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 9 - Update Account Snapshot
	err = a.updateAccountSnapshot(attempt)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 10 - Persist and Publish
	err = a.publishRecon(attempt)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 11 - Finalize Recon Outcome and Return
	return a.finalizeRecon(attempt, nil)
}

// Section 2 - Domain Helpers

// Section 2.1 - Initialization

func (a *Account) recordReconOutcome(refreshed bool, err error) {
	if err != nil {
		a.reconStats.Executed++
		a.reconStats.Failed++
		return
	}
	if !refreshed {
		a.reconStats.SkippedClean++
		return
	}
	a.reconStats.Executed++
	a.reconStats.Succeeded++
}

func (a *Account) bindInputs(cfg Config) {
	a.config = cfg
	if cfg.Nuubot != nil {
		a.log = cfg.Nuubot.Log
	}
}

func (a *Account) validateIdentity() error {
	var cfg = a.config
	if cfg.Nuubot == nil || a.log == nil || cfg.Nuubot.MarketData == nil ||
		cfg.SweepID == 0 || cfg.BotID == 0 || cfg.LedgerID == 0 ||
		cfg.CycleNumber <= 0 || cfg.ExecutorNumber <= 0 ||
		cfg.Name == "" || cfg.Symbol == "" {
		return fmt.Errorf("initialize Account: complete identity is required")
	}
	if cfg.Venue != "simulator" || cfg.Network != "simnet" {
		return fmt.Errorf("initialize Account: first trading tranche requires simulator simnet")
	}
	if cfg.Nuubot.Meta.Symbol != cfg.Symbol ||
		cfg.Nuubot.Meta.IsDelisted ||
		cfg.Nuubot.Meta.Retired {
		return fmt.Errorf("initialize Account: symbol Meta is unavailable")
	}
	if cfg.Nuubot.App.Hyperliquid.MinOrderNotionalUSDC == 0 ||
		!cfg.EquityUSDC.IsPositive() {
		return fmt.Errorf("initialize Account: notional floor and equity must be positive")
	}
	if cfg.PersistMode != "none" && cfg.PersistMode != "max" {
		return fmt.Errorf("initialize Account: invalid persistence mode")
	}
	return nil
}

func (a *Account) initializeLedger() error {
	var cfg = a.config
	var err = a.ledger.Init(ledger.Config{
		SweepID:        cfg.SweepID,
		BotID:          cfg.BotID,
		ID:             cfg.LedgerID,
		CycleNumber:    cfg.CycleNumber,
		ExecutorNumber: cfg.ExecutorNumber,
		Venue:          cfg.Venue,
		Account:        cfg.Name,
		Network:        cfg.Network,
		Symbol:         cfg.Symbol,
	})
	if err != nil {
		return fmt.Errorf("initialize Account: %w", err)
	}
	return nil
}

func (a *Account) initializeVenue() error {
	var cfg = a.config
	var err = a.venue.Init(venue.Config{
		MarketData: cfg.Nuubot.MarketData,
		MarketKey: market.Key{
			Venue:   cfg.Venue,
			Network: cfg.Network,
			Symbol:  cfg.Symbol,
		},
		Account:     cfg.Name,
		Asset:       int(cfg.Nuubot.Meta.AssetID),
		Symbol:      cfg.Symbol,
		Equity:      cfg.EquityUSDC,
		FeePct:      cfg.FeePct,
		SlippagePct: cfg.SlippagePct,
		PersistMode: cfg.PersistMode,
		Path:        cfg.Nuubot.RuntimePath,
	})
	if err != nil {
		return fmt.Errorf("initialize Account: %w", err)
	}
	return nil
}

func (a *Account) initializeAccount() {
	a.dirty = true
	a.started = true
}

// Section 2.2 - Reconciliation

func (a *Account) prepareRecon(nowMS uint64, forced bool) (*reconAttempt, bool, error) {
	var attempt = &reconAttempt{
		nowMS:   nowMS,
		started: time.Now(),
		stage:   "prepare",
	}
	if !a.started || a.stopped {
		return attempt, false, fmt.Errorf("reconcile Account: invalid lifecycle state")
	}
	if nowMS == 0 {
		return attempt, false, fmt.Errorf("reconcile Account: timestamp is required")
	}
	if a.lastReconMS != 0 && nowMS < a.lastReconMS {
		return attempt, false, fmt.Errorf("recon clock moved backward")
	}
	attempt.pendingOrders, attempt.pendingFills = a.ledger.PendingCounts()
	var sinceMS = nowMS - a.lastReconMS
	var dirty = a.dirty || a.ledger.HasPendingRecon()
	var cadence = a.config.Nuubot.Runtime
	if a.lastReconMS != 0 && !forced &&
		sinceMS < cadence.ReconSweepIntervalMS {
		if !dirty {
			return attempt, true, nil
		}
		if sinceMS < cadence.ReconIntervalMS {
			return attempt, true, nil
		}
	}
	a.log.Info("account running Recon 1")
	return attempt, false, nil
}

func (a *Account) downloadOrderEvidence(attempt *reconAttempt) error {
	attempt.stage = "download_orders"
	var payload, err = a.venue.OpenOrders(a.config.Name)
	if err != nil {
		return err
	}
	var openOrders []hyperliquid.OpenOrder
	openOrders, err = hyperliquid.DecodeOpenOrders(payload)
	if err != nil {
		return err
	}

	var active = a.ledger.ActiveOrders()
	var byCLOID = make(map[string]order.Order, len(active))
	var byOID = make(map[uint64]order.Order, len(active))
	for _, owned := range active {
		byCLOID[owned.CLOID] = owned
		if owned.VenueOrderID != 0 {
			byOID[owned.VenueOrderID] = owned
		}
	}

	var observed = make(map[uint64]struct{}, len(active))
	attempt.orders = make([]ledger.OrderUpdate, 0, len(active))
	for _, current := range openOrders {
		var owned, found, resolveErr = resolveActiveOrder(current, byCLOID, byOID)
		if resolveErr != nil {
			return resolveErr
		}
		if !found {
			continue
		}
		if _, duplicate := observed[owned.OrderID]; duplicate {
			return fmt.Errorf("download Order evidence: duplicate Order %d", owned.OrderID)
		}
		var update ledger.OrderUpdate
		update, err = a.orderUpdate(owned, current, "open", order.Open, current.TimestampMS, "")
		if err != nil {
			return err
		}
		attempt.orders = append(attempt.orders, update)
		observed[owned.OrderID] = struct{}{}
	}

	for _, owned := range active {
		if _, found := observed[owned.OrderID]; found {
			continue
		}
		attempt.orderStatusQueries++
		payload, err = a.venue.OrderStatus(a.config.Name, owned.CLOID)
		if err != nil {
			return err
		}
		var current hyperliquid.OrderStatus
		current, err = hyperliquid.DecodeOrderStatus(payload)
		if err != nil {
			return err
		}
		if current.Status == "unknownOid" {
			attempt.orders = append(attempt.orders, ledger.OrderUpdate{
				OrderID: owned.OrderID,
				Update: order.Update{
					VenueOrderID: owned.VenueOrderID,
					VenueStatus:  current.Status,
					Status:       owned.Status,
					UpdatedMS:    attempt.nowMS,
					RawJSON:      current.Raw,
				},
			})
			continue
		}
		if current.Order == nil {
			return fmt.Errorf("download Order evidence: exact Order is absent")
		}
		var status order.Status
		status, err = venueOrderStatus(current.OrderStatus)
		if err != nil {
			return err
		}
		var update ledger.OrderUpdate
		update, err = a.orderUpdate(
			owned,
			*current.Order,
			current.OrderStatus,
			status,
			current.StatusTimestamp,
			current.Raw,
		)
		if err != nil {
			return err
		}
		attempt.orders = append(attempt.orders, update)
	}
	return nil
}

func (a *Account) downloadFillEvidence(attempt *reconAttempt) error {
	attempt.stage = "download_fills"
	var merged = make(map[uint64]fill.Fill)
	var observed = make(map[uint64]fill.Fill)
	var err = a.pullFillEvidence(
		attempt,
		"discovery",
		a.lastReconMS,
		attempt.nowMS,
		merged,
		observed,
	)
	if err != nil {
		return err
	}
	var queried = make(map[[2]uint64]struct{})
	for _, pending := range a.ledger.PendingFillAnchors() {
		var startMS = pending.TimestampMS - min(pending.TimestampMS, uint64(feeRepairWindowMS))
		var endMS = pending.TimestampMS + feeRepairWindowMS
		if endMS < pending.TimestampMS {
			endMS = ^uint64(0)
		}
		var bounds = [2]uint64{startMS, endMS}
		if _, exists := queried[bounds]; exists {
			continue
		}
		queried[bounds] = struct{}{}
		err = a.pullFillEvidence(
			attempt,
			"repair",
			startMS,
			endMS,
			merged,
			observed,
		)
		if err != nil {
			return err
		}
	}
	attempt.fills = sortedFills(merged)
	return nil
}

func (a *Account) pullFillEvidence(
	attempt *reconAttempt,
	kind string,
	startMS uint64,
	endMS uint64,
	merged map[uint64]fill.Fill,
	observed map[uint64]fill.Fill,
) error {
	var started = time.Now()
	var payload, err = a.venue.Fills(a.config.Name, startMS, endMS)
	if err != nil {
		return err
	}
	var rows []hyperliquid.Fill
	rows, err = hyperliquid.DecodeFills(payload)
	if err != nil {
		return err
	}
	var query = FillQueryTelemetry{
		Kind:       kind,
		StartMS:    startMS,
		EndMS:      endMS,
		Rows:       len(rows),
		DurationMS: time.Since(started).Milliseconds(),
	}
	for _, row := range rows {
		var current fill.Fill
		current, err = a.venueFill(row)
		if err != nil {
			query.Error = err.Error()
			attempt.fillQueries = append(attempt.fillQueries, query)
			return err
		}
		var previous, seen = observed[current.VenueTID]
		if seen {
			var combined fill.Fill
			combined, err = mergeFill(previous, current)
			if err != nil {
				query.Error = err.Error()
				attempt.fillQueries = append(attempt.fillQueries, query)
				return err
			}
			observed[current.VenueTID] = combined
			merged[current.VenueTID] = combined
			query.FillsUnchanged++
			var existing, exists = a.ledger.Fill(current.VenueTID)
			if exists && !existing.HasFee() {
				query.PendingMatched++
			}
			continue
		}
		var existing, exists = a.ledger.Fill(current.VenueTID)
		switch {
		case !exists:
			query.FillsAdded++
		case !existing.HasFee() && current.Fee != nil:
			query.FeesEnriched++
			query.PendingMatched++
		case !existing.HasFee():
			query.FillsUnchanged++
			query.PendingMatched++
		default:
			query.FillsUnchanged++
		}
		observed[current.VenueTID] = current
		merged[current.VenueTID] = current
	}
	attempt.fillQueries = append(attempt.fillQueries, query)
	return nil
}

func (a *Account) downloadAccountState(attempt *reconAttempt) error {
	attempt.stage = "download_account"
	var payload, err = a.venue.AccountState(a.config.Name)
	if err != nil {
		return err
	}
	var state hyperliquid.AccountState
	state, err = hyperliquid.DecodeClearinghouseState(payload)
	if err != nil {
		return err
	}
	_, _, err = accountPosition(state, a.config.Symbol)
	if err != nil {
		return err
	}
	attempt.accountState = state
	attempt.accountRaw = string(payload)
	return nil
}

func (a *Account) updateOrderRecords(
	attempt *reconAttempt,
	updates []ledger.OrderUpdate,
) error {
	attempt.stage = "update_orders"
	if len(updates) == 0 {
		return nil
	}
	return a.ledger.UpdateOrders(updates)
}

func (a *Account) updateFillRecords(
	attempt *reconAttempt,
	fills []fill.Fill,
) ([]fill.Fill, error) {
	attempt.stage = "update_fills"
	var deferredFills = make([]fill.Fill, 0)
	for _, execution := range fills {
		var previous, existed = a.ledger.Fill(execution.VenueTID)
		var changed, deferred, err = a.ledger.UpdateReconFill(execution)
		if err != nil {
			return nil, err
		}
		if deferred {
			deferredFills = append(deferredFills, execution)
			continue
		}
		if !changed {
			continue
		}
		if existed && !previous.HasFee() && execution.Fee != nil {
			a.log.Info(fmt.Sprintf(
				"fill fee enriched venue=simulator network=%s account=%s symbol=%s venue_tid=%d previous=missing fee=%s",
				a.config.Network,
				a.config.Name,
				a.config.Symbol,
				execution.VenueTID,
				execution.Fee,
			))
			continue
		}
		a.log.Info(fmt.Sprintf(
			"fill added venue=simulator network=%s account=%s symbol=%s venue_tid=%d has_fee=%t",
			a.config.Network,
			a.config.Name,
			a.config.Symbol,
			execution.VenueTID,
			execution.Fee != nil,
		))
	}
	return deferredFills, nil
}

func (a *Account) searchFillOIDs(attempt *reconAttempt) error {
	attempt.stage = "oid_search"
	var matched, orderIDs, unmatched, err = a.ledger.ReconOIDSearch(attempt.deferredFills)
	attempt.oidSearchOrders = len(orderIDs)
	attempt.oidSearchFills = len(matched)
	if attempt.oidSearchFills == 0 {
		a.log.Info("Recon-OIDSearch found nothing")
	} else {
		a.log.Info(fmt.Sprintf(
			"Recon-OIDSearch found orders=%d fills=%d",
			attempt.oidSearchOrders,
			attempt.oidSearchFills,
		))
	}
	if err != nil {
		return err
	}
	if len(matched) != 0 {
		var deferredAgain []fill.Fill
		deferredAgain, err = a.updateFillRecords(attempt, matched)
		if err != nil {
			attempt.stage = "oid_search"
			return err
		}
		if len(deferredAgain) != 0 {
			attempt.stage = "oid_search"
			return fmt.Errorf("Recon-OIDSearch matched Fills remained unresolved")
		}
		var updates []ledger.OrderUpdate
		updates, err = selectOrderUpdates(attempt.orders, orderIDs)
		if err != nil {
			attempt.stage = "oid_search"
			return err
		}
		err = a.updateOrderRecords(attempt, updates)
		if err != nil {
			attempt.stage = "oid_search"
			return err
		}
	}
	attempt.stage = "oid_search"
	if unmatched != 0 {
		return fmt.Errorf("Recon-OIDSearch left %d Fills unmatched", unmatched)
	}
	if len(matched) != 0 {
		return fmt.Errorf("Recon-OIDSearch found Fills after Order updates")
	}
	return nil
}

func selectOrderUpdates(
	updates []ledger.OrderUpdate,
	orderIDs []uint64,
) ([]ledger.OrderUpdate, error) {
	var wanted = make(map[uint64]struct{}, len(orderIDs))
	for _, orderID := range orderIDs {
		wanted[orderID] = struct{}{}
	}
	var selected = make([]ledger.OrderUpdate, 0, len(wanted))
	for _, update := range updates {
		if _, found := wanted[update.OrderID]; !found {
			continue
		}
		selected = append(selected, update)
		delete(wanted, update.OrderID)
	}
	if len(wanted) != 0 {
		return nil, fmt.Errorf(
			"Recon-OIDSearch missing %d Order updates",
			len(wanted),
		)
	}
	return selected, nil
}

func (a *Account) updateTradeRecords(attempt *reconAttempt) error {
	attempt.stage = "update_trades"
	return a.ledger.UpdateMark(a.markPrice())
}

func (a *Account) updateAccountSnapshot(attempt *reconAttempt) error {
	attempt.stage = "update_account"
	var current = a.ledger.Summary()
	var positionSize, entryPrice, err = accountPosition(
		attempt.accountState,
		a.config.Symbol,
	)
	if err != nil {
		return err
	}
	attempt.ledgerSummary = current
	attempt.snapshot = Snapshot{
		Generation:       a.generation + 1,
		CycleNumber:      a.config.CycleNumber,
		ExecutorNumber:   a.config.ExecutorNumber,
		Account:          a.config.Name,
		Venue:            a.config.Venue,
		Network:          a.config.Network,
		Symbol:           a.config.Symbol,
		ObservedMS:       attempt.nowMS,
		AccountValue:     attempt.accountState.Margin.Equity,
		Withdrawable:     attempt.accountState.Withdrawable,
		PositionQuantity: positionSize,
		EntryPrice:       entryPrice,
		RealizedPnL:      current.RealizedPnL,
		UnrealizedPnL:    current.UnrealizedPnL,
		GrossPnL:         current.GrossPnL,
		Fees:             current.Fees,
		NetPnL:           current.NetPnL,
		OpenTrades:       current.OpenTrades,
		ActiveOrders:     current.ActiveOrders,
		Fills:            current.Fills,
		PendingOrders:    current.PendingOrders,
		PendingFills:     current.PendingFills,
	}
	return nil
}

func (a *Account) publishRecon(attempt *reconAttempt) error {
	attempt.stage = "publish"
	a.ledger.UpdateAccountPayload(attempt.accountRaw)
	a.generation = attempt.snapshot.Generation
	a.lastSnapshot = attempt.snapshot
	return a.persist(false)
}

func (a *Account) finalizeRecon(
	attempt *reconAttempt,
	err error,
) (Snapshot, bool, error) {
	var telemetry = ReconTelemetry{
		Kind:                "sweep",
		Outcome:             ReconFailed,
		Stage:               attempt.stage,
		ObservedMS:          attempt.nowMS,
		DurationMS:          time.Since(attempt.started).Milliseconds(),
		Orders:              len(attempt.orders),
		OrderStatusQueries:  attempt.orderStatusQueries,
		Fills:               len(attempt.fills),
		PendingOrdersBefore: attempt.pendingOrders,
		PendingFillsBefore:  attempt.pendingFills,
		OIDSearchOrders:     attempt.oidSearchOrders,
		OIDSearchFills:      attempt.oidSearchFills,
		FillQueries:         append([]FillQueryTelemetry(nil), attempt.fillQueries...),
	}
	if err != nil {
		a.dirty = true
		telemetry.Error = err.Error()
		a.reconTelemetry = telemetry
		return Snapshot{}, false, fmt.Errorf("reconcile Account: %w", err)
	}
	if attempt.snapshot.ObservedMS == 0 {
		telemetry.Outcome = ReconSkipped
		telemetry.Stage = "complete"
		telemetry.ConsecutiveFailures = a.failureCount
		telemetry.PendingOrders = attempt.pendingOrders
		telemetry.PendingFills = attempt.pendingFills
		a.reconTelemetry = telemetry
		return a.lastSnapshot, false, nil
	}
	a.dirty = false
	a.lastReconMS = attempt.nowMS
	a.stats.reconciles++
	telemetry.Outcome = ReconSucceeded
	telemetry.Stage = "complete"
	telemetry.PendingOrders = attempt.ledgerSummary.PendingOrders
	telemetry.PendingFills = attempt.ledgerSummary.PendingFills
	a.reconTelemetry = telemetry
	return attempt.snapshot, true, nil
}

// Section 2.3 - Venue Evidence

func resolveActiveOrder(
	current hyperliquid.OpenOrder,
	byCLOID map[string]order.Order,
	byOID map[uint64]order.Order,
) (order.Order, bool, error) {
	var byClient, hasClient = byCLOID[current.CLOID]
	var byVenue, hasVenue = byOID[current.VenueOrderID]
	if hasClient && hasVenue && byClient.OrderID != byVenue.OrderID {
		return order.Order{}, false, fmt.Errorf(
			"download Order evidence: CLOID and Venue OID resolve different Orders",
		)
	}
	if hasClient {
		return byClient, true, nil
	}
	if hasVenue {
		return byVenue, true, nil
	}
	return order.Order{}, false, nil
}

func (a *Account) orderUpdate(
	owned order.Order,
	current hyperliquid.OpenOrder,
	venueStatus string,
	status order.Status,
	timestampMS uint64,
	raw string,
) (ledger.OrderUpdate, error) {
	if current.Coin != a.config.Symbol ||
		current.Side != owned.Side ||
		(current.CLOID == "" && current.VenueOrderID == 0) ||
		timestampMS == 0 {
		return ledger.OrderUpdate{}, fmt.Errorf(
			"download Order evidence: invalid official identity",
		)
	}
	if current.CLOID != "" && current.CLOID != owned.CLOID {
		return ledger.OrderUpdate{}, fmt.Errorf(
			"download Order evidence: changed CLOID",
		)
	}
	if owned.VenueOrderID != 0 && current.VenueOrderID != 0 &&
		current.VenueOrderID != owned.VenueOrderID {
		return ledger.OrderUpdate{}, fmt.Errorf(
			"download Order evidence: changed Venue OID",
		)
	}
	if raw == "" {
		var payload, err = hyperliquid.Encode(current)
		if err != nil {
			return ledger.OrderUpdate{}, err
		}
		raw = string(payload)
	}
	return ledger.OrderUpdate{
		OrderID: owned.OrderID,
		Update: order.Update{
			VenueOrderID: current.VenueOrderID,
			VenueStatus:  venueStatus,
			Status:       status,
			UpdatedMS:    timestampMS,
			RawJSON:      raw,
		},
	}, nil
}

func (a *Account) venueFill(execution hyperliquid.Fill) (fill.Fill, error) {
	if execution.Coin != a.config.Symbol ||
		(execution.Side != order.Buy && execution.Side != order.Sell) ||
		execution.Direction == "" {
		return fill.Fill{}, fmt.Errorf(
			"download Fill evidence: invalid official identity",
		)
	}
	var price, err = decimal.NewFromString(execution.Price)
	if err != nil || !price.IsPositive() {
		return fill.Fill{}, fmt.Errorf("download Fill evidence: invalid price")
	}
	var quantity decimal.Decimal
	quantity, err = decimal.NewFromString(execution.Size)
	if err != nil || !quantity.IsPositive() {
		return fill.Fill{}, fmt.Errorf("download Fill evidence: invalid size")
	}
	if _, err = decimal.NewFromString(execution.StartPosition); err != nil {
		return fill.Fill{}, fmt.Errorf(
			"download Fill evidence: invalid start position",
		)
	}
	if _, err = decimal.NewFromString(execution.ClosedPnL); err != nil {
		return fill.Fill{}, fmt.Errorf(
			"download Fill evidence: invalid closed PnL",
		)
	}
	var fee *decimal.Decimal
	if execution.Fee != nil {
		if execution.FeeToken == "" {
			return fill.Fill{}, fmt.Errorf(
				"download Fill evidence: missing fee token",
			)
		}
		var value decimal.Decimal
		value, err = decimal.NewFromString(*execution.Fee)
		if err != nil {
			return fill.Fill{}, fmt.Errorf("download Fill evidence: invalid fee")
		}
		fee = &value
	}
	var raw []byte
	raw, err = hyperliquid.Encode(execution)
	if err != nil {
		return fill.Fill{}, err
	}
	var liquidity = "maker"
	if execution.Crossed {
		liquidity = "taker"
	}
	return fill.Fill{
		CLOID:        execution.CLOID,
		VenueOrderID: execution.VenueOrderID,
		VenueTID:     execution.VenueTID,
		Symbol:       execution.Coin,
		Side:         execution.Side,
		Quantity:     quantity,
		Price:        price,
		TimestampMS:  execution.TimestampMS,
		Fee:          fee,
		Liquidity:    liquidity,
		RawJSON:      string(raw),
	}, nil
}

func accountPosition(
	state hyperliquid.AccountState,
	symbol string,
) (decimal.Decimal, decimal.Decimal, error) {
	var size = decimal.Zero
	var entry = decimal.Zero
	var found bool
	for _, current := range state.Positions {
		if current.Symbol != symbol {
			continue
		}
		if found {
			return decimal.Zero, decimal.Zero, fmt.Errorf(
				"download Account state: duplicate symbol %s",
				symbol,
			)
		}
		found = true
		size = current.SignedSize
		if current.EntryPrice != nil {
			entry = *current.EntryPrice
		}
		if !size.IsZero() && (current.EntryPrice == nil || !entry.IsPositive()) {
			return decimal.Zero, decimal.Zero, fmt.Errorf(
				"download Account state: open position lacks entry price",
			)
		}
	}
	return size, entry, nil
}

func venueOrderStatus(value string) (order.Status, error) {
	switch value {
	case "open":
		return order.Open, nil
	case "filled":
		return order.Filled, nil
	case "canceled", "cancelled":
		return order.Canceled, nil
	case "rejected":
		return order.Rejected, nil
	case "expired":
		return order.Expired, nil
	default:
		return "", fmt.Errorf(
			"download Order evidence: unknown Venue status %q",
			value,
		)
	}
}

// Section 3 - Generic Helpers

func mergeFill(previous fill.Fill, current fill.Fill) (fill.Fill, error) {
	if previous.CLOID != current.CLOID ||
		previous.VenueOrderID != current.VenueOrderID ||
		previous.VenueTID != current.VenueTID ||
		previous.Symbol != current.Symbol ||
		previous.Side != current.Side ||
		!previous.Quantity.Equal(current.Quantity) ||
		!previous.Price.Equal(current.Price) ||
		previous.TimestampMS != current.TimestampMS {
		return fill.Fill{}, fmt.Errorf(
			"merge Fill evidence: changed execution for Venue TID %d",
			current.VenueTID,
		)
	}
	if previous.Fee != nil && current.Fee != nil &&
		!previous.Fee.Equal(*current.Fee) {
		return fill.Fill{}, fmt.Errorf(
			"merge Fill evidence: changed fee for Venue TID %d",
			current.VenueTID,
		)
	}
	var merged = previous
	if merged.Fee == nil && current.Fee != nil {
		var fee = *current.Fee
		merged.Fee = &fee
	}
	if merged.Liquidity == "" {
		merged.Liquidity = current.Liquidity
	} else if current.Liquidity != "" &&
		merged.Liquidity != current.Liquidity {
		return fill.Fill{}, fmt.Errorf(
			"merge Fill evidence: changed liquidity for Venue TID %d",
			current.VenueTID,
		)
	}
	if current.RawJSON != "" {
		merged.RawJSON = current.RawJSON
	}
	return merged, nil
}

func sortedFills(values map[uint64]fill.Fill) []fill.Fill {
	var ids = make([]uint64, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool { return ids[left] < ids[right] })
	var rows = make([]fill.Fill, 0, len(ids))
	for _, id := range ids {
		rows = append(rows, values[id])
	}
	return rows
}
