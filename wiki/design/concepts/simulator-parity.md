# Simulator Parity

Status: Internal lifecycle and official JSON boundary implemented. External fixture and testnet proof remain pending.
Covers: `internal/simulator`, `internal/hyperliquid`, and `internal/account`
Purpose: Keep simulated exchange truth separate from Nuubot domain truth.

## Contracts

```text
Official API
  request fields
  response JSON
  status meaning

Simulator
  matching
  canonical exchange state
  fresh official responses

Account and Ledger
  Nuubot ownership
  reconciliation
  finance
```

These contracts share values.

They never share objects or memory.

## Submission

Account submits official asset, side, price, size, reduce-only, order type, grouping, CLOID, and timestamp values.

CLOID is mandatory and opaque.

Simulator assigns OID once.

The ordered official response returns `resting`, `filled`, or `error`.

Private waiting and arming state never appears as a custom public status.

Waiting trigger children appear as official open Orders.

## Canonical Truth

Simulator keeps one canonical private Order record per accepted Order.

Matching, arming, filling, and cancellation update that record.

OID and CLOID indexes point to the same record.

The active index contains only matchable nonterminal Orders.

Each execution creates one canonical Fill record.

Terminal Orders leave the active index and never match again.

## Bracket Lifecycle

```text
submit entry, TP, SL
  Venue assigns three OIDs
  entry is active
  TP and SL wait privately

entry fills
  same entry record becomes filled
  TP and SL arm privately

TP or SL fills
  same child record becomes filled
  sibling record becomes canceled
  position becomes flat
```

Parent cancellation cancels waiting children.

Reduce-only execution cannot increase or reverse exposure.

## Public Boundary

Every mutation and query returns fresh detached official JSON.

Account decodes and validates that JSON.

Simulator exposes no `Result`, history slice, map, pointer, or diagnostic object through Account.

Open Orders return active official rows.

Order status returns one official envelope.

Fill history returns official rows and may omit CLOID.

Account state returns one official clearinghouse snapshot.

## Identity Resolution

Submit binds each request CLOID to the Venue-assigned OID.

Recon uses CLOID when an official row supplies it.

Fill rows may contain only OID and TID.

Account therefore resolves OID when CLOID is absent.

Conflicting CLOID and OID fail validation.

## Persistence

Simulator owns schema version 3.

One row is keyed by official simulated account and symbol.

The payload stores canonical Orders once, canonical Fills once, counters, and policy identity.

It stores no Ledger, Trade, local Order, role, or purpose identity.

Legacy Simulator payloads are not read or adapted.

The last BBO remains transient.

## Testing Boundary

`internal/simulator/simulator_test.go` is the internal Simulator parity suite.

It proves Simulator Venue semantics and pure exact comparison mechanics.

It does not replace Account-to-Ledger integration or complete Bot system proof.

## Current Proof

- Official request inputs contain no Account domain identity.
- Arbitrary shape-valid CLOID passes without domain decoding.
- Venue assigns ordered OIDs.
- Detached JSON mutation cannot alter Simulator truth.
- Bracket Fill and sibling cancellation update canonical records.
- Terminal Orders cannot create duplicate Fills.
- Version 3 persistence round-trips each record once.
- Failed durable mutation does not change memory truth.

External frozen-output and Hyperliquid testnet parity remain pending.

## Does Not Claim

- Mainnet mutation proof.
- Queue-position simulation.
- Order-book-depth partial fills.
- Exact fees, balances, or identifiers from testnet.
- Python implementation parity.
