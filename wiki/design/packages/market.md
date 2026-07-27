# Market Package

Status: Implemented.
Covers: `internal/market/market.go`, `internal/market/marketdata.go`
Purpose: Validate BBO values and own shared latest-value buffers and exact-key subscriptions.

## Scope

MarketData identifies one stream by:

```text
Venue + Network + Symbol
```

Physical Account is not market identity.

## Owner and Children

BtBot or Runner owns one MarketData object.

Nuubot carries its shared reference.

MarketData owns latest BBO values and subscription registrations.

## Responsibilities

- Validate finite positive BBO timestamp and price.
- Reject mismatched symbols and backward timestamps.
- Publish the latest detached BBO before callbacks run.
- Notify matching subscribers synchronously in registration order.
- Return callback failures to the producer.
- Remove subscriptions idempotently.

## Does Not

- Own transport, replay, strategy, matching, Account state, or persistence.
- Represent separate bid and ask values.
- Coalesce or silently lose subscribed updates.

## Lifecycle

```text
CreateMarketData
IngestBBO
LatestBBO or SubscribeBBO
Subscription.Stop
MarketData.Stop
```

BtBot publishes Parquet BBO values.

Runner will publish WebSocket BBO values when live transport is implemented.

## Concurrency

MarketData protects buffers and registrations with a mutex.

Callbacks run without the MarketData lock, allowing them to read LatestBBO.

## Persistence

None.

BtBot rebuilds current state from replay.

Runner rebuilds current state from new live WebSocket truth.

## Errors

Invalid identity or BBO values mutate nothing and invoke no callback.

All callback errors are returned to the producer.

## Program Flow

```text
IngestBBO
  validate market update
  publish latest BBO and copy subscribers
  notify matching subscribers
```

## Required Proof

- Exact-key latest reads.
- Buffer-before-callback ordering.
- Invalid-update atomicity.
- No-subscriber buffering.
- Idempotent subscription shutdown.
- Complete callback error propagation.

See [MarketData](../marketdata.md) for permanent cross-package ownership.
