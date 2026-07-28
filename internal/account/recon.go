package account

import (
	"fmt"
	"slices"
	"time"

	"github.com/shopspring/decimal"

	"nuubot/internal/hyperliquid"
	"nuubot/internal/ledger"
	"nuubot/internal/market"
	"nuubot/internal/order"
	"nuubot/internal/simulator"
)

const feeRepairWindowMS = 1000

type reconAttempt struct {
	nowMS              uint64
	started            time.Time
	stage              string
	orders             []ledger.OrderEvidence
	orderStatusQueries int
	fills              []ledger.FillEvidence
	fillQueries        []FillQueryTelemetry
	fillChanges        []ledger.FillChange
	pendingOrders      int
	pendingFills       int
	accountState       hyperliquid.AccountState
	accountRaw         string
	ledgerAttempt      *ledger.ReconAttempt
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

	// Step 3: validate persistence mode
	err = a.validatePersistence()
	if err != nil {
		return err
	}

	// Step 4: initialize Ledger with persistence mode
	err = a.initializeLedger()
	if err != nil {
		return err
	}

	// Step 5: initialize Venue with persistence mode
	err = a.initializeVenue()
	if err != nil {
		a.ledger.Stop()
		return err
	}

	// Step 6: initialize Account
	a.initializeAccount()
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
	// Step 1: prepare attempt
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

	// Step 5: update Fill records
	err = a.updateFillRecords(attempt)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 6: update Order records
	err = a.updateOrderRecords(attempt)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 7: update Trade records
	err = a.updateTradeRecords(attempt)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 8: update Account Snapshot
	err = a.updateAccountSnapshot(attempt)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 9: persist and publish
	err = a.persistAndPublishRecon(attempt)
	if err != nil {
		return a.finalizeRecon(attempt, err)
	}

	// Step 10: finalize Recon outcome and return
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
	a.log = cfg.Nuubot.Log
	a.config = cfg
}

