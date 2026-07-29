# Account Stack Coherency Audit

Date: 2026-07-29

## Verdict

FAIL.

The five migrated domain files are not internally coherent yet.

Reviewed:

- `internal/account/account.go`
- `internal/account/ledger/ledger.go`
- `internal/account/trade/trade.go`
- `internal/account/order/order.go`
- `internal/account/fill/fill.go`

External callers were excluded.

## Blockers

### 1. Account calls missing Ledger batch API *FIXED*

Severity: blocker.

Location: `internal/account/account.go:234`.

`Account.PlaceOrders` called `Ledger.NextBatchNo`.

Fixed by removing the non-domain Batch identity.

One transient Order set belongs to one Trade. Each Order uses its canonical
`OrderID`, and CLOID encodes `(LedgerID, OrderID)`.

### 2. Account reads removed Trade field *FIXED*

Severity: blocker.

Location: `internal/account/account.go:250`.

Account reads `tradeInput.TradeNo`.

The flat Trade field is `TradeNumber`.

Fixed: Account now reads the flat Trade field `TradeNumber`.

### 3. Ledger calls missing Store method *FIXED*

Severity: blocker.

Location: `internal/account/ledger/ledger.go:343`.

`Ledger.Persist` calls `ledgerStore.persist`.

That Store method does not exist yet.

Fixed: removed all persistence plumbing from the Account stack for later redesign.

### 4. Maximum persistence has no initialized Store *FIXED*

Severity: blocker.

Location: `internal/account/ledger/ledger.go:125-155`.

`Ledger.Init` initializes memory only.

It never opens, assigns, or loads `l.store`.

Fixed: removed Store ownership and persistence configuration from Ledger.

### 5. Account has no lifecycle initialization inside the reviewed stack *FIXED*

Severity: blocker.

Location: `internal/account/account.go:183-201`.

`PlaceOrders` and `CancelOrders` require `a.started`.

Fixed by verification: `Account.Init` initializes Ledger, initializes Venue,
then calls `initializeAccount`.

`initializeAccount` sets `dirty` and `started` only after both children
initialize.

No production change was required.

### 6. New Fill can lose valid Venue OID *FIXED*

Severity: blocker.

Location: `internal/account/ledger/ledger.go:301`.

Ledger overwrote Fill evidence with the parent Order OID.

Fixed: Ledger no longer copies parent OID into Fill evidence.

An Exchange-supplied Fill OID remains unchanged.

Fill OID is optional when CLOID resolves the parent Order.

Fill CLOID is optional when Exchange OID resolves the parent Order.

Ledger preserves both values exactly and rejects unknown or conflicting
supplied identity.

### 7. Order update can mutate before duplicate-OID rejection *FIXED*

Severity: blocker.

Location: `internal/account/ledger/ledger.go:212-244`.

Ledger called `Order.Update` before checking whether another Order owned the
new OID.

Fixed: Ledger now validates every proposed OID against existing indexes and
every other update before mutating any Order.

Changed OID proposals for one Order also fail during the validation pass.

### 8. New Order batch allows duplicate CLOIDs *FIXED*

Severity: blocker.

Location: `internal/account/ledger/ledger.go:537-551`.

Validation checks existing indexes only.

Two new Orders in the same batch may share one CLOID and overwrite the index.

Fixed: Ledger rejects duplicate CLOIDs and Order IDs inside the incoming batch.

### 9. Closed Trade count is wrong *FIXED*

Severity: blocker.

Location: `internal/account/ledger/ledger.go:650`.

Ledger calculates closed Trades as total minus active.

This counts canceled and error Trades as completed round trips.

Fixed: Ledger counts only Trades whose status is exactly `trade.Closed`.

### 10. Pending Order count omits partial Orders *FIXED*

Severity: major.

Location: `internal/account/ledger/ledger.go:495-505`.

Ledger counts pending Orders only when status is `filled`.

Partially filled Orders with missing fees are omitted.

Fixed: Ledger counts every Order with pending fees or unresolved filled completion.

### 11. Ledger refresh has a dead parameter *FIXED*

Severity: minor.

Location: `internal/account/ledger/ledger.go:604`.

`refreshTrade` accepts `*Summary` and discards it.

Fixed: deleted the parameter and updated every internal call.

### 12. Fill lookup index duplicates canonical Fill storage *FIXED*

Severity: minor.

Location: `internal/account/ledger/ledger.go:89-104`.

`fills` is already keyed by Venue TID.

`fillByTID` is written but never read.

Fixed: deleted `FillRef` and `fillByTID`.

Superseded after the canonical `FillID` decision.

Ledger now stores Fills by `FillID` and requires a VenueTID-to-FillID index.

### 13. Trade errors violate lowercase error contract *FIXED*

Severity: minor.

Location: `internal/account/trade/trade.go:186,201,232`.

Three returned errors start with uppercase domain names.

Fixed: all three errors now use lowercase text.

## Internally Coherent Contracts

- Fill has one flat representation.
- Order has one flat representation and owns no Fills.
- Trade has one flat representation and owns no Orders.
- Ledger owns the flat record maps.
- Account returns flat Order and Trade values.
- No compatibility aliases, wrappers, alternate records, or clone paths remain in the reviewed files.
- `HasFee`, Fill fee completion, Order closure, and Trade finance use the new flat fields.

## Requested Deletion Inventory

Delete after explicit execution confirmation:

```text
internal/account/account_test.go
internal/account/venue_test.go
internal/account/fill/fill_test.go
internal/account/order/order_test.go
internal/account/trade/trade_test.go
internal/account/ledger/ledger_test.go
```

## Proof Checked

- Read all five named production files.
- Compared every cross-package call against current declarations.
- Scanned for removed types, methods, and fields.
- Checked current formatting and whitespace.

## Proof Missing

- No package compilation.
- No external caller assessment.
- Store, Recon, Venue, executor, and result-publication coherence remain deferred.

## Assumptions

- Temporary external breakage is approved.
- Store remains outside this repair pass.
- The next pass repairs only internal coherence among the five named files.

## Open Questions

- Should `FillRef` remain for future Store or Recon use despite duplicating identity already stored on Fill?

## Bloat Check

No fake server, mock production path, fallback, compatibility bridge, or temporary adapter exists.

One dead parameter and one unused duplicate Fill index exist.

The main risks are incomplete wiring, mutation ordering, incorrect counts, and unfinished persistence ownership.
