package ledger

import (
	"fmt"

	"nuubot/internal/fill"
	"nuubot/internal/order"
)

// Section 1 - Program Flow

// Recon atomically applies one complete normalized Venue observation.
func (l *Ledger) Recon(input ReconInput) error {
	if !l.started || l.stopped {
		return fmt.Errorf("reconcile ledger: invalid lifecycle state")
	}

	// index active local Orders
	if input.ObservedMS == 0 || input.FillsThroughMS < l.fillsThroughMS {
		return fmt.Errorf("reconcile ledger: invalid observation or backward Fill cursor")
	}
	var staged = cloneTrades(l.trades)
	var byCLOID = indexCLOIDs(staged)

	// match incoming Venue evidence
	for _, evidence := range input.Orders {
		var owned = byCLOID[evidence.CLOID]
		if owned == nil {
			continue
		}
		var err = owned.ApplyVenueState(order.VenueState{
			VenueOrderID: evidence.VenueOrderID,
			Status:       evidence.Status,
			RejectReason: evidence.RejectReason,
			TimestampMS:  evidence.TimestampMS,
			Raw:          evidence.Raw,
		})
		if err != nil {
			return fmt.Errorf("reconcile ledger: %w", err)
		}
	}

	// validate complete recon batch
	for _, evidence := range input.Fills {
		var owned = byCLOID[evidence.CLOID]
		if owned == nil {
			continue
		}
		var snapshot = owned.Snapshot()
		var err = owned.ApplyFill(fill.Input{
			LedgerID:     snapshot.LedgerID,
			TradeID:      snapshot.TradeID,
			OrderID:      snapshot.OrderID,
			Account:      snapshot.Account,
			CycleNumber:  snapshot.CycleNumber,
			Symbol:       snapshot.Symbol,
			CLOID:        snapshot.CLOID,
			VenueOrderID: evidence.VenueOrderID,
			VenueTID:     evidence.VenueTID,
			Side:         evidence.Side,
			Quantity:     evidence.Quantity,
			Price:        evidence.Price,
			TimestampMS:  evidence.TimestampMS,
			Fee:          evidence.Fee,
			Liquidity:    evidence.Liquidity,
			Raw:          evidence.Raw,
		})
		if err != nil {
			return fmt.Errorf("reconcile ledger: %w", err)
		}
	}
	for _, ownedTrade := range staged {
		var err = ownedTrade.Refresh()
		if err != nil {
			return fmt.Errorf("reconcile ledger: %w", err)
		}
		for _, ownedOrder := range ownedTrade.Orders() {
			if ownedOrder.Status == order.Filled &&
				!ownedOrder.FilledQuantity.Equal(ownedOrder.RequestedQuantity) {
				return fmt.Errorf(
					"reconcile ledger: filled Order %d lacks complete Fill evidence",
					ownedOrder.OrderID,
				)
			}
		}
	}

	// persist affected trees and cursor when configured
	var state = candidate{
		trades:          staged,
		nextTradeID:     l.nextTradeID,
		nextTradeNo:     l.nextTradeNo,
		nextOrderID:     l.nextOrderID,
		fillsThroughMS:  input.FillsThroughMS,
		lastReconMS:     input.ObservedMS,
		accountStateRaw: input.AccountStateRaw,
	}
	var err = l.persistCandidate(state)
	if err != nil {
		return err
	}

	// publish recon result
	l.trades = staged
	l.fillsThroughMS = input.FillsThroughMS
	l.lastReconMS = input.ObservedMS
	l.accountStateRaw = input.AccountStateRaw
	return nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
