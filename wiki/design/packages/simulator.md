# Simulator Package

Status: Canonical exchange state and official JSON boundary implemented. External parity remains pending.
Covers: `internal/simulator/*.go`
Purpose: Simulate one Hyperliquid Venue without sharing Account domain state.

## Ownership

Account owns Simulator lifetime.

Simulator owns:

- accepted official Orders;
- Venue-assigned OIDs and TIDs;
- canonical Fills;
- private batch, waiting, and arming state;
- active Order indexes;
- position and finance state;
- matching policy;
- one MarketData subscription;
- transient BBO state; and
- schema version 3 persistence.

Simulator owns no Ledger, Trade, domain Order, domain Fill, role, or purpose.

## Inputs

`Config` contains simulated Venue identity, policy, one exact MarketData key, and one narrow Account change callback:

```text
account
asset
symbol
equity
fee percent
slippage percent
persist mode
store path
```

Place receives `hyperliquid.PlaceOrderAction`.

Cancel receives `hyperliquid.CancelByCLOIDAction`.

Simulator subscribes directly to MarketData during Init and reads the latest buffered BBO inside its callback.

CLOID is mandatory, shape-validated, stored unchanged, and never domain-decoded.

No caller supplies OID.

## Testing

Simulator tests are Venue parity tests.

They prove official request and response shape, canonical Order and Fill state,
matching, cancellation, position and finance behavior, detached JSON,
persistence, failure atomicity, and exact comparison mechanics.

External Hyperliquid fixture and testnet parity remain pending.

## Canonical State

Each accepted Order creates one private `simOrder`.

OID and CLOID indexes point to that record.

The active index points to the same record while it can match.

Arming, Fill, and cancellation mutate that record.

Terminal mutation removes it from the active index.

Terminal Orders never match again.

Each execution appends one private `simFill`.

Simulator never creates detached private Order history copies.

## Program Flow

```text
Init
  validate Simulator config
  initialize Simulator state
  restore durable Simulator state when configured
  mark Simulator started
  subscribe to MarketData
  read latest BBO

PlaceOrders
  validate official Order action
  stage Order mutation
  match marketable Orders
  persist and publish Order mutation
  return official Order response

CancelOrders
  validate official cancel action
  stage cancel mutation
  persist and publish cancel mutation
  return official cancel response

onBBO
  normalize BBO
  warm initial BBO state
  stage BBO matching
  persist changed Venue truth
  publish BBO outcome

Stop
  ignore repeated stop
  stop MarketData subscription
  persist Simulator state
  close Simulator store
  mark Simulator stopped
```

## Private Bracket State

`normalTpsl` groups one submitted batch privately.

Limit entry Orders start active.

Trigger children wait privately when their entry is present.

Entry Fill arms its trigger children.

TP or SL Fill cancels its sibling.

Public submit statuses remain official `resting`, `filled`, or `error`.

Private waiting and arming never become custom public statuses.

## Matching

- The first BBO warms market state.
- Marketable GTC and IOC Orders may fill during submission.
- Resting Orders match later crossing BBOs.
- Limit, TP, and SL crossings preserve existing exact-decimal behavior.
- Adverse slippage applies to the Fill basis.
- Reduce-only execution cannot open or reverse exposure.
- Unsupported reduce-only execution cancels without a Fill.
- One submitted private batch fills at most one leg per BBO.

Each canonical Order owns one transient exact comparison key.

Simulator builds one matching key per admitted BBO.

Matching-key comparison performs no allocation and changes no official value.

## Official Responses

Simulator returns fresh detached JSON for:

- submit acknowledgement;
- cancel acknowledgement;
- bulk open Orders;
- exact Order status;
- bounded Fill history; and
- clearinghouse state.

Submit returns each Venue-assigned OID in request order.

Open rows use remaining positive size.

Terminal exact status retains submitted size and terminal status.

Fill rows use OID and TID. They need not expose CLOID.

Clearinghouse time uses the latest canonical Venue event timestamp.

Returned bytes never alias canonical state.

## Position

Simulator updates signed size, entry price, realized PnL, and fees once per accepted Fill.

Reduce-only sizing reads this maintained state.

Clearinghouse output uses the latest admitted BBO for marked values.

A recovered open position requires one fresh BBO.

## Persistence

Simulator owns `simulator_venue_state`.

Its primary key is official simulated account plus symbol.

Schema version 3 stores:

- official identity and policy;
- Venue counters;
- latest canonical Venue timestamp;
- each canonical Order once; and
- each canonical Fill once.

No Ledger foreign key or local domain identity enters this payload.

Legacy Simulator payloads are not loaded or adapted.

`none` keeps state in memory.

`max` persists before changed memory becomes visible.

Persistence failure leaves memory and durable truth unchanged.

The last BBO remains transient.

Reload reconstructs indexes and maintained position from canonical records.

## No Result Escape

Simulator exposes no terminal `Result` through Account.

ResultPublisher receives reconciled Ledger evidence only.

Simulator private counters, records, and persistence never cross that boundary.

## Required Proof

- Official action validation.
- Mandatory opaque CLOID.
- Ordered Venue OID assignment.
- Detached response bytes.
- Canonical bracket mutation.
- One Fill per execution.
- Terminal no-rematch.
- Reduce-only protection.
- Position replay equality.
- Version 3 round-trip.
- Durable failure atomicity.
- Frozen official response fixtures.
- Controlled testnet parity.
