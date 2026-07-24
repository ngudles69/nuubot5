package simulator

// Section 1 - Program Flow

// Publish writes one complete terminal Simulator result.
func Publish(path string, result Result) error {
	var store, err = openSimulatorStore(path)
	if err != nil {
		return err
	}
	defer store.close()

	// write complete Simulator evidence
	return store.save(result.Config, storedState{
		SchemaVersion:    simulatorSchemaVersion,
		LedgerID:         result.LedgerID,
		Name:             result.Name,
		Account:          result.Account,
		CycleNumber:      result.CycleNumber,
		Symbol:           result.Symbol,
		Equity:           result.Equity.String(),
		FeePct:           result.FeePct.String(),
		SlippagePct:      result.SlippagePct.String(),
		NextVenueOrderID: result.NextVenueOrderID,
		NextVenueTID:     result.NextVenueTID,
		OrderHistory:     append([]OrderState(nil), result.Orders...),
		Fills:            append([]FillState(nil), result.Fills...),
	})
}

// Section 2 - Domain Helpers

// Section 3 - Generic Helpers
