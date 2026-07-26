package account

import (
	"encoding/json"
	"fmt"

	"nuubot/internal/ledger"
	"nuubot/internal/order"
)

// Section 1 - Program Flow

func (a *Account) reconcile2(nowMS uint64, forced bool) (Snapshot, bool, error) {
	if !a.started || a.stopped || nowMS == 0 {
		return Snapshot{}, false, fmt.Errorf("reconcile Account: invalid state or timestamp")
	}
	if !a.dirty && !forced {
		return a.lastSnapshot, false, nil
	}

	a.log.Info("account running Recon 2")

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
	err = a.ledger.Recon2(ledger.ReconInput{
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

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
