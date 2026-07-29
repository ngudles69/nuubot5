# Venue Package

Status: Simnet implemented. Mainnet and testnet pending.
Covers: `internal/venue/*.go`
Purpose: Own one configured execution path behind Account.

## Ownership

Account owns one concrete Venue lifecycle.

Venue owns network selection and network-specific resources.

For simnet, Venue owns one Simulator.

For future mainnet or testnet, Venue will own Exchange transport,
authentication, signing, and protocol events.

Account never accesses Simulator directly.

Simulator never receives Account or Ledger references.

## Interface

Venue exposes:

```text
Connect
PlaceOrders
CancelOrders
SetLeverage
GetOpenOrders
GetOrderHistory
GetFillHistory
GetOrderStatus
GetAccountState
Disconnect
```

Venue returns detached Hyperliquid protocol JSON.

Account validates and reconciles those responses identically across networks.

Nuubot5 source owns current behavior.

Nuubot3 and NautilusTrader provide reusable intent only.

Venue routes calls. It contains no Account or trading business logic.

## Simnet

Simulator subscribes directly to MarketData.

Simulator updates only Simulator-owned Order, Fill, position, and finance truth.

Simulator does not push changes to Account.

Account discovers changes through reconciliation.

Clean Accounts sweep every 60 seconds.

## Pending

Mainnet, testnet, credentials, signing, and WebSocket events remain unimplemented.

Future Simulator events must use the approved Venue protocol event path.
