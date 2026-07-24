# Simulator Package

Status: Implemented for the first Simulator trading tranche. External parity remains pending.
Covers: `internal/simulator/*.go`
Purpose: Provide Hyperliquid-shaped simulated Venue truth for Account reconciliation.

## Canonical Sources

- `D:/rust/nuubot3/nuubot/exchange/simulator.py`
- `D:/rust/nuubot3/wiki/account/simulator.md`
- `D:/rust/nuutrader6/src/nuubot/hcbots/simulator.py`
- Installed `async_hyperliquid` 0.4.8 output shapes
- [Simulator Parity](../concepts/simulator-parity.md)

## Ownership

Account owns one concrete Simulator.

Simulator owns simulated Orders, Fills, positions, counters, and transient BBO state.

Simulator owns no Ledger, Trade, Order, or Fill domain object.

## Lifecycle

Simulator is one concrete implementation. No one-product Venue interface exists.

`Init` validates identity, policy, persistence mode, and optional persisted state.

`Stop` releases owned resources.

Constructors perform no network, storage, matching, or background work.

## Program Flow

```text
Init
  bind Simulator inputs
  validate Simulator identity
  validate Simulator policy
  validate persistence mode
  initialize Simulator
  open Simulator state when configured
  load Simulator state when configured
  admit Simulator lifecycle

PlaceOrders
  validate Venue requests
  stage recoverable Simulator mutation
  allocate staged Venue identities
  execute staged market-like Orders
  persist staged state when configured
  commit staged state
  return admitted SDK-shaped submit response

CancelOrders
  stage recoverable Simulator mutation
  cancel staged simulated Orders
  persist staged state when configured
  commit staged state
  return admitted SDK-shaped cancel response

IngestBBO
  validate BBO identity
  warm transient market state
  stage recoverable Simulator mutation
  match staged eligible Orders
  persist staged state when configured
  commit staged state
  report changed truth

Result
  copy identity policy and counters
  copy Orders and Fill history
  return immutable Simulator result

Stop
  stop Simulator
```

Source action comments follow this flow.

## Parity Layers

Nuutrader6 supplies matching behavior.

`async_hyperliquid` supplies the client-visible response contract.

Simulator returns the same admitted shapes as the Hyperliquid adapter.

| Operation | Shape |
|---|---|
| Submit batch | `status`, `response.type`, ordered `statuses` |
| Cancel batch | `status`, `response.type`, ordered `statuses` |
| Open Orders | Hyperliquid frontend Order rows |
| Order status | Hyperliquid Order status envelope |
| User Fills | Hyperliquid Fill rows |
| Account state | Hyperliquid clearinghouse state |

Internal response meanings and ordered statuses follow the admitted reference.

Recorded external JSON parity remains pending.

The official API controls exchange semantics.

Account consumes the same public contract from both Venue implementations.

Simulator diagnostics remain outside public responses.

## Bracket States

| Event | Entry | TP | SL |
|---|---|---|---|
| Resting submission | `resting` | `waitingForFill` | `waitingForFill` |
| Immediate entry Fill | `filled` | `waitingForTrigger` | `waitingForTrigger` |
| Later entry Fill | `filled` | `waitingForTrigger` | `waitingForTrigger` |
| TP Fill | `filled` | `filled` | `canceled` |
| SL Fill | `filled` | `canceled` | `filled` |

Public Order queries may expose waiting children as `open`.

Internal activation must not add unsupported public history rows.

Parent cancellation cancels both waiting children.

## Matching Rules

- The first BBO only warms transient market state.
- Resting Orders cannot consume an already-processed BBO.
- Explicit market-like IOC Orders may execute from their supplied current reference.
- Later BBOs fill crossed resting Orders.
- Regular buy limits cross when ask is at or below requested price.
- Regular sell limits cross when bid is at or above requested price.
- TP and SL use trigger or requested price as Fill basis.
- Adverse slippage applies to the Fill basis.
- The first slice has no partial depth fills.
- One BBO fills at most one leg per Trade.
- Entry execution arms TP and SL children.
- TP or SL execution cancels its sibling.
- Reduce-only execution cannot increase or reverse exposure.
- Invalid reduce-only Orders cancel without a Fill.

## Fill Evidence

Simulator Fill rows include:

```text
coin
px
sz
side
time
startPosition
dir
closedPnl
hash
oid
crossed
fee
tid
cloid
feeToken
twapId
```

`closedPnl` is gross realized price PnL before fees.

Fees remain a separate field.

Ledger calculates its own domain PnL.

## Account State

Simulator reports Hyperliquid-shaped margin summaries and positions.

Position size is signed.

Position value uses the latest admitted BBO midpoint.

Account value includes realized PnL, unrealized PnL, and fees.

A recovered open position requires one fresh BBO before marked state.

## Persistence

Simulator state is one versioned JSON payload keyed by Ledger identity.

The payload contains identity, policy, counters, Orders, Order history, and Fill history.

The last BBO is transient and never restored.

Identity, version, and policy mismatch fail without overwriting state.

See [Trading Schema](../concepts/trading-schema.md).

Account passes the configured `persist_mode`.

`none` keeps state in memory until one successful final export.

`max` persists every Simulator state change.

Account passes store operations only for `max`.

Simulator never detects Runner, Sweep, paper, or live mode.

For `max`, durable SQLite success precedes publication of changed in-memory state.

Persistence failure leaves both memory and recoverable state at the prior version.

After loading persisted state, Account requires a fresh BBO before a marked snapshot.

## Terminal Result

`Result` returns one immutable `simulator.Result`.

It contains identity, policy, counters, Orders, Order history, and Fill history.

Every slice and map is newly owned. No value aliases mutable Simulator state.

## Does Not

- Mutate Ledger.
- Call Executor policy.
- Reconcile Account state.
- Import the Python client.
- Contact Hyperliquid.
- Hide corrupt state with defaults.
- Use BBO price as every Fill price.

## Required Proof

- Submit, cancel, Order, Fill, and account-state fixtures match admitted reference shapes.
- First BBO warms without matching.
- Resting Orders wait for a later BBO.
- Market-like IOC submission may fill immediately.
- TP and SL activation and OCO behavior are correct.
- Resting and immediate-filled batch statuses retain request order.
- Public history contains no Simulator-only activation event.
- Controlled testnet proof confirms TP, SL, and sibling-cancellation behavior.
- Reduce-only execution never reverses exposure.
- Duplicate BBO without new Order state performs no durable mutation.
- Persisted state round-trips exactly.
- Corrupt state fails without replacement.
