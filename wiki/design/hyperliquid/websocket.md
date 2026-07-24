# Hyperliquid WebSocket

Status: Approved design. Implementation pending.

Covers: `internal/hyperliquid` WebSocket transport.

Purpose: Maintain recoverable Hyperliquid streams using `gorilla/websocket`.

## In

- Explicit mainnet or testnet WebSocket URL.
- Connect and reconnect.
- Ping and Pong heartbeat.
- Read-deadline drop detection.
- Socket error detection.
- Bounded reconnect backoff.
- Backoff reset after a stable connection.
- Automatic restoration of active subscriptions.
- Context cancellation.
- Clean shutdown.
- Typed message decoding.

## Out

- Runtime decisions.
- Account reconciliation.
- Ledger mutation.
- Order placement or cancellation.
- Shared subscriber policy.
- Bot-local feed state.
- Callbacks executed while locks are held.
- Unbounded sleeps or reconnect loops.
- Panics or fatal process exits.

## Ownership

One supervisor owns each socket and its reconnect loop.

One reader reads each active socket.

One serialized writer sends subscriptions, Ping, and close frames.

Application owners provide desired subscriptions and consume typed events.

## Lifecycle

```text
start
  connect
  restore active subscriptions
  read events
  send heartbeat
  detect timeout, close, or read failure
  close failed socket
  wait bounded backoff
  reconnect

stop
  cancel supervisor
  send close frame when possible
  close socket
  wait for owned goroutines
```

## Concurrency

Callbacks run outside locks.

Subscription state is copied before dispatch or restoration.

Close is idempotent.

No goroutine survives successful shutdown.

## Required Proof

- Successful connection receives typed events.
- Missing Pong or read traffic triggers reconnect.
- Socket closure triggers one reconnect loop.
- Active subscriptions restore once.
- Stable connection resets backoff.
- Cancellation interrupts backoff and reads.
- Shutdown leaves no owned goroutines.
- Slow callbacks cannot deadlock subscription or shutdown.
