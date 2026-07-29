# Venue

Status: Implemented for simnet Accounts.
Covers: `internal/account`, `internal/venue`, `internal/hyperliquid`, and `internal/simulator`
Purpose: Separate Account domain state from official exchange operations.

## Ownership

```text
Account
└── Venue(config)
    ├── mainnet/testnet
    │   └── Exchange
    └── simnet/backtest
        └── simulated Venue/Exchange
```

Account owns one concrete Venue lifecycle.

Venue owns network selection and network-specific resources.

Venue owns Simulator for simnet.

Account connects to and disconnects from Venue.

Live Hyperliquid Venue behavior remains pending.

Venue is a pass-through routing boundary.

Venue contains no trading, cancellation, reconciliation, risk, or Account
business logic.

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

Set Leverage uses `hyperliquid.UpdateLeverageAction`.

Every batch item receives one ordered official status.

Malformed responses fail before acknowledgement advances.

Mutation responses acknowledge Venue admission.

Ledger lifecycle still advances through Recon.

## Queries

Get Open Orders, Get Order History, and Get Fill History are bulk official
calls.

Get Order Status queries exactly one Order by OID or CLOID.

Get Account State returns one official clearinghouse snapshot.

Each call constructs new JSON from current Venue truth.

Returned bytes never alias Venue memory.

## Implementation References

Nuubot5 source owns current Venue behavior.

Nuubot3 and NautilusTrader supply reusable intent only.

The function comparison is [NautilusTrader Intent Comparison](../nautilus.md).

Their useful read intent includes:

```text
GetOpenOrders
GetOrderHistory
GetFillHistory
GetOrderStatus
GetAccountState
```

Order History complements Open Orders and Fill history. It does not replace
either source.

NautilusTrader bulk Order reports are framework-normalized reports.

They are not a separate Hyperliquid endpoint from historical Orders.

Hyperliquid Active Asset Data is equivalent to Nuubot Meta intent.

Future `CancelAllOrders` is an Account convenience operation over known Orders.

It does not add cancellation policy or Order selection to Venue.

## Market Data

Venue exposes no BBO ingestion operation.

Simulator subscribes directly to shared MarketData and records private execution truth.

Simulator sends no callback or state to Account.

Account learns Simulator details only through Venue responses.

Clean reconciliation sweeps discover MarketData-driven Simulator changes every
60 seconds.


## Invariants

- Account owns Venue lifetime, not Venue truth.
- Venue owns accepted exchange state.
- Account never imports or calls Simulator.
- Simulator never receives Account or Ledger references.
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
