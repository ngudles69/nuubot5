package ledger

import (
	"fmt"

	"github.com/shopspring/decimal"

	"nuubot/internal/fill"
	"nuubot/internal/order"
	"nuubot/internal/trade"
)

// Section 1 - Program Flow

// Publish writes one complete terminal Ledger result.
func Publish(path string, result Result) error {
	var store, err = openLedgerStore(path)
	if err != nil {
		return err
	}
	defer store.close()

	// reconstruct validated Ledger evidence
	var state candidate
	state, err = candidateFromResult(result)
	if err != nil {
		return err
	}

	// write complete Ledger evidence
	err = store.save(result.Config, state)
	if err != nil {
		return err
	}
	return nil
}

// Section 2 - Domain Helpers

func candidateFromResult(result Result) (candidate, error) {
	var state = candidate{
		trades:          make(map[uint64]*trade.Trade, len(result.Trades)),
		nextTradeID:     1,
		nextTradeNo:     1,
		nextOrderID:     1,
		fillsThroughMS:  result.FillsThroughMS,
		lastReconMS:     result.LastReconMS,
		accountStateRaw: result.AccountState,
	}
	for _, tradeSnapshot := range result.Trades {
		var created, err = trade.New(tradeSnapshot.Input)
		if err != nil {
			return candidate{}, fmt.Errorf("publish Ledger: %w", err)
		}
		for _, orderSnapshot := range tradeSnapshot.Orders {
			var restored *order.Order
			restored, err = order.New(orderSnapshot.Input)
			if err != nil {
				return candidate{}, fmt.Errorf("publish Ledger: %w", err)
			}
			err = created.AddOrder(restored)
			if err != nil {
				return candidate{}, fmt.Errorf("publish Ledger: %w", err)
			}
			state.nextOrderID = max(state.nextOrderID, orderSnapshot.OrderID+1)
		}
		for _, orderSnapshot := range tradeSnapshot.Orders {
			var restored, exists = created.Order(orderSnapshot.OrderID)
			if !exists {
				return candidate{}, fmt.Errorf(
					"publish Ledger: Order %d was not admitted",
					orderSnapshot.OrderID,
				)
			}
			err = applyOrderSnapshot(restored, orderSnapshot)
			if err != nil {
				return candidate{}, err
			}
		}
		err = created.Refresh()
		if err != nil {
			return candidate{}, fmt.Errorf("publish Ledger: %w", err)
		}
		if created.State().Status != tradeSnapshot.Status {
			return candidate{}, fmt.Errorf(
				"publish Ledger: Trade %d status mismatch",
				tradeSnapshot.TradeID,
			)
		}
		state.trades[tradeSnapshot.TradeID] = created
		state.nextTradeID = max(state.nextTradeID, tradeSnapshot.TradeID+1)
		state.nextTradeNo = max(state.nextTradeNo, tradeSnapshot.TradeNo+1)
	}
	return state, nil
}

func applyOrderSnapshot(restored *order.Order, snapshot order.Snapshot) error {
	var err error
	if snapshot.Status != order.Created {
		err = restored.RecordSubmit(
			snapshot.VenueOrderID,
			snapshot.RejectReason,
			snapshot.Raw,
		)
		if err != nil {
			return fmt.Errorf("publish Ledger: %w", err)
		}
	}
	if snapshot.Status != order.Created && snapshot.Status != order.Submitted {
		var observedMS = max(snapshot.UpdatedMS, snapshot.TimestampMS)
		err = restored.ApplyVenueState(order.VenueState{
			VenueOrderID: snapshot.VenueOrderID,
			Status:       snapshot.Status,
			RejectReason: snapshot.RejectReason,
			TimestampMS:  observedMS,
			Raw:          snapshot.Raw,
		})
		if err != nil {
			return fmt.Errorf("publish Ledger: %w", err)
		}
	}
	for _, execution := range snapshot.Fills {
		var fee *decimal.Decimal
		if execution.HasFee {
			var value = execution.Fee
			fee = &value
		}
		err = restored.ApplyFill(fill.Input{
			LedgerID:     execution.LedgerID,
			TradeID:      execution.TradeID,
			OrderID:      execution.OrderID,
			Account:      execution.Account,
			CycleNumber:  execution.CycleNumber,
			Symbol:       execution.Symbol,
			CLOID:        execution.CLOID,
			VenueOrderID: execution.VenueOrderID,
			VenueTID:     execution.VenueTID,
			Side:         execution.Side,
			Quantity:     execution.Quantity,
			Price:        execution.Price,
			TimestampMS:  execution.TimestampMS,
			Fee:          fee,
			Liquidity:    execution.Liquidity,
			Raw:          execution.Raw,
		})
		if err != nil {
			return fmt.Errorf("publish Ledger: %w", err)
		}
	}
	if restored.Snapshot().Status != snapshot.Status {
		return fmt.Errorf(
			"publish Ledger: Order %d status mismatch",
			snapshot.OrderID,
		)
	}
	return nil
}

// Section 3 - Generic Helpers
