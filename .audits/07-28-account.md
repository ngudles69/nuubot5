# Account Review

Date: 2026-07-28
Status: Approved for implementation after the Account-stack redesign stabilizes.
Scope: `internal/account/account.go`, `internal/account/recon.go`, and direct callers.

## Boundary

The separate Account-stack session owns Ledger, Trade, Order, and Fill redesign.

This review owns the Account menu, lifecycle organization, Order preparation, and direct caller cleanup.

Re-read current source and callers before implementation because the Account-stack redesign may change types and APIs.

## Approved Ownership

```text
Executor
  -> supplies Order intent and OrderLevel

Account
  -> assigns Trade and batch identity
  -> creates the complete CLOID
  -> creates local Order intent
  -> creates the Venue request
  -> submits the Venue request

cloid
  -> performs mechanical encoding only
```

Executor does not construct CLOIDs.

Account owns dirty state. Ledger reports pending evidence only.

```text
Account dirty = Account local dirty state || Ledger pending evidence
```

Do not add a public `IsDirty` method without a confirmed caller.

## Keep in Account

- `Init` — initialize Account, Ledger, and Venue.
- `PlaceOrders` — prepare, persist, submit, and record one Order batch.
- `CancelOrders` — validate and cancel owned active Orders.
- `Result` — return terminal Account evidence.
- `Telemetry` — return the latest successful Account Snapshot.
- `Stop` — stop Venue, Ledger, and Account.
- `markPrice` — read the latest buffered MarketData price.
- `createCLOID` — create complete Account-owned submission identity.
- `venueOrderRequest` — translate normalized Order intent into a Venue request.
- `venueGrouping` — select the official Venue grouping.
- `Trade` — provide focused Trade state required by Executors.
- `OpenTrades` — provide exposure required by Grid shutdown.
- `ActiveOrders` — provide Order identities required by current shutdown behavior.

## Move to Recon

Keep the reconciliation process in `internal/account/recon.go`:

- `Reconcile`
- `ReconciliationTelemetry`
- `ReconStats`
- `ReconOutcome`
- `FillQueryTelemetry`
- `ReconTelemetry`
- reconciliation attempt types and helpers

Move `Init` from `recon.go` to `account.go`. It initializes the complete Account, not reconciliation.

## Rename or Adjust

- Rename `normalizeSpecs` to `prepareOrderBatch`.
- Keep validation, rounding, minimum-notional enforcement, and batch shaping together unless implementation proves a simpler split.
- Move `CountOrders` into terminal Account or Ledger results when practical. Its only production use is Grid round-trip reporting.

## Remove

- `Result.Clone` — currently returns the same value and performs no clone.
- `copySpec` — redundant because preparation replaces returned price pointers without mutating caller values.
- `TradeOrders` — no callers.
- `Order` — used only by Account tests; tests can inspect Account-owned Ledger state directly.
- `Fill` — used only by Account tests; tests can inspect Account-owned Ledger state directly.
- `PositionQuantity` — no callers; Snapshot already owns the value.
- `HasPendingRecon` — no external callers; Account reconciliation reads Ledger pending evidence directly.

Test-only access does not justify a production Account menu function. Same-package Account tests may traverse Account-owned Ledger records and values directly.

## Current Caller Reasons

- `ActiveOrders` — Trade and Grid shutdown cancellation and flatness proof.
- `Trade` — Trade completion and Grid-level state refresh.
- `OpenTrades` — Grid shutdown exposure closure.
- `CountOrders` — Grid terminal round-trip count.
- `TradeOrders` — no callers.
- `Order` — Account tests only.
- `Fill` — Account tests only.
- `PositionQuantity` — no callers.
- `HasPendingRecon` — no external callers.

## Lifecycle Decision Pending

Account currently marks itself started during `Init` and has no separate `Start` method.

Decide during implementation whether Account needs a distinct `Start`. Do not add lifecycle ceremony without separate startup work.

## Proof Required

- Re-trace all callers after the Account-stack redesign lands.
- Preserve Place, Cancel, shutdown, and reconciliation behavior.
- Run `gofmt`.
- Run focused Account and Executor proof with `-tags noasm`.
- Run full Go tests and vet with `-tags noasm`.
- Run Observer, Trade, and Grid system proof when behavior-facing APIs change.