func (a *Account) validateIdentity() error {
	var cfg = a.config
	if a.log == nil || cfg.Nuubot.MarketData == nil || cfg.LedgerID == 0 ||
		cfg.CycleNumber <= 0 || cfg.ExecutorNumber <= 0 || cfg.Name == "" || cfg.Symbol == "" {
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
	return nil
}

func (a *Account) validatePersistence() error {
	if a.config.PersistMode != ledger.None && a.config.PersistMode != ledger.Max {
		return fmt.Errorf(
			"initialize Account: invalid persistence mode %q",
			a.config.PersistMode,
		)
	}
	return nil
}

func (a *Account) initializeLedger() error {
	var cfg = a.config
	var err = a.ledger.Init(ledger.Config{
		ID:             cfg.LedgerID,
		CycleNumber:    cfg.CycleNumber,
		ExecutorNumber: cfg.ExecutorNumber,
		Account:        cfg.Name,
		Network:        cfg.Network,
		Symbol:         cfg.Symbol,
		PersistMode:    cfg.PersistMode,
		Path:           cfg.Nuubot.RuntimePath,
	})
	if err != nil {
		return fmt.Errorf("initialize Account: %w", err)
	}
	return nil
}

func (a *Account) initializeVenue() error {
	var cfg = a.config
	var simulated simulator.Simulator
	var err = simulated.Init(simulator.Config{
		MarketData: cfg.Nuubot.MarketData,
		MarketKey: market.Key{
			Venue:   cfg.Venue,
			Network: cfg.Network,
			Symbol:  cfg.Symbol,
		},
		OnChange: func() {
			a.dirty = true
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
	a.venue = &simulated
	return nil
}

func (a *Account) initializeAccount() {
	a.dirty = true
	a.started = true
}

// Section 2.2 - Reconciliation Pipeline

func (a *Account) prepareRecon(nowMS uint64, forced bool) (*reconAttempt, bool, error) {
	var attempt = &reconAttempt{
		nowMS:   nowMS,
		started: time.Now(),
		stage:   "prepare",
	}
	if !a.started || a.stopped || nowMS == 0 {
		return attempt, false, fmt.Errorf("invalid state or timestamp")
	}
	if a.lastReconMS > nowMS {
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
	var openCLOIDs = make(map[string]uint64, len(openOrders))
	var openOIDs = make(map[uint64]string, len(openOrders))
	attempt.orders = make([]ledger.OrderEvidence, 0, len(openOrders))
	for _, current := range openOrders {
		var evidence ledger.OrderEvidence
		evidence, err = a.orderEvidence(current, order.Open, current.TimestampMS, "")
		if err != nil {
			return err
		}
		if current.CLOID != "" {
			var _, exists = openCLOIDs[current.CLOID]
			if exists {
				return fmt.Errorf("download Order evidence: duplicate cloid %s", current.CLOID)
			}
			openCLOIDs[current.CLOID] = current.VenueOrderID
		}
		if current.VenueOrderID != 0 {
			var _, exists = openOIDs[current.VenueOrderID]
			if exists {
				return fmt.Errorf(
					"download Order evidence: duplicate Venue OID %d",
					current.VenueOrderID,
				)
			}
			openOIDs[current.VenueOrderID] = current.CLOID
		}
		attempt.orders = append(attempt.orders, evidence)
	}
	for _, active := range a.ledger.ActiveOrders() {
		var venueOID, open = openCLOIDs[active.CLOID]
		if open {
			if active.VenueOrderID != 0 && venueOID != 0 &&
				venueOID != active.VenueOrderID {
				return fmt.Errorf(
					"download Order evidence: cloid %s changed Venue OID",
					active.CLOID,
				)
			}
			continue
		}
		if active.VenueOrderID != 0 {
			var venueCLOID string
			venueCLOID, open = openOIDs[active.VenueOrderID]
			if open {
				if venueCLOID != "" && venueCLOID != active.CLOID {
					return fmt.Errorf(
						"download Order evidence: Venue OID %d changed CLOID",
						active.VenueOrderID,
					)
				}
				continue
			}
		}
		attempt.orderStatusQueries++
		payload, err = a.venue.OrderStatus(a.config.Name, active.CLOID)
		if err != nil {
			return err
		}
		var current hyperliquid.OrderStatus
		current, err = hyperliquid.DecodeOrderStatus(payload)
		if err != nil {
			return err
		}
		if current.Status == "unknownOid" {
			if active.VenueOrderID == 0 &&
				(active.Status == order.Created || active.Status == order.Submitted) {
				attempt.orders = append(attempt.orders, ledger.OrderEvidence{
					CLOID:        active.CLOID,
					Status:       order.Error,
					RejectReason: "Venue submission is absent",
					TimestampMS:  attempt.nowMS,
					Raw:          current.Raw,
				})
				continue
			}
			return fmt.Errorf("download Order evidence: unknown cloid %s", active.CLOID)
		}
		if current.Order == nil {
			return fmt.Errorf("download Order evidence: exact Order is absent")
		}
		if current.Order.CLOID != "" && current.Order.CLOID != active.CLOID {
			return fmt.Errorf(
				"download Order evidence: exact Order changed CLOID",
			)
		}
		if active.VenueOrderID != 0 && current.Order.VenueOrderID != 0 &&
			current.Order.VenueOrderID != active.VenueOrderID {
			return fmt.Errorf(
				"download Order evidence: exact Order changed Venue OID",
			)
		}
		var status order.Status
		status, err = venueOrderStatus(current.OrderStatus)
		if err != nil {
			return err
		}
		var evidence ledger.OrderEvidence
		evidence, err = a.orderEvidence(
			*current.Order,
			status,
			current.StatusTimestamp,
			current.Raw,
		)
		if err != nil {
			return err
		}
		attempt.orders = append(attempt.orders, evidence)
	}
	return nil
}

func (a *Account) downloadFillEvidence(attempt *reconAttempt) error {
	attempt.stage = "download_fills"
	var merged = make(map[uint64]ledger.FillEvidence)
	var observed = make(map[uint64]ledger.FillEvidence)
	var err = a.pullFillEvidence(
		attempt,
		"discovery",
		a.ledger.FillsThroughMS(),
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
	attempt.fills = sortedFillEvidence(merged)
	return nil
}

func (a *Account) pullFillEvidence(
	attempt *reconAttempt,
	kind string,
	startMS uint64,
	endMS uint64,
	merged map[uint64]ledger.FillEvidence,
	observed map[uint64]ledger.FillEvidence,
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
		var current ledger.FillEvidence
		current, err = a.fillEvidence(row)
		if err != nil {
			query.Error = err.Error()
			attempt.fillQueries = append(attempt.fillQueries, query)
			return err
		}
		var previous, seen = observed[current.VenueTID]
		if seen {
			var combined ledger.FillEvidence
			combined, err = mergeFillEvidence(previous, current)
			if err != nil {
				query.Error = err.Error()
				attempt.fillQueries = append(attempt.fillQueries, query)
				return err
			}
			observed[current.VenueTID] = combined
			merged[current.VenueTID] = combined
			query.FillsUnchanged++
			var existing, exists = a.ledger.Fill(current.VenueTID)
			if exists && !existing.HasFee {
				query.PendingMatched++
			}
			continue
		}
		var existing, exists = a.ledger.Fill(current.VenueTID)
		switch {
		case !exists:
			query.FillsAdded++
		case !existing.HasFee && current.Fee != nil:
			query.FeesEnriched++
			query.PendingMatched++
		case !existing.HasFee:
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

func (a *Account) updateFillRecords(attempt *reconAttempt) error {
	attempt.stage = "update_fills"
	var err = enrichFillCLOIDs(attempt.orders, attempt.fills)
	if err != nil {
		return err
	}
	var staged *ledger.ReconAttempt
	staged, err = a.ledger.PrepareRecon(ledger.ReconInput{
		FillsThroughMS:  attempt.nowMS,
		ObservedMS:      attempt.nowMS,
		AccountStateRaw: attempt.accountRaw,
	})
	if err != nil {
		return err
	}
	attempt.ledgerAttempt = staged
	err = a.ledger.UpdateReconFills(staged, attempt.fills)
	if err != nil {
		return err
	}
	attempt.fillChanges = a.ledger.ReconFillChanges(staged)
	return nil
}

func (a *Account) updateOrderRecords(attempt *reconAttempt) error {
	attempt.stage = "update_orders"
	return a.ledger.UpdateReconOrders(attempt.ledgerAttempt, attempt.orders)
}

func (a *Account) updateTradeRecords(attempt *reconAttempt) error {
	attempt.stage = "update_trades"
	return a.ledger.UpdateReconTrades(attempt.ledgerAttempt, a.markPrice())
}

func (a *Account) updateAccountSnapshot(attempt *reconAttempt) error {
	attempt.stage = "update_account"
	var current, err = a.ledger.ReconSummary(attempt.ledgerAttempt)
	if err != nil {
		return err
	}
	var positionSize decimal.Decimal
	var entryPrice decimal.Decimal
	positionSize, entryPrice, err = accountPosition(
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

func (a *Account) persistAndPublishRecon(attempt *reconAttempt) error {
	attempt.stage = "persist_publish"
	var err = a.ledger.CommitRecon(attempt.ledgerAttempt)
	if err != nil {
		return err
	}
	a.generation = attempt.snapshot.Generation
	a.lastSnapshot = attempt.snapshot
	a.logFillChanges(attempt.fillChanges)
	return nil
}

func (a *Account) logFillChanges(changes []ledger.FillChange) {
	for _, change := range changes {
		if change.Kind == ledger.FillAdded {
			a.log.Info(fmt.Sprintf(
				"fill added venue=simulator network=%s account=%s symbol=%s venue_tid=%d has_fee=%t",
				a.config.Network,
				a.config.Name,
				a.config.Symbol,
				change.VenueTID,
				change.HasFee,
			))
		} else if change.Kind == ledger.FillFeeEnriched {
			a.log.Info(fmt.Sprintf(
				"fill fee enriched venue=simulator network=%s account=%s symbol=%s venue_tid=%d previous=missing fee=%s",
				a.config.Network,
				a.config.Name,
				a.config.Symbol,
				change.VenueTID,
				change.Fee,
			))
		}
	}
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

func (a *Account) orderEvidence(
	current hyperliquid.OpenOrder,
	status order.Status,
	timestampMS uint64,
	raw string,
) (ledger.OrderEvidence, error) {
	if current.Coin != a.config.Symbol ||
		(current.Side != order.Buy && current.Side != order.Sell) ||
		(current.CLOID == "" && current.VenueOrderID == 0) || timestampMS == 0 {
		return ledger.OrderEvidence{}, fmt.Errorf(
			"download Order evidence: invalid official identity",
		)
	}
	var price, err = decimal.NewFromString(current.LimitPrice)
	if err != nil || !price.IsPositive() {
		return ledger.OrderEvidence{}, fmt.Errorf(
			"download Order evidence: invalid price",
		)
	}
	var size decimal.Decimal
	size, err = decimal.NewFromString(current.Size)
	if err != nil || !size.IsPositive() {
		return ledger.OrderEvidence{}, fmt.Errorf(
			"download Order evidence: invalid size %q",
			current.Size,
		)
	}
	if raw == "" {
		var payload []byte
		payload, err = hyperliquid.Encode(current)
		if err != nil {
			return ledger.OrderEvidence{}, err
		}
		raw = string(payload)
	}
	return ledger.OrderEvidence{
		CLOID:        current.CLOID,
		VenueOrderID: current.VenueOrderID,
		Status:       status,
		TimestampMS:  timestampMS,
		Raw:          raw,
	}, nil
}

func (a *Account) fillEvidence(
	execution hyperliquid.Fill,
) (ledger.FillEvidence, error) {
	if execution.Coin != a.config.Symbol ||
		(execution.Side != order.Buy && execution.Side != order.Sell) ||
		execution.Direction == "" {
		return ledger.FillEvidence{}, fmt.Errorf(
			"download Fill evidence: invalid official identity",
		)
	}
	var price, err = decimal.NewFromString(execution.Price)
	if err != nil || !price.IsPositive() {
		return ledger.FillEvidence{}, fmt.Errorf(
			"download Fill evidence: invalid price",
		)
	}
	var quantity decimal.Decimal
	quantity, err = decimal.NewFromString(execution.Size)
	if err != nil || !quantity.IsPositive() {
		return ledger.FillEvidence{}, fmt.Errorf(
			"download Fill evidence: invalid size",
		)
	}
	if _, err = decimal.NewFromString(execution.StartPosition); err != nil {
		return ledger.FillEvidence{}, fmt.Errorf(
			"download Fill evidence: invalid start position",
		)
	}
	if _, err = decimal.NewFromString(execution.ClosedPnL); err != nil {
		return ledger.FillEvidence{}, fmt.Errorf(
			"download Fill evidence: invalid closed PnL",
		)
	}
	var fee *decimal.Decimal
	if execution.Fee != nil {
		if execution.FeeToken == "" {
			return ledger.FillEvidence{}, fmt.Errorf(
				"download Fill evidence: missing fee token",
			)
		}
		var value decimal.Decimal
		value, err = decimal.NewFromString(*execution.Fee)
		if err != nil {
			return ledger.FillEvidence{}, fmt.Errorf(
				"download Fill evidence: invalid fee",
			)
		}
		fee = &value
	}
	var raw []byte
	raw, err = hyperliquid.Encode(execution)
	if err != nil {
		return ledger.FillEvidence{}, err
	}
	var liquidity = "maker"
	if execution.Crossed {
		liquidity = "taker"
	}
	return ledger.FillEvidence{
		CLOID:        execution.CLOID,
		VenueOrderID: execution.VenueOrderID,
		VenueTID:     execution.VenueTID,
		Side:         execution.Side,
		Quantity:     quantity,
		Price:        price,
		TimestampMS:  execution.TimestampMS,
		Fee:          fee,
		Liquidity:    liquidity,
		Raw:          string(raw),
	}, nil
}

func accountPosition(
	state hyperliquid.AccountState,
	symbol string,
) (decimal.Decimal, decimal.Decimal, error) {
	var size = decimal.Zero
	var entry = decimal.Zero
	var found = false
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

func enrichFillCLOIDs(
	orders []ledger.OrderEvidence,
	fills []ledger.FillEvidence,
) error {
	var cloids = make(map[uint64]string, len(orders))
	for _, current := range orders {
		if current.VenueOrderID == 0 || current.CLOID == "" {
			continue
		}
		var _, exists = cloids[current.VenueOrderID]
		if exists {
			return fmt.Errorf(
				"enrich Fill evidence: duplicate Venue OID %d",
				current.VenueOrderID,
			)
		}
		cloids[current.VenueOrderID] = current.CLOID
	}
	for index := range fills {
		var current = &fills[index]
		var value, exists = cloids[current.VenueOrderID]
		if !exists {
			continue
		}
		if current.CLOID == "" {
			current.CLOID = value
			continue
		}
		if current.CLOID != value {
			return fmt.Errorf(
				"enrich Fill evidence: Venue OID %d changed CLOID",
				current.VenueOrderID,
			)
		}
	}
	return nil
}

// Section 3 - Generic Helpers

func mergeFillEvidence(
	previous ledger.FillEvidence,
	current ledger.FillEvidence,
) (ledger.FillEvidence, error) {
	if previous.CLOID != current.CLOID || previous.VenueOrderID != current.VenueOrderID ||
		previous.VenueTID != current.VenueTID || previous.Side != current.Side ||
		!previous.Quantity.Equal(current.Quantity) || !previous.Price.Equal(current.Price) ||
		previous.TimestampMS != current.TimestampMS {
		return ledger.FillEvidence{}, fmt.Errorf(
			"merge Fill evidence: changed execution for Venue TID %d",
			current.VenueTID,
		)
	}
	if previous.Fee != nil && current.Fee != nil && !previous.Fee.Equal(*current.Fee) {
		return ledger.FillEvidence{}, fmt.Errorf(
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
	} else if current.Liquidity != "" && merged.Liquidity != current.Liquidity {
		return ledger.FillEvidence{}, fmt.Errorf(
			"merge Fill evidence: changed liquidity for Venue TID %d",
			current.VenueTID,
		)
	}
	if merged.Raw == "" {
		merged.Raw = current.Raw
	} else if current.Raw != "" && merged.Raw != current.Raw {
		return ledger.FillEvidence{}, fmt.Errorf(
			"merge Fill evidence: changed raw evidence for Venue TID %d",
			current.VenueTID,
		)
	}
	return merged, nil
}

func sortedFillEvidence(values map[uint64]ledger.FillEvidence) []ledger.FillEvidence {
	var ids = make([]uint64, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	var evidence = make([]ledger.FillEvidence, 0, len(ids))
	for _, id := range ids {
		evidence = append(evidence, values[id])
	}
	return evidence
}
