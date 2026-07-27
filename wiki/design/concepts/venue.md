# Venue

Status: Implemented for Simulator-backed Accounts.
Covers: `internal/account`, `internal/hyperliquid`, and `internal/simulator`
Purpose: Separate Account domain state from official exchange operations.

## Ownership

Account owns one Venue lifecycle.

Account and Venue remain separate entities.

Account owns the smallest interface it consumes:

```go
type venue interface {
    PlaceOrders(hyperliquid.PlaceOrderAction, uint64) ([]byte, error)
    CancelOrders(hyperliquid.CancelByCLOIDAction, uint64) ([]byte, error)
    OpenOrders(string) ([]byte, error)
    Fills(string, uint64, uint64) ([]byte, error)
    OrderStatus(string, string) ([]byte, error)
    AccountState(string) ([]byte, error)
    Stop() error
}
```

Simulator is the implemented Venue.

Live Hyperliquid remains pending.

## Boundary

Account sends only official operation values.

Venue returns fresh detached official JSON.

Account validates JSON through `internal/hyperliquid`.

Account then translates validated values into Ledger evidence.

These values never cross into Venue:

- Ledger, Trade, Order, or Fill objects;
- Ledger, Trade, or Account-owned Order IDs;
- Order roles or strategy purpose;
- Account scratch, callbacks, caches, or storage; and
- caller-owned response buffers.

Venue responses never expose private Simulator objects, indexes, persistence, or diagnostics.

## Identity

Every Nuubot Order request carries one mandatory official CLOID.

Venue validates CLOID shape and stores the value unchanged.

Venue treats CLOID as opaque.

Venue never decodes Nuubot identity from CLOID.

Venue assigns OID once after accepting the request.

The ordered submit response binds that OID to its request.

Account resolves acknowledgement by response order and CLOID.

Recon resolves evidence CLOID-first when CLOID exists.

Recon falls back to OID when official evidence omits CLOID.

If both exist, Account and Ledger require one consistent Order.

## Mutations

Place uses `hyperliquid.PlaceOrderAction`.

Cancel uses `hyperliquid.CancelByCLOIDAction`.

Every batch item receives one ordered official status.

Malformed responses fail before acknowledgement advances.

Mutation responses acknowledge Venue admission.

Ledger lifecycle still advances through Recon.

## Queries

Open Orders and Fill history are bulk official calls.

Exact Order status is exception handling for selected active Orders missing from the bulk response.

Account state is one official clearinghouse snapshot.

Each call constructs new JSON from current Venue truth.

Returned bytes never alias Venue memory.

## Market Data

Venue exposes no BBO ingestion operation.

Simulator subscribes directly to shared MarketData and records private execution truth.

Account receives only a narrow dirty-state notification and learns Venue details through official responses.

Live Venue implementations require no empty market-data method.


## Invariants

- Account owns Venue lifetime, not Venue truth.
- Venue owns accepted exchange state.
- CLOID is mandatory and opaque.
- OID is Venue-assigned.
- Official JSON remains untrusted until Account validates it.
- No implementation silently falls back to another Venue.
- No private state crosses either direction.
- Recon2 is retired.

## Does Not

- Mutate Ledger or domain objects.
- Decide Executor or Risk policy.
- Poll Account state.
- Cache public responses.
- Accept caller response storage.
- Fabricate domain identity.
