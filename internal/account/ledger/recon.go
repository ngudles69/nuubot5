package ledger

import (
	"fmt"
	"sort"

	"nuubot/internal/account/fill"
)

// Section 1 - Program Flow

// UpdateReconFill applies one downloaded Fill or defers unknown OID-only evidence.
func (l *Ledger) UpdateReconFill(input fill.Fill) (bool, bool, error) {
	// Step 1: defer OID-only evidence without a current Order index
	if input.CLOID == "" && input.VenueOrderID != 0 {
		if _, found := l.orderByOID[input.VenueOrderID]; !found {
			return false, true, nil
		}
	}

	// Step 2: apply resolvable Fill evidence
	var changed, err = l.AddFill(input)
	if err != nil {
		return false, false, err
	}
	return changed, false, nil
}

// ReconOIDSearch resolves deferred OID-only Fills after Order updates.
func (l *Ledger) ReconOIDSearch(
	deferred []fill.Fill,
) ([]fill.Fill, []uint64, int, error) {
	// Step 1: search updated Order OID indexes
	var orderIDs = make(map[uint64]struct{})
	var matched = make([]fill.Fill, 0, len(deferred))
	var unmatchedFills int
	for _, current := range deferred {
		if current.CLOID != "" || current.VenueOrderID == 0 {
			return nil, nil, 0, fmt.Errorf("search Recon OIDs: invalid deferred Fill")
		}
		var reference, found = l.orderByOID[current.VenueOrderID]
		if !found {
			unmatchedFills++
			continue
		}
		orderIDs[reference.OrderID] = struct{}{}
		matched = append(matched, current)
	}

	// Step 2: return matched Fills and distinct owning Orders
	var orders = make([]uint64, 0, len(orderIDs))
	for orderID := range orderIDs {
		orders = append(orders, orderID)
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i] < orders[j]
	})
	return matched, orders, unmatchedFills, nil
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
